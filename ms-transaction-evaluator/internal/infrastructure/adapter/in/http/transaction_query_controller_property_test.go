package http

import (
	"encoding/json"
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// Feature: transaction-finalization-latency, Property 4: API response includes latency if and only if finalized
// Validates: Requirements 2.2, 2.3, 2.4

func TestProperty_APIResponseIncludesLatencyIffFinalized(t *testing.T) {
	currencies := []entity.Currency{entity.USD, entity.COP, entity.EUR}
	paymentMethods := []entity.PaymentMethod{entity.CARD, entity.BANK_TRANSFER, entity.CRYPTO}
	statuses := []entity.TransactionStatus{entity.PENDING, entity.APPROVED, entity.DECLINED}

	rapid.Check(t, func(t *rapid.T) {
		// Generate a random transaction entity
		id := rapid.StringMatching(`^txn_[a-z0-9]{8,16}$`).Draw(t, "id")
		amountInCents := rapid.Int64Range(1, 100_000_000).Draw(t, "amountInCents")
		currency := currencies[rapid.IntRange(0, len(currencies)-1).Draw(t, "currencyIdx")]
		paymentMethod := paymentMethods[rapid.IntRange(0, len(paymentMethods)-1).Draw(t, "paymentMethodIdx")]
		customerID := rapid.StringMatching(`^cust_[a-z0-9]{4,12}$`).Draw(t, "customerID")
		customerName := rapid.StringMatching(`^[A-Z][a-z]{2,10} [A-Z][a-z]{2,10}$`).Draw(t, "customerName")
		customerEmail := rapid.StringMatching(`^[a-z]{3,8}@[a-z]{3,8}\.com$`).Draw(t, "customerEmail")
		customerPhone := rapid.StringMatching(`^\+[0-9]{7,15}$`).Draw(t, "customerPhone")
		customerIP := rapid.StringMatching(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`).Draw(t, "customerIP")
		status := statuses[rapid.IntRange(0, len(statuses)-1).Draw(t, "statusIdx")]

		baseUnix := rapid.Int64Range(1_000_000_000, 2_000_000_000).Draw(t, "createdAtUnix")
		offsetSecs := rapid.Int64Range(1, 3600).Draw(t, "updatedAtOffset")
		createdAt := time.Unix(baseUnix, 0).UTC()
		updatedAt := time.Unix(baseUnix+offsetSecs, 0).UTC()

		// Decide whether to set FinalizedAt based on a random boolean
		isFinalized := rapid.Bool().Draw(t, "isFinalized")

		var finalizedAt *time.Time
		if isFinalized {
			finalizedOffset := rapid.Int64Range(1, 7200).Draw(t, "finalizedAtOffset")
			ft := time.Unix(baseUnix+finalizedOffset, 0).UTC()
			finalizedAt = &ft
		}

		txn := entity.TransactionEntity{
			ID:                id,
			AmountInCents:     amountInCents,
			Currency:          currency,
			PaymentMethod:     paymentMethod,
			CustomerID:        customerID,
			CustomerName:      customerName,
			CustomerEmail:     customerEmail,
			CustomerPhone:     customerPhone,
			CustomerIPAddress: customerIP,
			Status:            status,
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
			FinalizedAt:       finalizedAt,
		}

		// Convert to response DTO
		resp := toTransactionResponse(txn)

		// Marshal to JSON
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
		}

		// Unmarshal into a generic map to check field presence
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("failed to unmarshal response to map: %v", err)
		}

		_, hasFinalizedAt := raw["finalized_at"]
		_, hasLatency := raw["finalization_latency_ms"]

		if isFinalized {
			// When FinalizedAt is set, both fields must be present
			if !hasFinalizedAt {
				t.Fatal("expected finalized_at in JSON response for finalized transaction")
			}
			if !hasLatency {
				t.Fatal("expected finalization_latency_ms in JSON response for finalized transaction")
			}

			// Verify latency value is positive
			latencyVal, ok := raw["finalization_latency_ms"].(float64)
			if !ok {
				t.Fatal("finalization_latency_ms is not a number")
			}
			if latencyVal <= 0 {
				t.Fatalf("expected positive finalization_latency_ms, got %f", latencyVal)
			}
		} else {
			// When FinalizedAt is nil, neither field should be present
			if hasFinalizedAt {
				t.Fatal("expected finalized_at to be omitted for non-finalized transaction")
			}
			if hasLatency {
				t.Fatal("expected finalization_latency_ms to be omitted for non-finalized transaction")
			}
		}
	})
}
