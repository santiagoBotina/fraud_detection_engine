package usecase

import (
	"context"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/repository"
	"time"
)

// GetTransactionStatsUseCase computes aggregated metrics across all transactions.
type GetTransactionStatsUseCase struct {
	transactionRepo repository.TransactionRepository
}

// NewGetTransactionStatsUseCase creates a new GetTransactionStatsUseCase.
func NewGetTransactionStatsUseCase(repo repository.TransactionRepository) *GetTransactionStatsUseCase {
	return &GetTransactionStatsUseCase{transactionRepo: repo}
}

// Execute retrieves all transactions and aggregates them into stats in a single pass.
func (uc *GetTransactionStatsUseCase) Execute(ctx context.Context) (*entity.TransactionStats, error) {
	transactions, err := uc.transactionRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	last7d := now.Add(-7 * 24 * time.Hour)
	last30d := now.Add(-30 * 24 * time.Hour)

	stats := &entity.TransactionStats{
		PaymentMethods: make(map[entity.PaymentMethod]int),
	}

	var latencySum float64

	for i := range transactions {
		txn := &transactions[i]
		stats.Total++

		// Time buckets
		if !txn.CreatedAt.Before(last24h) {
			stats.Today++
		}
		if !txn.CreatedAt.Before(last7d) {
			stats.ThisWeek++
		}
		if !txn.CreatedAt.Before(last30d) {
			stats.ThisMonth++
		}

		// Status counts
		switch txn.Status {
		case entity.APPROVED:
			stats.Approved++
		case entity.DECLINED:
			stats.Declined++
		case entity.PENDING:
			stats.Pending++
		}

		// Payment method counts
		stats.PaymentMethods[txn.PaymentMethod]++

		// Latency for finalized transactions
		if txn.FinalizedAt != nil && !txn.FinalizedAt.IsZero() {
			latencyMs := float64(txn.FinalizedAt.Sub(txn.CreatedAt).Milliseconds())
			latencySum += latencyMs
			stats.FinalizedCount++

			tier := entity.ClassifyLatency(latencyMs)
			switch tier {
			case entity.LatencyLow:
				stats.LatencyLow++
			case entity.LatencyMedium:
				stats.LatencyMedium++
			case entity.LatencyHigh:
				stats.LatencyHigh++
			}
		}
	}

	if stats.FinalizedCount > 0 {
		stats.AvgLatencyMs = latencySum / float64(stats.FinalizedCount)
	}

	return stats, nil
}
