package usecase

import (
	"context"
	"errors"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"pgregory.net/rapid"
)

// --- Hand-written mocks ---

type mockRuleRepository struct {
	findFunc    func(ctx context.Context) ([]entity.Rule, error)
	findAllFunc func(ctx context.Context) ([]entity.Rule, error)
}

func (m *mockRuleRepository) FindActiveRulesSortedByPriority(ctx context.Context) ([]entity.Rule, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx)
	}
	return nil, nil
}

func (m *mockRuleRepository) FindAll(ctx context.Context) ([]entity.Rule, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	if m.findFunc != nil {
		return m.findFunc(ctx)
	}
	return nil, nil
}

type mockDecisionPublisher struct {
	publishFunc func(ctx context.Context, result *entity.DecisionResult) error
	lastResult  *entity.DecisionResult
	called      bool
}

func (m *mockDecisionPublisher) Publish(ctx context.Context, result *entity.DecisionResult) error {
	m.called = true
	m.lastResult = result
	if m.publishFunc != nil {
		return m.publishFunc(ctx, result)
	}
	return nil
}

type mockFraudScoreRequestPublisher struct {
	publishFunc     func(ctx context.Context, transaction *entity.TransactionMessage) error
	lastTransaction *entity.TransactionMessage
	called          bool
}

func (m *mockFraudScoreRequestPublisher) Publish(ctx context.Context, transaction *entity.TransactionMessage) error {
	m.called = true
	m.lastTransaction = transaction
	if m.publishFunc != nil {
		return m.publishFunc(ctx, transaction)
	}
	return nil
}

type mockRuleEvaluationRepository struct {
	saveBatchFunc func(ctx context.Context, results []entity.RuleEvaluationResult) error
	findFunc      func(ctx context.Context, transactionID string) ([]entity.RuleEvaluationResult, error)
	lastResults   []entity.RuleEvaluationResult
	saveCalled    bool
}

func (m *mockRuleEvaluationRepository) SaveBatch(ctx context.Context, results []entity.RuleEvaluationResult) error {
	m.saveCalled = true
	m.lastResults = results
	if m.saveBatchFunc != nil {
		return m.saveBatchFunc(ctx, results)
	}
	return nil
}

func (m *mockRuleEvaluationRepository) FindByTransactionID(ctx context.Context, transactionID string) ([]entity.RuleEvaluationResult, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, transactionID)
	}
	return nil, nil
}

// --- Helpers ---

