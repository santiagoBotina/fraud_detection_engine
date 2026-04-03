package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"ms-transaction-evaluator/internal/domain/entity"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

type SaramaTransactionPublisher struct {
	producer sarama.SyncProducer
	topic    string
	logger   zerolog.Logger
}

func NewSaramaTransactionPublisher(producer sarama.SyncProducer, topic string, logger zerolog.Logger) *SaramaTransactionPublisher {
	return &SaramaTransactionPublisher{producer: producer, topic: topic, logger: logger}
}

func (p *SaramaTransactionPublisher) Publish(_ context.Context, transaction *entity.TransactionEntity) error {
	p.logger.Info().
		Str("transaction_id", transaction.ID).
		Str("topic", p.topic).
		Str("status", string(transaction.Status)).
		Int64("amount_in_cents", transaction.AmountInCents).
		Msg("publishing transaction to Kafka")

	payload, err := json.Marshal(transaction)
	if err != nil {
		p.logger.Error().Err(err).Str("transaction_id", transaction.ID).Msg("failed to marshal transaction")
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(transaction.ID),
		Value: sarama.ByteEncoder(payload),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("transaction_id", transaction.ID).
			Str("topic", p.topic).
			Msg("failed to publish transaction message")
		return fmt.Errorf("failed to publish transaction message: %w", err)
	}

	p.logger.Info().
		Str("transaction_id", transaction.ID).
		Str("topic", p.topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("transaction published to Kafka")

	return nil
}
