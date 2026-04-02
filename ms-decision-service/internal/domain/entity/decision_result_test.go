package entity

import (
	"encoding/json"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func genDecisionResult() gopter.Gen {
	return gopter.CombineGens(
		genUUID(),
		genDecisionStatus(),
	).Map(func(values []interface{}) DecisionResult {
		return DecisionResult{
			TransactionID: values[0].(string),
			Status:        values[1].(DecisionStatus),
		}
	})
}

// Feature: kafka-transaction-decision-service, Property 3: DecisionResult JSON round-trip
// Validates: Requirements 5.2, 7.3
func TestProperty_DecisionResultJSONRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("DecisionResult JSON round-trip preserves equality", prop.ForAll(
		func(original DecisionResult) bool {
			data, err := json.Marshal(original)
			if err != nil {
				return false
			}

			var decoded DecisionResult
			if err := json.Unmarshal(data, &decoded); err != nil {
				return false
			}

			return original == decoded
		},
		genDecisionResult(),
	))

	properties.TestingRun(t)
}