func newTestTransaction() *entity.TransactionMessage {
	return &entity.TransactionMessage{
		ID:                "tx-123",
		AmountInCents:     50000,
		Currency:          "USD",
		PaymentMethod:     "CARD",
		CustomerID:        "cust-1",
		CustomerName:      "John Doe",
		CustomerEmail:     "john@example.com",
		CustomerPhone:     "555-0100",
		CustomerIPAddress: "192.168.1.1",
		Status:            "PENDING",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// --- Tests ---

func TestEvaluateTransactionUseCase_Execute(t *testing.T) {
	tests := []struct {
		name                    string
		transaction             *entity.TransactionMessage
		ruleRepo                *mockRuleRepository
		publisher               *mockDecisionPublisher
		fraudScorePublisher     *mockFraudScoreRequestPublisher
		wantErr                 error
		wantStatus              entity.DecisionStatus
		wantDecisionPublished   bool
		wantFraudScorePublished bool
	}{
		{
			name:                "nil transaction returns ErrTransactionNil",
			transaction:         nil,
			ruleRepo:            &mockRuleRepository{},
			publisher:           &mockDecisionPublisher{},
			fraudScorePublisher: &mockFraudScoreRequestPublisher{},
			wantErr:             ErrTransactionNil,
		},
		{
			name:        "rule retrieval failure returns ErrRuleRetrievalFailed",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return nil, errors.New("dynamo timeout")
				},
			},
			publisher:           &mockDecisionPublisher{},
			fraudScorePublisher: &mockFraudScoreRequestPublisher{},
			wantErr:             ErrRuleRetrievalFailed,
		},
		{
			name:        "decision publish failure returns ErrDecisionPublishFailed",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{}, nil
				},
			},
			publisher: &mockDecisionPublisher{
				publishFunc: func(_ context.Context, _ *entity.DecisionResult) error {
					return errors.New("kafka unavailable")
				},
			},
			fraudScorePublisher: &mockFraudScoreRequestPublisher{},
			wantErr:             ErrDecisionPublishFailed,
		},
		{
			name:        "no rules match returns default APPROVED via decision publisher",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{}, nil
				},
			},
			publisher:               &mockDecisionPublisher{},
			fraudScorePublisher:     &mockFraudScoreRequestPublisher{},
			wantStatus:              entity.APPROVED,
			wantDecisionPublished:   true,
			wantFraudScorePublished: false,
		},
		{
			name:        "DECLINED result publishes to decision publisher",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{
						{
							RuleID:            "rule-1",
							RuleName:          "High amount",
							ConditionField:    entity.FieldAmountInCents,
							ConditionOperator: entity.OpGreaterThan,
							ConditionValue:    "10000",
							ResultStatus:      entity.DECLINED,
							Priority:          1,
							IsActive:          true,
						},
					}, nil
				},
			},
			publisher:               &mockDecisionPublisher{},
			fraudScorePublisher:     &mockFraudScoreRequestPublisher{},
			wantStatus:              entity.DECLINED,
			wantDecisionPublished:   true,
			wantFraudScorePublished: false,
		},
		{
			name:        "FRAUD_CHECK result routes to fraud score publisher, not decision publisher",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{
						{
							RuleID:            "rule-1",
							RuleName:          "Flag for fraud check",
							ConditionField:    entity.FieldAmountInCents,
							ConditionOperator: entity.OpGreaterThan,
							ConditionValue:    "10000",
							ResultStatus:      entity.FRAUDCHECK,
							Priority:          1,
							IsActive:          true,
						},
					}, nil
				},
			},
			publisher:               &mockDecisionPublisher{},
			fraudScorePublisher:     &mockFraudScoreRequestPublisher{},
			wantStatus:              entity.FRAUDCHECK,
			wantDecisionPublished:   false,
			wantFraudScorePublished: true,
		},
		{
			name:        "FRAUD_CHECK with publish failure returns ErrFraudScorePublishFailed",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{
						{
							RuleID:            "rule-1",
							RuleName:          "Flag for fraud check",
							ConditionField:    entity.FieldAmountInCents,
							ConditionOperator: entity.OpGreaterThan,
							ConditionValue:    "10000",
							ResultStatus:      entity.FRAUDCHECK,
							Priority:          1,
							IsActive:          true,
						},
					}, nil
				},
			},
			publisher: &mockDecisionPublisher{},
			fraudScorePublisher: &mockFraudScoreRequestPublisher{
				publishFunc: func(_ context.Context, _ *entity.TransactionMessage) error {
					return errors.New("kafka unavailable")
				},
			},
			wantErr: ErrFraudScorePublishFailed,
		},
		{
			name:        "multiple rules first-match-wins FRAUD_CHECK routes to fraud score publisher",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{
						{
							RuleID:            "rule-1",
							RuleName:          "Flag for fraud check",
							ConditionField:    entity.FieldAmountInCents,
							ConditionOperator: entity.OpGreaterThan,
							ConditionValue:    "10000",
							ResultStatus:      entity.FRAUDCHECK,
							Priority:          1,
							IsActive:          true,
						},
						{
							RuleID:            "rule-2",
							RuleName:          "Decline high amount",
							ConditionField:    entity.FieldAmountInCents,
							ConditionOperator: entity.OpGreaterThan,
							ConditionValue:    "10000",
							ResultStatus:      entity.DECLINED,
							Priority:          2,
							IsActive:          true,
						},
					}, nil
				},
			},
			publisher:               &mockDecisionPublisher{},
			fraudScorePublisher:     &mockFraudScoreRequestPublisher{},
			wantStatus:              entity.FRAUDCHECK,
			wantDecisionPublished:   false,
			wantFraudScorePublished: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewEvaluateTransactionUseCase(tc.ruleRepo, tc.publisher, tc.fraudScorePublisher, &mockRuleEvaluationRepository{}, zerolog.Nop())
			result, err := uc.Execute(context.Background(), tc.transaction)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error wrapping %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}
				if result != nil {
					t.Fatalf("expected nil result on error, got %+v", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.TransactionID != tc.transaction.ID {
				t.Errorf("expected TransactionID %q, got %q", tc.transaction.ID, result.TransactionID)
			}
			if result.Status != tc.wantStatus {
				t.Errorf("expected Status %q, got %q", tc.wantStatus, result.Status)
			}

			// Verify decision publisher routing
			if tc.wantDecisionPublished && !tc.publisher.called {
				t.Error("expected decision publisher to be called, but it was not")
			}
			if !tc.wantDecisionPublished && tc.publisher.called {
				t.Error("expected decision publisher NOT to be called, but it was")
			}
			if tc.wantDecisionPublished && tc.publisher.lastResult != nil {
				if tc.publisher.lastResult.TransactionID != tc.transaction.ID {
					t.Errorf("published TransactionID %q, want %q", tc.publisher.lastResult.TransactionID, tc.transaction.ID)
				}
				if tc.publisher.lastResult.Status != tc.wantStatus {
					t.Errorf("published Status %q, want %q", tc.publisher.lastResult.Status, tc.wantStatus)
				}
			}

			// Verify fraud score publisher routing
			if tc.wantFraudScorePublished && !tc.fraudScorePublisher.called {
				t.Error("expected fraud score publisher to be called, but it was not")
			}
			if !tc.wantFraudScorePublished && tc.fraudScorePublisher.called {
				t.Error("expected fraud score publisher NOT to be called, but it was")
			}
			if tc.wantFraudScorePublished && tc.fraudScorePublisher.lastTransaction != nil {
				if tc.fraudScorePublisher.lastTransaction.ID != tc.transaction.ID {
					t.Errorf("fraud score published transaction ID %q, want %q", tc.fraudScorePublisher.lastTransaction.ID, tc.transaction.ID)
				}
			}
		})
	}
}

