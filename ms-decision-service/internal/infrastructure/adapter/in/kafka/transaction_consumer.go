package kafka

import (
	"context"
	"encoding/json"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/usecase"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// TransactionConsumer implements sarama.ConsumerGroupHandler for processing transaction messages.
type TransactionConsumer struct {
	evaluateUseCase *usecase.EvaluateTransactionUseCase
	logger          zerolog.Logger
}

// NewTransactionConsumer creates a new consumer with the given use case and logger.
func NewTransactionConsumer(
	evaluateUseCase *usecase.EvaluateTransactionUseCase,
	logger zerolog.Logger,
) *TransactionConsumer {
	return &TransactionConsumer{
		evaluateUseCase: evaluateUseCase,
		logger:          logger,
	}
}

// Setup implements sarama.ConsumerGroupHandler.
func (c *TransactionConsumer) Setup(session sarama.ConsumerGroupSession) error {
	c.logger.Info().
		Str("member_id", session.MemberID()).
		Int32("generation_id", session.GenerationID()).
		Any("claims", session.Claims()).
		Msg("consumer group session started")
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler.
func (c *TransactionConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	c.logger.Info().
		Str("member_id", session.MemberID()).
		Int32("generation_id", session.GenerationID()).
		Msg("consumer group session ended")
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler.
func (c *TransactionConsumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	c.logger.Info().
		Str("topic", claim.Topic()).
		Int32("partition", claim.Partition()).
		Int64("initial_offset", claim.InitialOffset()).
		Msg("starting to consume partition")

	for msg := range claim.Messages() {
		c.logger.Info().
			Str("topic", msg.Topic).
			Int32("partition", msg.Partition).
			Int64("offset", msg.Offset).
			Str("key", string(msg.Key)).
			Msg("message received")

		var transaction entity.TransactionMessage
		if err := json.Unmarshal(msg.Value, &transaction); err != nil {
			c.logger.Error().
				Err(err).
				Str("topic", msg.Topic).
				Int64("offset", msg.Offset).
				Str("raw", string(msg.Value)).
				Msg("failed to deserialize message")
			session.MarkMessage(msg, "")
			continue
		}

		c.logger.Info().
			Str("transaction_id", transaction.ID).
			Int64("amount_in_cents", transaction.AmountInCents).
			Str("currency", transaction.Currency).
			Str("payment_method", transaction.PaymentMethod).
			Str("customer_id", transaction.CustomerID).
			Msg("evaluating transaction")

		result, err := c.evaluateUseCase.Execute(context.Background(), &transaction)
		if err != nil {
			c.logger.Error().
				Err(err).
				Str("transaction_id", transaction.ID).
				Msg("failed to evaluate transaction")
			session.MarkMessage(msg, "")
			continue
		}

		c.logger.Info().
			Str("transaction_id", result.TransactionID).
			Str("decision", string(result.Status)).
			Msg("transaction evaluated")

		session.MarkMessage(msg, "")
	}
	return nil
}
