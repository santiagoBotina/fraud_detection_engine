package kafka

import (
	"context"
	"encoding/json"
	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/usecase"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// FraudScoreConsumer implements sarama.ConsumerGroupHandler for processing fraud score calculated messages.
type FraudScoreConsumer struct {
	evaluateUseCase *usecase.EvaluateFraudScoreUseCase
	logger          zerolog.Logger
}

// NewFraudScoreConsumer creates a new consumer with the given use case and logger.
func NewFraudScoreConsumer(
	evaluateUseCase *usecase.EvaluateFraudScoreUseCase,
	logger zerolog.Logger,
) *FraudScoreConsumer {
	return &FraudScoreConsumer{
		evaluateUseCase: evaluateUseCase,
		logger:          logger,
	}
}

// Setup implements sarama.ConsumerGroupHandler.
func (c *FraudScoreConsumer) Setup(session sarama.ConsumerGroupSession) error {
	c.logger.Info().
		Str("member_id", session.MemberID()).
		Int32("generation_id", session.GenerationID()).
		Any("claims", session.Claims()).
		Msg("consumer group session started")
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler.
func (c *FraudScoreConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	c.logger.Info().
		Str("member_id", session.MemberID()).
		Int32("generation_id", session.GenerationID()).
		Msg("consumer group session ended")
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler.
func (c *FraudScoreConsumer) ConsumeClaim(
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

		var fraudScore entity.FraudScoreCalculatedMessage
		if err := json.Unmarshal(msg.Value, &fraudScore); err != nil {
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
			Str("transaction_id", fraudScore.TransactionID).
			Int("fraud_score", fraudScore.FraudScore).
			Msg("evaluating fraud score")

		result, err := c.evaluateUseCase.Execute(context.Background(), &fraudScore)
		if err != nil {
			c.logger.Error().
				Err(err).
				Str("transaction_id", fraudScore.TransactionID).
				Msg("failed to evaluate fraud score")
			session.MarkMessage(msg, "")
			continue
		}

		c.logger.Info().
			Str("transaction_id", result.TransactionID).
			Str("decision", string(result.Status)).
			Msg("fraud score evaluated")

		session.MarkMessage(msg, "")
	}
	return nil
}
