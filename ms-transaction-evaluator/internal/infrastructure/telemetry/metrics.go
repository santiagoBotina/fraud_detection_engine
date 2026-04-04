package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
)

// TransactionFinalizationDuration records the time between transaction creation
// and finalization (APPROVED or DECLINED) in seconds.
var TransactionFinalizationDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "transaction_finalization_duration_seconds",
		Help:    "Time between transaction creation and finalization",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	},
	[]string{"status"},
)

func init() {
	prometheus.MustRegister(TransactionFinalizationDuration)
}
