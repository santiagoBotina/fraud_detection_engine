package http

import (
	"errors"
	"log/slog"
	"net/http"

	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"

	"github.com/labstack/echo/v5"
)

type TransactionController struct {
	validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase
	saveUseCase     *usecase.SaveTransactionUseCase
	logger          *slog.Logger
}

func NewTransactionController(validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase, saveUseCase *usecase.SaveTransactionUseCase, logger *slog.Logger) *TransactionController {
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

	tc.logger.Info("received evaluate transaction request")

	// Bind the request body to the struct
	if err := c.Bind(&req); err != nil {
		tc.logger.Error("failed to bind request body", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
	}

	tc.logger.Info("request parsed",
		"amount_in_cents", req.AmountInCents,
		"currency", req.Currency,
		"payment_method", req.PaymentMethod,
		"customer_id", req.CustomerInfo.CustomerID,
	)

	// Validate the request
	if err := tc.validateUseCase.Execute(&req); err != nil {
		tc.logger.Warn("validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: err.Error(),
		})
	}

	// Save the transaction after validation succeeds
	transaction, err := tc.saveUseCase.Execute(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, usecase.ErrEventPublishFailed) {
			tc.logger.Error("transaction saved but Kafka publish failed",
				"error", err,
			)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Transaction saved but event publish failed",
				Details: err.Error(),
			})
		}

		tc.logger.Error("failed to save transaction", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save transaction",
			Details: err.Error(),
		})
	}

	tc.logger.Info("transaction processed successfully",
		"transaction_id", transaction.ID,
		"status", transaction.Status,
	)

	// Return success with the saved transaction
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "Transaction saved successfully",
		Data:    transaction,
	})
}

func (tc *TransactionController) RegisterRoutes(e *echo.Echo) {
	e.POST("/evaluate", tc.EvaluateTransaction)
}
