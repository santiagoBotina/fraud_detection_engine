package usecase

import (
	"context"
	"errors"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"sort"
	"testing"

	"pgregory.net/rapid"
)

func TestListRulesUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		ruleRepo  *mockRuleRepository
		wantErr   error
		wantCount int
	}{
		{
			name: "successful retrieval returns all rules",
			ruleRepo: &mockRuleRepository{
				findAllFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{
						{
							RuleID:   "rule-1",
							RuleName: "Block CRYPTO",
							Priority: 1,
							IsActive: true,
						},
						{
							RuleID:   "rule-2",
							RuleName: "Inactive rule",
							Priority: 2,
							IsActive: false,
						},
						{
							RuleID:   "rule-3",
							RuleName: "High amount",
							Priority: 3,
							IsActive: true,
						},
					}, nil
				},
			},
			wantCount: 3,
		},
		{
			name: "empty rules returns empty slice",
			ruleRepo: &mockRuleRepository{
				findAllFunc: func(_ context.Context) ([]entity.Rule, error) {
					return []entity.Rule{}, nil
				},
			},
			wantCount: 0,
		},
		{
			name: "repository error returns ErrRuleRetrievalFailed",
			ruleRepo: &mockRuleRepository{
				findAllFunc: func(_ context.Context) ([]entity.Rule, error) {
					return nil, errors.New("dynamo timeout")
				},
			},
			wantErr: ErrRuleRetrievalFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewListRulesUseCase(tc.ruleRepo)
			rules, err := uc.Execute(context.Background())

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error wrapping %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(rules) != tc.wantCount {
				t.Fatalf("expected %d rules, got %d", tc.wantCount, len(rules))
			}
		})
	}
}

// Feature: fraud-analyst-dashboard, Property 9: Rules list sorted by priority with all fields
// Validates: Requirements 6.1, 6.2
func TestProperty_RulesListSortedByPriorityWithAllFields(t *testing.T) {
	nonEmptyStr := rapid.StringMatching(`[a-zA-Z0-9]{1,20}`)

	validFields := []entity.ConditionField{
		entity.FieldAmountInCents,
		entity.FieldCurrency,
		entity.FieldPaymentMethod,
		entity.FieldCustomerID,
		entity.FieldCustomerIPAddress,
		entity.FieldFraudScore,
	}

	validOperators := []entity.ConditionOperator{
		entity.OpGreaterThan,
		entity.OpLessThan,
		entity.OpEqual,
		entity.OpNotEqual,
		entity.OpGreaterThanOrEqual,
		entity.OpLessThanOrEqual,
	}

	resultStatuses := []entity.DecisionStatus{
		entity.APPROVED,
		entity.DECLINED,
		entity.FRAUDCHECK,
	}

	rapid.Check(t, func(t *rapid.T) {
		// Generate 1–20 random rules with unique priorities
		ruleCount := rapid.IntRange(1, 20).Draw(t, "ruleCount")
		rules := make([]entity.Rule, ruleCount)

		for i := range ruleCount {
			rules[i] = entity.Rule{
				RuleID:            nonEmptyStr.Draw(t, fmt.Sprintf("ruleID_%d", i)),
				RuleName:          nonEmptyStr.Draw(t, fmt.Sprintf("ruleName_%d", i)),
				ConditionField:    rapid.SampledFrom(validFields).Draw(t, fmt.Sprintf("field_%d", i)),
				ConditionOperator: rapid.SampledFrom(validOperators).Draw(t, fmt.Sprintf("op_%d", i)),
				ConditionValue:    nonEmptyStr.Draw(t, fmt.Sprintf("condVal_%d", i)),
				ResultStatus:      rapid.SampledFrom(resultStatuses).Draw(t, fmt.Sprintf("status_%d", i)),
				Priority:          rapid.IntRange(1, 1000).Draw(t, fmt.Sprintf("priority_%d", i)),
				IsActive:          rapid.Bool().Draw(t, fmt.Sprintf("isActive_%d", i)),
			}
		}

		// Sort rules by priority ascending to simulate the repository contract
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].Priority < rules[j].Priority
		})

		ruleRepo := &mockRuleRepository{
			findAllFunc: func(_ context.Context) ([]entity.Rule, error) {
				return rules, nil
			},
		}

		uc := NewListRulesUseCase(ruleRepo)
		result, err := uc.Execute(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert: same number of rules returned
		if len(result) != len(rules) {
			t.Fatalf("expected %d rules, got %d", len(rules), len(result))
		}

		// Assert: rules are sorted by priority ascending
		for i := 1; i < len(result); i++ {
			if result[i].Priority < result[i-1].Priority {
				t.Fatalf("rules not sorted by priority: index %d has priority %d, index %d has priority %d",
					i-1, result[i-1].Priority, i, result[i].Priority)
			}
		}

		// Assert: each rule has all required fields non-empty
		for i, rule := range result {
			if rule.RuleID == "" {
				t.Errorf("rule[%d]: RuleID is empty", i)
			}
			if rule.RuleName == "" {
				t.Errorf("rule[%d]: RuleName is empty", i)
			}
			if rule.ConditionField == "" {
				t.Errorf("rule[%d]: ConditionField is empty", i)
			}
			if rule.ConditionOperator == "" {
				t.Errorf("rule[%d]: ConditionOperator is empty", i)
			}
			if rule.ConditionValue == "" {
				t.Errorf("rule[%d]: ConditionValue is empty", i)
			}
			if rule.ResultStatus == "" {
				t.Errorf("rule[%d]: ResultStatus is empty", i)
			}
			if rule.Priority == 0 {
				t.Errorf("rule[%d]: Priority is zero", i)
			}
		}
	})
}
