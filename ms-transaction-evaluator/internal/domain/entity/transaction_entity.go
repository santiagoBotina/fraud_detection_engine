package entity

type EvaluateTransactionRequest struct {
	AmountInCents int64         `json:"amount_in_cents" example:"10000"`
	Currency      Currency      `json:"currency" example:"USD"`
	PaymentMethod PaymentMethod `json:"payment_method" example:"CARD"`
	CustomerInfo  CustomerInfo  `json:"customer"`
}

type CustomerInfo struct {
	CustomerID string `json:"customer_id" example:"cust_123"`
	Name       string `json:"name" example:"John Doe"`
	Email      string `json:"email" example:"john@example.com"`
	Phone      string `json:"phone" example:"+1234567890"`
	IpAddress  string `json:"ip_address" example:"192.168.1.1"`
}

type Currency string

const (
	USD Currency = "USD"
	COP Currency = "COP"
	EUR Currency = "EUR"
)

type PaymentMethod string

const (
	CARD          PaymentMethod = "CARD"
	BANK_TRANSFER PaymentMethod = "BANK_TRANSFER"
	CRYPTO        PaymentMethod = "CRYPTO"
)

type TransactionEntity struct {
	ID                string        `json:"id"`
	AmountInCents     int64         `json:"amount_in_cents"`
	Currency          Currency      `json:"currency"`
	PaymentMethod     PaymentMethod `json:"payment_method"`
	CustomerID        string        `json:"customer_id"`
	CustomerName      string        `json:"customer_name"`
	CustomerEmail     string        `json:"customer_email"`
	CustomerPhone     string        `json:"customer_phone"`
	CustomerIPAddress string        `json:"customer_ip_address"`
}
