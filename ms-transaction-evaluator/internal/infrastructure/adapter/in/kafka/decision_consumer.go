package kafka

import (
	"context"
	"encoding/json"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// DecisionConsumer implements sarama.ConsumerGroupHandler for the Decision.Calculated topic.
type DecisionConsumer struct {
	useCase *usecase.UpdateTransactionStatusUseCase
	logger  zerolog.Logger
}

// NewDecisionConsumer creates a new consumer for decision results.
func NewDecisionConsumer(uc *usecase.UpdateTransactionStatusUseCase, logger zerolog.Logger) *DecisionConsumer {
	return &DecisionConsumer{useCase: uc, logger: logger}
}

func (c *DecisionConsumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (c *DecisionConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim processes messages from the Decision.Calculated topic.
func (c *DecisionConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.logger.Info().
			Str("topic", msg.Topic).
			Int32("partition", msg.Partition).
			Int64("offset", msg.Offset).
			Str("key", string(msg.Key)).
			Msg("decision message received")

		var decision entity.DecisionCalculatedMessage
		if err := json.Unmarshal(msg.Value, &decision); err != nil {
			c.logger.Error().Err(err).
				Str("raw", string(msg.Value)).
				Msg("failed to unmarshal decision message")
			session.MarkMessage(msg, "")
			continue
		}

		c.logger.Info().
			Str("transaction_id", decision.TransactionID).
			Str("status", decision.Status).
			Msg("updating transaction status")

		if err := c.useCase.Execute(context.Background(), &decision); err != nil {
			c.logger.Error().Err(err).
				Str("transaction_id", decision.TransactionID).
				Msg("failed to update transaction status")
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
