package entity

// DecisionResult represents the outcome of evaluating a transaction against the rules engine.
type DecisionResult struct {
	TransactionID string         `json:"transaction_id"`
	Status        DecisionStatus `json:"status"`
}
