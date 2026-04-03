package entity

import "strconv"

// ConditionField represents a transaction attribute that a rule can evaluate against.
type ConditionField string

const (
	FieldAmountInCents     ConditionField = "amount_in_cents"
	FieldCurrency          ConditionField = "currency"
	FieldPaymentMethod     ConditionField = "payment_method"
	FieldCustomerID        ConditionField = "customer_id"
	FieldCustomerIPAddress ConditionField = "customer_ip_address"
	FieldFraudScore        ConditionField = "fraud_score"
)

// ConditionOperator represents a comparison operator used in rule evaluation.
type ConditionOperator string

const (
	OpGreaterThan        ConditionOperator = "GREATER_THAN"
	OpLessThan           ConditionOperator = "LESS_THAN"
	OpEqual              ConditionOperator = "EQUAL"
	OpNotEqual           ConditionOperator = "NOT_EQUAL"
	OpGreaterThanOrEqual ConditionOperator = "GREATER_THAN_OR_EQUAL"
	OpLessThanOrEqual    ConditionOperator = "LESS_THAN_OR_EQUAL"
)

// DecisionStatus represents the outcome of a rule evaluation.
type DecisionStatus string

const (
	APPROVED   DecisionStatus = "APPROVED"
	DECLINED   DecisionStatus = "DECLINED"
	FRAUDCHECK DecisionStatus = "FRAUD_CHECK"
)

// Rule represents a single fraud detection rule stored in DynamoDB.
type Rule struct {
	RuleID            string            `json:"rule_id"`
	RuleName          string            `json:"rule_name"`
	ConditionField    ConditionField    `json:"condition_field"`
	ConditionOperator ConditionOperator `json:"condition_operator"`
	ConditionValue    string            `json:"condition_value"`
	ResultStatus      DecisionStatus    `json:"result_status"`
	Priority          int               `json:"priority"`
	IsActive          bool              `json:"is_active"`
}

// Compare evaluates fieldValue against conditionValue using the operator.
// For FieldAmountInCents, both values are parsed as int64 and compared numerically.
// For all other fields, only EQUAL and NOT_EQUAL are supported (string comparison).
func (op ConditionOperator) Compare(fieldValue, conditionValue string, field ConditionField) bool {
	if field == FieldAmountInCents || field == FieldFraudScore {
		return op.compareNumeric(fieldValue, conditionValue)
	}

	return op.compareString(fieldValue, conditionValue)
}

func (op ConditionOperator) compareNumeric(fieldValue, conditionValue string) bool {
	fv, err := strconv.ParseInt(fieldValue, 10, 64)
	if err != nil {
		return false
	}

	cv, err := strconv.ParseInt(conditionValue, 10, 64)
	if err != nil {
		return false
	}

	switch op {
	case OpGreaterThan:
		return fv > cv
	case OpLessThan:
		return fv < cv
	case OpEqual:
		return fv == cv
	case OpNotEqual:
		return fv != cv
	case OpGreaterThanOrEqual:
		return fv >= cv
	case OpLessThanOrEqual:
		return fv <= cv
	default:
		return false
	}
}

func (op ConditionOperator) compareString(fieldValue, conditionValue string) bool {
	switch op {
	case OpEqual:
		return fieldValue == conditionValue
	case OpNotEqual:
		return fieldValue != conditionValue
	default:
		return false
	}
}

// Matches checks whether the given transaction satisfies this rule's condition.
func (r *Rule) Matches(transaction *TransactionMessage) bool {
	fieldValue := transaction.GetFieldValue(r.ConditionField)
	return r.ConditionOperator.Compare(fieldValue, r.ConditionValue, r.ConditionField)
}
