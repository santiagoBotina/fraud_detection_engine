package usecase

import (
	"context"
	"fmt"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/repository"
)

// GetTransactionUseCase retrieves a single transaction by ID.
type GetTransactionUseCase struct {
	transactionRepo repository.TransactionRepository
}

// NewGetTransactionUseCase creates a new GetTransactionUseCase.
func NewGetTransactionUseCase(repo repository.TransactionRepository) *GetTransactionUseCase {
	return &GetTransactionUseCase{transactionRepo: repo}
}

// Execute retrieves a transaction by its ID.
func (uc *GetTransactionUseCase) Execute(ctx context.Context, id string) (*entity.TransactionEntity, error) {
	txn, err := uc.transactionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransactionNotFound, id)
	}

	if txn == nil {
		return nil, fmt.Errorf("%w: %s", ErrTransactionNotFound, id)
	}

	return txn, nil
}
