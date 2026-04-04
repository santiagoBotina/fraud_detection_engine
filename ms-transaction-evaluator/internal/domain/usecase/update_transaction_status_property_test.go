package usecase

import (
	"context"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/infrastructure/telemetry"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"pgregory.net/rapid"
)

// Feature: transaction-finalization-latency, Property 1: Finalized_at set if and only if terminal status
// Validates: Requirements 1.1, 1.2

// statusCaptureMockRepo is a hand-written mock implementing TransactionRepository
// that captures the finalizedAt parameter passed to UpdateStatus.
type statusCaptureMockRepo struct {
	capturedFinalizedAt *time.Time
	updateStatusCalled  bool
}

func (m *statusCaptureMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *statusCaptureMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus, finalizedAt *time.Time) error {
	m.updateStatusCalled = true
	m.capturedFinalizedAt = finalizedAt
	return nil
}

func (m *statusCaptureMockRepo) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return &entity.TransactionEntity{
		CreatedAt: time.Now().UTC().Add(-5 * time.Second),
	}, nil
}

func (m *statusCaptureMockRepo) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func (m *statusCaptureMockRepo) FindAll(_ context.Context) ([]entity.TransactionEntity, error) {
	return nil, nil
}

func TestProperty_FinalizedAtSetIffTerminalStatus(t *testing.T) {
	decisionStatuses := []string{"APPROVED", "DECLINED", "FRAUD_CHECK"}

	rapid.Check(t, func(t *rapid.T) {
		mock := &statusCaptureMockRepo{}
		uc := NewUpdateTransactionStatusUseCase(mock)

		txnID := rapid.StringMatching(`^txn_[a-z0-9]{8,16}$`).Draw(t, "transactionID")
		statusIdx := rapid.IntRange(0, len(decisionStatuses)-1).Draw(t, "statusIdx")
		status := decisionStatuses[statusIdx]

		msg := &entity.DecisionCalculatedMessage{
			TransactionID: txnID,
			Status:        status,
		}

		err := uc.Execute(context.Background(), msg)
		if err != nil {
			t.Fatalf("unexpected error from Execute: %v", err)
		}

		if !mock.updateStatusCalled {
			t.Fatal("expected UpdateStatus to be called")
		}

		isTerminal := status == "APPROVED" || status == "DECLINED"

		if isTerminal && mock.capturedFinalizedAt == nil {
			t.Fatalf("status %s is terminal but finalizedAt was nil", status)
		}

		if !isTerminal && mock.capturedFinalizedAt != nil {
			t.Fatalf("status %s is non-terminal but finalizedAt was %v", status, *mock.capturedFinalizedAt)
		}
	})
}

// Feature: transaction-finalization-latency, Property 5: Histogram observed with correct label on finalization
// Validates: Requirements 3.1, 3.2

// histogramMockRepo is a mock that returns a transaction with a configurable
// created_at so we can predict the approximate latency observed in the histogram.
type histogramMockRepo struct {
	createdAt time.Time
}

func (m *histogramMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *histogramMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus, _ *time.Time) error {
	return nil
}

func (m *histogramMockRepo) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return &entity.TransactionEntity{
		CreatedAt: m.createdAt,
	}, nil
}

func (m *histogramMockRepo) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func (m *histogramMockRepo) FindAll(_ context.Context) ([]entity.TransactionEntity, error) {
	return nil, nil
}

// getHistogramSampleCount collects metrics from the histogram for a given status
// label and returns the sample count and sample sum.
func getHistogramSampleCount(statusLabel string) (uint64, float64) {
	observer := telemetry.TransactionFinalizationDuration.WithLabelValues(statusLabel)

	// Get the underlying prometheus.Histogram from the observer
	hist, ok := observer.(prometheus.Histogram)
	if !ok {
		return 0, 0
	}

	var m dto.Metric
	if err := hist.Write(&m); err != nil {
		return 0, 0
	}

	h := m.GetHistogram()
	return h.GetSampleCount(), h.GetSampleSum()
}

func TestProperty_HistogramObservedWithCorrectLabelOnFinalization(t *testing.T) {
	terminalStatuses := []string{"APPROVED", "DECLINED"}

	rapid.Check(t, func(t *rapid.T) {
		// Reset the histogram to avoid accumulation across iterations
		telemetry.TransactionFinalizationDuration.Reset()

		// Use a created_at slightly in the past so latency is positive
		createdAt := time.Now().UTC().Add(-2 * time.Second)
		mock := &histogramMockRepo{createdAt: createdAt}
		uc := NewUpdateTransactionStatusUseCase(mock)

		statusIdx := rapid.IntRange(0, len(terminalStatuses)-1).Draw(t, "statusIdx")
		status := terminalStatuses[statusIdx]

		txnID := rapid.StringMatching(`^txn_[a-z0-9]{8,16}$`).Draw(t, "transactionID")

		beforeExec := time.Now().UTC()

		msg := &entity.DecisionCalculatedMessage{
			TransactionID: txnID,
			Status:        status,
		}

		err := uc.Execute(context.Background(), msg)
		if err != nil {
			t.Fatalf("unexpected error from Execute: %v", err)
		}

		afterExec := time.Now().UTC()

		// Verify histogram was observed exactly once for the given status label.
		// The use case uses the entity status string (APPROVED/DECLINED) as the label.
		count, sum := getHistogramSampleCount(status)

		if count != 1 {
			t.Fatalf("expected histogram sample count 1 for status %s, got %d", status, count)
		}

		// Verify the observed value is approximately correct:
		// The latency should be between (beforeExec - createdAt) and (afterExec - createdAt)
		minLatency := beforeExec.Sub(createdAt).Seconds()
		maxLatency := afterExec.Sub(createdAt).Seconds()

		if sum < minLatency || sum > maxLatency {
			t.Fatalf("histogram sum %f not in expected range [%f, %f] for status %s",
				sum, minLatency, maxLatency, status)
		}

		// Verify the other terminal status was NOT observed
		otherStatus := terminalStatuses[0]
		if status == terminalStatuses[0] {
			otherStatus = terminalStatuses[1]
		}

		otherCount, _ := getHistogramSampleCount(otherStatus)
		if otherCount != 0 {
			t.Fatalf("expected no histogram observation for status %s, got %d", otherStatus, otherCount)
		}
	})
}
