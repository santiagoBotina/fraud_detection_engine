package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"ms-decision-service/internal/domain/entity"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// SaramaFraudScoreRequestPublisher implements repository.FraudScoreRequestPublisher using Sarama.
type SaramaFraudScoreRequestPublisher struct {
	producer sarama.SyncProducer
	topic    string
	logger   zerolog.Logger
}

// NewSaramaFraudScoreRequestPublisher creates a new Kafka-backed fraud score request publisher.
func NewSaramaFraudScoreRequestPublisher(
	producer sarama.SyncProducer,
	topic string,
	logger zerolog.Logger,
) *SaramaFraudScoreRequestPublisher {
	return &SaramaFraudScoreRequestPublisher{producer: producer, topic: topic, logger: logger}
}

// Publish marshals the transaction message to JSON and sends it to Kafka with retry.
func (p *SaramaFraudScoreRequestPublisher) Publish(_ context.Context, transaction *entity.TransactionMessage) error {
	p.logger.Info().
		Str("transaction_id", transaction.ID).
		Str("topic", p.topic).
		Msg("publishing fraud score request to Kafka")

	payload, err := json.Marshal(transaction)
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("transaction_id", transaction.ID).
			Msg("failed to marshal transaction message")
		return fmt.Errorf("failed to marshal transaction message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(transaction.ID),
		Value: sarama.ByteEncoder(payload),
	}

	const maxRetries = 3
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		partition, offset, sendErr := p.producer.SendMessage(msg)
		if sendErr == nil {
			p.logger.Info().
				Str("transaction_id", transaction.ID).
				Str("topic", p.topic).
				Int32("partition", partition).
				Int64("offset", offset).
				Msg("fraud score request published to Kafka")
			return nil
		}
		lastErr = sendErr
		p.logger.Warn().
			Int("attempt", attempt).
			Int("max_retries", maxRetries).
			Err(sendErr).
			Str("transaction_id", transaction.ID).
			Msg("Kafka publish attempt failed")
	}

	p.logger.Error().
		Err(lastErr).
		Str("transaction_id", transaction.ID).
		Int("attempts", maxRetries).
		Msg("failed to publish fraud score request after retries")

	return fmt.Errorf("failed to publish fraud score request after %d attempts: %w", maxRetries, lastErr)
}
