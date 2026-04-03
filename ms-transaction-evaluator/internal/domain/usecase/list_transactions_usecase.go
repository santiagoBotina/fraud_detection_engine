package usecase

import (
	"context"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/repository"
)

const (
	minLimit = 1
	maxLimit = 100
)

// ListTransactionsUseCase retrieves a paginated list of transactions.
type ListTransactionsUseCase struct {
	transactionRepo repository.TransactionRepository
}

// NewListTransactionsUseCase creates a new ListTransactionsUseCase.
func NewListTransactionsUseCase(repo repository.TransactionRepository) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{transactionRepo: repo}
}

// Execute retrieves a paginated list of transactions sorted by created_at descending.
func (uc *ListTransactionsUseCase) Execute(ctx context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error) {
	if limit < minLimit || limit > maxLimit {
		return nil, "", ErrInvalidLimit
	}

	transactions, nextCursor, err := uc.transactionRepo.FindAllPaginated(ctx, limit, cursor)
	if err != nil {
		return nil, "", err
	}

	return transactions, nextCursor, nil
}
