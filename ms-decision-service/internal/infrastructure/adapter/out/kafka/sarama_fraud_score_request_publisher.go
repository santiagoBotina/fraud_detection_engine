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

// fraudScoreRequest is the Kafka message schema expected by the fraud-score consumer.
type fraudScoreRequest struct {
	TransactionID     string `json:"transaction_id"`
	AmountInCents     int64  `json:"amount_in_cents"`
	Currency          string `json:"currency"`
	PaymentMethod     string `json:"payment_method"`
	CustomerID        string `json:"customer_id"`
	CustomerIPAddress string `json:"customer_ip_address"`
	Timestamp         string `json:"timestamp"`
}

// Publish maps the transaction to the FraudScoreRequest schema and sends it to Kafka with retry.
func (p *SaramaFraudScoreRequestPublisher) Publish(_ context.Context, transaction *entity.TransactionMessage) error {
	p.logger.Info().
		Str("transaction_id", transaction.ID).
		Str("topic", p.topic).
		Msg("publishing fraud score request to Kafka")

	request := fraudScoreRequest{
		TransactionID:     transaction.ID,
		AmountInCents:     transaction.AmountInCents,
		Currency:          transaction.Currency,
		PaymentMethod:     transaction.PaymentMethod,
		CustomerID:        transaction.CustomerID,
		CustomerIPAddress: transaction.CustomerIPAddress,
		Timestamp:         transaction.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	payload, err := json.Marshal(request)
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
