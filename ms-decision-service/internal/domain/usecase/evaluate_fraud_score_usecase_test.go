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

// --- Generators ---

// genFraudScoreRule generates a fraud score rule with ConditionField=FieldFraudScore
// and ResultStatus restricted to APPROVED or DECLINED (never FRAUD_CHECK).
func genFraudScoreRule() *rapid.Generator[entity.Rule] {
	return rapid.Custom[entity.Rule](func(t *rapid.T) entity.Rule {
		operators := []entity.ConditionOperator{
			entity.OpGreaterThan,
			entity.OpLessThan,
			entity.OpEqual,
			entity.OpNotEqual,
			entity.OpGreaterThanOrEqual,
			entity.OpLessThanOrEqual,
		}
		statuses := []entity.DecisionStatus{entity.APPROVED, entity.DECLINED}

		return entity.Rule{
			RuleID:            rapid.String().Draw(t, "ruleID"),
			RuleName:          rapid.String().Draw(t, "ruleName"),
			ConditionField:    entity.FieldFraudScore,
			ConditionOperator: rapid.SampledFrom(operators).Draw(t, "operator"),
			ConditionValue:    fmt.Sprintf("%d", rapid.IntRange(0, 100).Draw(t, "conditionValue")),
			ResultStatus:      rapid.SampledFrom(statuses).Draw(t, "resultStatus"),
			Priority:          rapid.IntRange(1, 1000).Draw(t, "priority"),
			IsActive:          true,
		}
	})
}

// Feature: fraud-score-service, Property 7: Fraud score rules never produce FRAUD_CHECK
// Validates: Requirements 5.4
func TestProperty_FraudScoreRulesNeverProduceFraudCheck(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate an arbitrary fraud score in [0, 100]
		fraudScore := rapid.IntRange(0, 100).Draw(t, "fraudScore")

		// Generate a set of fraud score rules (0–10 rules), all with ConditionField=FieldFraudScore
		// and ResultStatus restricted to APPROVED or DECLINED
		numRules := rapid.IntRange(0, 10).Draw(t, "numRules")
		rules := make([]entity.Rule, 0, numRules)
		for i := 0; i < numRules; i++ {
			rules = append(rules, genFraudScoreRule().Draw(t, fmt.Sprintf("rule-%d", i)))
		}

		// Build a FraudScoreCalculatedMessage
		msg := &entity.FraudScoreCalculatedMessage{
			TransactionID: rapid.String().Draw(t, "transactionID"),
			FraudScore:    fraudScore,
			CalculatedAt:  time.Now(),
		}

		// Mock rule repo returns the generated fraud score rules
		ruleRepo := &mockRuleRepository{
			findFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}
		decisionPub := &mockDecisionPublisher{}

		uc := NewEvaluateFraudScoreUseCase(ruleRepo, decisionPub, &mockRuleEvaluationRepository{}, zerolog.Nop())
		result, err := uc.Execute(context.Background(), msg)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// The core property: result status must NEVER be FRAUD_CHECK
		if result.Status == entity.FRAUDCHECK {
			t.Fatalf("fraud score rules produced FRAUD_CHECK for score=%d with %d rules, but should only produce APPROVED or DECLINED",
				fraudScore, len(rules))
		}

		// Additionally verify it's one of the two valid statuses
		if result.Status != entity.APPROVED && result.Status != entity.DECLINED {
			t.Fatalf("expected APPROVED or DECLINED, got %q", result.Status)
		}
	})
}

// Feature: fraud-score-service, Property 8: Fraud score evaluation produces a final decision and publishes it
// Validates: Requirements 5.3, 5.5
func TestProperty_FraudScoreEvaluationProducesDecisionAndPublishes(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a valid FraudScoreCalculatedMessage with fraud score in [0, 100]
		msg := &entity.FraudScoreCalculatedMessage{
			TransactionID: rapid.String().Draw(t, "transactionID"),
			FraudScore:    rapid.IntRange(0, 100).Draw(t, "fraudScore"),
			CalculatedAt:  time.Now(),
		}

		// Generate a set of fraud score rules (0–10 rules)
		numRules := rapid.IntRange(0, 10).Draw(t, "numRules")
		rules := make([]entity.Rule, 0, numRules)
		for i := 0; i < numRules; i++ {
			rules = append(rules, genFraudScoreRule().Draw(t, fmt.Sprintf("rule-%d", i)))
		}

		ruleRepo := &mockRuleRepository{
			findFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}
		decisionPub := &mockDecisionPublisher{}

		uc := NewEvaluateFraudScoreUseCase(ruleRepo, decisionPub, &mockRuleEvaluationRepository{}, zerolog.Nop())
		result, err := uc.Execute(context.Background(), msg)

		// Assert no error returned
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Assert decision publisher was called
		if !decisionPub.called {
			t.Fatal("expected decision publisher to be called, but it was not")
		}

		// Assert published result has correct transaction ID
		if decisionPub.lastResult == nil {
			t.Fatal("expected decision publisher to receive a result")
		}
		if decisionPub.lastResult.TransactionID != msg.TransactionID {
			t.Fatalf("published TransactionID %q, want %q", decisionPub.lastResult.TransactionID, msg.TransactionID)
		}

		// Assert published result status is APPROVED or DECLINED
		if decisionPub.lastResult.Status != entity.APPROVED && decisionPub.lastResult.Status != entity.DECLINED {
			t.Fatalf("expected published status to be APPROVED or DECLINED, got %q", decisionPub.lastResult.Status)
		}

		// Assert result matches what was published
		if result.TransactionID != msg.TransactionID {
			t.Fatalf("result TransactionID %q, want %q", result.TransactionID, msg.TransactionID)
		}
		if result.Status != entity.APPROVED && result.Status != entity.DECLINED {
			t.Fatalf("expected result status to be APPROVED or DECLINED, got %q", result.Status)
		}
	})
}

