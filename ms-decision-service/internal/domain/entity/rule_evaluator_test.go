package entity

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: kafka-transaction-decision-service, Property 4: Condition operator comparison correctness
// Validates: Requirements 3.4, 4.5, 4.6
func TestProperty_ConditionOperatorComparisonCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	allOperators := []ConditionOperator{
		OpGreaterThan, OpLessThan, OpEqual,
		OpNotEqual, OpGreaterThanOrEqual, OpLessThanOrEqual,
	}

	// Numeric comparison: all 6 operators against native Go comparison
	properties.Property("numeric comparison matches native Go for all operators", prop.ForAll(
		func(a, b int64, opIdx int) bool {
			op := allOperators[opIdx%len(allOperators)]
			fieldVal := fmt.Sprintf("%d", a)
			condVal := fmt.Sprintf("%d", b)
			got := op.Compare(fieldVal, condVal, FieldAmountInCents)

			var expected bool
			switch op {
			case OpGreaterThan:
				expected = a > b
			case OpLessThan:
				expected = a < b
			case OpEqual:
				expected = a == b
			case OpNotEqual:
				expected = a != b
			case OpGreaterThanOrEqual:
				expected = a >= b
			case OpLessThanOrEqual:
				expected = a <= b
			}

			return got == expected
		},
		gen.Int64(),
		gen.Int64(),
		gen.IntRange(0, 5),
	))

	// String comparison: EQUAL/NOT_EQUAL match native Go, others return false
	properties.Property("string comparison matches native Go for EQUAL/NOT_EQUAL, false for others", prop.ForAll(
		func(a, b string, opIdx int) bool {
			op := allOperators[opIdx%len(allOperators)]
			got := op.Compare(a, b, FieldCurrency)

			switch op {
			case OpEqual:
				return got == (a == b)
			case OpNotEqual:
				return got == (a != b)
			default:
				return got == false
			}
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.IntRange(0, 5),
	))

	properties.TestingRun(t)
}

// Feature: kafka-transaction-decision-service, Property 5: First-match-wins by priority
// Validates: Requirements 4.2, 4.3
func TestProperty_FirstMatchWinsByPriority(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("EvaluateRules returns ResultStatus of lowest-priority matching rule", prop.ForAll(
		func(amountInCents int64, currency string, numRules int) bool {
			tx := &TransactionMessage{
				ID:                "tx-1",
				AmountInCents:     amountInCents,
				Currency:          currency,
				PaymentMethod:     "CARD",
				CustomerID:        "cust-1",
				CustomerIPAddress: "1.2.3.4",
			}

			actualAmount := fmt.Sprintf("%d", amountInCents)

			// Generate rules with unique priorities that all match the transaction
			rules := make([]Rule, 0, numRules)
			for i := 0; i < numRules; i++ {
				status := decisionStatuses[rand.Intn(len(decisionStatuses))]
				rules = append(rules, Rule{
					RuleID:            fmt.Sprintf("rule-%d", i),
					RuleName:          fmt.Sprintf("Rule %d", i),
					ConditionField:    FieldAmountInCents,
					ConditionOperator: OpEqual,
					ConditionValue:    actualAmount,
					ResultStatus:      status,
					Priority:          (i + 1) * 10,
					IsActive:          true,
				})
			}

			// Sort by priority ascending (first-match-wins)
			sort.Slice(rules, func(i, j int) bool {
				return rules[i].Priority < rules[j].Priority
			})

			result := EvaluateRules(tx, rules)

			// The first rule in sorted order should win
			return result == rules[0].ResultStatus
		},
		gen.Int64Range(1, 999999),
		gen.AlphaString(),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// Feature: kafka-transaction-decision-service, Property 6: Default APPROVED when no rules match
// Validates: Requirements 4.4
func TestProperty_DefaultApprovedWhenNoRulesMatch(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("EvaluateRules returns APPROVED when no rules match", prop.ForAll(
		func(amountInCents int64, currency string, numRules int) bool {
			tx := &TransactionMessage{
				ID:                "tx-1",
				AmountInCents:     amountInCents,
				Currency:          currency,
				PaymentMethod:     "CARD",
				CustomerID:        "cust-1",
				CustomerIPAddress: "1.2.3.4",
			}

			// Create rules that will never match: use EQUAL with a value
			// that's guaranteed to differ from the transaction's amount
			nonMatchingAmount := fmt.Sprintf("%d", amountInCents+1)

			rules := make([]Rule, 0, numRules)
			for i := 0; i < numRules; i++ {
				rules = append(rules, Rule{
					RuleID:            fmt.Sprintf("rule-%d", i),
					RuleName:          fmt.Sprintf("Rule %d", i),
					ConditionField:    FieldAmountInCents,
					ConditionOperator: OpEqual,
					ConditionValue:    nonMatchingAmount,
					ResultStatus:      DECLINED,
					Priority:          (i + 1) * 10,
					IsActive:          true,
				})
			}

			result := EvaluateRules(tx, rules)
			return result == APPROVED
		},
		gen.Int64Range(1, 999999),
		gen.AlphaString(),
		gen.IntRange(0, 10), // 0 allows empty rules list too
	))

	properties.TestingRun(t)
}
