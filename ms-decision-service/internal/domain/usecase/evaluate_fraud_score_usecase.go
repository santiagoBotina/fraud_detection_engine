package usecase

import (
	"context"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/repository"
	"strconv"
)

// EvaluateFraudScoreUseCase orchestrates fraud score evaluation against fraud-score-specific rules.
type EvaluateFraudScoreUseCase struct {
	ruleRepo          repository.RuleRepository
	decisionPublisher repository.DecisionPublisher
}

// NewEvaluateFraudScoreUseCase creates a new use case with the given ports.
func NewEvaluateFraudScoreUseCase(
	ruleRepo repository.RuleRepository,
	decisionPublisher repository.DecisionPublisher,
) *EvaluateFraudScoreUseCase {
	return &EvaluateFraudScoreUseCase{
		ruleRepo:          ruleRepo,
		decisionPublisher: decisionPublisher,
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

	result := &entity.DecisionResult{
		TransactionID: msg.TransactionID,
		Status:        status,
	}

	if err := uc.decisionPublisher.Publish(ctx, result); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecisionPublishFailed, err)
	}

	return result, nil
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
