package kafka

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"ms-transaction-evaluator/internal/domain/entity"
	"ms-transaction-evaluator/internal/domain/usecase"
	"time"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// DecisionConsumer implements sarama.ConsumerGroupHandler for the Decision.Calculated topic.
type DecisionConsumer struct {
	useCase       *usecase.UpdateTransactionStatusUseCase
	logger        zerolog.Logger
	maxDelayMs    int
	minDelayMs    int
}

// NewDecisionConsumer creates a new consumer for decision results.
// minDelayMs and maxDelayMs control an artificial processing delay (0 = disabled).
func NewDecisionConsumer(uc *usecase.UpdateTransactionStatusUseCase, logger zerolog.Logger, minDelayMs, maxDelayMs int) *DecisionConsumer {
	return &DecisionConsumer{useCase: uc, logger: logger, minDelayMs: minDelayMs, maxDelayMs: maxDelayMs}
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

		// Simulate processing delay when configured (for local/load testing)
		if c.maxDelayMs > 0 {
			delayMs := c.minDelayMs
			if c.maxDelayMs > c.minDelayMs {
				delayMs += rand.IntN(c.maxDelayMs - c.minDelayMs)
			}
			c.logger.Debug().Int("delay_ms", delayMs).Msg("simulating processing delay")
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		if err := c.useCase.Execute(context.Background(), &decision); err != nil {
			c.logger.Error().Err(err).
				Str("transaction_id", decision.TransactionID).
				Msg("failed to update transaction status")
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
