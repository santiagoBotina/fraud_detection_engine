package usecase

import (
	"errors"
	"regexp"
	"strings"

	"ms-transaction-evaluator/internal/domain/entity"
)

var (
	ErrAmountRequired        = errors.New("amount_in_cents is required")
	ErrAmountMustBePositive  = errors.New("amount_in_cents must be positive")
	ErrCurrencyRequired      = errors.New("currency is required")
	ErrCurrencyInvalid       = errors.New("currency is invalid")
	ErrPaymentMethodRequired = errors.New("payment_method is required")
	ErrPaymentMethodInvalid  = errors.New("payment_method is invalid")
	ErrCustomerRequired      = errors.New("customer is required")
	ErrCustomerIDRequired    = errors.New("customer_id is required")
	ErrCustomerNameRequired  = errors.New("customer name is required")
	ErrCustomerEmailRequired = errors.New("customer email is required")
	ErrCustomerEmailInvalid  = errors.New("customer email is invalid")
	ErrCustomerPhoneRequired = errors.New("customer phone is required")
	ErrCustomerIPRequired    = errors.New("customer ip_address is required")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type ValidateCreateTransactionPayloadUseCase struct{}

func NewValidateCreateTransactionPayloadUseCase() *ValidateCreateTransactionPayloadUseCase {
	return &ValidateCreateTransactionPayloadUseCase{}
}

func (uc *ValidateCreateTransactionPayloadUseCase) Execute(req *entity.EvaluateTransactionRequest) error {
	if req == nil {
		return errors.New("request is nil")
	}

	if req.AmountInCents == 0 {
		return ErrAmountRequired
	}
	if req.AmountInCents < 0 {
		return ErrAmountMustBePositive
	}

	if req.Currency == "" {
		return ErrCurrencyRequired
	}
	if !isValidCurrency(req.Currency) {
		return ErrCurrencyInvalid
	}

	if req.PaymentMethod == "" {
		return ErrPaymentMethodRequired
	}
	if !isValidPaymentMethod(req.PaymentMethod) {
		return ErrPaymentMethodInvalid
	}

	if err := validateCustomerInfo(&req.CustomerInfo); err != nil {
		return err
	}

	return nil
}

func validateCustomerInfo(customer *entity.CustomerInfo) error {
	if customer == nil {
		return ErrCustomerRequired
	}

	if strings.TrimSpace(customer.CustomerID) == "" {
		return ErrCustomerIDRequired
	}

	if strings.TrimSpace(customer.Name) == "" {
		return ErrCustomerNameRequired
	}

	if strings.TrimSpace(customer.Email) == "" {
		return ErrCustomerEmailRequired
	}
	if !emailRegex.MatchString(customer.Email) {
		return ErrCustomerEmailInvalid
	}

	if strings.TrimSpace(customer.Phone) == "" {
		return ErrCustomerPhoneRequired
	}

	if strings.TrimSpace(customer.IpAddress) == "" {
		return ErrCustomerIPRequired
	}

	return nil
}

func isValidCurrency(currency entity.Currency) bool {
	switch currency {
	case entity.USD, entity.COP, entity.EUR:
		return true
	default:
		return false
	}
}

func isValidPaymentMethod(method entity.PaymentMethod) bool {
	switch method {
	case entity.CARD, entity.BANK_TRANSFER, entity.CRYPTO:
		return true
	default:
		return false
	}
}
