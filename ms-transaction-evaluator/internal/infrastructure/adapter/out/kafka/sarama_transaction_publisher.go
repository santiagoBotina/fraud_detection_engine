package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"ms-transaction-evaluator/internal/domain/entity"

	"github.com/IBM/sarama"
)

type SaramaTransactionPublisher struct {
	producer sarama.SyncProducer
	topic    string
	logger   *slog.Logger
}

func NewSaramaTransactionPublisher(producer sarama.SyncProducer, topic string, logger *slog.Logger) *SaramaTransactionPublisher {
	return &SaramaTransactionPublisher{producer: producer, topic: topic, logger: logger}
}

func (p *SaramaTransactionPublisher) Publish(_ context.Context, transaction *entity.TransactionEntity) error {
	p.logger.Info("publishing transaction to Kafka",
		"transaction_id", transaction.ID,
		"topic", p.topic,
		"status", transaction.Status,
		"amount_in_cents", transaction.AmountInCents,
	)

	payload, err := json.Marshal(transaction)
	if err != nil {
		p.logger.Error("failed to marshal transaction", "error", err, "transaction_id", transaction.ID)
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(transaction.ID),
		Value: sarama.ByteEncoder(payload),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("failed to publish transaction message",
			"error", err,
			"transaction_id", transaction.ID,
			"topic", p.topic,
		)
		return fmt.Errorf("failed to publish transaction message: %w", err)
	}

	p.logger.Info("transaction published to Kafka",
		"transaction_id", transaction.ID,
		"topic", p.topic,
		"partition", partition,
		"offset", offset,
	)

	return nil
}
