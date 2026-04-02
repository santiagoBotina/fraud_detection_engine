package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"ms-decision-service/internal/domain/entity"

	"github.com/IBM/sarama"
)

// SaramaDecisionPublisher implements repository.DecisionPublisher using Sarama.
type SaramaDecisionPublisher struct {
	producer sarama.SyncProducer
	topic    string
	logger   *slog.Logger
}

// NewSaramaDecisionPublisher creates a new Kafka-backed decision publisher.
func NewSaramaDecisionPublisher(producer sarama.SyncProducer, topic string, logger *slog.Logger) *SaramaDecisionPublisher {
	return &SaramaDecisionPublisher{producer: producer, topic: topic, logger: logger}
}

// Publish marshals the decision result to JSON and sends it to Kafka with retry.
func (p *SaramaDecisionPublisher) Publish(_ context.Context, result *entity.DecisionResult) error {
	p.logger.Info("publishing decision result to Kafka",
		"transaction_id", result.TransactionID,
		"status", result.Status,
		"topic", p.topic,
	)

	payload, err := json.Marshal(result)
	if err != nil {
		p.logger.Error("failed to marshal decision result",
			"error", err,
			"transaction_id", result.TransactionID,
		)
		return fmt.Errorf("failed to marshal decision result: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(result.TransactionID),
		Value: sarama.ByteEncoder(payload),
	}

	const maxRetries = 3
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		partition, offset, sendErr := p.producer.SendMessage(msg)
		if sendErr == nil {
			p.logger.Info("decision result published to Kafka",
				"transaction_id", result.TransactionID,
				"status", result.Status,
				"topic", p.topic,
				"partition", partition,
				"offset", offset,
			)
			return nil
		}
		lastErr = sendErr
		p.logger.Warn("Kafka publish attempt failed",
			"attempt", attempt,
			"max_retries", maxRetries,
			"error", sendErr,
			"transaction_id", result.TransactionID,
		)
	}

	p.logger.Error("failed to publish decision result after retries",
		"error", lastErr,
		"transaction_id", result.TransactionID,
		"attempts", maxRetries,
	)

	return fmt.Errorf("failed to publish decision result after %d attempts: %w", maxRetries, lastErr)
}
