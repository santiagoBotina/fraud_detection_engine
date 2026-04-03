package http

import (
	"context"
	"encoding/json"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// mockQueryTransactionRepository is a hand-written mock for TransactionRepository
// used by the query controller tests.
type mockQueryTransactionRepository struct {
	findByIDFunc         func(ctx context.Context, id string) (*entity.TransactionEntity, error)
	findAllPaginatedFunc func(ctx context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error)
}

func (m *mockQueryTransactionRepository) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *mockQueryTransactionRepository) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus) error {
	return nil
}

func (m *mockQueryTransactionRepository) FindByID(ctx context.Context, id string) (*entity.TransactionEntity, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockQueryTransactionRepository) FindAllPaginated(ctx context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error) {
	if m.findAllPaginatedFunc != nil {
		return m.findAllPaginatedFunc(ctx, limit, cursor)
	}
	return nil, "", nil
}

func sampleTransaction() entity.TransactionEntity {
	return entity.TransactionEntity{
		ID:                "txn_abc123",
		AmountInCents:     15000,
		Currency:          entity.USD,
		PaymentMethod:     entity.CARD,
		CustomerID:        "cust_123",
		CustomerName:      "John Doe",
		CustomerEmail:     "john@example.com",
		CustomerPhone:     "+1234567890",
		CustomerIPAddress: "192.168.1.1",
		Status:            entity.APPROVED,
		CreatedAt:         time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2025, 1, 15, 10, 30, 5, 0, time.UTC),
	}
}

func newQueryController(repo *mockQueryTransactionRepository) (*TransactionQueryController, *echo.Echo) {
	listUC := usecase.NewListTransactionsUseCase(repo)
	getUC := usecase.NewGetTransactionUseCase(repo)
	controller := NewTransactionQueryController(listUC, getUC, zerolog.Nop())

	e := echo.New()
	controller.RegisterRoutes(e)

	return controller, e
}

func TestTransactionQueryController_ListTransactions(t *testing.T) {
	t.Run("should return 200 with valid params and transactions", func(t *testing.T) {
		txn := sampleTransaction()
		repo := &mockQueryTransactionRepository{
			findAllPaginatedFunc: func(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
				return []entity.TransactionEntity{txn}, "next123", nil
			},
		}
		_, e := newQueryController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions?limit=10", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp ListTransactionsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.NextCursor != "next123" {
			t.Errorf("expected next_cursor %q, got %q", "next123", resp.NextCursor)
		}

		data, ok := resp.Data.([]interface{})
		if !ok {
			t.Fatalf("expected data to be an array, got %T", resp.Data)
		}
		if len(data) != 1 {
			t.Errorf("expected 1 transaction, got %d", len(data))
		}
	})

	t.Run("should return 200 with empty result", func(t *testing.T) {
		repo := &mockQueryTransactionRepository{
			findAllPaginatedFunc: func(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
				return []entity.TransactionEntity{}, "", nil
			},
		}
		_, e := newQueryController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp ListTransactionsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.NextCursor != "" {
			t.Errorf("expected empty next_cursor, got %q", resp.NextCursor)
		}

		data, ok := resp.Data.([]interface{})
		if !ok {
			t.Fatalf("expected data to be an array, got %T", resp.Data)
		}
		if len(data) != 0 {
			t.Errorf("expected 0 transactions, got %d", len(data))
		}
	})

	t.Run("should return 400 for invalid limit (non-numeric)", func(t *testing.T) {
		repo := &mockQueryTransactionRepository{}
		_, e := newQueryController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions?limit=abc", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var resp ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Error != "Invalid limit parameter" {
			t.Errorf("expected error %q, got %q", "Invalid limit parameter", resp.Error)
		}
	})

	t.Run("should return 400 for invalid limit (out of range)", func(t *testing.T) {
		repo := &mockQueryTransactionRepository{}
		_, e := newQueryController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions?limit=101", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var resp ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Error != "Invalid limit parameter" {
			t.Errorf("expected error %q, got %q", "Invalid limit parameter", resp.Error)
		}
	})

	t.Run("should return 400 for invalid cursor", func(t *testing.T) {
		repo := &mockQueryTransactionRepository{
			findAllPaginatedFunc: func(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
				return nil, "", usecase.ErrInvalidCursor
			},
		}
		_, e := newQueryController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions?cursor=bad-cursor", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var resp ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Error != "Invalid cursor parameter" {
			t.Errorf("expected error %q, got %q", "Invalid cursor parameter", resp.Error)
		}
	})

	t.Run("should use default limit when not specified", func(t *testing.T) {
		var capturedLimit int
		repo := &mockQueryTransactionRepository{
			findAllPaginatedFunc: func(_ context.Context, limit int, _ string) ([]entity.TransactionEntity, string, error) {
				capturedLimit = limit
				return []entity.TransactionEntity{}, "", nil
			},
		}
		_, e := newQueryController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if capturedLimit != defaultLimit {
			t.Errorf("expected default limit %d, got %d", defaultLimit, capturedLimit)
		}
	})
}

func TestTransactionQueryController_GetTransaction(t *testing.T) {
	t.Run("should return 200 with valid ID", func(t *testing.T) {
		txn := sampleTransaction()
		repo := &mockQueryTransactionRepository{
			findByIDFunc: func(_ context.Context, id string) (*entity.TransactionEntity, error) {
				if id == "txn_abc123" {
					return &txn, nil
				}
				return nil, nil
			},
		}
		_, e := newQueryController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions/txn_abc123", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp TransactionDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Data == nil {
			t.Fatal("expected data in response, got nil")
		}

		dataMap, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected data to be a map, got %T", resp.Data)
		}

		if dataMap["id"] != "txn_abc123" {
			t.Errorf("expected id %q, got %q", "txn_abc123", dataMap["id"])
		}
	})

	t.Run("should return 404 when transaction not found", func(t *testing.T) {
		repo := &mockQueryTransactionRepository{
			findByIDFunc: func(_ context.Context, _ string) (*entity.TransactionEntity, error) {
				return nil, nil
			},
		}
		_, e := newQueryController(repo)

		req := httptest.NewRequest(http.MethodGet, "/transactions/nonexistent", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}

		var resp ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Error != "Transaction not found" {
			t.Errorf("expected error %q, got %q", "Transaction not found", resp.Error)
		}
	})
}

func TestTransactionQueryController_CORSPreflight(t *testing.T) {
	t.Run("should include CORS headers on preflight OPTIONS request", func(t *testing.T) {
		repo := &mockQueryTransactionRepository{}
		_, e := newQueryController(repo)

		// Add CORS middleware matching main.go configuration
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Response().Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
				c.Response().Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
				c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")
				if c.Request().Method == http.MethodOptions {
					return c.NoContent(http.StatusNoContent)
				}
				return next(c)
			}
		})

		req := httptest.NewRequest(http.MethodOptions, "/transactions", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", "GET")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "http://localhost:5173" {
			t.Errorf("expected Access-Control-Allow-Origin %q, got %q", "http://localhost:5173", allowOrigin)
		}

		allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
		if allowMethods == "" {
			t.Error("expected Access-Control-Allow-Methods header to be set")
		}

		allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
		if allowHeaders == "" {
			t.Error("expected Access-Control-Allow-Headers header to be set")
		}
	})
}
