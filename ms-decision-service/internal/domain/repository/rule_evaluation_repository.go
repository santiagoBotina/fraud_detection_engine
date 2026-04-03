package repository

import (
	"context"
	"ms-decision-service/internal/domain/entity"
)

// RuleEvaluationRepository defines the port for persisting and retrieving rule evaluation results.
type RuleEvaluationRepository interface {
	SaveBatch(ctx context.Context, results []entity.RuleEvaluationResult) error
	FindByTransactionID(ctx context.Context, transactionID string) ([]entity.RuleEvaluationResult, error)
}
