package entity

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func genUUID() gopter.Gen {
	return func(params *gopter.GenParameters) *gopter.GenResult {
		b := make([]byte, 16)
		for i := range b {
			b[i] = byte(rand.Intn(256))
		}
		uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
		return gopter.NewGenResult(uuid, gopter.NoShrinker)
	}
}

func genTimeUTC() gopter.Gen {
	return gen.Int64Range(0, 2000000000).Map(func(secs int64) time.Time {
		return time.Unix(secs, 0).UTC()
	})
}

func genNonEmptyAlpha() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	})
}

func genTransactionMessage() gopter.Gen {
	return gopter.CombineGens(
		genUUID(),
		gen.Int64Range(1, 999999999),
		genNonEmptyAlpha(),
		genNonEmptyAlpha(),
		genNonEmptyAlpha(),
		genNonEmptyAlpha(),
		genNonEmptyAlpha(),
		genNonEmptyAlpha(),
		genNonEmptyAlpha(),
		genNonEmptyAlpha(),
		genTimeUTC(),
		genTimeUTC(),
	).Map(func(values []interface{}) TransactionMessage {
		return TransactionMessage{
			ID:                values[0].(string),
			AmountInCents:     values[1].(int64),
			Currency:          values[2].(string),
			PaymentMethod:     values[3].(string),
			CustomerID:        values[4].(string),
			CustomerName:      values[5].(string),
			CustomerEmail:     values[6].(string),
			CustomerPhone:     values[7].(string),
			CustomerIPAddress: values[8].(string),
			Status:            values[9].(string),
			CreatedAt:         values[10].(time.Time),
			UpdatedAt:         values[11].(time.Time),
		}
	})
}

// Feature: kafka-transaction-decision-service, Property 1: TransactionMessage JSON round-trip
// Validates: Requirements 1.2, 2.1, 7.1
func TestProperty_TransactionMessageJSONRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("TransactionMessage JSON round-trip preserves equality", prop.ForAll(
		func(original TransactionMessage) bool {
			data, err := json.Marshal(original)
			if err != nil {
				return false
			}

			var decoded TransactionMessage
			if err := json.Unmarshal(data, &decoded); err != nil {
				return false
			}

			return original.ID == decoded.ID &&
				original.AmountInCents == decoded.AmountInCents &&
				original.Currency == decoded.Currency &&
				original.PaymentMethod == decoded.PaymentMethod &&
				original.CustomerID == decoded.CustomerID &&
				original.CustomerName == decoded.CustomerName &&
				original.CustomerEmail == decoded.CustomerEmail &&
				original.CustomerPhone == decoded.CustomerPhone &&
				original.CustomerIPAddress == decoded.CustomerIPAddress &&
				original.Status == decoded.Status &&
				original.CreatedAt.Equal(decoded.CreatedAt) &&
				original.UpdatedAt.Equal(decoded.UpdatedAt)
		},
		genTransactionMessage(),
	))

	properties.TestingRun(t)
}

// Feature: kafka-transaction-decision-service, Property 9: GetFieldValue covers all valid condition fields
// Validates: Requirements 3.5
func TestProperty_GetFieldValueCoversAllValidConditionFields(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("GetFieldValue returns non-empty string for all valid condition fields", prop.ForAll(
		func(tx TransactionMessage, fieldIdx int) bool {
			fields := []ConditionField{
				FieldAmountInCents,
				FieldCurrency,
				FieldPaymentMethod,
				FieldCustomerID,
				FieldCustomerIPAddress,
			}
			field := fields[fieldIdx%len(fields)]
			result := tx.GetFieldValue(field)
			return len(result) > 0
		},
		genTransactionMessage(),
		gen.IntRange(0, 4),
	))

	properties.TestingRun(t)
}
