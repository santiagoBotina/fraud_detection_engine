package usecase

import (
	"context"
	"errors"
	"fmt"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/repository"
)

var (
	ErrDecisionMessageNil = errors.New("decision message is nil")
	ErrInvalidStatus      = errors.New("invalid decision status")
	ErrStatusUpdateFailed = errors.New("failed to update transaction status")
)

// statusMap maps decision statuses from the Decision.Calculated topic to transaction statuses.
var statusMap = map[string]entity.TransactionStatus{
	"APPROVED":    entity.APPROVED,
	"DECLINED":    entity.DECLINED,
	"FRAUD_CHECK": entity.PENDING,
}

// UpdateTransactionStatusUseCase updates a transaction's status based on a decision result.
type UpdateTransactionStatusUseCase struct {
	transactionRepo repository.TransactionRepository
}

// NewUpdateTransactionStatusUseCase creates a new use case.
func NewUpdateTransactionStatusUseCase(repo repository.TransactionRepository) *UpdateTransactionStatusUseCase {
	return &UpdateTransactionStatusUseCase{transactionRepo: repo}
}

// Execute maps the decision status to a transaction status and updates the record.
func (uc *UpdateTransactionStatusUseCase) Execute(ctx context.Context, msg *entity.DecisionCalculatedMessage) error {
	if msg == nil {
		return ErrDecisionMessageNil
	}

	txnStatus, ok := statusMap[msg.Status]
	if !ok {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, msg.Status)
	}

	if err := uc.transactionRepo.UpdateStatus(ctx, msg.TransactionID, txnStatus); err != nil {
		return fmt.Errorf("%w: %w", ErrStatusUpdateFailed, err)
	}

	return nil
}
