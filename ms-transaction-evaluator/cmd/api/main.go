package main

import (
	"context"
	"io"
	_ "ms-transaction-evaluator/docs"
	"ms-transaction-evaluator/internal/domain/usecase"
	httpAdapter "ms-transaction-evaluator/internal/infrastructure/adapter/in/http"
	kafkaIn "ms-transaction-evaluator/internal/infrastructure/adapter/in/kafka"
	dynamodbAdapter "ms-transaction-evaluator/internal/infrastructure/adapter/out/aws/dynamodb"
	kafkaAdapter "ms-transaction-evaluator/internal/infrastructure/adapter/out/kafka"
	"ms-transaction-evaluator/internal/infrastructure/telemetry"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/dnwe/otelsarama"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
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

	// Parse log level from environment, default to info
	level, err := zerolog.ParseLevel(getEnvOrDefault("LOG_LEVEL", "info"))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout
	if os.Getenv("LOG_FORMAT") == "console" {
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	logger := zerolog.New(output).
		With().
		Timestamp().
		Str("service", "ms-transaction-evaluator").
		Logger()

	logger.Info().Msg("starting transaction evaluator")

	// Initialize AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(getEnvOrDefault("AWS_REGION", "us-east-1")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			getEnvOrDefault("AWS_ACCESS_KEY_ID", "dummy"),
			getEnvOrDefault("AWS_SECRET_ACCESS_KEY", "dummy"),
			"",
		)),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to load AWS SDK config")
	}

	// Instrument AWS SDK with OpenTelemetry
	otelaws.AppendMiddlewares(&cfg.APIOptions)

	// Initialize DynamoDB client with optional custom endpoint for local development
	var dynamoClient *dynamodb.Client
	endpoint := os.Getenv("DYNAMO_DB_ENDPOINT")
	tableName := os.Getenv("DYNAMO_DB_TRANSACTIONS_TABLE")

	if endpoint != "" {
		logger.Info().Str("endpoint", endpoint).Msg("using custom DynamoDB endpoint")
		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &endpoint
		})
	} else {
		logger.Info().Msg("using default AWS DynamoDB endpoint")
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}

	transactionRepo := dynamodbAdapter.NewDynamoDBTransactionRepository(dynamoClient, tableName, logger)
	logger.Info().Str("table", tableName).Msg("DynamoDB repository initialized")

	// Initialize Kafka producer
	brokerAddress := getEnvOrDefault("KAFKA_BROKER_ADDRESS", "localhost:9092")
	transactionTopic := getEnvOrDefault("KAFKA_TRANSACTION_CREATED_TOPIC", "Transaction.Created")

	logger.Info().Str("broker", brokerAddress).Msg("connecting to Kafka broker")
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{brokerAddress}, producerConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create Kafka producer")
	}
	producer = otelsarama.WrapSyncProducer(producerConfig, producer)
	defer producer.Close()
	logger.Info().Str("broker", brokerAddress).Str("topic", transactionTopic).Msg("Kafka producer connected")

	eventPublisher := kafkaAdapter.NewSaramaTransactionPublisher(producer, transactionTopic, logger)

	// Initialize use cases
	validateUseCase := usecase.NewValidateCreateTransactionPayloadUseCase()
	saveUseCase := usecase.NewSaveTransactionUseCase(transactionRepo, eventPublisher)
	updateStatusUseCase := usecase.NewUpdateTransactionStatusUseCase(transactionRepo)
	listTransactionsUseCase := usecase.NewListTransactionsUseCase(transactionRepo)
	getTransactionUseCase := usecase.NewGetTransactionUseCase(transactionRepo)
	getTransactionStatsUseCase := usecase.NewGetTransactionStatsUseCase(transactionRepo)

	e := echo.New()

	// Initialize OpenTelemetry
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4317"
	}
	shutdownTelemetry, err := telemetry.Init(context.Background(), "ms-transaction-evaluator", otelEndpoint)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize telemetry")
	}
	defer shutdownTelemetry(context.Background())

	e.Use(echootel.NewMiddleware("ms-transaction-evaluator"))
	e.Use(middleware.RequestLogger())

	// CORS middleware — allow dashboard origin
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{http.MethodGet, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType},
	}))

	// Initialize controllers
	transactionController := httpAdapter.NewTransactionController(validateUseCase, saveUseCase, logger)
	transactionStatsController := httpAdapter.NewTransactionStatsController(getTransactionStatsUseCase, logger)
	transactionQueryController := httpAdapter.NewTransactionQueryController(listTransactionsUseCase, getTransactionUseCase, logger)

	// Register routes — stats BEFORE query so /transactions/stats doesn't match /transactions/:id
	transactionController.RegisterRoutes(e)
	transactionStatsController.RegisterRoutes(e)
	transactionQueryController.RegisterRoutes(e)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Prometheus metrics endpoint
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Decision.Calculated consumer
	decisionTopic := getEnvOrDefault("KAFKA_DECISION_CALCULATED_TOPIC", "Decision.Calculated")
	decisionConsumerGroup := "transaction-evaluator-decision-group"

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup([]string{brokerAddress}, decisionConsumerGroup, saramaConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create decision consumer group")
	}
	defer group.Close()
	logger.Info().
		Str("group", decisionConsumerGroup).
		Str("topic", decisionTopic).
		Msg("decision consumer group connected")

	decisionConsumer := kafkaIn.NewDecisionConsumer(updateStatusUseCase, logger, getEnvAsInt("DECISION_MIN_DELAY_MS", 0), getEnvAsInt("DECISION_MAX_DELAY_MS", 0))
	wrappedConsumer := otelsarama.WrapConsumerGroupHandler(decisionConsumer)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info().Msg("shutting down transaction evaluator")
		cancel()
	}()

	// Start decision consumer in background
	go func() {
		for {
			if err := group.Consume(ctx, []string{decisionTopic}, wrappedConsumer); err != nil {
				logger.Error().Err(err).Msg("decision consumer group error")
			}
			if ctx.Err() != nil {
				return
			}
			logger.Info().Msg("rebalancing decision consumer group")
		}
	}()

	port := os.Getenv("EVALUATOR_APP_PORT")

	if err := e.Start(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("failed to start server")
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return defaultValue
}
