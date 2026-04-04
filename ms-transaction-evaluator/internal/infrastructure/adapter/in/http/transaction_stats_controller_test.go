package http

import (
	"context"
	"encoding/json"
	"errors"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// mockStatsTransactionRepository is a hand-written mock for TransactionRepository
// used by the stats controller tests.
type mockStatsTransactionRepository struct {
	findAllFunc func(ctx context.Context) ([]entity.TransactionEntity, error)
}

func (m *mockStatsTransactionRepository) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *mockStatsTransactionRepository) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus, _ *time.Time) error {
	return nil
}

func (m *mockStatsTransactionRepository) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return nil, nil
}

func (m *mockStatsTransactionRepository) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func (m *mockStatsTransactionRepository) FindAll(ctx context.Context) ([]entity.TransactionEntity, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, nil
}

func newStatsController(repo *mockStatsTransactionRepository) (*TransactionStatsController, *echo.Echo) {
	statsUC := usecase.NewGetTransactionStatsUseCase(repo)
	controller := NewTransactionStatsController(statsUC, zerolog.Nop())

	e := echo.New()
	controller.RegisterRoutes(e)

	return controller, e
}

func TestTransactionStatsController_GetStats(t *testing.T) {
	t.Run("should return 200 with correct stats response", func(t *testing.T) {
		now := time.Now().UTC()
		finalizedAt := now.Add(3 * time.Second)

		repo := &mockStatsTransactionRepository{
			findAllFunc: func(_ context.Context) ([]entity.TransactionEntity, error) {
				return []entity.TransactionEntity{
					{
						ID:            "txn_1",
						AmountInCents: 10000,
						Currency:      entity.USD,
						PaymentMethod: entity.CARD,
						Status:        entity.APPROVED,
						CreatedAt:     now,
						UpdatedAt:     now,
						FinalizedAt:   &finalizedAt,
					},
					{
						ID:            "txn_2",
						AmountInCents: 20000,
						Currency:      entity.EUR,
						PaymentMethod: entity.BANK_TRANSFER,
						Status:        entity.DECLINED,
						CreatedAt:     now,
						UpdatedAt:     now,
					},
					{
						ID:            "txn_3",
						AmountInCents: 5000,
						Currency:      entity.COP,
						PaymentMethod: entity.CARD,
						Status:        entity.PENDING,
						CreatedAt:     now,
						UpdatedAt:     now,
					},
				}, nil
			},
		}
		_, e := newStatsController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions/stats", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp TransactionStatsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		// Total count
		if resp.Total != 3 {
			t.Errorf("expected total 3, got %d", resp.Total)
		}

		// Status counts
		if resp.Approved != 1 {
			t.Errorf("expected approved 1, got %d", resp.Approved)
		}
		if resp.Declined != 1 {
			t.Errorf("expected declined 1, got %d", resp.Declined)
		}
		if resp.Pending != 1 {
			t.Errorf("expected pending 1, got %d", resp.Pending)
		}

		// Time buckets — all transactions created "now" should be in all buckets
		if resp.Today != 3 {
			t.Errorf("expected today 3, got %d", resp.Today)
		}
		if resp.ThisWeek != 3 {
			t.Errorf("expected this_week 3, got %d", resp.ThisWeek)
		}
		if resp.ThisMonth != 3 {
			t.Errorf("expected this_month 3, got %d", resp.ThisMonth)
		}

		// Payment methods — CARD: 2, BANK_TRANSFER: 1
		if resp.PaymentMethods["CARD"] != 2 {
			t.Errorf("expected CARD count 2, got %d", resp.PaymentMethods["CARD"])
		}
		if resp.PaymentMethods["BANK_TRANSFER"] != 1 {
			t.Errorf("expected BANK_TRANSFER count 1, got %d", resp.PaymentMethods["BANK_TRANSFER"])
		}

		// Latency — only txn_1 is finalized (3s = 3000ms → MEDIUM tier)
		if resp.FinalizedCount != 1 {
			t.Errorf("expected finalized_count 1, got %d", resp.FinalizedCount)
		}
		if resp.AvgLatencyMs != 3000 {
			t.Errorf("expected avg_latency_ms 3000, got %f", resp.AvgLatencyMs)
		}
		if resp.LatencyMedium != 1 {
			t.Errorf("expected latency_medium 1, got %d", resp.LatencyMedium)
		}
		if resp.LatencyLow != 0 {
			t.Errorf("expected latency_low 0, got %d", resp.LatencyLow)
		}
		if resp.LatencyHigh != 0 {
			t.Errorf("expected latency_high 0, got %d", resp.LatencyHigh)
		}
	})

	t.Run("should return 500 on database error", func(t *testing.T) {
		repo := &mockStatsTransactionRepository{
			findAllFunc: func(_ context.Context) ([]entity.TransactionEntity, error) {
				return nil, errors.New("DynamoDB scan failed")
			},
		}
		_, e := newStatsController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions/stats", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		var resp ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Error != "Internal server error" {
			t.Errorf("expected error %q, got %q", "Internal server error", resp.Error)
		}

		if resp.Details != "DynamoDB scan failed" {
			t.Errorf("expected details %q, got %q", "DynamoDB scan failed", resp.Details)
		}
	})
}
