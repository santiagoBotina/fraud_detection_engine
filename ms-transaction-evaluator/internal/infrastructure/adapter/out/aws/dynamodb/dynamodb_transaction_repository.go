package dynamodb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"ms-transaction-evaluator/internal/domain/entity"
	"sort"
	"time"

	"github.com/rs/zerolog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBTransactionRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    zerolog.Logger
}

func NewDynamoDBTransactionRepository(client *dynamodb.Client, tableName string, logger zerolog.Logger) *DynamoDBTransactionRepository {
	return &DynamoDBTransactionRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
	}
}

type transactionItem struct {
	ID                string                   `dynamodbav:"id"`
	AmountInCents     int64                    `dynamodbav:"amount_in_cents"`
	Currency          string                   `dynamodbav:"currency"`
	PaymentMethod     string                   `dynamodbav:"payment_method"`
	CustomerID        string                   `dynamodbav:"customer_id"`
	CustomerName      string                   `dynamodbav:"customer_name"`
	CustomerEmail     string                   `dynamodbav:"customer_email"`
	CustomerPhone     string                   `dynamodbav:"customer_phone"`
	CustomerIPAddress string                   `dynamodbav:"customer_ip_address"`
	Status            entity.TransactionStatus `dynamodbav:"status"`
	CreatedAt         string                   `dynamodbav:"created_at"`
	UpdatedAt         string                   `dynamodbav:"updated_at"`
}

func (r *DynamoDBTransactionRepository) Save(ctx context.Context, transaction *entity.TransactionEntity) error {
	r.logger.Info().
		Str("transaction_id", transaction.ID).
		Str("table", r.tableName).
		Msg("saving transaction to DynamoDB")

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
		r.logger.Error().
			Err(err).
			Str("transaction_id", transaction.ID).
			Msg("failed to marshal transaction for DynamoDB")
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			r.logger.Warn().
				Str("transaction_id", transaction.ID).
				Str("table", r.tableName).
				Msg("duplicate transaction")
			return fmt.Errorf("transaction with id %s already exists", transaction.ID)
		}

		r.logger.Error().
			Err(err).
			Str("transaction_id", transaction.ID).
			Str("table", r.tableName).
			Msg("failed to save transaction to DynamoDB")
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	r.logger.Info().
		Str("transaction_id", transaction.ID).
		Str("table", r.tableName).
		Msg("transaction saved to DynamoDB")

	return nil
}

// UpdateStatus updates the status and updated_at fields of a transaction in DynamoDB.
func (r *DynamoDBTransactionRepository) UpdateStatus(ctx context.Context, id string, status entity.TransactionStatus) error {
	r.logger.Info().
		Str("transaction_id", id).
		Str("status", string(status)).
		Str("table", r.tableName).
		Msg("updating transaction status in DynamoDB")

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET #s = :status, updated_at = :now"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
			":now":    &types.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("transaction_id", id).
			Str("table", r.tableName).
			Msg("failed to update transaction status")
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	r.logger.Info().
		Str("transaction_id", id).
		Str("status", string(status)).
		Str("table", r.tableName).
		Msg("transaction status updated")

	return nil
}

type paginationCursor struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

func (r *DynamoDBTransactionRepository) mapItemToEntity(item transactionItem) (entity.TransactionEntity, error) {
	createdAt, err := time.Parse("2006-01-02T15:04:05Z07:00", item.CreatedAt)
	if err != nil {
		return entity.TransactionEntity{}, fmt.Errorf("failed to parse created_at: %w", err)
	}

	updatedAt, err := time.Parse("2006-01-02T15:04:05Z07:00", item.UpdatedAt)
	if err != nil {
		return entity.TransactionEntity{}, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return entity.TransactionEntity{
		ID:                item.ID,
		AmountInCents:     item.AmountInCents,
		Currency:          entity.Currency(item.Currency),
		PaymentMethod:     entity.PaymentMethod(item.PaymentMethod),
		CustomerID:        item.CustomerID,
		CustomerName:      item.CustomerName,
		CustomerEmail:     item.CustomerEmail,
		CustomerPhone:     item.CustomerPhone,
		CustomerIPAddress: item.CustomerIPAddress,
		Status:            item.Status,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}, nil
}

func (r *DynamoDBTransactionRepository) FindByID(ctx context.Context, id string) (*entity.TransactionEntity, error) {
	r.logger.Info().
		Str("transaction_id", id).
		Str("table", r.tableName).
		Msg("finding transaction by ID in DynamoDB")

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("transaction_id", id).
			Str("table", r.tableName).
			Msg("failed to get transaction from DynamoDB")
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var item transactionItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		r.logger.Error().
			Err(err).
			Str("transaction_id", id).
			Msg("failed to unmarshal transaction from DynamoDB")
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	txn, err := r.mapItemToEntity(item)
	if err != nil {
		return nil, err
	}

	return &txn, nil
}

func (r *DynamoDBTransactionRepository) FindAllPaginated(ctx context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error) {
	r.logger.Info().
		Int("limit", limit).
		Str("table", r.tableName).
		Msg("scanning transactions from DynamoDB")

	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Limit:     aws.Int32(int32(limit)),
	}

	if cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			r.logger.Error().
				Err(err).
				Msg("failed to decode cursor")
			return nil, "", errors.New("invalid cursor: failed to decode base64")
		}

		var cur paginationCursor
		if err := json.Unmarshal(decoded, &cur); err != nil {
			r.logger.Error().
				Err(err).
				Msg("failed to unmarshal cursor JSON")
			return nil, "", errors.New("invalid cursor: failed to parse JSON")
		}

		if cur.ID == "" || cur.CreatedAt == "" {
			return nil, "", errors.New("invalid cursor: missing required fields")
		}

		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: cur.ID},
		}
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("table", r.tableName).
			Msg("failed to scan transactions from DynamoDB")
		return nil, "", fmt.Errorf("failed to scan transactions: %w", err)
	}

	transactions := make([]entity.TransactionEntity, 0, len(result.Items))
	for _, item := range result.Items {
		var ddbItem transactionItem
		if err := attributevalue.UnmarshalMap(item, &ddbItem); err != nil {
			r.logger.Error().
				Err(err).
				Msg("failed to unmarshal transaction item")
			return nil, "", fmt.Errorf("failed to unmarshal transaction: %w", err)
		}

		txn, err := r.mapItemToEntity(ddbItem)
		if err != nil {
			return nil, "", err
		}

		transactions = append(transactions, txn)
	}

	// Sort client-side by created_at descending
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
	})

	// Build next_cursor from DynamoDB's LastEvaluatedKey
	var nextCursor string
	if result.LastEvaluatedKey != nil && len(result.LastEvaluatedKey) > 0 {
		// Use the last item in our sorted results for the cursor
		lastTxn := transactions[len(transactions)-1]
		cur := paginationCursor{
			ID:        lastTxn.ID,
			CreatedAt: lastTxn.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		curJSON, err := json.Marshal(cur)
		if err != nil {
			r.logger.Error().
				Err(err).
				Msg("failed to marshal cursor")
			return nil, "", fmt.Errorf("failed to marshal cursor: %w", err)
		}

		nextCursor = base64.StdEncoding.EncodeToString(curJSON)
	}

	return transactions, nextCursor, nil
}
