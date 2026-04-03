package usecase

import (
	"context"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/repository"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// EvaluateFraudScoreUseCase orchestrates fraud score evaluation against fraud-score-specific rules.
type EvaluateFraudScoreUseCase struct {
	ruleRepo          repository.RuleRepository
	decisionPublisher repository.DecisionPublisher
	ruleEvalRepo      repository.RuleEvaluationRepository
	logger            zerolog.Logger
}

// NewEvaluateFraudScoreUseCase creates a new use case with the given ports.
func NewEvaluateFraudScoreUseCase(
	ruleRepo repository.RuleRepository,
	decisionPublisher repository.DecisionPublisher,
	ruleEvalRepo repository.RuleEvaluationRepository,
	logger zerolog.Logger,
) *EvaluateFraudScoreUseCase {
	return &EvaluateFraudScoreUseCase{
		ruleRepo:          ruleRepo,
		decisionPublisher: decisionPublisher,
		ruleEvalRepo:      ruleEvalRepo,
		logger:            logger,
	}
}

// Execute evaluates the fraud score against fraud-score rules and publishes the final decision.
// If no fraud-score rule matches, it defaults to APPROVED (fail-open).
func (uc *EvaluateFraudScoreUseCase) Execute(
	ctx context.Context,
	msg *entity.FraudScoreCalculatedMessage,
) (*entity.DecisionResult, error) {
	if msg == nil {
		return nil, ErrFraudScoreMessageNil
	}

	rules, err := uc.ruleRepo.FindActiveRulesSortedByPriority(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRuleRetrievalFailed, err)
	}

	status := evaluateFraudScoreRules(msg.FraudScore, rules)

	// Persist fraud-score rule evaluation results (non-fatal — log error but do not block)
	uc.persistFraudScoreRuleEvaluations(ctx, msg, rules)

	result := &entity.DecisionResult{
		TransactionID: msg.TransactionID,
		Status:        status,
	}

	if err := uc.decisionPublisher.Publish(ctx, result); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecisionPublishFailed, err)
	}

	return result, nil
}

// persistFraudScoreRuleEvaluations builds RuleEvaluationResult records for each fraud-score
// rule evaluated and persists them via SaveBatch. Errors are logged but do not block the flow.
func (uc *EvaluateFraudScoreUseCase) persistFraudScoreRuleEvaluations(
	ctx context.Context,
	msg *entity.FraudScoreCalculatedMessage,
	rules []entity.Rule,
) {
	// Only persist evaluations for fraud_score rules
	fraudScoreRules := filterFraudScoreRules(rules)
	if len(fraudScoreRules) == 0 {
		return
	}

	now := time.Now()
	scoreStr := strconv.Itoa(msg.FraudScore)
	results := make([]entity.RuleEvaluationResult, 0, len(fraudScoreRules))

	for _, rule := range fraudScoreRules {
		matched := rule.ConditionOperator.Compare(
			scoreStr,
			rule.ConditionValue,
			entity.FieldFraudScore,
		)

		results = append(results, entity.RuleEvaluationResult{
			TransactionID:     msg.TransactionID,
			RuleID:            rule.RuleID,
			RuleName:          rule.RuleName,
			ConditionField:    string(rule.ConditionField),
			ConditionOperator: string(rule.ConditionOperator),
			ConditionValue:    rule.ConditionValue,
			ActualFieldValue:  scoreStr,
			Matched:           matched,
			ResultStatus:      string(rule.ResultStatus),
			EvaluatedAt:       now,
			Priority:          rule.Priority,
		})
	}

	if err := uc.ruleEvalRepo.SaveBatch(ctx, results); err != nil {
		uc.logger.Error().Err(err).
			Str("transaction_id", msg.TransactionID).
			Int("rule_count", len(results)).
			Msg("failed to persist fraud score rule evaluation results")
	}
}

// filterFraudScoreRules returns only rules with ConditionField == FieldFraudScore.
func filterFraudScoreRules(rules []entity.Rule) []entity.Rule {
	var filtered []entity.Rule
	for _, r := range rules {
		if r.ConditionField == entity.FieldFraudScore {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// evaluateFraudScoreRules evaluates the fraud score against rules where ConditionField is fraud_score.
// Returns the ResultStatus of the first matching rule. If no rule matches, returns APPROVED (fail-open).
func evaluateFraudScoreRules(fraudScore int, rules []entity.Rule) entity.DecisionStatus {
	scoreStr := strconv.Itoa(fraudScore)
	for _, rule := range rules {
		if rule.ConditionField != entity.FieldFraudScore {
			continue
		}
		if rule.ConditionOperator.Compare(scoreStr, rule.ConditionValue, entity.FieldFraudScore) {
			return rule.ResultStatus
		}
	}
	return entity.APPROVED
}
