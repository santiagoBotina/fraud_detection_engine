package repository

import (
	"context"
	"ms-decision-service/internal/domain/entity"
)

// DecisionPublisher defines the port for publishing decision results.
type DecisionPublisher interface {
	Publish(ctx context.Context, result *entity.DecisionResult) error
}
