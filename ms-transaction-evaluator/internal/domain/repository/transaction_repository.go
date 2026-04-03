package repository

import (
	"context"
	"ms-transaction-evaluator/internal/domain/entity"
)

type TransactionRepository interface {
	Save(ctx context.Context, transaction *entity.TransactionEntity) error
	UpdateStatus(ctx context.Context, id string, status entity.TransactionStatus) error
}
