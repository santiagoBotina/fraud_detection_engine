package usecase

import (
	"context"
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// Feature: transaction-finalization-latency, Property 2: New transactions have no finalized_at
// Validates: Requirements 1.4

// saveCaptureMockRepo captures the transaction entity passed to Save.
type saveCaptureMockRepo struct {
	captured *entity.TransactionEntity
}

func (m *saveCaptureMockRepo) Save(_ context.Context, txn *entity.TransactionEntity) error {
	m.captured = txn
	return nil
}

func (m *saveCaptureMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus, _ *time.Time) error {
	return nil
}

func (m *saveCaptureMockRepo) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return nil, nil
}

func (m *saveCaptureMockRepo) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func (m *saveCaptureMockRepo) FindAll(_ context.Context) ([]entity.TransactionEntity, error) {
	return nil, nil
}

// noopEventPublisher is a mock event publisher that always succeeds.
type noopEventPublisher struct{}

func (m *noopEventPublisher) Publish(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func TestProperty_NewTransactionsHaveNoFinalizedAt(t *testing.T) {
	currencies := []entity.Currency{entity.USD, entity.COP, entity.EUR}
	paymentMethods := []entity.PaymentMethod{entity.CARD, entity.BANK_TRANSFER, entity.CRYPTO}

	rapid.Check(t, func(t *rapid.T) {
		mock := &saveCaptureMockRepo{}
		pub := &noopEventPublisher{}
		uc := NewSaveTransactionUseCase(mock, pub)

		currency := currencies[rapid.IntRange(0, len(currencies)-1).Draw(t, "currencyIdx")]
		paymentMethod := paymentMethods[rapid.IntRange(0, len(paymentMethods)-1).Draw(t, "paymentMethodIdx")]

		req := &entity.EvaluateTransactionRequest{
			AmountInCents: rapid.Int64Range(1, 999999999).Draw(t, "amountInCents"),
			Currency:      currency,
			PaymentMethod: paymentMethod,
			CustomerInfo: entity.CustomerInfo{
				CustomerID: rapid.StringMatching(`^cust_[a-z0-9]{4,12}$`).Draw(t, "customerID"),
				Name:       rapid.StringMatching(`^[A-Z][a-z]{2,10} [A-Z][a-z]{2,10}$`).Draw(t, "name"),
				Email:      rapid.StringMatching(`^[a-z]{3,8}@[a-z]{3,8}\.[a-z]{2,4}$`).Draw(t, "email"),
				Phone:      rapid.StringMatching(`^\+[0-9]{7,15}$`).Draw(t, "phone"),
				IpAddress:  rapid.StringMatching(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`).Draw(t, "ipAddress"),
			},
		}

		result, err := uc.Execute(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error from Execute: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.FinalizedAt != nil {
			t.Fatalf("expected FinalizedAt to be nil for new transaction, got %v", *result.FinalizedAt)
		}

		if result.Status != entity.PENDING {
			t.Fatalf("expected status PENDING for new transaction, got %s", result.Status)
		}
	})
}
