package http

import (
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// TransactionStatsResponse is the API response DTO for transaction statistics.
// PaymentMethods uses map[string]int (not map[PaymentMethod]int) since this is the HTTP layer.
type TransactionStatsResponse struct {
	Today          int            `json:"today"`
	ThisWeek       int            `json:"this_week"`
	ThisMonth      int            `json:"this_month"`
	Total          int            `json:"total"`
	Approved       int            `json:"approved"`
	Declined       int            `json:"declined"`
	Pending        int            `json:"pending"`
	PaymentMethods map[string]int `json:"payment_methods"`
	AvgLatencyMs   float64        `json:"avg_latency_ms"`
	FinalizedCount int            `json:"finalized_count"`
	LatencyLow     int            `json:"latency_low"`
	LatencyMedium  int            `json:"latency_medium"`
	LatencyHigh    int            `json:"latency_high"`
}

// toTransactionStatsResponse maps a domain TransactionStats entity to the HTTP response DTO.
func toTransactionStatsResponse(stats *entity.TransactionStats) TransactionStatsResponse {
	paymentMethods := make(map[string]int, len(stats.PaymentMethods))
	for method, count := range stats.PaymentMethods {
		paymentMethods[string(method)] = count
	}

	return TransactionStatsResponse{
		Today:          stats.Today,
		ThisWeek:       stats.ThisWeek,
		ThisMonth:      stats.ThisMonth,
		Total:          stats.Total,
		Approved:       stats.Approved,
		Declined:       stats.Declined,
		Pending:        stats.Pending,
		PaymentMethods: paymentMethods,
		AvgLatencyMs:   stats.AvgLatencyMs,
		FinalizedCount: stats.FinalizedCount,
		LatencyLow:     stats.LatencyLow,
		LatencyMedium:  stats.LatencyMedium,
		LatencyHigh:    stats.LatencyHigh,
	}
}

// TransactionStatsController handles the transaction stats endpoint.
type TransactionStatsController struct {
	statsUseCase *usecase.GetTransactionStatsUseCase
	logger       zerolog.Logger
}

// NewTransactionStatsController creates a new TransactionStatsController.
func NewTransactionStatsController(
	statsUseCase *usecase.GetTransactionStatsUseCase,
	logger zerolog.Logger,
) *TransactionStatsController {
	return &TransactionStatsController{
		statsUseCase: statsUseCase,
		logger:       logger,
	}
}

// GetStats handles GET /transactions/stats.
func (tsc *TransactionStatsController) GetStats(c *echo.Context) error {
	stats, err := tsc.statsUseCase.Execute(c.Request().Context())
	if err != nil {
		tsc.logger.Error().Err(err).Msg("failed to get transaction stats")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
	}

	tsc.logger.Info().Int("total", stats.Total).Msg("transaction stats retrieved")

	return c.JSON(http.StatusOK, toTransactionStatsResponse(stats))
}

// RegisterRoutes registers the transaction stats routes on the Echo instance.
func (tsc *TransactionStatsController) RegisterRoutes(e *echo.Echo) {
	e.GET("/transactions/stats", tsc.GetStats)
}
