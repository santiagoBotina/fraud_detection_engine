package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"ms-decision-service/internal/domain/entity"
	"ms-decision-service/internal/domain/repository"
	"ms-decision-service/internal/domain/usecase"

	"github.com/IBM/sarama"
)

// --- Mock RuleRepository ---

type mockRuleRepository struct {
	findFunc func(ctx context.Context) ([]entity.Rule, error)
}

func (m *mockRuleRepository) FindActiveRulesSortedByPriority(ctx context.Context) ([]entity.Rule, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx)
	}
	return nil, nil
}

// --- Mock DecisionPublisher ---

type mockDecisionPublisher struct {
	publishFunc func(ctx context.Context, result *entity.DecisionResult) error
	published   []*entity.DecisionResult
}

func (m *mockDecisionPublisher) Publish(ctx context.Context, result *entity.DecisionResult) error {
	m.published = append(m.published, result)
	if m.publishFunc != nil {
		return m.publishFunc(ctx, result)
	}
	return nil
}

// --- Mock ConsumerGroupSession ---

type mockConsumerGroupSession struct {
	markedMessages []*sarama.ConsumerMessage
}

func (m *mockConsumerGroupSession) Claims() map[string][]int32          { return nil }
func (m *mockConsumerGroupSession) MemberID() string                    { return "test-member" }
func (m *mockConsumerGroupSession) GenerationID() int32                 { return 1 }
func (m *mockConsumerGroupSession) MarkOffset(string, int32, int64, string) {}
func (m *mockConsumerGroupSession) Commit()                             {}
func (m *mockConsumerGroupSession) ResetOffset(string, int32, int64, string) {}
func (m *mockConsumerGroupSession) Context() context.Context            { return context.Background() }

func (m *mockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, _ string) {
	m.markedMessages = append(m.markedMessages, msg)
}

// --- Mock ConsumerGroupClaim ---

type mockConsumerGroupClaim struct {
	messages chan *sarama.ConsumerMessage
}

func (m *mockConsumerGroupClaim) Topic() string             { return "test-topic" }
func (m *mockConsumerGroupClaim) Partition() int32          { return 0 }
func (m *mockConsumerGroupClaim) InitialOffset() int64      { return 0 }
func (m *mockConsumerGroupClaim) HighWaterMarkOffset() int64 { return 0 }
func (m *mockConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage { return m.messages }

// --- Helper ---

func buildUseCase(ruleRepo repository.RuleRepository, publisher repository.DecisionPublisher) *usecase.EvaluateTransactionUseCase {
	return usecase.NewEvaluateTransactionUseCase(ruleRepo, publisher)
}

func validTransactionJSON() []byte {
	tx := entity.TransactionMessage{
		ID:                "tx-123",
		AmountInCents:     5000,
		Currency:          "USD",
		PaymentMethod:     "CARD",
		CustomerID:        "cust-1",
		CustomerName:      "John Doe",
		CustomerEmail:     "john@example.com",
		CustomerPhone:     "555-0100",
		CustomerIPAddress: "192.168.1.1",
		Status:            "PENDING",
		CreatedAt:         time.Now().UTC().Truncate(time.Second),
		UpdatedAt:         time.Now().UTC().Truncate(time.Second),
	}
	data, _ := json.Marshal(tx)
	return data
}

func TestConsumeClaim_ValidMessage(t *testing.T) {
	ruleRepo := &mockRuleRepository{}
	publisher := &mockDecisionPublisher{}
	uc := buildUseCase(ruleRepo, publisher)
	logger := slog.Default()

	consumer := NewTransactionConsumer(uc, logger)

	session := &mockConsumerGroupSession{}
	msgChan := make(chan *sarama.ConsumerMessage, 1)
	claim := &mockConsumerGroupClaim{messages: msgChan}

	msgChan <- &sarama.ConsumerMessage{Value: validTransactionJSON()}
	close(msgChan)

	err := consumer.ConsumeClaim(session, claim)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(session.markedMessages) != 1 {
		t.Fatalf("expected 1 marked message, got %d", len(session.markedMessages))
	}

	if len(publisher.published) != 1 {
		t.Fatalf("expected 1 published decision, got %d", len(publisher.published))
	}

	if publisher.published[0].TransactionID != "tx-123" {
		t.Errorf("expected transaction ID tx-123, got %s", publisher.published[0].TransactionID)
	}
}

func TestConsumeClaim_MalformedJSON(t *testing.T) {
	ruleRepo := &mockRuleRepository{}
	publisher := &mockDecisionPublisher{}
	uc := buildUseCase(ruleRepo, publisher)
	logger := slog.Default()

	consumer := NewTransactionConsumer(uc, logger)

	session := &mockConsumerGroupSession{}
	msgChan := make(chan *sarama.ConsumerMessage, 1)
	claim := &mockConsumerGroupClaim{messages: msgChan}

	msgChan <- &sarama.ConsumerMessage{Value: []byte("not valid json!!!")}
	close(msgChan)

	err := consumer.ConsumeClaim(session, claim)
	if err != nil {
		t.Fatalf("expected no error (should not crash), got %v", err)
	}

	// Message should still be marked even though deserialization failed
	if len(session.markedMessages) != 1 {
		t.Fatalf("expected 1 marked message (malformed skipped), got %d", len(session.markedMessages))
	}

	// No decision should have been published
	if len(publisher.published) != 0 {
		t.Fatalf("expected 0 published decisions for malformed message, got %d", len(publisher.published))
	}
}

func TestConsumeClaim_UseCaseError(t *testing.T) {
	ruleRepo := &mockRuleRepository{
		findFunc: func(_ context.Context) ([]entity.Rule, error) {
			return nil, context.DeadlineExceeded
		},
	}
	publisher := &mockDecisionPublisher{}
	uc := buildUseCase(ruleRepo, publisher)
	logger := slog.Default()

	consumer := NewTransactionConsumer(uc, logger)

	session := &mockConsumerGroupSession{}
	msgChan := make(chan *sarama.ConsumerMessage, 1)
	claim := &mockConsumerGroupClaim{messages: msgChan}

	msgChan <- &sarama.ConsumerMessage{Value: validTransactionJSON()}
	close(msgChan)

	err := consumer.ConsumeClaim(session, claim)
	if err != nil {
		t.Fatalf("expected no error (use case error should be logged, not returned), got %v", err)
	}

	// Message should still be marked even when use case fails
	if len(session.markedMessages) != 1 {
		t.Fatalf("expected 1 marked message despite use case error, got %d", len(session.markedMessages))
	}

	// No decision should have been published since rule retrieval failed
	if len(publisher.published) != 0 {
		t.Fatalf("expected 0 published decisions when use case errors, got %d", len(publisher.published))
	}
}
