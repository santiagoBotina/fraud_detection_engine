package usecase

import "errors"

var (
	ErrRuleRetrievalFailed        = errors.New("failed to retrieve rules")
	ErrDecisionPublishFailed      = errors.New("failed to publish decision result")
	ErrFraudScorePublishFailed    = errors.New("failed to publish fraud score request")
	ErrTransactionNil             = errors.New("transaction is nil")
	ErrFraudScoreMessageNil       = errors.New("fraud score message is nil")
	ErrTransactionIDEmpty         = errors.New("transaction ID is empty")
	ErrEvaluationRetrievalFailed  = errors.New("failed to retrieve rule evaluations")
)
