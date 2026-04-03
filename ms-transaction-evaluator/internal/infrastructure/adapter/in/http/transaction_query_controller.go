package http

import (
	"errors"
	"ms-transaction-evaluator/internal/domain/usecase"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// ListTransactionsResponse represents the response for GET /transactions.
type ListTransactionsResponse struct {
	Data       interface{} `json:"data"`
	NextCursor string      `json:"next_cursor"`
}

// TransactionDetailResponse represents the response for GET /transactions/:id.
type TransactionDetailResponse struct {
	Data interface{} `json:"data"`
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

	return c.JSON(http.StatusOK, ListTransactionsResponse{
		Data:       transactions,
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
		Data: transaction,
	})
}

// RegisterRoutes registers the transaction query routes on the Echo instance.
func (tqc *TransactionQueryController) RegisterRoutes(e *echo.Echo) {
	e.GET("/transactions", tqc.ListTransactions)
	e.GET("/transactions/:id", tqc.GetTransaction)
}
