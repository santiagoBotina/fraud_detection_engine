package usecase

import (
	"context"
	"math"
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// Feature: dashboard-pagination-metrics, Property 3: Stats aggregation correctness
// Validates: Requirements 3.2, 3.3, 3.4, 3.5

// statsMockRepo is a hand-written mock implementing TransactionRepository
// that returns a pre-configured slice of transactions from FindAll.
type statsMockRepo struct {
	transactions []entity.TransactionEntity
}

func (m *statsMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *statsMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus, _ *time.Time) error {
	return nil
}

func (m *statsMockRepo) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return nil, nil
}

func (m *statsMockRepo) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func (m *statsMockRepo) FindAll(_ context.Context) ([]entity.TransactionEntity, error) {
	return m.transactions, nil
}

// genStatus draws a random TransactionStatus.
func genStatus(t *rapid.T, label string) entity.TransactionStatus {
	statuses := []entity.TransactionStatus{entity.APPROVED, entity.DECLINED, entity.PENDING}
	idx := rapid.IntRange(0, len(statuses)-1).Draw(t, label)
	return statuses[idx]
}

// genPaymentMethod draws a random PaymentMethod.
func genPaymentMethod(t *rapid.T, label string) entity.PaymentMethod {
	methods := []entity.PaymentMethod{entity.CARD, entity.BANK_TRANSFER, entity.CRYPTO}
	idx := rapid.IntRange(0, len(methods)-1).Draw(t, label)
	return methods[idx]
}

// timeZone represents a discrete time zone that is clearly inside or outside
// each time bucket, avoiding boundary races between test and use case.
// Zones are defined with a 1-hour buffer around each boundary.
type timeZone int

const (
	zoneToday    timeZone = iota // 0–23 hours ago (clearly within 24h)
	zoneThisWeek                // 2–6 days ago (within 7d but not 24h)
	zoneMonth                   // 8–29 days ago (within 30d but not 7d)
	zoneOlder                   // 31–59 days ago (outside 30d)
)

// offsetForZone returns a time.Duration offset from now for the given zone.
func offsetForZone(t *rapid.T, zone timeZone) time.Duration {
	switch zone {
	case zoneToday:
		// 1 to 22 hours ago — safely within 24h
		hours := rapid.IntRange(1, 22).Draw(t, "hoursAgo")
		return time.Duration(hours) * time.Hour
	case zoneThisWeek:
		// 2 to 6 days ago — safely within 7d but outside 24h
		days := rapid.IntRange(2, 6).Draw(t, "daysAgo")
		return time.Duration(days) * 24 * time.Hour
	case zoneMonth:
		// 8 to 29 days ago — safely within 30d but outside 7d
		days := rapid.IntRange(8, 29).Draw(t, "daysAgo")
		return time.Duration(days) * 24 * time.Hour
	default:
		// 31 to 59 days ago — safely outside 30d
		days := rapid.IntRange(31, 59).Draw(t, "daysAgo")
		return time.Duration(days) * 24 * time.Hour
	}
}

