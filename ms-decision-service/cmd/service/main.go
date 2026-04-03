package main

import (
	"context"
	"io"
	"ms-decision-service/internal/domain/usecase"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafkaIn "ms-decision-service/internal/infrastructure/adapter/in/kafka"
	dynamodbAdapter "ms-decision-service/internal/infrastructure/adapter/out/aws/dynamodb"
	kafkaOut "ms-decision-service/internal/infrastructure/adapter/out/kafka"

	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	godotenv.Load()

	logger := initLogger()

	logger.Info().Msg("starting decision service")

	// AWS SDK config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(getEnvOrDefault("AWS_REGION", "us-east-1")),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to load AWS SDK config")
	}
	logger.Info().Str("region", getEnvOrDefault("AWS_REGION", "us-east-1")).Msg("AWS SDK config loaded")

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

	// Kafka producer for decision results
	brokerAddress := getEnvOrDefault("KAFKA_BROKER_ADDRESS", "localhost:9092")
	decisionTopic := getEnvOrDefault("KAFKA_DECISION_RESULTS_TOPIC", "decision-results")

	logger.Info().Str("broker", brokerAddress).Msg("connecting to Kafka broker")
	producer, err := sarama.NewSyncProducer([]string{brokerAddress}, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create Kafka producer")
	}
	defer producer.Close()
	logger.Info().Str("broker", brokerAddress).Str("topic", decisionTopic).Msg("Kafka producer connected")

	decisionPublisher := kafkaOut.NewSaramaDecisionPublisher(producer, decisionTopic, logger)

	// Fraud score request publisher
	fraudScoreRequestTopic := getEnvOrDefault("KAFKA_FRAUD_SCORE_REQUEST_TOPIC", "FraudScore.Request")
	fraudScorePublisher := kafkaOut.NewSaramaFraudScoreRequestPublisher(producer, fraudScoreRequestTopic, logger)
	logger.Info().Str("topic", fraudScoreRequestTopic).Msg("fraud score request publisher initialized")

	// Use cases
	evaluateUC := usecase.NewEvaluateTransactionUseCase(ruleRepo, decisionPublisher, fraudScorePublisher)
	evaluateFraudScoreUC := usecase.NewEvaluateFraudScoreUseCase(ruleRepo, decisionPublisher)

	// Kafka consumer
	consumerGroup := getEnvOrDefault("KAFKA_CONSUMER_GROUP", "decision-service-group")
	pendingTopic := getEnvOrDefault("KAFKA_TRANSACTION_PENDING_TOPIC", "transaction-pending")

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

	// Start fraud score consumer in a goroutine
	go func() {
		for {
			if err := fsGroup.Consume(ctx, []string{fraudScoreCalculatedTopic}, fraudScoreConsumer); err != nil {
				logger.Error().Err(err).Msg("fraud score consumer group error")
			}
			if ctx.Err() != nil {
				return
			}
			logger.Info().Msg("rebalancing fraud score consumer group")
		}
	}()

	for {
		if err := group.Consume(ctx, []string{pendingTopic}, consumer); err != nil {
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
