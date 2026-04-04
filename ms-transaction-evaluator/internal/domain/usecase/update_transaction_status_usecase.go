package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/repository"
	"ms-transaction-evaluator/internal/infrastructure/telemetry"
	"time"
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
// For terminal statuses (APPROVED, DECLINED), it records the finalized_at timestamp
// and observes the finalization latency in the Prometheus histogram.
func (uc *UpdateTransactionStatusUseCase) Execute(ctx context.Context, msg *entity.DecisionCalculatedMessage) error {
	if msg == nil {
		return ErrDecisionMessageNil
	}

	txnStatus, ok := statusMap[msg.Status]
	if !ok {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, msg.Status)
	}

	var finalizedAt *time.Time

	if txnStatus == entity.APPROVED || txnStatus == entity.DECLINED {
		now := time.Now().UTC()
		finalizedAt = &now
	}

	if err := uc.transactionRepo.UpdateStatus(ctx, msg.TransactionID, txnStatus, finalizedAt); err != nil {
		return fmt.Errorf("%w: %w", ErrStatusUpdateFailed, err)
	}

	if finalizedAt != nil {
		uc.observeLatency(ctx, msg.TransactionID, *finalizedAt, string(txnStatus))
	}

	return nil
}

// observeLatency fetches the transaction to get created_at, computes latency,
// and observes the histogram. Failures are logged but do not fail the status update.
func (uc *UpdateTransactionStatusUseCase) observeLatency(
	ctx context.Context, transactionID string, finalizedAt time.Time, status string,
) {
	txn, err := uc.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		log.Printf("failed to fetch transaction for latency observation: %v", err)
		return
	}

	latencySeconds := finalizedAt.Sub(txn.CreatedAt).Seconds()

	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("prometheus histogram observation panicked: %v", r)
			}
		}()
		telemetry.TransactionFinalizationDuration.WithLabelValues(status).Observe(latencySeconds)
	}()
}