func TestProperty_StatsAggregationCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		now := time.Now()
		numTxns := rapid.IntRange(0, 60).Draw(t, "numTransactions")

		zones := []timeZone{zoneToday, zoneThisWeek, zoneMonth, zoneOlder}

		txns := make([]entity.TransactionEntity, numTxns)
		for i := range numTxns {
			// Pick a random zone and generate an offset within it
			zoneIdx := rapid.IntRange(0, len(zones)-1).Draw(t, "zone")
			zone := zones[zoneIdx]
			offset := offsetForZone(t, zone)
			createdAt := now.Add(-offset)

			status := genStatus(t, "status")
			pm := genPaymentMethod(t, "paymentMethod")

			// Decide whether this transaction is finalized
			finalized := rapid.Bool().Draw(t, "finalized")
			var finalizedAt *time.Time
			if finalized {
				// Random latency between 100ms and 10000ms
				latencyMs := rapid.IntRange(100, 10000).Draw(t, "latencyMs")
				ft := createdAt.Add(time.Duration(latencyMs) * time.Millisecond)
				finalizedAt = &ft
			}

			txns[i] = entity.TransactionEntity{
				ID:            rapid.StringMatching(`^txn_[a-z0-9]{8}$`).Draw(t, "txnID"),
				AmountInCents: rapid.Int64Range(1, 10_000_000).Draw(t, "amount"),
				Currency:      entity.USD,
				PaymentMethod: pm,
				CustomerID:    "cust_test",
				Status:        status,
				CreatedAt:     createdAt,
				UpdatedAt:     createdAt,
				FinalizedAt:   finalizedAt,
			}
		}

		// Execute the use case
		repo := &statsMockRepo{transactions: txns}
		uc := NewGetTransactionStatsUseCase(repo)
		stats, err := uc.Execute(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Reference computation — use a now that is slightly after the use case's
		// now to account for execution time. Since we use zones with 1-hour buffers,
		// a few milliseconds difference won't affect results.
		refNow := time.Now()
		last24h := refNow.Add(-24 * time.Hour)
		last7d := refNow.Add(-7 * 24 * time.Hour)
		last30d := refNow.Add(-30 * 24 * time.Hour)

		wantToday := 0
		wantWeek := 0
		wantMonth := 0
		wantApproved := 0
		wantDeclined := 0
		wantPending := 0
		wantPM := make(map[entity.PaymentMethod]int)
		wantFinalizedCount := 0
		wantLatencyLow := 0
		wantLatencyMedium := 0
		wantLatencyHigh := 0
		var wantLatencySum float64

		for i := range txns {
			txn := &txns[i]

			if !txn.CreatedAt.Before(last24h) {
				wantToday++
			}
			if !txn.CreatedAt.Before(last7d) {
				wantWeek++
			}
			if !txn.CreatedAt.Before(last30d) {
				wantMonth++
			}

			switch txn.Status {
			case entity.APPROVED:
				wantApproved++
			case entity.DECLINED:
				wantDeclined++
			case entity.PENDING:
				wantPending++
			}

			wantPM[txn.PaymentMethod]++

			if txn.FinalizedAt != nil && !txn.FinalizedAt.IsZero() {
				latencyMs := float64(txn.FinalizedAt.Sub(txn.CreatedAt).Milliseconds())
				wantLatencySum += latencyMs
				wantFinalizedCount++

				switch {
				case latencyMs <= entity.LatencyLowThresholdMs:
					wantLatencyLow++
				case latencyMs <= entity.LatencyMediumThresholdMs:
					wantLatencyMedium++
				default:
					wantLatencyHigh++
				}
			}
		}

		var wantAvgLatency float64
		if wantFinalizedCount > 0 {
			wantAvgLatency = wantLatencySum / float64(wantFinalizedCount)
		}

		// (a) Time-bucket counts match
		if stats.Today != wantToday {
			t.Fatalf("Today: got %d, want %d", stats.Today, wantToday)
		}
		if stats.ThisWeek != wantWeek {
			t.Fatalf("ThisWeek: got %d, want %d", stats.ThisWeek, wantWeek)
		}
		if stats.ThisMonth != wantMonth {
			t.Fatalf("ThisMonth: got %d, want %d", stats.ThisMonth, wantMonth)
		}

		// (b) Status counts sum to total and match distribution
		if stats.Total != numTxns {
			t.Fatalf("Total: got %d, want %d", stats.Total, numTxns)
		}
		if stats.Approved+stats.Declined+stats.Pending != stats.Total {
			t.Fatalf("status counts (%d+%d+%d=%d) do not sum to total %d",
				stats.Approved, stats.Declined, stats.Pending,
				stats.Approved+stats.Declined+stats.Pending, stats.Total)
		}
		if stats.Approved != wantApproved {
			t.Fatalf("Approved: got %d, want %d", stats.Approved, wantApproved)
		}
		if stats.Declined != wantDeclined {
			t.Fatalf("Declined: got %d, want %d", stats.Declined, wantDeclined)
		}
		if stats.Pending != wantPending {
			t.Fatalf("Pending: got %d, want %d", stats.Pending, wantPending)
		}

		// (c) Payment method counts match
		for pm, want := range wantPM {
			got := stats.PaymentMethods[pm]
			if got != want {
				t.Fatalf("PaymentMethod %s: got %d, want %d", pm, got, want)
			}
		}
		for pm, got := range stats.PaymentMethods {
			if _, exists := wantPM[pm]; !exists {
				t.Fatalf("unexpected PaymentMethod %s with count %d", pm, got)
			}
		}

		// (d) Avg latency and tier counts match thresholds
		if stats.FinalizedCount != wantFinalizedCount {
			t.Fatalf("FinalizedCount: got %d, want %d", stats.FinalizedCount, wantFinalizedCount)
		}
		if stats.LatencyLow != wantLatencyLow {
			t.Fatalf("LatencyLow: got %d, want %d", stats.LatencyLow, wantLatencyLow)
		}
		if stats.LatencyMedium != wantLatencyMedium {
			t.Fatalf("LatencyMedium: got %d, want %d", stats.LatencyMedium, wantLatencyMedium)
		}
		if stats.LatencyHigh != wantLatencyHigh {
			t.Fatalf("LatencyHigh: got %d, want %d", stats.LatencyHigh, wantLatencyHigh)
		}
		if math.Abs(stats.AvgLatencyMs-wantAvgLatency) > 0.01 {
			t.Fatalf("AvgLatencyMs: got %f, want %f", stats.AvgLatencyMs, wantAvgLatency)
		}
	})
}
