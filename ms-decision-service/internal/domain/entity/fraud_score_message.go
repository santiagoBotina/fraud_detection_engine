package entity

import "time"

// FraudScoreCalculatedMessage represents the payload consumed from the FraudScore.Calculated Kafka topic.
type FraudScoreCalculatedMessage struct {
	TransactionID string    `json:"transaction_id"`
	FraudScore    int       `json:"fraud_score"`
	CalculatedAt  time.Time `json:"calculated_at"`
}
