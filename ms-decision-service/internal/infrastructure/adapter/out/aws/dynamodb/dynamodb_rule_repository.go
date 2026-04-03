package dynamodb

import (
	"context"
	"fmt"
	"ms-decision-service/internal/domain/entity"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

type ruleItem struct {
	RuleID            string `dynamodbav:"rule_id"`
	RuleName          string `dynamodbav:"rule_name"`
	ConditionField    string `dynamodbav:"condition_field"`
	ConditionOperator string `dynamodbav:"condition_operator"`
	ConditionValue    string `dynamodbav:"condition_value"`
	ResultStatus      string `dynamodbav:"result_status"`
	Priority          int    `dynamodbav:"priority"`
	IsActive          bool   `dynamodbav:"is_active"`
}

// DynamoDBRuleRepository implements repository.RuleRepository using AWS DynamoDB.
type DynamoDBRuleRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    zerolog.Logger
}

// NewDynamoDBRuleRepository creates a new DynamoDB-backed rule repository.
func NewDynamoDBRuleRepository(
	client *dynamodb.Client,
	tableName string,
	logger zerolog.Logger,
) *DynamoDBRuleRepository {
	return &DynamoDBRuleRepository{client: client, tableName: tableName, logger: logger}
}

// FindActiveRulesSortedByPriority scans the rules table for active rules and sorts by priority ascending.
func (r *DynamoDBRuleRepository) FindActiveRulesSortedByPriority(ctx context.Context) ([]entity.Rule, error) {
	r.logger.Info().Str("table", r.tableName).Msg("scanning rules table for active rules")

	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("is_active = :active"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":active": &types.AttributeValueMemberBOOL{Value: true},
		},
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		r.logger.Error().Err(err).Str("table", r.tableName).Msg("failed to scan rules table")
		return nil, fmt.Errorf("failed to scan rules table: %w", err)
	}

	var items []ruleItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		r.logger.Error().Err(err).Str("table", r.tableName).
			Int("item_count", len(result.Items)).Msg("failed to unmarshal rules")
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	rules := make([]entity.Rule, len(items))
	for i, item := range items {
		rules[i] = entity.Rule{
			RuleID:            item.RuleID,
			RuleName:          item.RuleName,
			ConditionField:    entity.ConditionField(item.ConditionField),
			ConditionOperator: entity.ConditionOperator(item.ConditionOperator),
			ConditionValue:    item.ConditionValue,
			ResultStatus:      entity.DecisionStatus(item.ResultStatus),
			Priority:          item.Priority,
			IsActive:          item.IsActive,
		}
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	r.logger.Info().Str("table", r.tableName).Int("active_count", len(rules)).Msg("active rules loaded")

	return rules, nil
}

// FindAll scans the rules table for all rules (active and inactive) and sorts by priority ascending.
func (r *DynamoDBRuleRepository) FindAll(ctx context.Context) ([]entity.Rule, error) {
	r.logger.Info().Str("table", r.tableName).Msg("scanning rules table for all rules")

	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		r.logger.Error().Err(err).Str("table", r.tableName).Msg("failed to scan rules table")
		return nil, fmt.Errorf("failed to scan rules table: %w", err)
	}

	var items []ruleItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		r.logger.Error().Err(err).Str("table", r.tableName).
			Int("item_count", len(result.Items)).Msg("failed to unmarshal rules")
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	rules := make([]entity.Rule, len(items))
	for i, item := range items {
		rules[i] = entity.Rule{
			RuleID:            item.RuleID,
			RuleName:          item.RuleName,
			ConditionField:    entity.ConditionField(item.ConditionField),
			ConditionOperator: entity.ConditionOperator(item.ConditionOperator),
			ConditionValue:    item.ConditionValue,
			ResultStatus:      entity.DecisionStatus(item.ResultStatus),
			Priority:          item.Priority,
			IsActive:          item.IsActive,
		}
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	r.logger.Info().Str("table", r.tableName).Int("total_count", len(rules)).Msg("all rules loaded")

	return rules, nil
}

// FilterAndSortActiveRules filters rules to only active ones and sorts by priority ascending.
// This is exported for testing purposes.
func FilterAndSortActiveRules(rules []entity.Rule) []entity.Rule {
	var active []entity.Rule
	for _, r := range rules {
		if r.IsActive {
			active = append(active, r)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].Priority < active[j].Priority
	})
	return active
}
