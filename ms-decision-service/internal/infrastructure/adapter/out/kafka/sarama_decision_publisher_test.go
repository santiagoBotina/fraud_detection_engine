package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"ms-decision-service/internal/domain/entity"
	"testing"

	"github.com/rs/zerolog"

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
			publisher := NewSaramaDecisionPublisher(mockProducer, "test-topic", zerolog.Nop())

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

		properties.Property("success log lines contain transaction_id, status, topic, partition, offset", prop.ForAll(
			func(result *entity.DecisionResult) bool {
				var buf bytes.Buffer
				logger := zerolog.New(&buf)

				mockProducer := &capturingSyncProducer{}
				publisher := NewSaramaDecisionPublisher(mockProducer, "test-topic", logger)

				err := publisher.Publish(context.Background(), result)
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

				requiredFields := []string{"transaction_id", "status", "topic", "partition", "offset"}
				for _, f := range requiredFields {
					if _, ok := fields[f]; !ok {
						return false
					}
				}

				return true
			},
			genDecisionResult(),
		))

		properties.TestingRun(t)
	})

	t.Run("error path contains error field", func(t *testing.T) {
		properties := gopter.NewProperties(parameters)

		properties.Property("error log lines contain error field", prop.ForAll(
			func(result *entity.DecisionResult) bool {
				var buf bytes.Buffer
				logger := zerolog.New(&buf)

				mockProducer := &failingSyncProducer{err: errors.New("kafka send failed")}
				publisher := NewSaramaDecisionPublisher(mockProducer, "test-topic", logger)

				err := publisher.Publish(context.Background(), result)
				if err == nil {
					return false
				}

				lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
				if len(lines) == 0 {
					return false
				}

				// The last line should be the final error log after retries exhausted
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

				return true
			},
			genDecisionResult(),
		))

		properties.TestingRun(t)
	})
}
