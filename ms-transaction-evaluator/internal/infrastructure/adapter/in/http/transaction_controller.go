package http

import (
	"net/http"

	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"

	"github.com/labstack/echo/v5"
)

type TransactionController struct {
	validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase
}

func NewTransactionController(validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase) *TransactionController {
	return &TransactionController{
		validateUseCase: validateUseCase,
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
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate the request
	if err := tc.validateUseCase.Execute(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// If validation passes, return success (for now)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Transaction validation successful",
		"data":    req,
	})
}

func (tc *TransactionController) RegisterRoutes(e *echo.Echo) {
	e.POST("/evaluate", tc.EvaluateTransaction)
}
