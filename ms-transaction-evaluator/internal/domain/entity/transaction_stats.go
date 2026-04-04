package entity

// LatencyTier classifies transaction finalization latency into buckets.
type LatencyTier string

const (
	LatencyLow    LatencyTier = "LOW"    // ≤ 2000ms
	LatencyMedium LatencyTier = "MEDIUM" // ≤ 5000ms
	LatencyHigh   LatencyTier = "HIGH"   // > 5000ms
)

// Latency tier thresholds in milliseconds.
const (
	LatencyLowThresholdMs    = 2000
	LatencyMediumThresholdMs = 5000
)

// ClassifyLatency returns the LatencyTier for a given latency in milliseconds.
func ClassifyLatency(latencyMs float64) LatencyTier {
	switch {
	case latencyMs <= LatencyLowThresholdMs:
		return LatencyLow
	case latencyMs <= LatencyMediumThresholdMs:
		return LatencyMedium
	default:
		return LatencyHigh
	}
}

// TransactionStats holds aggregated metrics across all transactions.
type TransactionStats struct {
	Today          int                   `json:"today"`
	ThisWeek       int                   `json:"this_week"`
	ThisMonth      int                   `json:"this_month"`
	Total          int                   `json:"total"`
	Approved       int                   `json:"approved"`
	Declined       int                   `json:"declined"`
	Pending        int                   `json:"pending"`
	PaymentMethods map[PaymentMethod]int `json:"payment_methods"`
	AvgLatencyMs   float64               `json:"avg_latency_ms"`
	FinalizedCount int                   `json:"finalized_count"`
	LatencyLow     int                   `json:"latency_low"`
	LatencyMedium  int                   `json:"latency_medium"`
	LatencyHigh    int                   `json:"latency_high"`
}
