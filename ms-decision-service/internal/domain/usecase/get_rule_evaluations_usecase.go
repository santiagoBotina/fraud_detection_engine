package usecase

import (
	"context"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/repository"
)

// GetRuleEvaluationsUseCase retrieves rule evaluation results for a given transaction.
type GetRuleEvaluationsUseCase struct {
	ruleEvalRepo repository.RuleEvaluationRepository
}

// NewGetRuleEvaluationsUseCase creates a new use case with the given repository.
func NewGetRuleEvaluationsUseCase(
	ruleEvalRepo repository.RuleEvaluationRepository,
) *GetRuleEvaluationsUseCase {
	return &GetRuleEvaluationsUseCase{
		ruleEvalRepo: ruleEvalRepo,
	}
}

// Execute retrieves all rule evaluation results for the specified transaction ID,
// sorted by priority ascending.
func (uc *GetRuleEvaluationsUseCase) Execute(
	ctx context.Context,
	transactionID string,
) ([]entity.RuleEvaluationResult, error) {
	if transactionID == "" {
		return nil, ErrTransactionIDEmpty
	}

	results, err := uc.ruleEvalRepo.FindByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEvaluationRetrievalFailed, err)
	}

	return results, nil
}
