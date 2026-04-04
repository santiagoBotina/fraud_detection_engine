package dynamodb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"ms-transaction-evaluator/internal/domain/entity"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithymiddleware "github.com/aws/smithy-go/middleware"
	"github.com/rs/zerolog"
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

// fakeHTTPClient returns an HTTP 200 with an empty JSON body for every request,
// preventing any real network call during tests.
type fakeHTTPClient struct{}

func (f *fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/x-amz-json-1.0"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(`{}`))),
	}, nil
}

// newCapturingDynamoDBClient creates a DynamoDB client that captures the
// UpdateItemInput parameters via middleware and uses a fake HTTP client so
// no real DynamoDB call is made.
func newCapturingDynamoDBClient(captured *dynamodb.UpdateItemInput) *dynamodb.Client {
	cfg := aws.Config{
		Region: "us-east-1",
	}
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:8000")
		o.HTTPClient = &fakeHTTPClient{}
		o.APIOptions = append(o.APIOptions, func(stack *smithymiddleware.Stack) error {
			return stack.Initialize.Add(smithymiddleware.InitializeMiddlewareFunc(
				"CaptureUpdateItem",
				func(ctx context.Context, in smithymiddleware.InitializeInput, next smithymiddleware.InitializeHandler) (smithymiddleware.InitializeOutput, smithymiddleware.Metadata, error) {
					if input, ok := in.Parameters.(*dynamodb.UpdateItemInput); ok {
						*captured = *input
					}
					return next.HandleInitialize(ctx, in)
				},
			), smithymiddleware.Before)
		})
	})
	return client
}

func TestUpdateStatus_FinalizedAt(t *testing.T) {
	t.Run("should include finalized_at in UpdateExpression when finalizedAt is non-nil", func(t *testing.T) {
		var captured dynamodb.UpdateItemInput
		client := newCapturingDynamoDBClient(&captured)
		logger := zerolog.Nop()
		repo := NewDynamoDBTransactionRepository(client, "transactions", logger)

		finalizedAt := time.Date(2025, 1, 15, 10, 0, 2, 0, time.UTC)
		err := repo.UpdateStatus(context.Background(), "txn_001", entity.APPROVED, &finalizedAt)
		if err != nil {
			t.Fatalf("UpdateStatus returned unexpected error: %v", err)
		}

		// Verify UpdateExpression contains finalized_at
		if captured.UpdateExpression == nil {
			t.Fatal("Expected UpdateExpression to be set")
		}
		expr := *captured.UpdateExpression
		if !strings.Contains(expr, "finalized_at") {
			t.Errorf("Expected UpdateExpression to contain 'finalized_at', got: %s", expr)
		}

		// Verify ExpressionAttributeValues contains :finalized_at
		if captured.ExpressionAttributeValues == nil {
			t.Fatal("Expected ExpressionAttributeValues to be set")
		}
		if _, ok := captured.ExpressionAttributeValues[":finalized_at"]; !ok {
			t.Error("Expected ExpressionAttributeValues to contain ':finalized_at' key")
		}
	})

	t.Run("should NOT include finalized_at in UpdateExpression when finalizedAt is nil", func(t *testing.T) {
		var captured dynamodb.UpdateItemInput
		client := newCapturingDynamoDBClient(&captured)
		logger := zerolog.Nop()
		repo := NewDynamoDBTransactionRepository(client, "transactions", logger)

		err := repo.UpdateStatus(context.Background(), "txn_002", entity.PENDING, nil)
		if err != nil {
			t.Fatalf("UpdateStatus returned unexpected error: %v", err)
		}

		// Verify UpdateExpression does NOT contain finalized_at
		if captured.UpdateExpression == nil {
			t.Fatal("Expected UpdateExpression to be set")
		}
		expr := *captured.UpdateExpression
		if strings.Contains(expr, "finalized_at") {
			t.Errorf("Expected UpdateExpression to NOT contain 'finalized_at', got: %s", expr)
		}

		// Verify ExpressionAttributeValues does NOT contain :finalized_at
		if _, ok := captured.ExpressionAttributeValues[":finalized_at"]; ok {
			t.Error("Expected ExpressionAttributeValues to NOT contain ':finalized_at' key")
		}
	})
}

// sequentialHTTPClient returns a different HTTP response for each successive
// request, allowing multi-page DynamoDB Scan simulation.
type sequentialHTTPClient struct {
	responses []string
	callCount atomic.Int32
}

func (s *sequentialHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	idx := int(s.callCount.Add(1)) - 1
	if idx >= len(s.responses) {
		idx = len(s.responses) - 1
	}
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/x-amz-json-1.0"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(s.responses[idx]))),
	}, nil
}

// errorHTTPClient always returns an HTTP 500 to simulate a DynamoDB service error.
type errorHTTPClient struct{}

func (e *errorHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	body := `{"__type":"InternalServerError","message":"Service Unavailable"}`
	return &http.Response{
		StatusCode: 500,
		Header:     http.Header{"Content-Type": []string{"application/x-amz-json-1.0"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}, nil
}

func newScanDynamoDBClient(httpClient aws.HTTPClient) *dynamodb.Client {
	cfg := aws.Config{
		Region: "us-east-1",
	}
	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:8000")
		o.HTTPClient = httpClient
	})
}

