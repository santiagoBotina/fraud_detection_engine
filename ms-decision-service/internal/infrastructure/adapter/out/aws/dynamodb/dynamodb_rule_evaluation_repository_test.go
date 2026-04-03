package dynamodb

import (
	"ms-decision-service/internal/domain/entity"
	"testing"
	"time"
)

func TestToRuleEvaluationItem(t *testing.T) {
	evaluatedAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	result := entity.RuleEvaluationResult{
		TransactionID:     "txn-001",
		RuleID:            "rule-001",
		RuleName:          "Block CRYPTO",
		ConditionField:    "payment_method",
		ConditionOperator: "EQUAL",
		ConditionValue:    "CRYPTO",
		ActualFieldValue:  "CARD",
		Matched:           false,
		ResultStatus:      "DECLINED",
		EvaluatedAt:       evaluatedAt,
		Priority:          1,
	}

	item := toRuleEvaluationItem(result)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"TransactionID", item.TransactionID, "txn-001"},
		{"RuleID", item.RuleID, "rule-001"},
		{"RuleName", item.RuleName, "Block CRYPTO"},
		{"ConditionField", item.ConditionField, "payment_method"},
		{"ConditionOperator", item.ConditionOperator, "EQUAL"},
		{"ConditionValue", item.ConditionValue, "CRYPTO"},
		{"ActualFieldValue", item.ActualFieldValue, "CARD"},
		{"ResultStatus", item.ResultStatus, "DECLINED"},
		{"EvaluatedAt", item.EvaluatedAt, "2025-01-15T10:30:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}

	if item.Matched {
		t.Error("expected Matched to be false")
	}

	if item.Priority != 1 {
		t.Errorf("expected Priority 1, got %d", item.Priority)
	}
}

func TestToRuleEvaluationResult(t *testing.T) {
	item := ruleEvaluationItem{
		TransactionID:     "txn-002",
		RuleID:            "rule-002",
		RuleName:          "High Amount",
		ConditionField:    "amount_in_cents",
		ConditionOperator: "GREATER_THAN",
		ConditionValue:    "100000",
		ActualFieldValue:  "150000",
		Matched:           true,
		ResultStatus:      "FRAUD_CHECK",
		EvaluatedAt:       "2025-01-15T10:30:01Z",
		Priority:          2,
	}

	result := toRuleEvaluationResult(item)

	if result.TransactionID != "txn-002" {
		t.Errorf("TransactionID: got %q, want %q", result.TransactionID, "txn-002")
	}

	if result.RuleID != "rule-002" {
		t.Errorf("RuleID: got %q, want %q", result.RuleID, "rule-002")
	}

	if !result.Matched {
		t.Error("expected Matched to be true")
	}

	if result.Priority != 2 {
		t.Errorf("Priority: got %d, want 2", result.Priority)
	}

	expectedTime := time.Date(2025, 1, 15, 10, 30, 1, 0, time.UTC)
	if !result.EvaluatedAt.Equal(expectedTime) {
		t.Errorf("EvaluatedAt: got %v, want %v", result.EvaluatedAt, expectedTime)
	}
}

func TestToRuleEvaluationResultInvalidTime(t *testing.T) {
	item := ruleEvaluationItem{
		TransactionID: "txn-003",
		RuleID:        "rule-003",
		EvaluatedAt:   "not-a-valid-time",
	}

	result := toRuleEvaluationResult(item)

	if !result.EvaluatedAt.IsZero() {
		t.Errorf("expected zero time for invalid EvaluatedAt, got %v", result.EvaluatedAt)
	}
}

func TestRoundTripConversion(t *testing.T) {
	evaluatedAt := time.Date(2025, 6, 20, 14, 0, 0, 0, time.UTC)

	original := entity.RuleEvaluationResult{
		TransactionID:     "txn-round",
		RuleID:            "rule-round",
		RuleName:          "Test Rule",
		ConditionField:    "currency",
		ConditionOperator: "EQUAL",
		ConditionValue:    "USD",
		ActualFieldValue:  "USD",
		Matched:           true,
		ResultStatus:      "APPROVED",
		EvaluatedAt:       evaluatedAt,
		Priority:          5,
	}

	item := toRuleEvaluationItem(original)
	restored := toRuleEvaluationResult(item)

	if restored.TransactionID != original.TransactionID {
		t.Errorf("TransactionID mismatch: got %q, want %q", restored.TransactionID, original.TransactionID)
	}

	if restored.RuleID != original.RuleID {
		t.Errorf("RuleID mismatch: got %q, want %q", restored.RuleID, original.RuleID)
	}

	if restored.RuleName != original.RuleName {
		t.Errorf("RuleName mismatch: got %q, want %q", restored.RuleName, original.RuleName)
	}

	if restored.ConditionField != original.ConditionField {
		t.Errorf("ConditionField mismatch: got %q, want %q", restored.ConditionField, original.ConditionField)
	}

	if restored.ConditionOperator != original.ConditionOperator {
		t.Errorf("ConditionOperator mismatch: got %q, want %q", restored.ConditionOperator, original.ConditionOperator)
	}

	if restored.ConditionValue != original.ConditionValue {
		t.Errorf("ConditionValue mismatch: got %q, want %q", restored.ConditionValue, original.ConditionValue)
	}

	if restored.ActualFieldValue != original.ActualFieldValue {
		t.Errorf("ActualFieldValue mismatch: got %q, want %q", restored.ActualFieldValue, original.ActualFieldValue)
	}

	if restored.Matched != original.Matched {
		t.Errorf("Matched mismatch: got %v, want %v", restored.Matched, original.Matched)
	}

	if restored.ResultStatus != original.ResultStatus {
		t.Errorf("ResultStatus mismatch: got %q, want %q", restored.ResultStatus, original.ResultStatus)
	}

	if !restored.EvaluatedAt.Equal(original.EvaluatedAt) {
		t.Errorf("EvaluatedAt mismatch: got %v, want %v", restored.EvaluatedAt, original.EvaluatedAt)
	}

	if restored.Priority != original.Priority {
		t.Errorf("Priority mismatch: got %d, want %d", restored.Priority, original.Priority)
	}
}
