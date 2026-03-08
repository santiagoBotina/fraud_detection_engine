package usecase

import (
	"context"
	"errors"
	"time"

	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/repository"

	"github.com/google/uuid"
)

var (
	ErrSaveTransactionFailed = errors.New("failed to save transaction")
)

type SaveTransactionUseCase struct {
	transactionRepo repository.TransactionRepository
}

func NewSaveTransactionUseCase(transactionRepo repository.TransactionRepository) *SaveTransactionUseCase {
	return &SaveTransactionUseCase{
		transactionRepo: transactionRepo,
	}
}

func (uc *SaveTransactionUseCase) Execute(ctx context.Context, req *entity.EvaluateTransactionRequest) (*entity.TransactionEntity, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	// Create transaction entity from request
	transaction := &entity.TransactionEntity{
		ID:                uuid.New().String(),
		AmountInCents:     req.AmountInCents,
		Currency:          req.Currency,
		PaymentMethod:     req.PaymentMethod,
		CustomerID:        req.CustomerInfo.CustomerID,
		CustomerName:      req.CustomerInfo.Name,
		CustomerEmail:     req.CustomerInfo.Email,
		CustomerPhone:     req.CustomerInfo.Phone,
		CustomerIPAddress: req.CustomerInfo.IpAddress,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// Save to repository
	if err := uc.transactionRepo.Save(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}
