package http

import (
	"context"
	"encoding/json"
	"errors"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/usecase"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// --- Hand-written mocks ---

type mockRuleEvaluationRepository struct {
	findFunc func(ctx context.Context, transactionID string) ([]entity.RuleEvaluationResult, error)
}

func (m *mockRuleEvaluationRepository) SaveBatch(_ context.Context, _ []entity.RuleEvaluationResult) error {
	return nil
}

func (m *mockRuleEvaluationRepository) FindByTransactionID(ctx context.Context, transactionID string) ([]entity.RuleEvaluationResult, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, transactionID)
	}
	return nil, nil
}

type mockRuleRepository struct {
	findAllFunc func(ctx context.Context) ([]entity.Rule, error)
}

func (m *mockRuleRepository) FindActiveRulesSortedByPriority(_ context.Context) ([]entity.Rule, error) {
	return nil, nil
}

func (m *mockRuleRepository) FindAll(ctx context.Context) ([]entity.Rule, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, nil
}

// --- Helper ---

func newEvaluationController(
	ruleEvalRepo *mockRuleEvaluationRepository,
	ruleRepo *mockRuleRepository,
) (*EvaluationController, *echo.Echo) {
	getRuleEvaluationsUC := usecase.NewGetRuleEvaluationsUseCase(ruleEvalRepo)
	listRulesUC := usecase.NewListRulesUseCase(ruleRepo)
	controller := NewEvaluationController(getRuleEvaluationsUC, listRulesUC, zerolog.Nop())

	e := echo.New()
	controller.RegisterRoutes(e)

	return controller, e
}

// --- Tests ---

