package dynamodb

import (
	"context"
	"fmt"

	"ms-transaction-evaluator/internal/domain/entity"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBTransactionRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBTransactionRepository(client *dynamodb.Client, tableName string) *DynamoDBTransactionRepository {
	return &DynamoDBTransactionRepository{
		client:    client,
		tableName: tableName,
	}
}

type transactionItem struct {
	ID                string `dynamodbav:"id"`
	AmountInCents     int64  `dynamodbav:"amount_in_cents"`
	Currency          string `dynamodbav:"currency"`
	PaymentMethod     string `dynamodbav:"payment_method"`
	CustomerID        string `dynamodbav:"customer_id"`
	CustomerName      string `dynamodbav:"customer_name"`
	CustomerEmail     string `dynamodbav:"customer_email"`
	CustomerPhone     string `dynamodbav:"customer_phone"`
	CustomerIPAddress string `dynamodbav:"customer_ip_address"`
	Status  		  entity.TransactionStatus `dynamodbav:"status"`
	CreatedAt         string `dynamodbav:"created_at"`
	UpdatedAt         string `dynamodbav:"updated_at"`
}

func (r *DynamoDBTransactionRepository) Save(ctx context.Context, transaction *entity.TransactionEntity) error {
	// Convert entity to DynamoDB item
	item := transactionItem{
		ID:                transaction.ID,
		AmountInCents:     transaction.AmountInCents,
		Currency:          string(transaction.Currency),
		PaymentMethod:     string(transaction.PaymentMethod),
		CustomerID:        transaction.CustomerID,
		CustomerName:      transaction.CustomerName,
		CustomerEmail:     transaction.CustomerEmail,
		CustomerPhone:     transaction.CustomerPhone,
		CustomerIPAddress: transaction.CustomerIPAddress,
		Status:            transaction.Status,
		CreatedAt:         transaction.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         transaction.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if err != nil && err.Error() != "" {
			return fmt.Errorf("transaction with id %s already exists", transaction.ID)
		}
		_ = ccf
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	return nil
}
