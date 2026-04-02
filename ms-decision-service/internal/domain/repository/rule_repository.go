package repository

import (
	"context"

	"ms-decision-service/internal/domain/entity"
)

// RuleRepository defines the port for retrieving fraud detection rules.
type RuleRepository interface {
	FindActiveRulesSortedByPriority(ctx context.Context) ([]entity.Rule, error)
}
