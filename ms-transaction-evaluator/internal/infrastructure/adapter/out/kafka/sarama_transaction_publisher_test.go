package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/rs/zerolog"
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
			publisher := NewSaramaTransactionPublisher(mockProducer, "test-topic", zerolog.Nop())

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

// failingSyncProducer is a mock sarama.SyncProducer that always returns an error from SendMessage.
type failingSyncProducer struct {
	err error
}

func (p *failingSyncProducer) SendMessage(_ *sarama.ProducerMessage) (int32, int64, error) {
	return 0, 0, p.err
}

func (p *failingSyncProducer) SendMessages(_ []*sarama.ProducerMessage) error {
	return p.err
}

func (p *failingSyncProducer) Close() error {
	return nil
}

func (p *failingSyncProducer) IsTransactional() bool {
	return false
}

func (p *failingSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	return sarama.ProducerTxnFlagReady
}

func (p *failingSyncProducer) BeginTxn() error {
	return nil
}

func (p *failingSyncProducer) CommitTxn() error {
	return nil
}

func (p *failingSyncProducer) AbortTxn() error {
	return nil
}

func (p *failingSyncProducer) AddOffsetsToTxn(_ map[string][]*sarama.PartitionOffsetMetadata, _ string) error {
	return nil
}

func (p *failingSyncProducer) AddMessageToTxn(_ *sarama.ConsumerMessage, _ string, _ *string) error {
	return nil
}

// Feature: zerolog-logging-refactor, Property 1: Structured log field preservation
// **Validates: Requirements 5.1, 5.3**
func TestProperty1_StructuredLogFieldPreservation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	t.Run("success path contains all expected fields", func(t *testing.T) {
		properties := gopter.NewProperties(parameters)

		properties.Property("success log lines contain transaction_id, topic, status, partition, offset", prop.ForAll(
			func(tx *entity.TransactionEntity) bool {
				var buf bytes.Buffer
				logger := zerolog.New(&buf)

				mockProducer := &capturingSyncProducer{}
				publisher := NewSaramaTransactionPublisher(mockProducer, "test-topic", logger)

				err := publisher.Publish(context.Background(), tx)
				if err != nil {
					return false
				}

				// Parse the last JSON log line (the "published" success message)
				lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
				if len(lines) == 0 {
					return false
				}

				lastLine := lines[len(lines)-1]
				var fields map[string]interface{}
				if jsonErr := json.Unmarshal(lastLine, &fields); jsonErr != nil {
					return false
				}

				requiredFields := []string{"transaction_id", "topic", "partition", "offset"}
				for _, f := range requiredFields {
					if _, ok := fields[f]; !ok {
						return false
					}
				}

				// Also check the first log line has status (the "publishing" message)
				firstLine := lines[0]
				var firstFields map[string]interface{}
				if jsonErr := json.Unmarshal(firstLine, &firstFields); jsonErr != nil {
					return false
				}

				if _, ok := firstFields["status"]; !ok {
					return false
				}

				return true
			},
			genTransactionEntity(),
		))

		properties.TestingRun(t)
	})

	t.Run("error path contains error field", func(t *testing.T) {
		properties := gopter.NewProperties(parameters)

		properties.Property("error log lines contain error field", prop.ForAll(
			func(tx *entity.TransactionEntity) bool {
				var buf bytes.Buffer
				logger := zerolog.New(&buf)

				mockProducer := &failingSyncProducer{err: errors.New("kafka send failed")}
				publisher := NewSaramaTransactionPublisher(mockProducer, "test-topic", logger)

				err := publisher.Publish(context.Background(), tx)
				if err == nil {
					return false
				}

				lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
				if len(lines) == 0 {
					return false
				}

				// Find the error log line (last line should be the error)
				lastLine := lines[len(lines)-1]
				var fields map[string]interface{}
				if jsonErr := json.Unmarshal(lastLine, &fields); jsonErr != nil {
					return false
				}

				if _, ok := fields["error"]; !ok {
					return false
				}

				if _, ok := fields["transaction_id"]; !ok {
					return false
				}

				if _, ok := fields["topic"]; !ok {
					return false
				}

				return true
			},
			genTransactionEntity(),
		))

		properties.TestingRun(t)
	})
}