// --- Tests for rule evaluation persistence (Task 3.4) ---
// Validates: Requirements 3.1, 3.2

func TestEvaluateTransactionUseCase_RuleEvaluationPersistence(t *testing.T) {
	t.Run("SaveBatch called with correct number of results matching active rules count", func(t *testing.T) {
		tx := newTestTransaction()
		rules := []entity.Rule{
			{
				RuleID:            "rule-1",
				RuleName:          "High amount",
				ConditionField:    entity.FieldAmountInCents,
				ConditionOperator: entity.OpGreaterThan,
				ConditionValue:    "10000",
				ResultStatus:      entity.DECLINED,
				Priority:          1,
				IsActive:          true,
			},
			{
				RuleID:            "rule-2",
				RuleName:          "Block CRYPTO",
				ConditionField:    entity.FieldPaymentMethod,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    "CRYPTO",
				ResultStatus:      entity.DECLINED,
				Priority:          2,
				IsActive:          true,
			},
			{
				RuleID:            "rule-3",
				RuleName:          "Flag USD",
				ConditionField:    entity.FieldCurrency,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    "USD",
				ResultStatus:      entity.FRAUDCHECK,
				Priority:          3,
				IsActive:          true,
			},
		}

		ruleRepo := &mockRuleRepository{
			findFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}
		ruleEvalRepo := &mockRuleEvaluationRepository{}

		uc := NewEvaluateTransactionUseCase(ruleRepo, &mockDecisionPublisher{}, &mockFraudScoreRequestPublisher{}, ruleEvalRepo, zerolog.Nop())
		_, err := uc.Execute(context.Background(), tx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ruleEvalRepo.saveCalled {
			t.Fatal("expected SaveBatch to be called")
		}
		if len(ruleEvalRepo.lastResults) != len(rules) {
			t.Fatalf("expected %d results, got %d", len(rules), len(ruleEvalRepo.lastResults))
		}
	})

	t.Run("each result has all required fields populated", func(t *testing.T) {
		tx := newTestTransaction()
		rules := []entity.Rule{
			{
				RuleID:            "rule-1",
				RuleName:          "High amount",
				ConditionField:    entity.FieldAmountInCents,
				ConditionOperator: entity.OpGreaterThan,
				ConditionValue:    "10000",
				ResultStatus:      entity.DECLINED,
				Priority:          1,
				IsActive:          true,
			},
			{
				RuleID:            "rule-2",
				RuleName:          "Block CRYPTO",
				ConditionField:    entity.FieldPaymentMethod,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    "CRYPTO",
				ResultStatus:      entity.DECLINED,
				Priority:          2,
				IsActive:          true,
			},
		}

		ruleRepo := &mockRuleRepository{
			findFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}
		ruleEvalRepo := &mockRuleEvaluationRepository{}

		uc := NewEvaluateTransactionUseCase(ruleRepo, &mockDecisionPublisher{}, &mockFraudScoreRequestPublisher{}, ruleEvalRepo, zerolog.Nop())
		_, err := uc.Execute(context.Background(), tx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, result := range ruleEvalRepo.lastResults {
			if result.TransactionID == "" {
				t.Errorf("result[%d]: TransactionID is empty", i)
			}
			if result.TransactionID != tx.ID {
				t.Errorf("result[%d]: TransactionID = %q, want %q", i, result.TransactionID, tx.ID)
			}
			if result.RuleID == "" {
				t.Errorf("result[%d]: RuleID is empty", i)
			}
			if result.RuleName == "" {
				t.Errorf("result[%d]: RuleName is empty", i)
			}
			if result.ConditionField == "" {
				t.Errorf("result[%d]: ConditionField is empty", i)
			}
			if result.ConditionOperator == "" {
				t.Errorf("result[%d]: ConditionOperator is empty", i)
			}
			if result.ConditionValue == "" {
				t.Errorf("result[%d]: ConditionValue is empty", i)
			}
			if result.ActualFieldValue == "" {
				t.Errorf("result[%d]: ActualFieldValue is empty", i)
			}
			if result.ResultStatus == "" {
				t.Errorf("result[%d]: ResultStatus is empty", i)
			}
			if result.EvaluatedAt.IsZero() {
				t.Errorf("result[%d]: EvaluatedAt is zero", i)
			}
			if result.Priority == 0 {
				t.Errorf("result[%d]: Priority is zero", i)
			}
			// Verify rule-specific fields match
			if result.RuleID != rules[i].RuleID {
				t.Errorf("result[%d]: RuleID = %q, want %q", i, result.RuleID, rules[i].RuleID)
			}
			if result.RuleName != rules[i].RuleName {
				t.Errorf("result[%d]: RuleName = %q, want %q", i, result.RuleName, rules[i].RuleName)
			}
			if result.Priority != rules[i].Priority {
				t.Errorf("result[%d]: Priority = %d, want %d", i, result.Priority, rules[i].Priority)
			}
		}
	})

	t.Run("persistence failure does not block decision publishing", func(t *testing.T) {
		tx := newTestTransaction()
		rules := []entity.Rule{
			{
				RuleID:            "rule-1",
				RuleName:          "High amount",
				ConditionField:    entity.FieldAmountInCents,
				ConditionOperator: entity.OpGreaterThan,
				ConditionValue:    "10000",
				ResultStatus:      entity.DECLINED,
				Priority:          1,
				IsActive:          true,
			},
		}

		ruleRepo := &mockRuleRepository{
			findFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}
		decisionPub := &mockDecisionPublisher{}
		ruleEvalRepo := &mockRuleEvaluationRepository{
			saveBatchFunc: func(_ context.Context, _ []entity.RuleEvaluationResult) error {
				return errors.New("dynamodb write failure")
			},
		}

		uc := NewEvaluateTransactionUseCase(ruleRepo, decisionPub, &mockFraudScoreRequestPublisher{}, ruleEvalRepo, zerolog.Nop())
		result, err := uc.Execute(context.Background(), tx)

		if err != nil {
			t.Fatalf("expected no error despite persistence failure, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.Status != entity.DECLINED {
			t.Errorf("expected DECLINED status, got %q", result.Status)
		}
		if !decisionPub.called {
			t.Error("expected decision publisher to be called despite persistence failure")
		}
		if !ruleEvalRepo.saveCalled {
			t.Error("expected SaveBatch to have been attempted")
		}
	})
}

// Feature: fraud-score-service, Property 1: FRAUD_CHECK routes to fraud score request, not decision results
// Validates: Requirements 1.1, 1.2
func TestProperty_FraudCheckRoutesToFraudScoreRequest(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate an arbitrary transaction
		tx := &entity.TransactionMessage{
			ID:                rapid.String().Draw(t, "id"),
			AmountInCents:     rapid.Int64Range(0, 10_000_000).Draw(t, "amount"),
			Currency:          rapid.SampledFrom([]string{"USD", "COP", "EUR"}).Draw(t, "currency"),
			PaymentMethod:     rapid.SampledFrom([]string{"CARD", "BANK_TRANSFER", "CRYPTO"}).Draw(t, "paymentMethod"),
			CustomerID:        rapid.String().Draw(t, "customerID"),
			CustomerName:      rapid.String().Draw(t, "customerName"),
			CustomerEmail:     rapid.String().Draw(t, "customerEmail"),
			CustomerPhone:     rapid.String().Draw(t, "customerPhone"),
			CustomerIPAddress: rapid.String().Draw(t, "customerIP"),
			Status:            "PENDING",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Build a rule that matches this transaction and produces FRAUD_CHECK.
		// Strategy: pick a field from the transaction and create a rule that matches it.
		fieldChoice := rapid.IntRange(0, 4).Draw(t, "fieldChoice")

		var rule entity.Rule
		switch fieldChoice {
		case 0:
			// Match on amount using GREATER_THAN_OR_EQUAL so it always matches
			rule = entity.Rule{
				RuleID:            "rule-fc",
				RuleName:          "Fraud check rule",
				ConditionField:    entity.FieldAmountInCents,
				ConditionOperator: entity.OpGreaterThanOrEqual,
				ConditionValue:    fmt.Sprintf("%d", tx.AmountInCents),
				ResultStatus:      entity.FRAUDCHECK,
				Priority:          1,
				IsActive:          true,
			}
		case 1:
			rule = entity.Rule{
				RuleID:            "rule-fc",
				RuleName:          "Fraud check rule",
				ConditionField:    entity.FieldCurrency,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    tx.Currency,
				ResultStatus:      entity.FRAUDCHECK,
				Priority:          1,
				IsActive:          true,
			}
		case 2:
			rule = entity.Rule{
				RuleID:            "rule-fc",
				RuleName:          "Fraud check rule",
				ConditionField:    entity.FieldPaymentMethod,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    tx.PaymentMethod,
				ResultStatus:      entity.FRAUDCHECK,
				Priority:          1,
				IsActive:          true,
			}
		case 3:
			rule = entity.Rule{
				RuleID:            "rule-fc",
				RuleName:          "Fraud check rule",
				ConditionField:    entity.FieldCustomerID,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    tx.CustomerID,
				ResultStatus:      entity.FRAUDCHECK,
				Priority:          1,
				IsActive:          true,
			}
		default:
			rule = entity.Rule{
				RuleID:            "rule-fc",
				RuleName:          "Fraud check rule",
				ConditionField:    entity.FieldCustomerIPAddress,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    tx.CustomerIPAddress,
				ResultStatus:      entity.FRAUDCHECK,
				Priority:          1,
				IsActive:          true,
			}
		}

		rules := []entity.Rule{rule}

		decisionPub := &mockDecisionPublisher{}
		fraudScorePub := &mockFraudScoreRequestPublisher{}
		ruleRepo := &mockRuleRepository{
			findFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}

		uc := NewEvaluateTransactionUseCase(ruleRepo, decisionPub, fraudScorePub, &mockRuleEvaluationRepository{}, zerolog.Nop())
		result, err := uc.Execute(context.Background(), tx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.Status != entity.FRAUDCHECK {
			t.Fatalf("expected FRAUD_CHECK status, got %q", result.Status)
		}

		// Assert: fraud score publisher IS called with the correct transaction
		if !fraudScorePub.called {
			t.Fatal("expected fraud score publisher to be called, but it was not")
		}
		if fraudScorePub.lastTransaction == nil {
			t.Fatal("expected fraud score publisher to receive the transaction")
		}
		if fraudScorePub.lastTransaction.ID != tx.ID {
			t.Fatalf("fraud score publisher got transaction ID %q, want %q", fraudScorePub.lastTransaction.ID, tx.ID)
		}

		// Assert: decision publisher is NOT called
		if decisionPub.called {
			t.Fatal("expected decision publisher NOT to be called, but it was")
		}
	})
}

// Feature: fraud-score-service, Property 2: Non-FRAUD_CHECK results route to decision results, not fraud score request
// Validates: Requirements 1.3, 1.4
func TestProperty_NonFraudCheckRoutesToDecisionResults(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate an arbitrary transaction
		tx := &entity.TransactionMessage{
			ID:                rapid.String().Draw(t, "id"),
			AmountInCents:     rapid.Int64Range(0, 10_000_000).Draw(t, "amount"),
			Currency:          rapid.SampledFrom([]string{"USD", "COP", "EUR"}).Draw(t, "currency"),
			PaymentMethod:     rapid.SampledFrom([]string{"CARD", "BANK_TRANSFER", "CRYPTO"}).Draw(t, "paymentMethod"),
			CustomerID:        rapid.String().Draw(t, "customerID"),
			CustomerName:      rapid.String().Draw(t, "customerName"),
			CustomerEmail:     rapid.String().Draw(t, "customerEmail"),
			CustomerPhone:     rapid.String().Draw(t, "customerPhone"),
			CustomerIPAddress: rapid.String().Draw(t, "customerIP"),
			Status:            "PENDING",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Pick either APPROVED or DECLINED as the result status
		resultStatus := rapid.SampledFrom([]entity.DecisionStatus{entity.APPROVED, entity.DECLINED}).Draw(t, "resultStatus")

		// Build a rule that matches this transaction and produces the chosen non-FRAUD_CHECK status
		fieldChoice := rapid.IntRange(0, 4).Draw(t, "fieldChoice")

		var rule entity.Rule
		switch fieldChoice {
		case 0:
			rule = entity.Rule{
				RuleID:            "rule-nfc",
				RuleName:          "Non fraud check rule",
				ConditionField:    entity.FieldAmountInCents,
				ConditionOperator: entity.OpGreaterThanOrEqual,
				ConditionValue:    fmt.Sprintf("%d", tx.AmountInCents),
				ResultStatus:      resultStatus,
				Priority:          1,
				IsActive:          true,
			}
		case 1:
			rule = entity.Rule{
				RuleID:            "rule-nfc",
				RuleName:          "Non fraud check rule",
				ConditionField:    entity.FieldCurrency,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    tx.Currency,
				ResultStatus:      resultStatus,
				Priority:          1,
				IsActive:          true,
			}
		case 2:
			rule = entity.Rule{
				RuleID:            "rule-nfc",
				RuleName:          "Non fraud check rule",
				ConditionField:    entity.FieldPaymentMethod,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    tx.PaymentMethod,
				ResultStatus:      resultStatus,
				Priority:          1,
				IsActive:          true,
			}
		case 3:
			rule = entity.Rule{
				RuleID:            "rule-nfc",
				RuleName:          "Non fraud check rule",
				ConditionField:    entity.FieldCustomerID,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    tx.CustomerID,
				ResultStatus:      resultStatus,
				Priority:          1,
				IsActive:          true,
			}
		default:
			rule = entity.Rule{
				RuleID:            "rule-nfc",
				RuleName:          "Non fraud check rule",
				ConditionField:    entity.FieldCustomerIPAddress,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    tx.CustomerIPAddress,
				ResultStatus:      resultStatus,
				Priority:          1,
				IsActive:          true,
			}
		}

		rules := []entity.Rule{rule}

		decisionPub := &mockDecisionPublisher{}
		fraudScorePub := &mockFraudScoreRequestPublisher{}
		ruleRepo := &mockRuleRepository{
			findFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}

		uc := NewEvaluateTransactionUseCase(ruleRepo, decisionPub, fraudScorePub, &mockRuleEvaluationRepository{}, zerolog.Nop())
		result, err := uc.Execute(context.Background(), tx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Assert: status is APPROVED or DECLINED
		if result.Status != entity.APPROVED && result.Status != entity.DECLINED {
			t.Fatalf("expected APPROVED or DECLINED status, got %q", result.Status)
		}
		if result.Status != resultStatus {
			t.Fatalf("expected status %q, got %q", resultStatus, result.Status)
		}

		// Assert: decision publisher IS called with the correct transaction ID and status
		if !decisionPub.called {
			t.Fatal("expected decision publisher to be called, but it was not")
		}
		if decisionPub.lastResult == nil {
			t.Fatal("expected decision publisher to receive a result")
		}
		if decisionPub.lastResult.TransactionID != tx.ID {
			t.Fatalf("decision publisher got transaction ID %q, want %q", decisionPub.lastResult.TransactionID, tx.ID)
		}
		if decisionPub.lastResult.Status != resultStatus {
			t.Fatalf("decision publisher got status %q, want %q", decisionPub.lastResult.Status, resultStatus)
		}

		// Assert: fraud score publisher is NOT called
		if fraudScorePub.called {
			t.Fatal("expected fraud score publisher NOT to be called, but it was")
		}
	})
}

// Feature: fraud-analyst-dashboard, Property 5: Rule evaluation persistence completeness
// Validates: Requirements 3.1, 3.2
func TestProperty_RuleEvaluationPersistenceCompleteness(t *testing.T) {
	// Generator for a non-empty alphanumeric string (ensures fields are never empty)
	nonEmptyStr := rapid.StringMatching(`[a-zA-Z0-9]{1,20}`)

	// Valid condition fields that produce non-empty ActualFieldValue from TransactionMessage
	validFields := []entity.ConditionField{
		entity.FieldAmountInCents,
		entity.FieldCurrency,
		entity.FieldPaymentMethod,
		entity.FieldCustomerID,
		entity.FieldCustomerIPAddress,
	}

	// Operators valid for string fields
	stringOps := []entity.ConditionOperator{entity.OpEqual, entity.OpNotEqual}
	// Operators valid for numeric fields
	numericOps := []entity.ConditionOperator{
		entity.OpGreaterThan, entity.OpLessThan, entity.OpEqual,
		entity.OpNotEqual, entity.OpGreaterThanOrEqual, entity.OpLessThanOrEqual,
	}

	resultStatuses := []entity.DecisionStatus{entity.APPROVED, entity.DECLINED, entity.FRAUDCHECK}

	rapid.Check(t, func(t *rapid.T) {
		// Generate a random transaction with non-empty fields
		tx := &entity.TransactionMessage{
			ID:                nonEmptyStr.Draw(t, "txID"),
			AmountInCents:     rapid.Int64Range(1, 10_000_000).Draw(t, "amount"),
			Currency:          rapid.SampledFrom([]string{"USD", "COP", "EUR"}).Draw(t, "currency"),
			PaymentMethod:     rapid.SampledFrom([]string{"CARD", "BANK_TRANSFER", "CRYPTO"}).Draw(t, "paymentMethod"),
			CustomerID:        nonEmptyStr.Draw(t, "customerID"),
			CustomerName:      nonEmptyStr.Draw(t, "customerName"),
			CustomerEmail:     nonEmptyStr.Draw(t, "customerEmail"),
			CustomerPhone:     nonEmptyStr.Draw(t, "customerPhone"),
			CustomerIPAddress: nonEmptyStr.Draw(t, "customerIP"),
			Status:            "PENDING",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Generate 1–10 active rules with unique IDs and priorities >= 1
		ruleCount := rapid.IntRange(1, 10).Draw(t, "ruleCount")
		rules := make([]entity.Rule, ruleCount)

		for i := range ruleCount {
			field := rapid.SampledFrom(validFields).Draw(t, fmt.Sprintf("field_%d", i))

			var op entity.ConditionOperator
			if field == entity.FieldAmountInCents {
				op = rapid.SampledFrom(numericOps).Draw(t, fmt.Sprintf("op_%d", i))
			} else {
				op = rapid.SampledFrom(stringOps).Draw(t, fmt.Sprintf("op_%d", i))
			}

			var condValue string
			if field == entity.FieldAmountInCents {
				condValue = fmt.Sprintf("%d", rapid.Int64Range(1, 10_000_000).Draw(t, fmt.Sprintf("condVal_%d", i)))
			} else {
				condValue = nonEmptyStr.Draw(t, fmt.Sprintf("condVal_%d", i))
			}

			rules[i] = entity.Rule{
				RuleID:            fmt.Sprintf("rule-%d", i+1),
				RuleName:          fmt.Sprintf("Rule %d", i+1),
				ConditionField:    field,
				ConditionOperator: op,
				ConditionValue:    condValue,
				ResultStatus:      rapid.SampledFrom(resultStatuses).Draw(t, fmt.Sprintf("status_%d", i)),
				Priority:          i + 1,
				IsActive:          true,
			}
		}

		ruleRepo := &mockRuleRepository{
			findFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}
		ruleEvalRepo := &mockRuleEvaluationRepository{}

		uc := NewEvaluateTransactionUseCase(
			ruleRepo, &mockDecisionPublisher{}, &mockFraudScoreRequestPublisher{},
			ruleEvalRepo, zerolog.Nop(),
		)
		_, _ = uc.Execute(context.Background(), tx)

		// Assert: SaveBatch was called
		if !ruleEvalRepo.saveCalled {
			t.Fatal("expected SaveBatch to be called")
		}

		// Assert: exactly one RuleEvaluationResult per active rule
		if len(ruleEvalRepo.lastResults) != len(rules) {
			t.Fatalf("expected %d results, got %d", len(rules), len(ruleEvalRepo.lastResults))
		}

		// Assert: each result has all required fields non-empty/non-zero
		for i, result := range ruleEvalRepo.lastResults {
			if result.TransactionID == "" {
				t.Errorf("result[%d]: TransactionID is empty", i)
			}
			if result.RuleID == "" {
				t.Errorf("result[%d]: RuleID is empty", i)
			}
			if result.RuleName == "" {
				t.Errorf("result[%d]: RuleName is empty", i)
			}
			if result.ConditionField == "" {
				t.Errorf("result[%d]: ConditionField is empty", i)
			}
			if result.ConditionOperator == "" {
				t.Errorf("result[%d]: ConditionOperator is empty", i)
			}
			if result.ConditionValue == "" {
				t.Errorf("result[%d]: ConditionValue is empty", i)
			}
			if result.ActualFieldValue == "" {
				t.Errorf("result[%d]: ActualFieldValue is empty", i)
			}
			if result.ResultStatus == "" {
				t.Errorf("result[%d]: ResultStatus is empty", i)
			}
			if result.EvaluatedAt.IsZero() {
				t.Errorf("result[%d]: EvaluatedAt is zero", i)
			}
			if result.Priority == 0 {
				t.Errorf("result[%d]: Priority is zero", i)
			}
		}
	})
}
