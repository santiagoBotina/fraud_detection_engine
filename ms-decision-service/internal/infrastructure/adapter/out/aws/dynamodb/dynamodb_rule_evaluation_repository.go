package dynamodb

import (
	"context"
	"fmt"
	"sort"
	"time"

	"ms-decision-service/internal/domain/entity"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

const maxBatchWriteItems = 25

type ruleEvaluationItem struct {
	TransactionID     string `dynamodbav:"transaction_id"`
	RuleID            string `dynamodbav:"rule_id"`
	RuleName          string `dynamodbav:"rule_name"`
	ConditionField    string `dynamodbav:"condition_field"`
	ConditionOperator string `dynamodbav:"condition_operator"`
	ConditionValue    string `dynamodbav:"condition_value"`
	ActualFieldValue  string `dynamodbav:"actual_field_value"`
	Matched           bool   `dynamodbav:"matched"`
	ResultStatus      string `dynamodbav:"result_status"`
	EvaluatedAt       string `dynamodbav:"evaluated_at"`
	Priority          int    `dynamodbav:"priority"`
}

// DynamoDBRuleEvaluationRepository implements repository.RuleEvaluationRepository using AWS DynamoDB.
type DynamoDBRuleEvaluationRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    zerolog.Logger
}

// NewDynamoDBRuleEvaluationRepository creates a new DynamoDB-backed rule evaluation repository.
func NewDynamoDBRuleEvaluationRepository(
	client *dynamodb.Client,
	tableName string,
	logger zerolog.Logger,
) *DynamoDBRuleEvaluationRepository {
	return &DynamoDBRuleEvaluationRepository{client: client, tableName: tableName, logger: logger}
}

// SaveBatch persists rule evaluation results using BatchWriteItem, handling the 25-item limit per batch.
func (r *DynamoDBRuleEvaluationRepository) SaveBatch(ctx context.Context, results []entity.RuleEvaluationResult) error {
	if len(results) == 0 {
		return nil
	}

	r.logger.Info().
		Str("table", r.tableName).
		Int("count", len(results)).
		Msg("saving rule evaluation results")

	for i := 0; i < len(results); i += maxBatchWriteItems {
		end := i + maxBatchWriteItems
		if end > len(results) {
			end = len(results)
		}

		chunk := results[i:end]

		writeRequests := make([]types.WriteRequest, 0, len(chunk))
		for _, result := range chunk {
			item := toRuleEvaluationItem(result)

			av, err := attributevalue.MarshalMap(item)
			if err != nil {
				r.logger.Error().Err(err).
					Str("transaction_id", result.TransactionID).
					Str("rule_id", result.RuleID).
					Msg("failed to marshal rule evaluation item")
				return fmt.Errorf("failed to marshal rule evaluation item: %w", err)
			}

			writeRequests = append(writeRequests, types.WriteRequest{
				PutRequest: &types.PutRequest{Item: av},
			})
		}

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				r.tableName: writeRequests,
			},
		}

		output, err := r.client.BatchWriteItem(ctx, input)
		if err != nil {
			r.logger.Error().Err(err).
				Str("table", r.tableName).
				Int("batch_size", len(chunk)).
				Msg("failed to batch write rule evaluation items")
			return fmt.Errorf("failed to batch write rule evaluation items: %w", err)
		}

		if len(output.UnprocessedItems) > 0 {
			r.logger.Warn().
				Str("table", r.tableName).
				Int("unprocessed_count", len(output.UnprocessedItems[r.tableName])).
				Msg("some rule evaluation items were not processed")
			return fmt.Errorf("unprocessed items remain after batch write")
		}
	}

	r.logger.Info().
		Str("table", r.tableName).
		Int("count", len(results)).
		Msg("rule evaluation results saved successfully")

	return nil
}

// FindByTransactionID queries rule evaluation results by transaction ID and returns them sorted by priority ascending.
func (r *DynamoDBRuleEvaluationRepository) FindByTransactionID(ctx context.Context, transactionID string) ([]entity.RuleEvaluationResult, error) {
	r.logger.Info().
		Str("table", r.tableName).
		Str("transaction_id", transactionID).
		Msg("querying rule evaluations by transaction ID")

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("transaction_id = :txnId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":txnId": &types.AttributeValueMemberS{Value: transactionID},
		},
	}

	output, err := r.client.Query(ctx, input)
	if err != nil {
		r.logger.Error().Err(err).
			Str("table", r.tableName).
			Str("transaction_id", transactionID).
			Msg("failed to query rule evaluations")
		return nil, fmt.Errorf("failed to query rule evaluations: %w", err)
	}

	var items []ruleEvaluationItem
	if err := attributevalue.UnmarshalListOfMaps(output.Items, &items); err != nil {
		r.logger.Error().Err(err).
			Str("table", r.tableName).
			Int("item_count", len(output.Items)).
			Msg("failed to unmarshal rule evaluation items")
		return nil, fmt.Errorf("failed to unmarshal rule evaluation items: %w", err)
	}

	results := make([]entity.RuleEvaluationResult, len(items))
	for i, item := range items {
		results[i] = toRuleEvaluationResult(item)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Priority < results[j].Priority
	})

	r.logger.Info().
		Str("table", r.tableName).
		Str("transaction_id", transactionID).
		Int("result_count", len(results)).
		Msg("rule evaluations retrieved")

	return results, nil
}

func toRuleEvaluationItem(r entity.RuleEvaluationResult) ruleEvaluationItem {
	return ruleEvaluationItem{
		TransactionID:     r.TransactionID,
		RuleID:            r.RuleID,
		RuleName:          r.RuleName,
		ConditionField:    r.ConditionField,
		ConditionOperator: r.ConditionOperator,
		ConditionValue:    r.ConditionValue,
		ActualFieldValue:  r.ActualFieldValue,
		Matched:           r.Matched,
		ResultStatus:      r.ResultStatus,
		EvaluatedAt:       r.EvaluatedAt.Format(time.RFC3339),
		Priority:          r.Priority,
	}
}

func toRuleEvaluationResult(item ruleEvaluationItem) entity.RuleEvaluationResult {
	evaluatedAt, _ := time.Parse(time.RFC3339, item.EvaluatedAt)

	return entity.RuleEvaluationResult{
		TransactionID:     item.TransactionID,
		RuleID:            item.RuleID,
		RuleName:          item.RuleName,
		ConditionField:    item.ConditionField,
		ConditionOperator: item.ConditionOperator,
		ConditionValue:    item.ConditionValue,
		ActualFieldValue:  item.ActualFieldValue,
		Matched:           item.Matched,
		ResultStatus:      item.ResultStatus,
		EvaluatedAt:       evaluatedAt,
		Priority:          item.Priority,
	}
}
