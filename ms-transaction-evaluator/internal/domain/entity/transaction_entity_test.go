package entity

import (
	"encoding/json"
	"testing"
)

func TestEvaluateTransactionRequest_JSONMarshaling(t *testing.T) {
	t.Run("should marshal to JSON correctly", func(t *testing.T) {
		req := EvaluateTransactionRequest{
			AmountInCents: 10000,
			Currency:      USD,
			PaymentMethod: CARD,
			CustomerInfo: CustomerInfo{
				CustomerID: "cust_123",
				Name:       "John Doe",
				Email:      "john@example.com",
				Phone:      "+1234567890",
				IpAddress:  "192.168.1.1",
			},
		}

		jsonData, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		expected := `{"amount_in_cents":10000,"currency":"USD","payment_method":"CARD","customer":{"customer_id":"cust_123","name":"John Doe","email":"john@example.com","phone":"+1234567890","ip_address":"192.168.1.1"}}`
		if string(jsonData) != expected {
			t.Errorf("JSON marshaling failed.\nExpected: %s\nGot: %s", expected, string(jsonData))
		}
	})

	t.Run("should unmarshal from JSON correctly", func(t *testing.T) {
		jsonData := `{
			"amount_in_cents": 15000,
			"currency": "EUR",
			"payment_method": "BANK_TRANSFER",
			"customer": {
				"customer_id": "cust_456",
				"name": "Jane Smith",
				"email": "jane@example.com",
				"phone": "+9876543210",
				"ip_address": "10.0.0.1"
			}
		}`

		var req EvaluateTransactionRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		if err != nil {
			t.Fatalf("Failed to unmarshal request: %v", err)
		}

		if req.AmountInCents != 15000 {
			t.Errorf("Expected AmountInCents to be 15000, got %d", req.AmountInCents)
		}
		if req.Currency != EUR {
			t.Errorf("Expected Currency to be EUR, got %s", req.Currency)
		}
		if req.PaymentMethod != BANK_TRANSFER {
			t.Errorf("Expected PaymentMethod to be BANK_TRANSFER, got %s", req.PaymentMethod)
		}
		if req.CustomerInfo.CustomerID != "cust_456" {
			t.Errorf("Expected CustomerID to be cust_456, got %s", req.CustomerInfo.CustomerID)
		}
		if req.CustomerInfo.Name != "Jane Smith" {
			t.Errorf("Expected Name to be Jane Smith, got %s", req.CustomerInfo.Name)
		}
		if req.CustomerInfo.Email != "jane@example.com" {
			t.Errorf("Expected Email to be jane@example.com, got %s", req.CustomerInfo.Email)
		}
		if req.CustomerInfo.Phone != "+9876543210" {
			t.Errorf("Expected Phone to be +9876543210, got %s", req.CustomerInfo.Phone)
		}
		if req.CustomerInfo.IpAddress != "10.0.0.1" {
			t.Errorf("Expected IpAddress to be 10.0.0.1, got %s", req.CustomerInfo.IpAddress)
		}
	})

	t.Run("should handle zero values", func(t *testing.T) {
		req := EvaluateTransactionRequest{}
		jsonData, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal empty request: %v", err)
		}

		var unmarshaled EvaluateTransactionRequest
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal empty request: %v", err)
		}

		if unmarshaled.AmountInCents != 0 {
			t.Errorf("Expected AmountInCents to be 0, got %d", unmarshaled.AmountInCents)
		}
	})
}

func TestCustomerInfo_JSONMarshaling(t *testing.T) {
	t.Run("should marshal to JSON correctly", func(t *testing.T) {
		customer := CustomerInfo{
			CustomerID: "cust_789",
			Name:       "Alice Johnson",
			Email:      "alice@example.com",
			Phone:      "+1122334455",
			IpAddress:  "172.16.0.1",
		}

		jsonData, err := json.Marshal(customer)
		if err != nil {
			t.Fatalf("Failed to marshal customer: %v", err)
		}

		expected := `{"customer_id":"cust_789","name":"Alice Johnson","email":"alice@example.com","phone":"+1122334455","ip_address":"172.16.0.1"}`
		if string(jsonData) != expected {
			t.Errorf("JSON marshaling failed.\nExpected: %s\nGot: %s", expected, string(jsonData))
		}
	})

	t.Run("should unmarshal from JSON correctly", func(t *testing.T) {
		jsonData := `{
			"customer_id": "cust_999",
			"name": "Bob Wilson",
			"email": "bob@example.com",
			"phone": "+5544332211",
			"ip_address": "192.168.100.50"
		}`

		var customer CustomerInfo
		err := json.Unmarshal([]byte(jsonData), &customer)
		if err != nil {
			t.Fatalf("Failed to unmarshal customer: %v", err)
		}

		if customer.CustomerID != "cust_999" {
			t.Errorf("Expected CustomerID to be cust_999, got %s", customer.CustomerID)
		}
		if customer.Name != "Bob Wilson" {
			t.Errorf("Expected Name to be Bob Wilson, got %s", customer.Name)
		}
		if customer.Email != "bob@example.com" {
			t.Errorf("Expected Email to be bob@example.com, got %s", customer.Email)
		}
		if customer.Phone != "+5544332211" {
			t.Errorf("Expected Phone to be +5544332211, got %s", customer.Phone)
		}
		if customer.IpAddress != "192.168.100.50" {
			t.Errorf("Expected IpAddress to be 192.168.100.50, got %s", customer.IpAddress)
		}
	})
}

