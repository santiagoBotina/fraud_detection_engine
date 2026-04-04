package usecase

import (
	"context"
	"errors"
	"math"
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"
)

// statsErrorMockRepo is a mock that returns an error from FindAll.
type statsErrorMockRepo struct {
	err error
}

func (m *statsErrorMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *statsErrorMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus, _ *time.Time) error {
	return nil
}

func (m *statsErrorMockRepo) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return nil, nil
}

func (m *statsErrorMockRepo) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func (m *statsErrorMockRepo) FindAll(_ context.Context) ([]entity.TransactionEntity, error) {
	return nil, m.err
}

func TestGetTransactionStatsUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		transactions []entity.TransactionEntity
		wantStats    entity.TransactionStats
	}{
		{
			name:         "empty dataset",
			transactions: []entity.TransactionEntity{},
			wantStats: entity.TransactionStats{
				PaymentMethods: map[entity.PaymentMethod]int{},
			},
		},
		{
			name: "single APPROVED CARD transaction created today, not finalized",
			transactions: []entity.TransactionEntity{
				{
					ID:            "txn_001",
					AmountInCents: 5000,
					Currency:      entity.USD,
					PaymentMethod: entity.CARD,
					CustomerID:    "cust_1",
					Status:        entity.APPROVED,
					CreatedAt:     now.Add(-1 * time.Hour),
					UpdatedAt:     now.Add(-1 * time.Hour),
					FinalizedAt:   nil,
				},
			},
			wantStats: entity.TransactionStats{
				Today:          1,
				ThisWeek:       1,
				ThisMonth:      1,
				Total:          1,
				Approved:       1,
				PaymentMethods: map[entity.PaymentMethod]int{entity.CARD: 1},
			},
		},
		{
			name: "mixed statuses",
			transactions: func() []entity.TransactionEntity {
				return []entity.TransactionEntity{
					{
						ID:            "txn_a",
						PaymentMethod: entity.CARD,
						Status:        entity.APPROVED,
						CreatedAt:     now.Add(-2 * time.Hour),
						UpdatedAt:     now.Add(-2 * time.Hour),
					},
					{
						ID:            "txn_b",
						PaymentMethod: entity.BANK_TRANSFER,
						Status:        entity.DECLINED,
						CreatedAt:     now.Add(-3 * 24 * time.Hour),
						UpdatedAt:     now.Add(-3 * 24 * time.Hour),
					},
					{
						ID:            "txn_c",
						PaymentMethod: entity.CRYPTO,
						Status:        entity.PENDING,
						CreatedAt:     now.Add(-10 * 24 * time.Hour),
						UpdatedAt:     now.Add(-10 * 24 * time.Hour),
					},
					{
						ID:            "txn_d",
						PaymentMethod: entity.CARD,
						Status:        entity.APPROVED,
						CreatedAt:     now.Add(-40 * 24 * time.Hour),
						UpdatedAt:     now.Add(-40 * 24 * time.Hour),
					},
				}
			}(),
			wantStats: entity.TransactionStats{
				Today:    1,
				ThisWeek: 2,
				ThisMonth: 3,
				Total:    4,
				Approved: 2,
				Declined: 1,
				Pending:  1,
				PaymentMethods: map[entity.PaymentMethod]int{
					entity.CARD:          2,
					entity.BANK_TRANSFER: 1,
					entity.CRYPTO:        1,
				},
			},
		},
		{
			name: "all finalized with latency tiers",
			transactions: func() []entity.TransactionEntity {
				// LOW: 1000ms, MEDIUM: 3000ms, HIGH: 7000ms
				lowFin := now.Add(-2 * time.Hour).Add(1000 * time.Millisecond)
				medFin := now.Add(-3 * time.Hour).Add(3000 * time.Millisecond)
				highFin := now.Add(-4 * time.Hour).Add(7000 * time.Millisecond)
				return []entity.TransactionEntity{
					{
						ID:            "txn_low",
						PaymentMethod: entity.CARD,
						Status:        entity.APPROVED,
						CreatedAt:     now.Add(-2 * time.Hour),
						UpdatedAt:     now.Add(-2 * time.Hour),
						FinalizedAt:   &lowFin,
					},
					{
						ID:            "txn_med",
						PaymentMethod: entity.BANK_TRANSFER,
						Status:        entity.DECLINED,
						CreatedAt:     now.Add(-3 * time.Hour),
						UpdatedAt:     now.Add(-3 * time.Hour),
						FinalizedAt:   &medFin,
					},
					{
						ID:            "txn_high",
						PaymentMethod: entity.CRYPTO,
						Status:        entity.APPROVED,
						CreatedAt:     now.Add(-4 * time.Hour),
						UpdatedAt:     now.Add(-4 * time.Hour),
						FinalizedAt:   &highFin,
					},
				}
			}(),
			wantStats: entity.TransactionStats{
				Today:          3,
				ThisWeek:       3,
				ThisMonth:      3,
				Total:          3,
				Approved:       2,
				Declined:       1,
				PaymentMethods: map[entity.PaymentMethod]int{
					entity.CARD:          1,
					entity.BANK_TRANSFER: 1,
					entity.CRYPTO:        1,
				},
				AvgLatencyMs:   (1000 + 3000 + 7000) / 3.0,
				FinalizedCount: 3,
				LatencyLow:     1,
				LatencyMedium:  1,
				LatencyHigh:    1,
			},
		},
		{
			name: "none finalized",
			transactions: []entity.TransactionEntity{
				{
					ID:            "txn_p1",
					PaymentMethod: entity.CARD,
					Status:        entity.PENDING,
					CreatedAt:     now.Add(-5 * time.Hour),
					UpdatedAt:     now.Add(-5 * time.Hour),
					FinalizedAt:   nil,
				},
				{
					ID:            "txn_p2",
					PaymentMethod: entity.BANK_TRANSFER,
					Status:        entity.PENDING,
					CreatedAt:     now.Add(-6 * time.Hour),
					UpdatedAt:     now.Add(-6 * time.Hour),
					FinalizedAt:   nil,
				},
			},
			wantStats: entity.TransactionStats{
				Today:    2,
				ThisWeek: 2,
				ThisMonth: 2,
				Total:    2,
				Pending:  2,
				PaymentMethods: map[entity.PaymentMethod]int{
					entity.CARD:          1,
					entity.BANK_TRANSFER: 1,
				},
				FinalizedCount: 0,
				AvgLatencyMs:   0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &statsMockRepo{transactions: tc.transactions}
			uc := NewGetTransactionStatsUseCase(repo)

			stats, err := uc.Execute(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if stats.Today != tc.wantStats.Today {
				t.Errorf("Today: got %d, want %d", stats.Today, tc.wantStats.Today)
			}
			if stats.ThisWeek != tc.wantStats.ThisWeek {
				t.Errorf("ThisWeek: got %d, want %d", stats.ThisWeek, tc.wantStats.ThisWeek)
			}
			if stats.ThisMonth != tc.wantStats.ThisMonth {
				t.Errorf("ThisMonth: got %d, want %d", stats.ThisMonth, tc.wantStats.ThisMonth)
			}
			if stats.Total != tc.wantStats.Total {
				t.Errorf("Total: got %d, want %d", stats.Total, tc.wantStats.Total)
			}
			if stats.Approved != tc.wantStats.Approved {
				t.Errorf("Approved: got %d, want %d", stats.Approved, tc.wantStats.Approved)
			}
			if stats.Declined != tc.wantStats.Declined {
				t.Errorf("Declined: got %d, want %d", stats.Declined, tc.wantStats.Declined)
			}
			if stats.Pending != tc.wantStats.Pending {
				t.Errorf("Pending: got %d, want %d", stats.Pending, tc.wantStats.Pending)
			}

			// Payment methods
			if len(stats.PaymentMethods) != len(tc.wantStats.PaymentMethods) {
				t.Errorf("PaymentMethods length: got %d, want %d", len(stats.PaymentMethods), len(tc.wantStats.PaymentMethods))
			}
			for pm, want := range tc.wantStats.PaymentMethods {
				if got := stats.PaymentMethods[pm]; got != want {
					t.Errorf("PaymentMethod %s: got %d, want %d", pm, got, want)
				}
			}

			if stats.FinalizedCount != tc.wantStats.FinalizedCount {
				t.Errorf("FinalizedCount: got %d, want %d", stats.FinalizedCount, tc.wantStats.FinalizedCount)
			}
			if math.Abs(stats.AvgLatencyMs-tc.wantStats.AvgLatencyMs) > 0.01 {
				t.Errorf("AvgLatencyMs: got %f, want %f", stats.AvgLatencyMs, tc.wantStats.AvgLatencyMs)
			}
			if stats.LatencyLow != tc.wantStats.LatencyLow {
				t.Errorf("LatencyLow: got %d, want %d", stats.LatencyLow, tc.wantStats.LatencyLow)
			}
			if stats.LatencyMedium != tc.wantStats.LatencyMedium {
				t.Errorf("LatencyMedium: got %d, want %d", stats.LatencyMedium, tc.wantStats.LatencyMedium)
			}
			if stats.LatencyHigh != tc.wantStats.LatencyHigh {
				t.Errorf("LatencyHigh: got %d, want %d", stats.LatencyHigh, tc.wantStats.LatencyHigh)
			}
		})
	}
}

func TestGetTransactionStatsUseCase_Execute_ErrorPropagation(t *testing.T) {
	repoErr := errors.New("dynamodb scan failed")
	repo := &statsErrorMockRepo{err: repoErr}
	uc := NewGetTransactionStatsUseCase(repo)

	stats, err := uc.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected error %v, got %v", repoErr, err)
	}
	if stats != nil {
		t.Errorf("expected nil stats on error, got %+v", stats)
	}
}
