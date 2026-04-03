package usecase

import (
	"context"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/repository"
)

// ListRulesUseCase retrieves all fraud detection rules for the dashboard.
type ListRulesUseCase struct {
	ruleRepo repository.RuleRepository
}

// NewListRulesUseCase creates a new use case with the given repository.
func NewListRulesUseCase(
	ruleRepo repository.RuleRepository,
) *ListRulesUseCase {
	return &ListRulesUseCase{
		ruleRepo: ruleRepo,
	}
}

// Execute retrieves all rules (including inactive) sorted by priority ascending.
func (uc *ListRulesUseCase) Execute(
	ctx context.Context,
) ([]entity.Rule, error) {
	rules, err := uc.ruleRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRuleRetrievalFailed, err)
	}

	return rules, nil
}
