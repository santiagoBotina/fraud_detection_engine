package main

import (
	"net/http"
	"os"

	_ "ms-transaction-evaluator/docs"
	"ms-transaction-evaluator/internal/domain/usecase"
	httpAdapter "ms-transaction-evaluator/internal/infrastructure/adapter/in/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Transaction Evaluator API
// @version 1.0
// @description API for evaluating transactions and detecting potential fraud
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

func main() {
	godotenv.Load()

	e := echo.New()
	e.Use(middleware.RequestLogger())

	// Initialize use case
	validateUseCase := usecase.NewValidateCreateTransactionPayloadUseCase()

	// Initialize controller
	transactionController := httpAdapter.NewTransactionController(validateUseCase)

	// Register routes
	transactionController.RegisterRoutes(e)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	port := os.Getenv("EVALUATOR_APP_PORT")

	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
