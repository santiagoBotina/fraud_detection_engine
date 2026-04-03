package usecase

import (
	"context"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/repository"
)

// EvaluateTransactionUseCase orchestrates transaction evaluation against the rules engine.
type EvaluateTransactionUseCase struct {
	ruleRepo            repository.RuleRepository
	decisionPublisher   repository.DecisionPublisher
	fraudScorePublisher repository.FraudScoreRequestPublisher
}

// NewEvaluateTransactionUseCase creates a new use case with the given ports.
func NewEvaluateTransactionUseCase(
	ruleRepo repository.RuleRepository,
	decisionPublisher repository.DecisionPublisher,
	fraudScorePublisher repository.FraudScoreRequestPublisher,
) *EvaluateTransactionUseCase {
	return &EvaluateTransactionUseCase{
		ruleRepo:            ruleRepo,
		decisionPublisher:   decisionPublisher,
		fraudScorePublisher: fraudScorePublisher,
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
