package repository

import (
	"context"
	"ms-transaction-evaluator/internal/domain/entity"
)

type TransactionEventPublisher interface {
	Publish(ctx context.Context, transaction *entity.TransactionEntity) error
}
