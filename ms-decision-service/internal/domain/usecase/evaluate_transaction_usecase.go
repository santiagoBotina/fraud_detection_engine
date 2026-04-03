package usecase

import (
	"context"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/repository"
	"time"

	"github.com/rs/zerolog"
)

// EvaluateTransactionUseCase orchestrates transaction evaluation against the rules engine.
type EvaluateTransactionUseCase struct {
	ruleRepo            repository.RuleRepository
	decisionPublisher   repository.DecisionPublisher
	fraudScorePublisher repository.FraudScoreRequestPublisher
	ruleEvalRepo        repository.RuleEvaluationRepository
	logger              zerolog.Logger
}

// NewEvaluateTransactionUseCase creates a new use case with the given ports.
func NewEvaluateTransactionUseCase(
	ruleRepo repository.RuleRepository,
	decisionPublisher repository.DecisionPublisher,
	fraudScorePublisher repository.FraudScoreRequestPublisher,
	ruleEvalRepo repository.RuleEvaluationRepository,
	logger zerolog.Logger,
) *EvaluateTransactionUseCase {
	return &EvaluateTransactionUseCase{
		ruleRepo:            ruleRepo,
		decisionPublisher:   decisionPublisher,
		fraudScorePublisher: fraudScorePublisher,
		ruleEvalRepo:        ruleEvalRepo,
		logger:              logger,
	}
}

// Execute evaluates the transaction against active rules and publishes the decision result.
// When the rule evaluation yields FRAUD_CHECK, the transaction is published to the fraud
// score request topic instead of the decision results topic.
func (uc *EvaluateTransactionUseCase) Execute(
	ctx context.Context,
	transaction *entity.TransactionMessage,
) (*entity.DecisionResult, error) {
	if transaction == nil {
		return nil, ErrTransactionNil
	}

	rules, err := uc.ruleRepo.FindActiveRulesSortedByPriority(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRuleRetrievalFailed, err)
	}

	status := entity.EvaluateRules(transaction, rules)

	// Persist rule evaluation results (non-fatal — log error but do not block)
	uc.persistTransactionRuleEvaluations(ctx, transaction, rules)

	if status == entity.FRAUDCHECK {
		if err := uc.fraudScorePublisher.Publish(ctx, transaction); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFraudScorePublishFailed, err)
		}

		return &entity.DecisionResult{
			TransactionID: transaction.ID,
			Status:        status,
		}, nil
	}

	result := &entity.DecisionResult{
		TransactionID: transaction.ID,
		Status:        status,
	}

	if err := uc.decisionPublisher.Publish(ctx, result); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecisionPublishFailed, err)
	}

	return result, nil
}

// persistTransactionRuleEvaluations builds RuleEvaluationResult records for each rule
// evaluated and persists them via SaveBatch. Errors are logged but do not block the flow.
func (uc *EvaluateTransactionUseCase) persistTransactionRuleEvaluations(
	ctx context.Context,
	transaction *entity.TransactionMessage,
	rules []entity.Rule,
) {
	if len(rules) == 0 {
		return
	}

	now := time.Now()
	results := make([]entity.RuleEvaluationResult, 0, len(rules))

	for _, rule := range rules {
		actualValue := transaction.GetFieldValue(rule.ConditionField)
		matched := rule.ConditionOperator.Compare(
			actualValue,
			rule.ConditionValue,
			rule.ConditionField,
		)

		results = append(results, entity.RuleEvaluationResult{
			TransactionID:     transaction.ID,
			RuleID:            rule.RuleID,
			RuleName:          rule.RuleName,
			ConditionField:    string(rule.ConditionField),
			ConditionOperator: string(rule.ConditionOperator),
			ConditionValue:    rule.ConditionValue,
			ActualFieldValue:  actualValue,
			Matched:           matched,
			ResultStatus:      string(rule.ResultStatus),
			EvaluatedAt:       now,
			Priority:          rule.Priority,
		})
	}

	if err := uc.ruleEvalRepo.SaveBatch(ctx, results); err != nil {
		uc.logger.Error().Err(err).
			Str("transaction_id", transaction.ID).
			Int("rule_count", len(results)).
			Msg("failed to persist rule evaluation results")
	}
}
