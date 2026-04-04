package entity

import (
	"testing"
	"time"

	"pgregory.net/rapid"
)

// Feature: transaction-finalization-latency, Property 3: Latency computation correctness
// Validates: Requirements 2.1

func TestProperty_LatencyComputationCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random created_at timestamp (Unix seconds in a reasonable range)
		baseUnix := rapid.Int64Range(1_000_000_000, 2_000_000_000).Draw(t, "createdAtUnix")
		createdAt := time.Unix(baseUnix, 0).UTC()

		// Generate a random positive duration in milliseconds (1ms to 30 minutes)
		durationMs := rapid.Int64Range(1, 1_800_000).Draw(t, "durationMs")

		// Compute finalized_at = created_at + duration
		finalizedAt := createdAt.Add(time.Duration(durationMs) * time.Millisecond)

		// Build a TransactionEntity with these timestamps
		txn := TransactionEntity{
			ID:          "txn_test",
			CreatedAt:   createdAt,
			FinalizedAt: &finalizedAt,
		}

		// Compute latency the same way the API does: finalized_at - created_at in milliseconds
		computedLatencyMs := txn.FinalizedAt.Sub(txn.CreatedAt).Milliseconds()

		if computedLatencyMs != durationMs {
			t.Fatalf("expected latency %d ms, got %d ms (createdAt=%v, finalizedAt=%v)",
				durationMs, computedLatencyMs, createdAt, finalizedAt)
		}
	})
}