func TestCurrency_Constants(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		expected string
	}{
		{"USD constant", USD, "USD"},
		{"COP constant", COP, "COP"},
		{"EUR constant", EUR, "EUR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.currency) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.currency))
			}
		})
	}
}

func TestCurrency_JSONMarshaling(t *testing.T) {
	t.Run("should marshal currency to JSON string", func(t *testing.T) {
		currencies := []Currency{USD, COP, EUR}
		expected := []string{`"USD"`, `"COP"`, `"EUR"`}

		for i, currency := range currencies {
			jsonData, err := json.Marshal(currency)
			if err != nil {
				t.Fatalf("Failed to marshal currency %s: %v", currency, err)
			}
			if string(jsonData) != expected[i] {
				t.Errorf("Expected %s, got %s", expected[i], string(jsonData))
			}
		}
	})

	t.Run("should unmarshal currency from JSON string", func(t *testing.T) {
		jsonStrings := []string{`"USD"`, `"COP"`, `"EUR"`}
		expected := []Currency{USD, COP, EUR}

		for i, jsonStr := range jsonStrings {
			var currency Currency
			err := json.Unmarshal([]byte(jsonStr), &currency)
			if err != nil {
				t.Fatalf("Failed to unmarshal currency from %s: %v", jsonStr, err)
			}
			if currency != expected[i] {
				t.Errorf("Expected %s, got %s", expected[i], currency)
			}
		}
	})
}

func TestPaymentMethod_Constants(t *testing.T) {
	tests := []struct {
		name          string
		paymentMethod PaymentMethod
		expected      string
	}{
		{"CARD constant", CARD, "CARD"},
		{"BANK_TRANSFER constant", BANK_TRANSFER, "BANK_TRANSFER"},
		{"CRYPTO constant", CRYPTO, "CRYPTO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.paymentMethod) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.paymentMethod))
			}
		})
	}
}

func TestPaymentMethod_JSONMarshaling(t *testing.T) {
	t.Run("should marshal payment method to JSON string", func(t *testing.T) {
		methods := []PaymentMethod{CARD, BANK_TRANSFER, CRYPTO}
		expected := []string{`"CARD"`, `"BANK_TRANSFER"`, `"CRYPTO"`}

		for i, method := range methods {
			jsonData, err := json.Marshal(method)
			if err != nil {
				t.Fatalf("Failed to marshal payment method %s: %v", method, err)
			}
			if string(jsonData) != expected[i] {
				t.Errorf("Expected %s, got %s", expected[i], string(jsonData))
			}
		}
	})

	t.Run("should unmarshal payment method from JSON string", func(t *testing.T) {
		jsonStrings := []string{`"CARD"`, `"BANK_TRANSFER"`, `"CRYPTO"`}
		expected := []PaymentMethod{CARD, BANK_TRANSFER, CRYPTO}

		for i, jsonStr := range jsonStrings {
			var method PaymentMethod
			err := json.Unmarshal([]byte(jsonStr), &method)
			if err != nil {
				t.Fatalf("Failed to unmarshal payment method from %s: %v", jsonStr, err)
			}
			if method != expected[i] {
				t.Errorf("Expected %s, got %s", expected[i], method)
			}
		}
	})
}

