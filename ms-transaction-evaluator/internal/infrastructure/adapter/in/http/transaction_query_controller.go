package http

import (
	"errors"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// TransactionResponse is the API response DTO for a single transaction.
// It maps from TransactionEntity and adds a computed finalization_latency_ms field.
type TransactionResponse struct {
	ID                    string                   `json:"id"`
	AmountInCents         int64                    `json:"amount_in_cents"`
	Currency              entity.Currency          `json:"currency"`
	PaymentMethod         entity.PaymentMethod     `json:"payment_method"`
	CustomerID            string                   `json:"customer_id"`
	CustomerName          string                   `json:"customer_name"`
	CustomerEmail         string                   `json:"customer_email"`
	CustomerPhone         string                   `json:"customer_phone"`
	CustomerIPAddress     string                   `json:"customer_ip_address"`
	Status                entity.TransactionStatus `json:"status"`
	CreatedAt             time.Time                `json:"created_at"`
	UpdatedAt             time.Time                `json:"updated_at"`
	FinalizedAt           *time.Time               `json:"finalized_at,omitempty"`
	FinalizationLatencyMs *int64                   `json:"finalization_latency_ms,omitempty"`
}

// toTransactionResponse maps a TransactionEntity to a TransactionResponse,
// computing finalization_latency_ms when the transaction has been finalized.
func toTransactionResponse(e entity.TransactionEntity) TransactionResponse {
	resp := TransactionResponse{
		ID:                e.ID,
		AmountInCents:     e.AmountInCents,
		Currency:          e.Currency,
		PaymentMethod:     e.PaymentMethod,
		CustomerID:        e.CustomerID,
		CustomerName:      e.CustomerName,
		CustomerEmail:     e.CustomerEmail,
		CustomerPhone:     e.CustomerPhone,
		CustomerIPAddress: e.CustomerIPAddress,
		Status:            e.Status,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
		FinalizedAt:       e.FinalizedAt,
	}

	if e.FinalizedAt != nil {
		latencyMs := e.FinalizedAt.Sub(e.CreatedAt).Milliseconds()
		resp.FinalizationLatencyMs = &latencyMs
	}

	return resp
}

// ListTransactionsResponse represents the response for GET /transactions.
type ListTransactionsResponse struct {
	Data       []TransactionResponse `json:"data"`
	NextCursor string                `json:"next_cursor"`
}

// TransactionDetailResponse represents the response for GET /transactions/:id.
type TransactionDetailResponse struct {
	Data TransactionResponse `json:"data"`
}

// TransactionQueryController handles read-only transaction query endpoints.
type TransactionQueryController struct {
	listUseCase *usecase.ListTransactionsUseCase
	getUseCase  *usecase.GetTransactionUseCase
	logger      zerolog.Logger
}

// NewTransactionQueryController creates a new TransactionQueryController.
func NewTransactionQueryController(
	listUseCase *usecase.ListTransactionsUseCase,
	getUseCase *usecase.GetTransactionUseCase,
	logger zerolog.Logger,
) *TransactionQueryController {
	return &TransactionQueryController{
		listUseCase: listUseCase,
		getUseCase:  getUseCase,
		logger:      logger,
	}
}

const defaultLimit = 20

// ListTransactions handles GET /transactions.
func (tqc *TransactionQueryController) ListTransactions(c *echo.Context) error {
	limitStr := c.QueryParam("limit")
	cursor := c.QueryParam("cursor")

	limit := defaultLimit
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil {
			tqc.logger.Warn().Str("limit", limitStr).Msg("invalid limit parameter")
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid limit parameter",
				Details: "limit must be a positive integer between 1 and 100",
			})
		}
		limit = parsed
	}

	transactions, nextCursor, err := tqc.listUseCase.Execute(c.Request().Context(), limit, cursor)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidLimit) {
			tqc.logger.Warn().Int("limit", limit).Msg("invalid limit value")
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid limit parameter",
				Details: err.Error(),
			})
		}
		if errors.Is(err, usecase.ErrInvalidCursor) {
			tqc.logger.Warn().Str("cursor", cursor).Msg("invalid cursor value")
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid cursor parameter",
				Details: err.Error(),
			})
		}
		tqc.logger.Error().Err(err).Msg("failed to list transactions")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
	}

	tqc.logger.Info().
		Int("count", len(transactions)).
		Str("next_cursor", nextCursor).
		Msg("transactions listed")

	responses := make([]TransactionResponse, len(transactions))
	for i, txn := range transactions {
		responses[i] = toTransactionResponse(txn)
	}

	return c.JSON(http.StatusOK, ListTransactionsResponse{
		Data:       responses,
		NextCursor: nextCursor,
	})
}

// GetTransaction handles GET /transactions/:id.
func (tqc *TransactionQueryController) GetTransaction(c *echo.Context) error {
	id := c.Param("id")

	transaction, err := tqc.getUseCase.Execute(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrTransactionNotFound) {
			tqc.logger.Warn().Str("id", id).Msg("transaction not found")
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Transaction not found",
				Details: err.Error(),
			})
		}
		tqc.logger.Error().Err(err).Str("id", id).Msg("failed to get transaction")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
	}

	tqc.logger.Info().Str("id", id).Msg("transaction retrieved")

	return c.JSON(http.StatusOK, TransactionDetailResponse{
		Data: toTransactionResponse(*transaction),
	})
}

// RegisterRoutes registers the transaction query routes on the Echo instance.
func (tqc *TransactionQueryController) RegisterRoutes(e *echo.Echo) {
	e.GET("/transactions", tqc.ListTransactions)
	e.GET("/transactions/:id", tqc.GetTransaction)
}
