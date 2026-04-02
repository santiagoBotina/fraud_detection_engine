package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/usecase"

	"github.com/IBM/sarama"
)

// TransactionConsumer implements sarama.ConsumerGroupHandler for processing transaction messages.
type TransactionConsumer struct {
	evaluateUseCase *usecase.EvaluateTransactionUseCase
	logger          *slog.Logger
}

// NewTransactionConsumer creates a new consumer with the given use case and logger.
func NewTransactionConsumer(
	evaluateUseCase *usecase.EvaluateTransactionUseCase,
	logger *slog.Logger,
) *TransactionConsumer {
	return &TransactionConsumer{
		evaluateUseCase: evaluateUseCase,
		logger:          logger,
	}
}

// Setup implements sarama.ConsumerGroupHandler.
func (c *TransactionConsumer) Setup(session sarama.ConsumerGroupSession) error {
	c.logger.Info("consumer group session started",
		"member_id", session.MemberID(),
		"generation_id", session.GenerationID(),
		"claims", session.Claims(),
	)
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler.
func (c *TransactionConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	c.logger.Info("consumer group session ended",
		"member_id", session.MemberID(),
		"generation_id", session.GenerationID(),
	)
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler.
func (c *TransactionConsumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	c.logger.Info("starting to consume partition",
		"topic", claim.Topic(),
		"partition", claim.Partition(),
		"initial_offset", claim.InitialOffset(),
	)

	for msg := range claim.Messages() {
		c.logger.Info("message received",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"key", string(msg.Key),
		)

		var transaction entity.TransactionMessage
		if err := json.Unmarshal(msg.Value, &transaction); err != nil {
			c.logger.Error("failed to deserialize message",
				"error", err,
				"topic", msg.Topic,
				"offset", msg.Offset,
				"raw", string(msg.Value),
			)
			session.MarkMessage(msg, "")
			continue
		}

		c.logger.Info("evaluating transaction",
			"transaction_id", transaction.ID,
			"amount_in_cents", transaction.AmountInCents,
			"currency", transaction.Currency,
			"payment_method", transaction.PaymentMethod,
			"customer_id", transaction.CustomerID,
		)

		result, err := c.evaluateUseCase.Execute(context.Background(), &transaction)
		if err != nil {
			c.logger.Error("failed to evaluate transaction",
				"error", err,
				"transaction_id", transaction.ID,
			)
			session.MarkMessage(msg, "")
			continue
		}

		c.logger.Info("transaction evaluated",
			"transaction_id", result.TransactionID,
			"decision", result.Status,
		)

		session.MarkMessage(msg, "")
	}
	return nil
}
