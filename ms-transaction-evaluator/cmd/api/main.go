package main

import (
	"context"
	"log"
	_ "ms-transaction-evaluator/docs"
	"ms-transaction-evaluator/internal/domain/usecase"
	httpAdapter "ms-transaction-evaluator/internal/infrastructure/adapter/in/http"
	dynamodbAdapter "ms-transaction-evaluator/internal/infrastructure/adapter/out/aws/dynamodb"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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

// @host localhost:3000
// @BasePath /
// @schemes http https

func main() {
	godotenv.Load()

	// Initialize AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(getEnvOrDefault("AWS_REGION", "us-east-1")),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}

	// Initialize DynamoDB client with optional custom endpoint for local development
	var dynamoClient *dynamodb.Client
	endpoint := os.Getenv("DYNAMO_DB_ENDPOINT")
	tableName := os.Getenv("DYNAMO_DB_TRANSACTIONS_TABLE")

	if endpoint != "" {
		log.Printf("Using custom DynamoDB endpoint: %s", endpoint)
		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &endpoint
		})
	} else {
		log.Printf("Using default AWS DynamoDB endpoint")
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}

	transactionRepo := dynamodbAdapter.NewDynamoDBTransactionRepository(dynamoClient, tableName)

	// Initialize use cases
	validateUseCase := usecase.NewValidateCreateTransactionPayloadUseCase()
	saveUseCase := usecase.NewSaveTransactionUseCase(transactionRepo)

	e := echo.New()
	e.Use(middleware.RequestLogger())

	// Initialize controller
	transactionController := httpAdapter.NewTransactionController(validateUseCase, saveUseCase)

	// Register routes
	transactionController.RegisterRoutes(e)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	port := os.Getenv("EVALUATOR_APP_PORT")

	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
