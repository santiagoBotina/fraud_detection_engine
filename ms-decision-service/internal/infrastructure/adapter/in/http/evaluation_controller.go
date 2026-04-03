package http

import (
	"errors"
	"ms-decision-service/internal/domain/usecase"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// DataResponse represents a generic JSON response wrapping data in a "data" field.
type DataResponse struct {
	Data interface{} `json:"data"`
}

// ErrorResponse represents an error API response.
type ErrorResponse struct {
	Error   string `json:"error" example:"Internal server error"`
	Details string `json:"details" example:"failed to query DynamoDB"`
}

// EvaluationController handles HTTP endpoints for rule evaluations and rules.
type EvaluationController struct {
	getRuleEvaluationsUseCase *usecase.GetRuleEvaluationsUseCase
	listRulesUseCase          *usecase.ListRulesUseCase
	logger                    zerolog.Logger
}

// NewEvaluationController creates a new EvaluationController.
func NewEvaluationController(
	getRuleEvaluationsUseCase *usecase.GetRuleEvaluationsUseCase,
	listRulesUseCase *usecase.ListRulesUseCase,
	logger zerolog.Logger,
) *EvaluationController {
	return &EvaluationController{
		getRuleEvaluationsUseCase: getRuleEvaluationsUseCase,
		listRulesUseCase:          listRulesUseCase,
		logger:                    logger,
	}
}

// GetEvaluations handles GET /evaluations/:transaction_id.
func (ec *EvaluationController) GetEvaluations(c *echo.Context) error {
	transactionID := c.Param("transaction_id")

	results, err := ec.getRuleEvaluationsUseCase.Execute(c.Request().Context(), transactionID)
	if err != nil {
		if errors.Is(err, usecase.ErrTransactionIDEmpty) {
			ec.logger.Warn().Msg("empty transaction_id parameter")
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid transaction_id parameter",
				Details: err.Error(),
			})
		}
		ec.logger.Error().Err(err).Str("transaction_id", transactionID).Msg("failed to get evaluations")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
	}

	// Ensure empty array in JSON instead of null
	data := interface{}(results)
	if results == nil {
		data = []struct{}{}
	}

	ec.logger.Info().
		Str("transaction_id", transactionID).
		Int("count", len(results)).
		Msg("evaluations retrieved")

	return c.JSON(http.StatusOK, DataResponse{Data: data})
}

// ListRules handles GET /rules.
func (ec *EvaluationController) ListRules(c *echo.Context) error {
	rules, err := ec.listRulesUseCase.Execute(c.Request().Context())
	if err != nil {
		ec.logger.Error().Err(err).Msg("failed to list rules")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
	}

	// Ensure empty array in JSON instead of null
	data := interface{}(rules)
	if rules == nil {
		data = []struct{}{}
	}

	ec.logger.Info().Int("count", len(rules)).Msg("rules listed")

	return c.JSON(http.StatusOK, DataResponse{Data: data})
}

// RegisterRoutes registers the evaluation and rules routes on the Echo instance.
func (ec *EvaluationController) RegisterRoutes(e *echo.Echo) {
	e.GET("/evaluations/:transaction_id", ec.GetEvaluations)
	e.GET("/rules", ec.ListRules)
}
