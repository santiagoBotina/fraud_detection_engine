package entity

import (
	"fmt"
	"time"
)

// TransactionMessage represents the JSON payload consumed from the Transaction.Created Kafka topic.
type TransactionMessage struct {
	ID                string    `json:"id"`
	AmountInCents     int64     `json:"amount_in_cents"`
	Currency          string    `json:"currency"`
	PaymentMethod     string    `json:"payment_method"`
	CustomerID        string    `json:"customer_id"`
	CustomerName      string    `json:"customer_name"`
	CustomerEmail     string    `json:"customer_email"`
	CustomerPhone     string    `json:"customer_phone"`
	CustomerIPAddress string    `json:"customer_ip_address"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// GetFieldValue returns the string representation of the transaction field
// identified by the given ConditionField. Returns empty string for unknown fields.
func (t *TransactionMessage) GetFieldValue(field ConditionField) string {
	switch field {
	case FieldAmountInCents:
		return fmt.Sprintf("%d", t.AmountInCents)
	case FieldCurrency:
		return t.Currency
	case FieldPaymentMethod:
		return t.PaymentMethod
	case FieldCustomerID:
		return t.CustomerID
	case FieldCustomerIPAddress:
		return t.CustomerIPAddress
	default:
		return ""
	}
}
