package entity

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

var conditionFields = []ConditionField{
	FieldAmountInCents,
	FieldCurrency,
	FieldPaymentMethod,
	FieldCustomerID,
	FieldCustomerIPAddress,
}

var conditionOperators = []ConditionOperator{
	OpGreaterThan,
	OpLessThan,
	OpEqual,
	OpNotEqual,
	OpGreaterThanOrEqual,
	OpLessThanOrEqual,
}

var decisionStatuses = []DecisionStatus{
	APPROVED,
	DECLINED,
	FRAUDCHECK,
}

func genConditionField() gopter.Gen {
	return func(params *gopter.GenParameters) *gopter.GenResult {
		idx := rand.Intn(len(conditionFields))
		return gopter.NewGenResult(conditionFields[idx], gopter.NoShrinker)
	}
}

func genConditionOperator() gopter.Gen {
	return func(params *gopter.GenParameters) *gopter.GenResult {
		idx := rand.Intn(len(conditionOperators))
		return gopter.NewGenResult(conditionOperators[idx], gopter.NoShrinker)
	}
}

func genDecisionStatus() gopter.Gen {
	return func(params *gopter.GenParameters) *gopter.GenResult {
		idx := rand.Intn(len(decisionStatuses))
		return gopter.NewGenResult(decisionStatuses[idx], gopter.NoShrinker)
	}
}

func genRule() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString(),
		gen.AlphaString(),
		genConditionField(),
		genConditionOperator(),
		gen.AlphaString(),
		genDecisionStatus(),
		gen.IntRange(1, 999),
		gen.Bool(),
	).Map(func(values []interface{}) Rule {
		return Rule{
			RuleID:            values[0].(string),
			RuleName:          values[1].(string),
			ConditionField:    values[2].(ConditionField),
			ConditionOperator: values[3].(ConditionOperator),
			ConditionValue:    values[4].(string),
			ResultStatus:      values[5].(DecisionStatus),
			Priority:          values[6].(int),
			IsActive:          values[7].(bool),
		}
	})
}

// Feature: kafka-transaction-decision-service, Property 2: Rule JSON round-trip
// Validates: Requirements 3.1, 7.2
func TestProperty_RuleJSONRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Rule JSON round-trip preserves equality", prop.ForAll(
		func(original Rule) bool {
			data, err := json.Marshal(original)
			if err != nil {
				return false
			}

			var decoded Rule
			if err := json.Unmarshal(data, &decoded); err != nil {
				return false
			}

			return original == decoded
		},
		genRule(),
	))

	properties.TestingRun(t)
}
