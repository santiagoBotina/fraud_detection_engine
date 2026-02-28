package usecase

import (
	"testing"

	"ms-transaction-evaluator/internal/domain/entity"
)

func TestValidateCreateTransactionPayloadUseCase_Execute(t *testing.T) {
	uc := NewValidateCreateTransactionPayloadUseCase()

	t.Run("should return error when request is nil", func(t *testing.T) {
		err := uc.Execute(nil)
		if err == nil {
			t.Error("Expected error when request is nil, got nil")
		}
		if err.Error() != "request is nil" {
			t.Errorf("Expected 'request is nil' error, got: %v", err)
		}
	})

	t.Run("should return error when amount is zero", func(t *testing.T) {
		req := &entity.EvaluateTransactionRequest{
			AmountInCents: 0,
			Currency:      entity.USD,
			PaymentMethod: entity.CARD,
			CustomerInfo:  createValidCustomerInfo(),
		}
		err := uc.Execute(req)
		if err != ErrAmountRequired {
			t.Errorf("Expected ErrAmountRequired, got: %v", err)
		}
	})

	t.Run("should return error when amount is negative", func(t *testing.T) {
		req := &entity.EvaluateTransactionRequest{
			AmountInCents: -100,
			Currency:      entity.USD,
			PaymentMethod: entity.CARD,
			CustomerInfo:  createValidCustomerInfo(),
		}
		err := uc.Execute(req)
		if err != ErrAmountMustBePositive {
			t.Errorf("Expected ErrAmountMustBePositive, got: %v", err)
		}
	})

	t.Run("should return error when currency is empty", func(t *testing.T) {
		req := &entity.EvaluateTransactionRequest{
			AmountInCents: 1000,
			Currency:      "",
			PaymentMethod: entity.CARD,
			CustomerInfo:  createValidCustomerInfo(),
		}
		err := uc.Execute(req)
		if err != ErrCurrencyRequired {
			t.Errorf("Expected ErrCurrencyRequired, got: %v", err)
		}
	})

	t.Run("should return error when currency is invalid", func(t *testing.T) {
		req := &entity.EvaluateTransactionRequest{
			AmountInCents: 1000,
			Currency:      "INVALID",
			PaymentMethod: entity.CARD,
			CustomerInfo:  createValidCustomerInfo(),
		}
		err := uc.Execute(req)
		if err != ErrCurrencyInvalid {
			t.Errorf("Expected ErrCurrencyInvalid, got: %v", err)
		}
	})

	t.Run("should accept valid currency USD", func(t *testing.T) {
		req := createValidRequest()
		req.Currency = entity.USD
		err := uc.Execute(req)
		if err != nil {
			t.Errorf("Expected no error for valid USD currency, got: %v", err)
		}
	})

	t.Run("should accept valid currency COP", func(t *testing.T) {
		req := createValidRequest()
		req.Currency = entity.COP
		err := uc.Execute(req)
		if err != nil {
			t.Errorf("Expected no error for valid COP currency, got: %v", err)
		}
	})

	t.Run("should accept valid currency EUR", func(t *testing.T) {
		req := createValidRequest()
		req.Currency = entity.EUR
		err := uc.Execute(req)
		if err != nil {
			t.Errorf("Expected no error for valid EUR currency, got: %v", err)
		}
	})

	t.Run("should return error when payment method is empty", func(t *testing.T) {
		req := &entity.EvaluateTransactionRequest{
			AmountInCents: 1000,
			Currency:      entity.USD,
			PaymentMethod: "",
			CustomerInfo:  createValidCustomerInfo(),
		}
		err := uc.Execute(req)
		if err != ErrPaymentMethodRequired {
			t.Errorf("Expected ErrPaymentMethodRequired, got: %v", err)
		}
	})

	t.Run("should return error when payment method is invalid", func(t *testing.T) {
		req := &entity.EvaluateTransactionRequest{
			AmountInCents: 1000,
			Currency:      entity.USD,
			PaymentMethod: "INVALID",
			CustomerInfo:  createValidCustomerInfo(),
		}
		err := uc.Execute(req)
		if err != ErrPaymentMethodInvalid {
			t.Errorf("Expected ErrPaymentMethodInvalid, got: %v", err)
		}
	})

	t.Run("should accept valid payment method CARD", func(t *testing.T) {
		req := createValidRequest()
		req.PaymentMethod = entity.CARD
		err := uc.Execute(req)
		if err != nil {
			t.Errorf("Expected no error for valid CARD payment method, got: %v", err)
		}
	})

	t.Run("should accept valid payment method BANK_TRANSFER", func(t *testing.T) {
		req := createValidRequest()
		req.PaymentMethod = entity.BANK_TRANSFER
		err := uc.Execute(req)
		if err != nil {
			t.Errorf("Expected no error for valid BANK_TRANSFER payment method, got: %v", err)
		}
	})

	t.Run("should accept valid payment method CRYPTO", func(t *testing.T) {
		req := createValidRequest()
		req.PaymentMethod = entity.CRYPTO
		err := uc.Execute(req)
		if err != nil {
			t.Errorf("Expected no error for valid CRYPTO payment method, got: %v", err)
		}
	})

	t.Run("should return error when customer ID is empty", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.CustomerID = ""
		err := uc.Execute(req)
		if err != ErrCustomerIDRequired {
			t.Errorf("Expected ErrCustomerIDRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer ID is only whitespace", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.CustomerID = "   "
		err := uc.Execute(req)
		if err != ErrCustomerIDRequired {
			t.Errorf("Expected ErrCustomerIDRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer name is empty", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Name = ""
		err := uc.Execute(req)
		if err != ErrCustomerNameRequired {
			t.Errorf("Expected ErrCustomerNameRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer name is only whitespace", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Name = "   "
		err := uc.Execute(req)
		if err != ErrCustomerNameRequired {
			t.Errorf("Expected ErrCustomerNameRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer email is empty", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Email = ""
		err := uc.Execute(req)
		if err != ErrCustomerEmailRequired {
			t.Errorf("Expected ErrCustomerEmailRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer email is only whitespace", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Email = "   "
		err := uc.Execute(req)
		if err != ErrCustomerEmailRequired {
			t.Errorf("Expected ErrCustomerEmailRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer email is invalid - no @", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Email = "invalidemail.com"
		err := uc.Execute(req)
		if err != ErrCustomerEmailInvalid {
			t.Errorf("Expected ErrCustomerEmailInvalid, got: %v", err)
		}
	})

	t.Run("should return error when customer email is invalid - no domain", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Email = "invalid@"
		err := uc.Execute(req)
		if err != ErrCustomerEmailInvalid {
			t.Errorf("Expected ErrCustomerEmailInvalid, got: %v", err)
		}
	})

	t.Run("should return error when customer email is invalid - no TLD", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Email = "invalid@domain"
		err := uc.Execute(req)
		if err != ErrCustomerEmailInvalid {
			t.Errorf("Expected ErrCustomerEmailInvalid, got: %v", err)
		}
	})

	t.Run("should accept valid email formats", func(t *testing.T) {
		validEmails := []string{
			"user@example.com",
			"test.user@example.com",
			"user+tag@example.co.uk",
			"user_name@example-domain.com",
			"123@example.com",
		}

		for _, email := range validEmails {
			req := createValidRequest()
			req.CustomerInfo.Email = email
			err := uc.Execute(req)
			if err != nil {
				t.Errorf("Expected no error for valid email '%s', got: %v", email, err)
			}
		}
	})

	t.Run("should return error when customer phone is empty", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Phone = ""
		err := uc.Execute(req)
		if err != ErrCustomerPhoneRequired {
			t.Errorf("Expected ErrCustomerPhoneRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer phone is only whitespace", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.Phone = "   "
		err := uc.Execute(req)
		if err != ErrCustomerPhoneRequired {
			t.Errorf("Expected ErrCustomerPhoneRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer IP address is empty", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.IpAddress = ""
		err := uc.Execute(req)
		if err != ErrCustomerIPRequired {
			t.Errorf("Expected ErrCustomerIPRequired, got: %v", err)
		}
	})

	t.Run("should return error when customer IP address is only whitespace", func(t *testing.T) {
		req := createValidRequest()
		req.CustomerInfo.IpAddress = "   "
		err := uc.Execute(req)
		if err != ErrCustomerIPRequired {
			t.Errorf("Expected ErrCustomerIPRequired, got: %v", err)
		}
	})

	t.Run("should validate successfully with all valid fields", func(t *testing.T) {
		req := createValidRequest()
		err := uc.Execute(req)
		if err != nil {
			t.Errorf("Expected no error for valid request, got: %v", err)
		}
	})

	t.Run("should validate successfully with positive amount", func(t *testing.T) {
		req := createValidRequest()
		req.AmountInCents = 999999
		err := uc.Execute(req)
		if err != nil {
			t.Errorf("Expected no error for valid positive amount, got: %v", err)
		}
	})
}

func TestValidateCustomerInfo(t *testing.T) {
	t.Run("should return error when customer info is nil", func(t *testing.T) {
		err := validateCustomerInfo(nil)
		if err != ErrCustomerRequired {
			t.Errorf("Expected ErrCustomerRequired, got: %v", err)
		}
	})

	t.Run("should validate successfully with all valid fields", func(t *testing.T) {
		customer := createValidCustomerInfo()
		err := validateCustomerInfo(&customer)
		if err != nil {
			t.Errorf("Expected no error for valid customer info, got: %v", err)
		}
	})
}

func TestIsValidCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency entity.Currency
		expected bool
	}{
		{"USD is valid", entity.USD, true},
		{"COP is valid", entity.COP, true},
		{"EUR is valid", entity.EUR, true},
		{"invalid currency", "GBP", false},
		{"empty currency", "", false},
		{"random string", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCurrency(tt.currency)
			if result != tt.expected {
				t.Errorf("isValidCurrency(%s) = %v, expected %v", tt.currency, result, tt.expected)
			}
		})
	}
}

func TestIsValidPaymentMethod(t *testing.T) {
	tests := []struct {
		name          string
		paymentMethod entity.PaymentMethod
		expected      bool
	}{
		{"CARD is valid", entity.CARD, true},
		{"BANK_TRANSFER is valid", entity.BANK_TRANSFER, true},
		{"CRYPTO is valid", entity.CRYPTO, true},
		{"invalid payment method", "PAYPAL", false},
		{"empty payment method", "", false},
		{"random string", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPaymentMethod(tt.paymentMethod)
			if result != tt.expected {
				t.Errorf("isValidPaymentMethod(%s) = %v, expected %v", tt.paymentMethod, result, tt.expected)
			}
		})
	}
}

func createValidRequest() *entity.EvaluateTransactionRequest {
	return &entity.EvaluateTransactionRequest{
		AmountInCents: 10000,
		Currency:      entity.USD,
		PaymentMethod: entity.CARD,
		CustomerInfo:  createValidCustomerInfo(),
	}
}

func createValidCustomerInfo() entity.CustomerInfo {
	return entity.CustomerInfo{
		CustomerID: "cust_12345",
		Name:       "John Doe",
		Email:      "john.doe@example.com",
		Phone:      "+1234567890",
		IpAddress:  "192.168.1.1",
	}
}
