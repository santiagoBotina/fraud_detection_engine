package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ms-decision-service/internal/domain/usecase"
	kafkaIn "ms-decision-service/internal/infrastructure/adapter/in/kafka"
	dynamodbAdapter "ms-decision-service/internal/infrastructure/adapter/out/aws/dynamodb"
	kafkaOut "ms-decision-service/internal/infrastructure/adapter/out/kafka"

	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	logger := slog.Default()

	logger.Info("starting decision service")

	// AWS SDK config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(getEnvOrDefault("AWS_REGION", "us-east-1")),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}
	logger.Info("AWS SDK config loaded", "region", getEnvOrDefault("AWS_REGION", "us-east-1"))

	// DynamoDB client
	var dynamoClient *dynamodb.Client

	endpoint := os.Getenv("DYNAMO_DB_ENDPOINT")
	if endpoint != "" {
		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &endpoint
		})
		logger.Info("DynamoDB client initialized", "endpoint", endpoint)
	} else {
		dynamoClient = dynamodb.NewFromConfig(cfg)
		logger.Info("DynamoDB client initialized", "endpoint", "default AWS")
	}

	rulesTable := getEnvOrDefault("DYNAMO_DB_RULES_TABLE", "ddb-rules")
	ruleRepo := dynamodbAdapter.NewDynamoDBRuleRepository(dynamoClient, rulesTable, logger)
	logger.Info("rules repository initialized", "table", rulesTable)

	// Kafka producer for decision results
	brokerAddress := getEnvOrDefault("KAFKA_BROKER_ADDRESS", "localhost:9092")
	decisionTopic := getEnvOrDefault("KAFKA_DECISION_RESULTS_TOPIC", "decision-results")

	logger.Info("connecting to Kafka broker", "broker", brokerAddress)
	producer, err := sarama.NewSyncProducer([]string{brokerAddress}, nil)
	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}
	defer producer.Close()
	logger.Info("Kafka producer connected", "broker", brokerAddress, "topic", decisionTopic)

	decisionPublisher := kafkaOut.NewSaramaDecisionPublisher(producer, decisionTopic, logger)

	// Use case
	evaluateUC := usecase.NewEvaluateTransactionUseCase(ruleRepo, decisionPublisher)

	// Kafka consumer
	consumerGroup := getEnvOrDefault("KAFKA_CONSUMER_GROUP", "decision-service-group")
	pendingTopic := getEnvOrDefault("KAFKA_TRANSACTION_PENDING_TOPIC", "transaction-pending")

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	logger.Info("creating consumer group", "group", consumerGroup, "topic", pendingTopic)
	group, err := sarama.NewConsumerGroup([]string{brokerAddress}, consumerGroup, saramaConfig)
	if err != nil {
		log.Fatalf("failed to create consumer group: %v", err)
	}
	defer group.Close()
	logger.Info("Kafka consumer group connected", "group", consumerGroup, "broker", brokerAddress)

	consumer := kafkaIn.NewTransactionConsumer(evaluateUC, logger)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("shutting down decision service")
		cancel()
	}()

	logger.Info("decision service started, consuming messages",
		"consumer_group", consumerGroup,
		"topic", pendingTopic,
	)

	for {
		if err := group.Consume(ctx, []string{pendingTopic}, consumer); err != nil {
			logger.Error("consumer group error", "error", err)
		}
		if ctx.Err() != nil {
			break
		}
		logger.Info("rebalancing consumer group")
	}

	logger.Info("decision service stopped")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
