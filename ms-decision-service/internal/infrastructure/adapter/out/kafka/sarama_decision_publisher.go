package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"ms-decision-service/internal/domain/entity"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// SaramaDecisionPublisher implements repository.DecisionPublisher using Sarama.
type SaramaDecisionPublisher struct {
	producer sarama.SyncProducer
	topic    string
	logger   zerolog.Logger
}

// NewSaramaDecisionPublisher creates a new Kafka-backed decision publisher.
func NewSaramaDecisionPublisher(
	producer sarama.SyncProducer,
	topic string,
	logger zerolog.Logger,
) *SaramaDecisionPublisher {
	return &SaramaDecisionPublisher{producer: producer, topic: topic, logger: logger}
}

// Publish marshals the decision result to JSON and sends it to Kafka with retry.
func (p *SaramaDecisionPublisher) Publish(_ context.Context, result *entity.DecisionResult) error {
	p.logger.Info().
		Str("transaction_id", result.TransactionID).
		Str("status", string(result.Status)).
		Str("topic", p.topic).
		Msg("publishing decision result to Kafka")

	payload, err := json.Marshal(result)
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("transaction_id", result.TransactionID).
			Msg("failed to marshal decision result")
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
			p.logger.Info().
				Str("transaction_id", result.TransactionID).
				Str("status", string(result.Status)).
				Str("topic", p.topic).
				Int32("partition", partition).
				Int64("offset", offset).
				Msg("decision result published to Kafka")
			return nil
		}
		lastErr = sendErr
		p.logger.Warn().
			Int("attempt", attempt).
			Int("max_retries", maxRetries).
			Err(sendErr).
			Str("transaction_id", result.TransactionID).
			Msg("Kafka publish attempt failed")
	}

	p.logger.Error().
		Err(lastErr).
		Str("transaction_id", result.TransactionID).
		Int("attempts", maxRetries).
		Msg("failed to publish decision result after retries")

	return fmt.Errorf("failed to publish decision result after %d attempts: %w", maxRetries, lastErr)
}
