package usecase

import (
	"context"
	"errors"
	"fmt"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/repository"
	"time"

	"github.com/google/uuid"
)

var ErrSaveTransactionFailed = errors.New("failed to save transaction")

type SaveTransactionUseCase struct {
	transactionRepo repository.TransactionRepository
	eventPublisher  repository.TransactionEventPublisher
}

func NewSaveTransactionUseCase(transactionRepo repository.TransactionRepository, eventPublisher repository.TransactionEventPublisher) *SaveTransactionUseCase {
	return &SaveTransactionUseCase{
		transactionRepo: transactionRepo,
		eventPublisher:  eventPublisher,
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
		Status:            entity.PENDING,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// Save to repository
	if err := uc.transactionRepo.Save(ctx, transaction); err != nil {
		return nil, err
	}

	// Publish transaction event to Kafka
	if err := uc.eventPublisher.Publish(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to publish transaction event: %w", ErrEventPublishFailed)
	}

	return transaction, nil
}
