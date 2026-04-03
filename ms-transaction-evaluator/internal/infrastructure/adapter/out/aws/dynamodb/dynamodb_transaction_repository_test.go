package dynamodb

import (
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

func TestTransactionItem_StatusMapping(t *testing.T) {
	t.Run("should map entity status to DynamoDB item", func(t *testing.T) {
		now := time.Now().UTC()
		txn := &entity.TransactionEntity{
			ID:            "txn_123",
			AmountInCents: 5000,
			Currency:      entity.USD,
			PaymentMethod: entity.CARD,
			Status:        entity.PENDING,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		item := transactionItem{
			ID:            txn.ID,
			AmountInCents: txn.AmountInCents,
			Currency:      string(txn.Currency),
			PaymentMethod: string(txn.PaymentMethod),
			Status:        txn.Status,
			CreatedAt:     txn.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     txn.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if item.Status != entity.PENDING {
			t.Errorf("Expected status PENDING, got %s", item.Status)
		}
	})

	t.Run("should marshal status to DynamoDB attribute value", func(t *testing.T) {
		item := transactionItem{
			ID:     "txn_456",
			Status: entity.PENDING,
		}

		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			t.Fatalf("Failed to marshal item: %v", err)
		}

		statusAttr, ok := av["status"]
		if !ok {
			t.Fatal("Expected 'status' attribute in marshaled map")
		}

		var status string
		err = attributevalue.Unmarshal(statusAttr, &status)
		if err != nil {
			t.Fatalf("Failed to unmarshal status attribute: %v", err)
		}

		if status != "PENDING" {
			t.Errorf("Expected status attribute to be PENDING, got %s", status)
		}
	})

	t.Run("should preserve all status values through marshaling", func(t *testing.T) {
		statuses := []entity.TransactionStatus{
			entity.PENDING,
			entity.APPROVED,
			entity.DECLINED,
		}

		for _, s := range statuses {
			item := transactionItem{
				ID:     "txn_test",
				Status: s,
			}

			av, err := attributevalue.MarshalMap(item)
			if err != nil {
				t.Fatalf("Failed to marshal item with status %s: %v", s, err)
			}

			var status string
			err = attributevalue.Unmarshal(av["status"], &status)
			if err != nil {
				t.Fatalf("Failed to unmarshal status %s: %v", s, err)
			}

			if status != string(s) {
				t.Errorf("Expected %s, got %s", s, status)
			}
		}
	})
}
