package dynamodb

import (
	"testing"

	"ms-decision-service/internal/domain/entity"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func genRule() gopter.Gen {
	conditionFields := []entity.ConditionField{
		entity.FieldAmountInCents,
		entity.FieldCurrency,
		entity.FieldPaymentMethod,
		entity.FieldCustomerID,
		entity.FieldCustomerIPAddress,
	}
	conditionOperators := []entity.ConditionOperator{
		entity.OpGreaterThan,
		entity.OpLessThan,
		entity.OpEqual,
		entity.OpNotEqual,
		entity.OpGreaterThanOrEqual,
		entity.OpLessThanOrEqual,
	}
	decisionStatuses := []entity.DecisionStatus{
		entity.APPROVED,
		entity.DECLINED,
		entity.FRAUDCHECK,
	}

	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString(),
		gen.IntRange(0, len(conditionFields)-1),
		gen.IntRange(0, len(conditionOperators)-1),
		gen.AlphaString(),
		gen.IntRange(0, len(decisionStatuses)-1),
		gen.IntRange(1, 999),
		gen.Bool(),
	).Map(func(v []interface{}) entity.Rule {
		return entity.Rule{
			RuleID:            v[0].(string),
			RuleName:          v[1].(string),
			ConditionField:    conditionFields[v[2].(int)],
			ConditionOperator: conditionOperators[v[3].(int)],
			ConditionValue:    v[4].(string),
			ResultStatus:      decisionStatuses[v[5].(int)],
			Priority:          v[6].(int),
			IsActive:          v[7].(bool),
		}
	})
}

func genRuleSlice() gopter.Gen {
	return gen.SliceOf(genRule())
}

// Feature: kafka-transaction-decision-service, Property 8: Active rules filtering and sort invariant
// **Validates: Requirements 3.2, 4.1**
func TestProperty8_ActiveRulesFilteringAndSortInvariant(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("only active rules returned and sorted by priority ascending", prop.ForAll(
		func(rules []entity.Rule) bool {
			result := FilterAndSortActiveRules(rules)

			// All returned rules must be active
			for _, r := range result {
				if !r.IsActive {
					return false
				}
			}

			// Count active rules in input matches result length
			activeCount := 0
			for _, r := range rules {
				if r.IsActive {
					activeCount++
				}
			}
			if len(result) != activeCount {
				return false
			}

			// Result must be sorted by priority ascending
			for i := 1; i < len(result); i++ {
				if result[i].Priority < result[i-1].Priority {
					return false
				}
			}

			return true
		},
		genRuleSlice(),
	))

	properties.TestingRun(t)
}