// --- Tests for rule evaluation persistence (Task 3.4) ---
// Validates: Requirements 3.1, 3.2

func TestEvaluateFraudScoreUseCase_RuleEvaluationPersistence(t *testing.T) {
	t.Run("SaveBatch called with correct number of results matching fraud score rules count", func(t *testing.T) {
		msg := &entity.FraudScoreCalculatedMessage{
			TransactionID: "tx-persist-1",
			FraudScore:    75,
			CalculatedAt:  time.Now(),
		}
		rules := []entity.Rule{
			{
				RuleID:            "rule-fs-1",
				RuleName:          "High score decline",
				ConditionField:    entity.FieldFraudScore,
				ConditionOperator: entity.OpGreaterThan,
				ConditionValue:    "70",
				ResultStatus:      entity.DECLINED,
				Priority:          1,
				IsActive:          true,
			},
			{
				RuleID:            "rule-fs-2",
				RuleName:          "Medium score approve",
				ConditionField:    entity.FieldFraudScore,
				ConditionOperator: entity.OpLessThanOrEqual,
				ConditionValue:    "50",
				ResultStatus:      entity.APPROVED,
				Priority:          2,
				IsActive:          true,
			},
			{
				// Non-fraud-score rule — should NOT be persisted by fraud score use case
				RuleID:            "rule-tx-1",
				RuleName:          "Block CRYPTO",
				ConditionField:    entity.FieldPaymentMethod,
				ConditionOperator: entity.OpEqual,
				ConditionValue:    "CRYPTO",
				ResultStatus:      entity.DECLINED,
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

		uc := NewEvaluateFraudScoreUseCase(ruleRepo, &mockDecisionPublisher{}, ruleEvalRepo, zerolog.Nop())
		_, err := uc.Execute(context.Background(), msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ruleEvalRepo.saveCalled {
			t.Fatal("expected SaveBatch to be called")
		}
		// Only fraud_score rules should be persisted (2 out of 3)
		expectedCount := 2
		if len(ruleEvalRepo.lastResults) != expectedCount {
			t.Fatalf("expected %d results (fraud_score rules only), got %d", expectedCount, len(ruleEvalRepo.lastResults))
		}
	})

	t.Run("each result has all required fields populated", func(t *testing.T) {
		msg := &entity.FraudScoreCalculatedMessage{
			TransactionID: "tx-persist-2",
			FraudScore:    45,
			CalculatedAt:  time.Now(),
		}
		rules := []entity.Rule{
			{
				RuleID:            "rule-fs-1",
				RuleName:          "High score decline",
				ConditionField:    entity.FieldFraudScore,
				ConditionOperator: entity.OpGreaterThan,
				ConditionValue:    "70",
				ResultStatus:      entity.DECLINED,
				Priority:          1,
				IsActive:          true,
			},
			{
				RuleID:            "rule-fs-2",
				RuleName:          "Low score approve",
				ConditionField:    entity.FieldFraudScore,
				ConditionOperator: entity.OpLessThanOrEqual,
				ConditionValue:    "50",
				ResultStatus:      entity.APPROVED,
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

		uc := NewEvaluateFraudScoreUseCase(ruleRepo, &mockDecisionPublisher{}, ruleEvalRepo, zerolog.Nop())
		_, err := uc.Execute(context.Background(), msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, result := range ruleEvalRepo.lastResults {
			if result.TransactionID == "" {
				t.Errorf("result[%d]: TransactionID is empty", i)
			}
			if result.TransactionID != msg.TransactionID {
				t.Errorf("result[%d]: TransactionID = %q, want %q", i, result.TransactionID, msg.TransactionID)
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
			// Verify the actual field value is the fraud score string
			if result.ActualFieldValue != fmt.Sprintf("%d", msg.FraudScore) {
				t.Errorf("result[%d]: ActualFieldValue = %q, want %q", i, result.ActualFieldValue, fmt.Sprintf("%d", msg.FraudScore))
			}
		}
	})

	t.Run("persistence failure does not block decision publishing", func(t *testing.T) {
		msg := &entity.FraudScoreCalculatedMessage{
			TransactionID: "tx-persist-3",
			FraudScore:    80,
			CalculatedAt:  time.Now(),
		}
		rules := []entity.Rule{
			{
				RuleID:            "rule-fs-1",
				RuleName:          "High score decline",
				ConditionField:    entity.FieldFraudScore,
				ConditionOperator: entity.OpGreaterThan,
				ConditionValue:    "70",
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

		uc := NewEvaluateFraudScoreUseCase(ruleRepo, decisionPub, ruleEvalRepo, zerolog.Nop())
		result, err := uc.Execute(context.Background(), msg)

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

// --- Unit Tests for EvaluateFraudScoreUseCase edge cases ---
// Validates: Requirements 5.3, 5.4, 5.5

func TestEvaluateFraudScoreUseCase_Execute(t *testing.T) {
	tests := []struct {
		name       string
		msg        *entity.FraudScoreCalculatedMessage
		ruleRepo   *mockRuleRepository
		publisher  *mockDecisionPublisher
		wantErr    error
		wantStatus entity.DecisionStatus
	}{
		{
			name: "fraud score 0 with LESS_THAN_OR_EQUAL 50 rule returns APPROVED",
			msg: &entity.FraudScoreCalculatedMessage{
				TransactionID: "tx-001",
				FraudScore:    0,
				CalculatedAt:  time.Now(),
			},
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{
						{
							RuleID:            "rule-1",
							RuleName:          "Low fraud score approve",
							ConditionField:    entity.FieldFraudScore,
							ConditionOperator: entity.OpLessThanOrEqual,
							ConditionValue:    "50",
							ResultStatus:      entity.APPROVED,
							Priority:          1,
							IsActive:          true,
						},
					}, nil
				},
			},
			publisher:  &mockDecisionPublisher{},
			wantStatus: entity.APPROVED,
		},
		{
			name: "fraud score 100 with GREATER_THAN 70 rule returns DECLINED",
			msg: &entity.FraudScoreCalculatedMessage{
				TransactionID: "tx-002",
				FraudScore:    100,
				CalculatedAt:  time.Now(),
			},
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{
						{
							RuleID:            "rule-2",
							RuleName:          "High fraud score decline",
							ConditionField:    entity.FieldFraudScore,
							ConditionOperator: entity.OpGreaterThan,
							ConditionValue:    "70",
							ResultStatus:      entity.DECLINED,
							Priority:          1,
							IsActive:          true,
						},
					}, nil
				},
			},
			publisher:  &mockDecisionPublisher{},
			wantStatus: entity.DECLINED,
		},
		{
			name: "no matching fraud score rules defaults to APPROVED",
			msg: &entity.FraudScoreCalculatedMessage{
				TransactionID: "tx-003",
				FraudScore:    50,
				CalculatedAt:  time.Now(),
			},
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{}, nil
				},
			},
			publisher:  &mockDecisionPublisher{},
			wantStatus: entity.APPROVED,
		},
		{
			name:      "nil message returns ErrFraudScoreMessageNil",
			msg:       nil,
			ruleRepo:  &mockRuleRepository{},
			publisher: &mockDecisionPublisher{},
			wantErr:   ErrFraudScoreMessageNil,
		},
		{
			name: "rule retrieval failure returns ErrRuleRetrievalFailed",
			msg: &entity.FraudScoreCalculatedMessage{
				TransactionID: "tx-004",
				FraudScore:    50,
				CalculatedAt:  time.Now(),
			},
			ruleRepo: &mockRuleRepository{
				findFunc: func(_ context.Context) ([]entity.Rule, error) {
					return nil, errors.New("dynamo timeout")
				},
			},
			publisher: &mockDecisionPublisher{},
			wantErr:   ErrRuleRetrievalFailed,
		},
		{
			name: "decision publish failure returns ErrDecisionPublishFailed",
			msg: &entity.FraudScoreCalculatedMessage{
				TransactionID: "tx-005",
				FraudScore:    50,
				CalculatedAt:  time.Now(),
			},
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewEvaluateFraudScoreUseCase(tc.ruleRepo, tc.publisher, &mockRuleEvaluationRepository{}, zerolog.Nop())
			result, err := uc.Execute(context.Background(), tc.msg)

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
			if result.TransactionID != tc.msg.TransactionID {
				t.Errorf("expected TransactionID %q, got %q", tc.msg.TransactionID, result.TransactionID)
			}
			if result.Status != tc.wantStatus {
				t.Errorf("expected Status %q, got %q", tc.wantStatus, result.Status)
			}

			// Verify decision publisher was called for successful cases
			if !tc.publisher.called {
				t.Error("expected decision publisher to be called, but it was not")
			}
			if tc.publisher.lastResult != nil {
				if tc.publisher.lastResult.TransactionID != tc.msg.TransactionID {
					t.Errorf("published TransactionID %q, want %q", tc.publisher.lastResult.TransactionID, tc.msg.TransactionID)
				}
				if tc.publisher.lastResult.Status != tc.wantStatus {
					t.Errorf("published Status %q, want %q", tc.publisher.lastResult.Status, tc.wantStatus)
				}
			}
		})
	}
}
