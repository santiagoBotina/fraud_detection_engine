package http

import (
	"net/http"

	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"

	"github.com/labstack/echo/v5"
)

type TransactionController struct {
	validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase
	saveUseCase     *usecase.SaveTransactionUseCase
}

func NewTransactionController(validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase, saveUseCase *usecase.SaveTransactionUseCase) *TransactionController {
	return &TransactionController{
		validateUseCase: validateUseCase,
		saveUseCase:     saveUseCase,
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

	// Bind the request body to the struct
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
	}

	// Validate the request
	if err := tc.validateUseCase.Execute(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: err.Error(),
		})
	}

	// Save the transaction after validation succeeds
	transaction, err := tc.saveUseCase.Execute(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save transaction",
			Details: err.Error(),
		})
	}

	// Return success with the saved transaction
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "Transaction saved successfully",
		Data:    transaction,
	})
}

func (tc *TransactionController) RegisterRoutes(e *echo.Echo) {
	e.POST("/evaluate", tc.EvaluateTransaction)
}
