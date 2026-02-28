package repository

import "ms-transaction-evaluator/internal/domain/entity"

type TransactionRepository interface {
	createTransaction() *entity.TransactionEntity
}
