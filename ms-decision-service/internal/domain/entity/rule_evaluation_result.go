package entity

import "time"

// RuleEvaluationResult represents the outcome of evaluating a single rule against a transaction.
type RuleEvaluationResult struct {
	TransactionID     string    `json:"transaction_id"`
	RuleID            string    `json:"rule_id"`
	RuleName          string    `json:"rule_name"`
	ConditionField    string    `json:"condition_field"`
	ConditionOperator string    `json:"condition_operator"`
	ConditionValue    string    `json:"condition_value"`
	ActualFieldValue  string    `json:"actual_field_value"`
	Matched           bool      `json:"matched"`
	ResultStatus      string    `json:"result_status"`
	EvaluatedAt       time.Time `json:"evaluated_at"`
	Priority          int       `json:"priority"`
}
