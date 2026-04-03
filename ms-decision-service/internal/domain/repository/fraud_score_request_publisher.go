package repository

import (
	"context"
	"ms-decision-service/internal/domain/entity"
)

// FraudScoreRequestPublisher defines the port for publishing transactions to the fraud score request topic.
type FraudScoreRequestPublisher interface {
	Publish(ctx context.Context, transaction *entity.TransactionMessage) error
}
