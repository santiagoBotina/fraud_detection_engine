package usecase

import (
	"context"
	"errors"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/infrastructure/telemetry"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// updateStatusMockRepo is a hand-written mock implementing TransactionRepository
// that captures the finalizedAt parameter and supports configurable FindByID behavior.
type updateStatusMockRepo struct {
	capturedFinalizedAt *time.Time
	capturedStatus      entity.TransactionStatus
	updateStatusCalled  bool
	updateStatusErr     error
	findByIDFunc        func(ctx context.Context, id string) (*entity.TransactionEntity, error)
}

func (m *updateStatusMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *updateStatusMockRepo) UpdateStatus(_ context.Context, _ string, status entity.TransactionStatus, finalizedAt *time.Time) error {
	m.updateStatusCalled = true
	m.capturedStatus = status
	m.capturedFinalizedAt = finalizedAt
	return m.updateStatusErr
}

func (m *updateStatusMockRepo) FindByID(ctx context.Context, id string) (*entity.TransactionEntity, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *updateStatusMockRepo) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func (m *updateStatusMockRepo) FindAll(_ context.Context) ([]entity.TransactionEntity, error) {
	return nil, nil
}

// histogramSampleCount collects metrics from the histogram for a given status
// label and returns the sample count.
func histogramSampleCount(statusLabel string) uint64 {
	observer := telemetry.TransactionFinalizationDuration.WithLabelValues(statusLabel)
	hist, ok := observer.(prometheus.Histogram)
	if !ok {
		return 0
	}
	var m dto.Metric
	if err := hist.Write(&m); err != nil {
		return 0
	}
	return m.GetHistogram().GetSampleCount()
}

func TestUpdateTransactionStatusUseCase_Execute_Finalization(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name                string
		decisionStatus      string
		expectFinalizedAt   bool
		expectHistogramObs  bool
		expectedTxnStatus   entity.TransactionStatus
		expectedHistLabel   string
	}{
		{
			name:               "APPROVED sets finalized_at and observes histogram",
			decisionStatus:     "APPROVED",
			expectFinalizedAt:  true,
			expectHistogramObs: true,
			expectedTxnStatus:  entity.APPROVED,
			expectedHistLabel:  "APPROVED",
		},
		{
			name:               "DECLINED sets finalized_at and observes histogram",
			decisionStatus:     "DECLINED",
			expectFinalizedAt:  true,
			expectHistogramObs: true,
			expectedTxnStatus:  entity.DECLINED,
			expectedHistLabel:  "DECLINED",
		},
		{
			name:               "FRAUD_CHECK leaves finalized_at nil and does not observe histogram",
			decisionStatus:     "FRAUD_CHECK",
			expectFinalizedAt:  false,
			expectHistogramObs: false,
			expectedTxnStatus:  entity.PENDING,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset histogram before each test case
			telemetry.TransactionFinalizationDuration.Reset()

			mock := &updateStatusMockRepo{
				findByIDFunc: func(_ context.Context, _ string) (*entity.TransactionEntity, error) {
					return &entity.TransactionEntity{
						CreatedAt: createdAt,
					}, nil
				},
			}
			uc := NewUpdateTransactionStatusUseCase(mock)

			msg := &entity.DecisionCalculatedMessage{
				TransactionID: "txn_test_001",
				Status:        tt.decisionStatus,
			}

			err := uc.Execute(context.Background(), msg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !mock.updateStatusCalled {
				t.Fatal("expected UpdateStatus to be called")
			}

			if mock.capturedStatus != tt.expectedTxnStatus {
				t.Errorf("expected status %s, got %s", tt.expectedTxnStatus, mock.capturedStatus)
			}

			// Verify finalized_at
			if tt.expectFinalizedAt {
				if mock.capturedFinalizedAt == nil {
					t.Fatal("expected finalized_at to be set, got nil")
				}
				if mock.capturedFinalizedAt.IsZero() {
					t.Fatal("expected finalized_at to be non-zero")
				}
			} else {
				if mock.capturedFinalizedAt != nil {
					t.Fatalf("expected finalized_at to be nil, got %v", *mock.capturedFinalizedAt)
				}
			}

			// Verify histogram observation
			if tt.expectHistogramObs {
				count := histogramSampleCount(tt.expectedHistLabel)
				if count != 1 {
					t.Errorf("expected histogram sample count 1 for status %s, got %d", tt.expectedHistLabel, count)
				}
			} else {
				// For FRAUD_CHECK, verify no histogram observations for any terminal status
				approvedCount := histogramSampleCount("APPROVED")
				declinedCount := histogramSampleCount("DECLINED")
				if approvedCount != 0 || declinedCount != 0 {
					t.Errorf("expected no histogram observations, got APPROVED=%d DECLINED=%d", approvedCount, declinedCount)
				}
			}
		})
	}
}

func TestUpdateTransactionStatusUseCase_Execute_NilMessage(t *testing.T) {
	mock := &updateStatusMockRepo{}
	uc := NewUpdateTransactionStatusUseCase(mock)

	err := uc.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil message")
	}
	if !errors.Is(err, ErrDecisionMessageNil) {
		t.Errorf("expected ErrDecisionMessageNil, got: %v", err)
	}
}

func TestUpdateTransactionStatusUseCase_Execute_InvalidStatus(t *testing.T) {
	mock := &updateStatusMockRepo{}
	uc := NewUpdateTransactionStatusUseCase(mock)

	msg := &entity.DecisionCalculatedMessage{
		TransactionID: "txn_test_002",
		Status:        "UNKNOWN_STATUS",
	}

	err := uc.Execute(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus, got: %v", err)
	}
}

func TestUpdateTransactionStatusUseCase_Execute_UpdateStatusError(t *testing.T) {
	mock := &updateStatusMockRepo{
		updateStatusErr: errors.New("dynamodb connection failed"),
	}
	uc := NewUpdateTransactionStatusUseCase(mock)

	msg := &entity.DecisionCalculatedMessage{
		TransactionID: "txn_test_003",
		Status:        "APPROVED",
	}

	err := uc.Execute(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error when UpdateStatus fails")
	}
	if !errors.Is(err, ErrStatusUpdateFailed) {
		t.Errorf("expected ErrStatusUpdateFailed, got: %v", err)
	}
}
