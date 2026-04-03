package entity

// DecisionCalculatedMessage represents the payload consumed from the Decision.Calculated Kafka topic.
type DecisionCalculatedMessage struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}
