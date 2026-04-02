package kafka

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"ms-transaction-evaluator/internal/domain/entity"

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

func genTransactionEntity() gopter.Gen {
	currencies := []entity.Currency{entity.USD, entity.COP, entity.EUR}
	paymentMethods := []entity.PaymentMethod{entity.CARD, entity.BANK_TRANSFER, entity.CRYPTO}
	statuses := []entity.TransactionStatus{entity.PENDING, entity.APPROVED, entity.REJECTED, entity.FAILED, entity.CANCELLED}

	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.Int64Range(1, 999999999),
		gen.IntRange(0, len(currencies)-1),
		gen.IntRange(0, len(paymentMethods)-1),
		gen.IntRange(0, len(statuses)-1),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
	).Map(func(v []interface{}) *entity.TransactionEntity {
		now := time.Now().UTC().Truncate(time.Second)
		return &entity.TransactionEntity{
			ID:                v[0].(string),
			AmountInCents:     v[1].(int64),
			Currency:          currencies[v[2].(int)],
			PaymentMethod:     paymentMethods[v[3].(int)],
			Status:            statuses[v[4].(int)],
			CustomerID:        v[5].(string),
			CustomerName:      v[6].(string),
			CustomerEmail:     v[7].(string),
			CustomerPhone:     v[8].(string),
			CustomerIPAddress: v[9].(string),
			CreatedAt:         now,
			UpdatedAt:         now,
		}
	})
}

// Feature: kafka-transaction-decision-service, Property 7: Kafka message key equals transaction identifier
// **Validates: Requirements 1.3, 5.3**
func TestProperty7_KafkaMessageKeyEqualsTransactionID(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("message key equals transaction ID", prop.ForAll(
		func(tx *entity.TransactionEntity) bool {
			mockProducer := &capturingSyncProducer{}
			publisher := NewSaramaTransactionPublisher(mockProducer, "test-topic", slog.Default())

			err := publisher.Publish(context.Background(), tx)
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

			return string(key) == tx.ID
		},
		genTransactionEntity(),
	))

	properties.TestingRun(t)
}