// scanResponseJSON builds a DynamoDB Scan JSON response body with the given
// transaction items and an optional LastEvaluatedKey.
func scanResponseJSON(items []transactionItem, hasMore bool, lastKeyID string) string {
	itemsJSON := ""
	for i, item := range items {
		if i > 0 {
			itemsJSON += ","
		}
		itemsJSON += fmt.Sprintf(`{
			"id":{"S":"%s"},
			"amount_in_cents":{"N":"%d"},
			"currency":{"S":"%s"},
			"payment_method":{"S":"%s"},
			"customer_id":{"S":"%s"},
			"customer_name":{"S":"%s"},
			"customer_email":{"S":"%s"},
			"customer_phone":{"S":"%s"},
			"customer_ip_address":{"S":"%s"},
			"status":{"S":"%s"},
			"created_at":{"S":"%s"},
			"updated_at":{"S":"%s"}
		}`,
			item.ID,
			item.AmountInCents,
			item.Currency,
			item.PaymentMethod,
			item.CustomerID,
			item.CustomerName,
			item.CustomerEmail,
			item.CustomerPhone,
			item.CustomerIPAddress,
			string(item.Status),
			item.CreatedAt,
			item.UpdatedAt,
		)
	}

	lastKeyJSON := ""
	if hasMore {
		lastKeyJSON = fmt.Sprintf(`,"LastEvaluatedKey":{"id":{"S":"%s"}}`, lastKeyID)
	}

	return fmt.Sprintf(`{"Count":%d,"Items":[%s],"ScannedCount":%d%s}`,
		len(items), itemsJSON, len(items), lastKeyJSON)
}

func TestFindAll(t *testing.T) {
	now := time.Now().UTC()
	timeStr := now.Format("2006-01-02T15:04:05Z07:00")

	t.Run("should collect items across multiple scan pages", func(t *testing.T) {
		page1Items := []transactionItem{
			{
				ID: "txn_001", AmountInCents: 1000, Currency: "USD",
				PaymentMethod: "CARD", CustomerID: "cust_1", CustomerName: "Alice",
				CustomerEmail: "alice@test.com", CustomerPhone: "+1111111111",
				CustomerIPAddress: "10.0.0.1", Status: entity.APPROVED,
				CreatedAt: timeStr, UpdatedAt: timeStr,
			},
			{
				ID: "txn_002", AmountInCents: 2000, Currency: "EUR",
				PaymentMethod: "BANK_TRANSFER", CustomerID: "cust_2", CustomerName: "Bob",
				CustomerEmail: "bob@test.com", CustomerPhone: "+2222222222",
				CustomerIPAddress: "10.0.0.2", Status: entity.PENDING,
				CreatedAt: timeStr, UpdatedAt: timeStr,
			},
		}
		page2Items := []transactionItem{
			{
				ID: "txn_003", AmountInCents: 3000, Currency: "COP",
				PaymentMethod: "CRYPTO", CustomerID: "cust_3", CustomerName: "Charlie",
				CustomerEmail: "charlie@test.com", CustomerPhone: "+3333333333",
				CustomerIPAddress: "10.0.0.3", Status: entity.DECLINED,
				CreatedAt: timeStr, UpdatedAt: timeStr,
			},
		}

		httpClient := &sequentialHTTPClient{
			responses: []string{
				scanResponseJSON(page1Items, true, "txn_002"),
				scanResponseJSON(page2Items, false, ""),
			},
		}

		client := newScanDynamoDBClient(httpClient)
		logger := zerolog.Nop()
		repo := NewDynamoDBTransactionRepository(client, "transactions", logger)

		results, err := repo.FindAll(context.Background())
		if err != nil {
			t.Fatalf("FindAll returned unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 transactions, got %d", len(results))
		}

		expectedIDs := []string{"txn_001", "txn_002", "txn_003"}
		for i, expected := range expectedIDs {
			if results[i].ID != expected {
				t.Errorf("Expected results[%d].ID = %s, got %s", i, expected, results[i].ID)
			}
		}

		// Verify fields on first item
		if results[0].AmountInCents != 1000 {
			t.Errorf("Expected AmountInCents 1000, got %d", results[0].AmountInCents)
		}
		if results[0].Currency != entity.USD {
			t.Errorf("Expected Currency USD, got %s", results[0].Currency)
		}
		if results[0].Status != entity.APPROVED {
			t.Errorf("Expected Status APPROVED, got %s", results[0].Status)
		}

		// Verify two scan calls were made
		callCount := int(httpClient.callCount.Load())
		if callCount != 2 {
			t.Errorf("Expected 2 scan calls for pagination, got %d", callCount)
		}
	})

	t.Run("should return empty slice for empty table", func(t *testing.T) {
		httpClient := &sequentialHTTPClient{
			responses: []string{
				`{"Count":0,"Items":[],"ScannedCount":0}`,
			},
		}

		client := newScanDynamoDBClient(httpClient)
		logger := zerolog.Nop()
		repo := NewDynamoDBTransactionRepository(client, "transactions", logger)

		results, err := repo.FindAll(context.Background())
		if err != nil {
			t.Fatalf("FindAll returned unexpected error: %v", err)
		}

		if results == nil {
			// FindAll returns nil when no items appended; both nil and empty are acceptable
			results = []entity.TransactionEntity{}
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 transactions, got %d", len(results))
		}
	})

	t.Run("should propagate scan error", func(t *testing.T) {
		httpClient := &errorHTTPClient{}

		client := newScanDynamoDBClient(httpClient)
		logger := zerolog.Nop()
		repo := NewDynamoDBTransactionRepository(client, "transactions", logger)

		results, err := repo.FindAll(context.Background())
		if err == nil {
			t.Fatal("Expected FindAll to return an error, got nil")
		}

		if results != nil {
			t.Errorf("Expected nil results on error, got %v", results)
		}

		if !strings.Contains(err.Error(), "failed to scan transactions") {
			t.Errorf("Expected error to contain 'failed to scan transactions', got: %s", err.Error())
		}
	})
}
