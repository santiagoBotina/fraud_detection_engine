package usecase

import "errors"

var (
	ErrRuleRetrievalFailed   = errors.New("failed to retrieve rules")
	ErrDecisionPublishFailed = errors.New("failed to publish decision result")
	ErrTransactionNil        = errors.New("transaction is nil")
)
