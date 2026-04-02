package usecase

import (
	"context"
	"fmt"

	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/repository"
)

// EvaluateTransactionUseCase orchestrates transaction evaluation against the rules engine.
type EvaluateTransactionUseCase struct {
	ruleRepo          repository.RuleRepository
	decisionPublisher repository.DecisionPublisher
}

// NewEvaluateTransactionUseCase creates a new use case with the given ports.
func NewEvaluateTransactionUseCase(
	ruleRepo repository.RuleRepository,
	decisionPublisher repository.DecisionPublisher,
) *EvaluateTransactionUseCase {
	return &EvaluateTransactionUseCase{
		ruleRepo:          ruleRepo,
		decisionPublisher: decisionPublisher,
	}
}

// Execute evaluates the transaction against active rules and publishes the decision result.
func (uc *EvaluateTransactionUseCase) Execute(
	ctx context.Context,
	transaction *entity.TransactionMessage,
) (*entity.DecisionResult, error) {
	if transaction == nil {
		return nil, ErrTransactionNil
	}

	rules, err := uc.ruleRepo.FindActiveRulesSortedByPriority(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRuleRetrievalFailed, err)
	}

	status := entity.EvaluateRules(transaction, rules)

	result := &entity.DecisionResult{
		TransactionID: transaction.ID,
		Status:        status,
	}

	if err := uc.decisionPublisher.Publish(ctx, result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecisionPublishFailed, err)
	}

	return result, nil
}
