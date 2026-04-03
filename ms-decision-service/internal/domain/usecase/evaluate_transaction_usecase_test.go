package usecase

import (
	"context"
	"errors"
	"ms-decision-service/internal/domain/entity"
	"testing"
	"time"
)

// --- Hand-written mocks ---

type mockRuleRepository struct {
	findFunc func(ctx context.Context) ([]entity.Rule, error)
}

func (m *mockRuleRepository) FindActiveRulesSortedByPriority(ctx context.Context) ([]entity.Rule, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx)
	}
	return nil, nil
}

type mockDecisionPublisher struct {
	publishFunc func(ctx context.Context, result *entity.DecisionResult) error
	lastResult  *entity.DecisionResult
}

func (m *mockDecisionPublisher) Publish(ctx context.Context, result *entity.DecisionResult) error {
	m.lastResult = result
	if m.publishFunc != nil {
		return m.publishFunc(ctx, result)
	}
	return nil
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
		name          string
		transaction   *entity.TransactionMessage
		ruleRepo      *mockRuleRepository
		publisher     *mockDecisionPublisher
		wantErr       error
		wantStatus    entity.DecisionStatus
		wantPublished bool
	}{
		{
			name:        "nil transaction returns ErrTransactionNil",
			transaction: nil,
			ruleRepo:    &mockRuleRepository{},
			publisher:   &mockDecisionPublisher{},
			wantErr:     ErrTransactionNil,
		},
		{
			name:        "rule retrieval failure returns ErrRuleRetrievalFailed",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return nil, errors.New("dynamo timeout")
				},
			},
			publisher: &mockDecisionPublisher{},
			wantErr:   ErrRuleRetrievalFailed,
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
			wantErr: ErrDecisionPublishFailed,
		},
		{
			name:        "no rules match returns default APPROVED",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{}, nil
				},
			},
			publisher:     &mockDecisionPublisher{},
			wantStatus:    entity.APPROVED,
			wantPublished: true,
		},
		{
			name:        "single rule match returns that rule ResultStatus",
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
			publisher:     &mockDecisionPublisher{},
			wantStatus:    entity.DECLINED,
			wantPublished: true,
		},
		{
			name:        "multiple rules first-match-wins by priority order",
			transaction: newTestTransaction(),
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					// Rules already sorted by priority ascending
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
			publisher:     &mockDecisionPublisher{},
			wantStatus:    entity.FRAUDCHECK,
			wantPublished: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewEvaluateTransactionUseCase(tc.ruleRepo, tc.publisher)
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
			if tc.wantPublished && tc.publisher.lastResult == nil {
				t.Error("expected decision to be published, but publisher was not called")
			}
			if tc.wantPublished && tc.publisher.lastResult != nil {
				if tc.publisher.lastResult.TransactionID != tc.transaction.ID {
					t.Errorf("published TransactionID %q, want %q", tc.publisher.lastResult.TransactionID, tc.transaction.ID)
				}
				if tc.publisher.lastResult.Status != tc.wantStatus {
					t.Errorf("published Status %q, want %q", tc.publisher.lastResult.Status, tc.wantStatus)
				}
			}
		})
	}
}