func TestTransactionEntity_JSONMarshaling(t *testing.T) {
	t.Run("should marshal to JSON correctly", func(t *testing.T) {
		entity := TransactionEntity{
			ID:                "txn_12345",
			AmountInCents:     25000,
			Currency:          COP,
			PaymentMethod:     CRYPTO,
			CustomerID:        "cust_abc",
			CustomerName:      "Carlos Rodriguez",
			CustomerEmail:     "carlos@example.com",
			CustomerPhone:     "+573001234567",
			CustomerIPAddress: "200.50.100.25",
		}

		jsonData, err := json.Marshal(entity)
		if err != nil {
			t.Fatalf("Failed to marshal transaction entity: %v", err)
		}

		expected := `{"id":"txn_12345","amount_in_cents":25000,"currency":"COP","payment_method":"CRYPTO","customer_id":"cust_abc","customer_name":"Carlos Rodriguez","customer_email":"carlos@example.com","customer_phone":"+573001234567","customer_ip_address":"200.50.100.25"}`
		if string(jsonData) != expected {
			t.Errorf("JSON marshaling failed.\nExpected: %s\nGot: %s", expected, string(jsonData))
		}
	})

	t.Run("should unmarshal from JSON correctly", func(t *testing.T) {
		jsonData := `{
			"id": "txn_67890",
			"amount_in_cents": 50000,
			"currency": "USD",
			"payment_method": "CARD",
			"customer_id": "cust_xyz",
			"customer_name": "Maria Garcia",
			"customer_email": "maria@example.com",
			"customer_phone": "+14155551234",
			"customer_ip_address": "198.51.100.42"
		}`

		var entity TransactionEntity
		err := json.Unmarshal([]byte(jsonData), &entity)
		if err != nil {
			t.Fatalf("Failed to unmarshal transaction entity: %v", err)
		}

		if entity.ID != "txn_67890" {
			t.Errorf("Expected ID to be txn_67890, got %s", entity.ID)
		}
		if entity.AmountInCents != 50000 {
			t.Errorf("Expected AmountInCents to be 50000, got %d", entity.AmountInCents)
		}
		if entity.Currency != USD {
			t.Errorf("Expected Currency to be USD, got %s", entity.Currency)
		}
		if entity.PaymentMethod != CARD {
			t.Errorf("Expected PaymentMethod to be CARD, got %s", entity.PaymentMethod)
		}
		if entity.CustomerID != "cust_xyz" {
			t.Errorf("Expected CustomerID to be cust_xyz, got %s", entity.CustomerID)
		}
		if entity.CustomerName != "Maria Garcia" {
			t.Errorf("Expected CustomerName to be Maria Garcia, got %s", entity.CustomerName)
		}
		if entity.CustomerEmail != "maria@example.com" {
			t.Errorf("Expected CustomerEmail to be maria@example.com, got %s", entity.CustomerEmail)
		}
		if entity.CustomerPhone != "+14155551234" {
			t.Errorf("Expected CustomerPhone to be +14155551234, got %s", entity.CustomerPhone)
		}
		if entity.CustomerIPAddress != "198.51.100.42" {
			t.Errorf("Expected CustomerIPAddress to be 198.51.100.42, got %s", entity.CustomerIPAddress)
		}
	})

	t.Run("should handle zero values", func(t *testing.T) {
		entity := TransactionEntity{}
		jsonData, err := json.Marshal(entity)
		if err != nil {
			t.Fatalf("Failed to marshal empty transaction entity: %v", err)
		}

		var unmarshaled TransactionEntity
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal empty transaction entity: %v", err)
		}

		if unmarshaled.ID != "" {
			t.Errorf("Expected ID to be empty, got %s", unmarshaled.ID)
		}
		if unmarshaled.AmountInCents != 0 {
			t.Errorf("Expected AmountInCents to be 0, got %d", unmarshaled.AmountInCents)
		}
	})
}

func TestEvaluateTransactionRequest_AllPaymentMethodsAndCurrencies(t *testing.T) {
	t.Run("should work with all combinations of currency and payment method", func(t *testing.T) {
		currencies := []Currency{USD, COP, EUR}
		methods := []PaymentMethod{CARD, BANK_TRANSFER, CRYPTO}

		for _, currency := range currencies {
			for _, method := range methods {
				req := EvaluateTransactionRequest{
					AmountInCents: 1000,
					Currency:      currency,
					PaymentMethod: method,
					CustomerInfo: CustomerInfo{
						CustomerID: "test_customer",
						Name:       "Test User",
						Email:      "test@example.com",
						Phone:      "+1234567890",
						IpAddress:  "127.0.0.1",
					},
				}

				jsonData, err := json.Marshal(req)
				if err != nil {
					t.Fatalf("Failed to marshal request with %s and %s: %v", currency, method, err)
				}

				var unmarshaled EvaluateTransactionRequest
				err = json.Unmarshal(jsonData, &unmarshaled)
				if err != nil {
					t.Fatalf("Failed to unmarshal request with %s and %s: %v", currency, method, err)
				}

				if unmarshaled.Currency != currency {
					t.Errorf("Expected currency %s, got %s", currency, unmarshaled.Currency)
				}
				if unmarshaled.PaymentMethod != method {
					t.Errorf("Expected payment method %s, got %s", method, unmarshaled.PaymentMethod)
				}
			}
		}
	})
}
