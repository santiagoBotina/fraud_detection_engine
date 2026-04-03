package http

import (
	"errors"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

type TransactionController struct {
	validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase
	saveUseCase     *usecase.SaveTransactionUseCase
	logger          zerolog.Logger
}

func NewTransactionController(validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase, saveUseCase *usecase.SaveTransactionUseCase, logger zerolog.Logger) *TransactionController {
	return &TransactionController{
		validateUseCase: validateUseCase,
		saveUseCase:     saveUseCase,
		logger:          logger,
	}
}

// EvaluateTransaction godoc
// @Summary Evaluate a transaction for fraud detection
// @Description Validates and evaluates a transaction request to detect potential fraud
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body entity.EvaluateTransactionRequest true "Transaction evaluation request"
// @Success 200 {object} SuccessResponse "Transaction validation successful"
// @Failure 400 {object} ErrorResponse "Invalid request or validation failed"
// @Router /evaluate [post]
func (tc *TransactionController) EvaluateTransaction(c *echo.Context) error {
	var req entity.EvaluateTransactionRequest

	tc.logger.Info().Msg("received evaluate transaction request")

	// Bind the request body to the struct
	if err := c.Bind(&req); err != nil {
		tc.logger.Error().Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
	}

	tc.logger.Info().
		Int64("amount_in_cents", req.AmountInCents).
		Str("currency", string(req.Currency)).
		Str("payment_method", string(req.PaymentMethod)).
		Str("customer_id", req.CustomerInfo.CustomerID).
		Msg("request parsed")

	// Validate the request
	if err := tc.validateUseCase.Execute(&req); err != nil {
		tc.logger.Warn().Err(err).Msg("validation failed")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: err.Error(),
		})
	}

	// Save the transaction after validation succeeds
	transaction, err := tc.saveUseCase.Execute(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, usecase.ErrEventPublishFailed) {
			tc.logger.Error().Err(err).Msg("transaction saved but Kafka publish failed")
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Transaction saved but event publish failed",
				Details: err.Error(),
			})
		}

		tc.logger.Error().Err(err).Msg("failed to save transaction")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save transaction",
			Details: err.Error(),
		})
	}

	tc.logger.Info().
		Str("transaction_id", transaction.ID).
		Str("status", string(transaction.Status)).
		Msg("transaction processed successfully")

	// Return success with the saved transaction
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "Transaction saved successfully",
		Data:    transaction,
	})
}

func (tc *TransactionController) RegisterRoutes(e *echo.Echo) {
	e.POST("/evaluate", tc.EvaluateTransaction)
}
