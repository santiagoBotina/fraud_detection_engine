package main

import (
	"context"
	"io"
	"ms-decision-service/internal/domain/usecase"
	"ms-decision-service/internal/infrastructure/telemetry"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "ms-decision-service/internal/infrastructure/adapter/in/http"
	kafkaIn "ms-decision-service/internal/infrastructure/adapter/in/kafka"
	dynamodbAdapter "ms-decision-service/internal/infrastructure/adapter/out/aws/dynamodb"
	kafkaOut "ms-decision-service/internal/infrastructure/adapter/out/kafka"

	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/dnwe/otelsarama"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func main() {
	godotenv.Load()

	logger := initLogger()

	logger.Info().Msg("starting decision service")

	// AWS SDK config
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
	logger.Info().Str("region", getEnvOrDefault("AWS_REGION", "us-east-1")).Msg("AWS SDK config loaded")

	// Instrument AWS SDK with OpenTelemetry
	otelaws.AppendMiddlewares(&cfg.APIOptions)

	// DynamoDB client
	var dynamoClient *dynamodb.Client

	endpoint := os.Getenv("DYNAMO_DB_ENDPOINT")
	if endpoint != "" {
		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &endpoint
		})
		logger.Info().Str("endpoint", endpoint).Msg("DynamoDB client initialized")
	} else {
		dynamoClient = dynamodb.NewFromConfig(cfg)
		logger.Info().Str("endpoint", "default AWS").Msg("DynamoDB client initialized")
	}

	rulesTable := getEnvOrDefault("DYNAMO_DB_RULES_TABLE", "ddb-rules")
	ruleRepo := dynamodbAdapter.NewDynamoDBRuleRepository(dynamoClient, rulesTable, logger)
	logger.Info().Str("table", rulesTable).Msg("rules repository initialized")

	ruleEvalsTable := getEnvOrDefault("DYNAMO_DB_RULE_EVALUATIONS_TABLE", "ddb-rule-evaluations")
	ruleEvalRepo := dynamodbAdapter.NewDynamoDBRuleEvaluationRepository(dynamoClient, ruleEvalsTable, logger)
	logger.Info().Str("table", ruleEvalsTable).Msg("rule evaluations repository initialized")

	// Kafka producer for decision results
	brokerAddress := getEnvOrDefault("KAFKA_BROKER_ADDRESS", "localhost:9092")
	decisionTopic := getEnvOrDefault("KAFKA_DECISION_CALCULATED_TOPIC", "Decision.Calculated")

	logger.Info().Str("broker", brokerAddress).Msg("connecting to Kafka broker")
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{brokerAddress}, producerConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create Kafka producer")
	}
	producer = otelsarama.WrapSyncProducer(producerConfig, producer)
	defer producer.Close()
	logger.Info().Str("broker", brokerAddress).Str("topic", decisionTopic).Msg("Kafka producer connected")

	decisionPublisher := kafkaOut.NewSaramaDecisionPublisher(producer, decisionTopic, logger)

	// Fraud score request publisher
	fraudScoreRequestTopic := getEnvOrDefault("KAFKA_FRAUD_SCORE_REQUEST_TOPIC", "FraudScore.Request")
	fraudScorePublisher := kafkaOut.NewSaramaFraudScoreRequestPublisher(producer, fraudScoreRequestTopic, logger)
	logger.Info().Str("topic", fraudScoreRequestTopic).Msg("fraud score request publisher initialized")

	// Use cases
	evaluateUC := usecase.NewEvaluateTransactionUseCase(ruleRepo, decisionPublisher, fraudScorePublisher, ruleEvalRepo, logger)
	evaluateFraudScoreUC := usecase.NewEvaluateFraudScoreUseCase(ruleRepo, decisionPublisher, ruleEvalRepo, logger)
	getRuleEvaluationsUC := usecase.NewGetRuleEvaluationsUseCase(ruleEvalRepo)
	listRulesUC := usecase.NewListRulesUseCase(ruleRepo)

	// Echo HTTP server
	e := echo.New()

	// Initialize OpenTelemetry
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4317"
	}
	shutdownTelemetry, err := telemetry.Init(context.Background(), "ms-decision-service", otelEndpoint)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize telemetry")
	}
	defer shutdownTelemetry(context.Background())

	e.Use(echootel.NewMiddleware("ms-decision-service"))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{http.MethodGet, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType},
	}))

	evaluationController := httpAdapter.NewEvaluationController(getRuleEvaluationsUC, listRulesUC, logger)
	evaluationController.RegisterRoutes(e)

	// Prometheus metrics endpoint
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	port := getEnvOrDefault("DECISION_APP_PORT", "3001")

	// Kafka consumer
	consumerGroup := getEnvOrDefault("KAFKA_CONSUMER_GROUP", "decision-service-group")
	pendingTopic := getEnvOrDefault("KAFKA_TRANSACTION_CREATED_TOPIC", "Transaction.Created")

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	logger.Info().Str("group", consumerGroup).Str("topic", pendingTopic).Msg("creating consumer group")
	group, err := sarama.NewConsumerGroup([]string{brokerAddress}, consumerGroup, saramaConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create consumer group")
	}
	defer group.Close()
	logger.Info().Str("group", consumerGroup).Str("broker", brokerAddress).Msg("Kafka consumer group connected")

	consumer := kafkaIn.NewTransactionConsumer(evaluateUC, logger)
	wrappedConsumer := otelsarama.WrapConsumerGroupHandler(consumer)

	// Fraud score consumer group
	fraudScoreCalculatedTopic := getEnvOrDefault("KAFKA_FRAUD_SCORE_CALCULATED_TOPIC", "FraudScore.Calculated")
	fraudScoreConsumerGroup := "fraud-score-consumer-group"

	logger.Info().Str("group", fraudScoreConsumerGroup).Str("topic", fraudScoreCalculatedTopic).Msg("creating fraud score consumer group")
	fsGroup, err := sarama.NewConsumerGroup([]string{brokerAddress}, fraudScoreConsumerGroup, saramaConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create fraud score consumer group")
	}
	defer fsGroup.Close()
	logger.Info().Str("group", fraudScoreConsumerGroup).Str("broker", brokerAddress).Msg("fraud score consumer group connected")

	fraudScoreConsumer := kafkaIn.NewFraudScoreConsumer(evaluateFraudScoreUC, logger)
	wrappedFraudScoreConsumer := otelsarama.WrapConsumerGroupHandler(fraudScoreConsumer)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info().Msg("shutting down decision service")
		cancel()
	}()

	logger.Info().
		Str("consumer_group", consumerGroup).
		Str("topic", pendingTopic).
		Msg("decision service started, consuming messages")

	// Start Echo HTTP server in a goroutine
	go func() {
		logger.Info().Str("port", port).Msg("starting HTTP server")
		if err := e.Start(":" + port); err != nil {
			logger.Error().Err(err).Msg("HTTP server error")
		}
	}()

	// Start fraud score consumer in a goroutine
	go func() {
		for {
			if err := fsGroup.Consume(ctx, []string{fraudScoreCalculatedTopic}, wrappedFraudScoreConsumer); err != nil {
				logger.Error().Err(err).Msg("fraud score consumer group error")
			}
			if ctx.Err() != nil {
				return
			}
			logger.Info().Msg("rebalancing fraud score consumer group")
		}
	}()

	for {
		if err := group.Consume(ctx, []string{pendingTopic}, wrappedConsumer); err != nil {
			logger.Error().Err(err).Msg("consumer group error")
		}
		if ctx.Err() != nil {
			break
		}
		logger.Info().Msg("rebalancing consumer group")
	}

	logger.Info().Msg("decision service stopped")
}

func initLogger() zerolog.Logger {
	level, err := zerolog.ParseLevel(getEnvOrDefault("LOG_LEVEL", "info"))
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout
	if os.Getenv("LOG_FORMAT") == "console" {
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	return zerolog.New(output).
		With().
		Timestamp().
		Str("service", "ms-decision-service").
		Logger()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
