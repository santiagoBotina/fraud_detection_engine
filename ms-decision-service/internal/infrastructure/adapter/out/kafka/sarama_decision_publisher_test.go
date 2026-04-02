package kafka

import (
	"context"
	"log/slog"
	"testing"

	"ms-decision-service/internal/domain/entity"

	"github.com/IBM/sarama"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// capturingSyncProducer is a mock sarama.SyncProducer that captures the last sent message.
type capturingSyncProducer struct {
	lastMessage *sarama.ProducerMessage
}

func (p *capturingSyncProducer) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	p.lastMessage = msg
	return 0, 0, nil
}

func (p *capturingSyncProducer) SendMessages(_ []*sarama.ProducerMessage) error {
	return nil
}

func (p *capturingSyncProducer) Close() error {
	return nil
}

func (p *capturingSyncProducer) IsTransactional() bool {
	return false
}

func (p *capturingSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	return sarama.ProducerTxnFlagReady
}

func (p *capturingSyncProducer) BeginTxn() error {
	return nil
}

func (p *capturingSyncProducer) CommitTxn() error {
	return nil
}

func (p *capturingSyncProducer) AbortTxn() error {
	return nil
}

func (p *capturingSyncProducer) AddOffsetsToTxn(_ map[string][]*sarama.PartitionOffsetMetadata, _ string) error {
	return nil
}

func (p *capturingSyncProducer) AddMessageToTxn(_ *sarama.ConsumerMessage, _ string, _ *string) error {
	return nil
}

func genDecisionResult() gopter.Gen {
	statuses := []entity.DecisionStatus{
		entity.APPROVED,
		entity.DECLINED,
		entity.FRAUDCHECK,
	}

	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(0, len(statuses)-1),
	).Map(func(v []interface{}) *entity.DecisionResult {
		return &entity.DecisionResult{
			TransactionID: v[0].(string),
			Status:        statuses[v[1].(int)],
		}
	})
}

// Feature: kafka-transaction-decision-service, Property 7: Kafka message key equals transaction identifier
// **Validates: Requirements 5.3**
func TestProperty7_KafkaMessageKeyEqualsTransactionID(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("message key equals transaction ID", prop.ForAll(
		func(result *entity.DecisionResult) bool {
			mockProducer := &capturingSyncProducer{}
			publisher := NewSaramaDecisionPublisher(mockProducer, "test-topic", slog.Default())

			err := publisher.Publish(context.Background(), result)
			if err != nil {
				return false
			}

			if mockProducer.lastMessage == nil {
				return false
			}

			key, keyErr := mockProducer.lastMessage.Key.Encode()
			if keyErr != nil {
				return false
			}

			return string(key) == result.TransactionID
		},
		genDecisionResult(),
	))

	properties.TestingRun(t)
}