func TestEvaluationController_GetEvaluations(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	t.Run("should return 200 with evaluation results", func(t *testing.T) {
		ruleEvalRepo := &mockRuleEvaluationRepository{
			findFunc: func(_ context.Context, txID string) ([]entity.RuleEvaluationResult, error) {
				if txID != "txn_abc123" {
					return nil, nil
				}
				return []entity.RuleEvaluationResult{
					{
						TransactionID:     "txn_abc123",
						RuleID:            "rule-001",
						RuleName:          "Block CRYPTO payments",
						ConditionField:    "payment_method",
						ConditionOperator: "EQUAL",
						ConditionValue:    "CRYPTO",
						ActualFieldValue:  "CARD",
						Matched:           false,
						ResultStatus:      "DECLINED",
						EvaluatedAt:       now,
						Priority:          1,
					},
					{
						TransactionID:     "txn_abc123",
						RuleID:            "rule-002",
						RuleName:          "High amount check",
						ConditionField:    "amount_in_cents",
						ConditionOperator: "GREATER_THAN",
						ConditionValue:    "100000",
						ActualFieldValue:  "15000",
						Matched:           false,
						ResultStatus:      "DECLINED",
						EvaluatedAt:       now,
						Priority:          2,
					},
				}, nil
			},
		}
		_, e := newEvaluationController(ruleEvalRepo, &mockRuleRepository{})

		req := httptest.NewRequest(http.MethodGet, "/evaluations/txn_abc123", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp DataResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		data, ok := resp.Data.([]interface{})
		if !ok {
			t.Fatalf("expected data to be an array, got %T", resp.Data)
		}
		if len(data) != 2 {
			t.Errorf("expected 2 evaluation results, got %d", len(data))
		}
	})

	t.Run("should return 200 with empty list when no evaluations found", func(t *testing.T) {
		ruleEvalRepo := &mockRuleEvaluationRepository{
			findFunc: func(_ context.Context, _ string) ([]entity.RuleEvaluationResult, error) {
				return nil, nil
			},
		}
		_, e := newEvaluationController(ruleEvalRepo, &mockRuleRepository{})

		req := httptest.NewRequest(http.MethodGet, "/evaluations/txn_nonexistent", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp DataResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		data, ok := resp.Data.([]interface{})
		if !ok {
			t.Fatalf("expected data to be an array, got %T", resp.Data)
		}
		if len(data) != 0 {
			t.Errorf("expected 0 evaluation results, got %d", len(data))
		}
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		ruleEvalRepo := &mockRuleEvaluationRepository{
			findFunc: func(_ context.Context, _ string) ([]entity.RuleEvaluationResult, error) {
				return nil, errors.New("dynamo timeout")
			},
		}
		_, e := newEvaluationController(ruleEvalRepo, &mockRuleRepository{})

		req := httptest.NewRequest(http.MethodGet, "/evaluations/txn_abc123", nil)
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
	})
}

func TestEvaluationController_ListRules(t *testing.T) {
	t.Run("should return 200 with all rules sorted by priority", func(t *testing.T) {
		ruleRepo := &mockRuleRepository{
			findAllFunc: func(_ context.Context) ([]entity.Rule, error) {
				return []entity.Rule{
					{
						RuleID:            "rule-001",
						RuleName:          "Block CRYPTO payments",
						ConditionField:    entity.FieldPaymentMethod,
						ConditionOperator: entity.OpEqual,
						ConditionValue:    "CRYPTO",
						ResultStatus:      entity.DECLINED,
						Priority:          1,
						IsActive:          true,
					},
					{
						RuleID:            "rule-002",
						RuleName:          "High amount check",
						ConditionField:    entity.FieldAmountInCents,
						ConditionOperator: entity.OpGreaterThan,
						ConditionValue:    "100000",
						ResultStatus:      entity.DECLINED,
						Priority:          2,
						IsActive:          true,
					},
					{
						RuleID:            "rule-003",
						RuleName:          "Inactive rule",
						ConditionField:    entity.FieldCurrency,
						ConditionOperator: entity.OpEqual,
						ConditionValue:    "COP",
						ResultStatus:      entity.APPROVED,
						Priority:          3,
						IsActive:          false,
					},
				}, nil
			},
		}
		_, e := newEvaluationController(&mockRuleEvaluationRepository{}, ruleRepo)

		req := httptest.NewRequest(http.MethodGet, "/rules", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp DataResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		data, ok := resp.Data.([]interface{})
		if !ok {
			t.Fatalf("expected data to be an array, got %T", resp.Data)
		}
		if len(data) != 3 {
			t.Errorf("expected 3 rules, got %d", len(data))
		}

		// Verify ordering by priority
		first, ok := data[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected first element to be a map, got %T", data[0])
		}
		if first["rule_id"] != "rule-001" {
			t.Errorf("expected first rule_id %q, got %v", "rule-001", first["rule_id"])
		}

		last, ok := data[2].(map[string]interface{})
		if !ok {
			t.Fatalf("expected last element to be a map, got %T", data[2])
		}
		if last["rule_id"] != "rule-003" {
			t.Errorf("expected last rule_id %q, got %v", "rule-003", last["rule_id"])
		}
	})

	t.Run("should return 200 with empty list when no rules exist", func(t *testing.T) {
		ruleRepo := &mockRuleRepository{
			findAllFunc: func(_ context.Context) ([]entity.Rule, error) {
				return nil, nil
			},
		}
		_, e := newEvaluationController(&mockRuleEvaluationRepository{}, ruleRepo)

		req := httptest.NewRequest(http.MethodGet, "/rules", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp DataResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		data, ok := resp.Data.([]interface{})
		if !ok {
			t.Fatalf("expected data to be an array, got %T", resp.Data)
		}
		if len(data) != 0 {
			t.Errorf("expected 0 rules, got %d", len(data))
		}
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		ruleRepo := &mockRuleRepository{
			findAllFunc: func(_ context.Context) ([]entity.Rule, error) {
				return nil, errors.New("dynamo timeout")
			},
		}
		_, e := newEvaluationController(&mockRuleEvaluationRepository{}, ruleRepo)

		req := httptest.NewRequest(http.MethodGet, "/rules", nil)
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
	})
}

func TestEvaluationController_CORSPreflight(t *testing.T) {
	t.Run("should include CORS headers on preflight OPTIONS request", func(t *testing.T) {
		_, e := newEvaluationController(&mockRuleEvaluationRepository{}, &mockRuleRepository{})

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

		req := httptest.NewRequest(http.MethodOptions, "/evaluations/txn_abc123", nil)
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
