package usecase

import (
	"context"
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// Feature: fraud-analyst-dashboard, Property 4: Transaction retrieval round-trip
// Validates: Requirements 2.1

// roundTripMockRepo is a hand-written mock implementing TransactionRepository
// that stores transactions in a map for round-trip retrieval testing.
type roundTripMockRepo struct {
	store map[string]*entity.TransactionEntity
}

func newRoundTripMockRepo() *roundTripMockRepo {
	return &roundTripMockRepo{store: make(map[string]*entity.TransactionEntity)}
}

func (m *roundTripMockRepo) Save(_ context.Context, txn *entity.TransactionEntity) error {
	m.store[txn.ID] = txn
	return nil
}

func (m *roundTripMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus, _ *time.Time) error {
	return nil
}

func (m *roundTripMockRepo) FindByID(_ context.Context, id string) (*entity.TransactionEntity, error) {
	txn, ok := m.store[id]
	if !ok {
		return nil, nil
	}
	return txn, nil
}

func (m *roundTripMockRepo) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func (m *roundTripMockRepo) FindAll(_ context.Context) ([]entity.TransactionEntity, error) {
	return nil, nil
}

func TestProperty_TransactionRetrievalRoundTrip(t *testing.T) {
	currencies := []entity.Currency{entity.USD, entity.COP, entity.EUR}
	paymentMethods := []entity.PaymentMethod{entity.CARD, entity.BANK_TRANSFER, entity.CRYPTO}
	statuses := []entity.TransactionStatus{entity.PENDING, entity.APPROVED, entity.DECLINED}

	rapid.Check(t, func(t *rapid.T) {
		repo := newRoundTripMockRepo()
		uc := NewGetTransactionUseCase(repo)

		// Generate a random TransactionEntity with all fields
		id := rapid.StringMatching(`^txn_[a-z0-9]{8,16}$`).Draw(t, "id")
		amountInCents := rapid.Int64Range(1, 100_000_000).Draw(t, "amountInCents")
		currency := currencies[rapid.IntRange(0, len(currencies)-1).Draw(t, "currencyIdx")]
		paymentMethod := paymentMethods[rapid.IntRange(0, len(paymentMethods)-1).Draw(t, "paymentMethodIdx")]
		customerID := rapid.StringMatching(`^cust_[a-z0-9]{4,12}$`).Draw(t, "customerID")
		customerName := rapid.StringMatching(`^[A-Z][a-z]{2,10} [A-Z][a-z]{2,10}$`).Draw(t, "customerName")
		customerEmail := rapid.StringMatching(`^[a-z]{3,8}@[a-z]{3,8}\\.com$`).Draw(t, "customerEmail")
		customerPhone := rapid.StringMatching(`^\\+[0-9]{7,15}$`).Draw(t, "customerPhone")
		customerIP := rapid.StringMatching(`^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$`).Draw(t, "customerIP")
		status := statuses[rapid.IntRange(0, len(statuses)-1).Draw(t, "statusIdx")]

		// Generate timestamps: createdAt in a reasonable range, updatedAt >= createdAt
		baseUnix := rapid.Int64Range(1_000_000_000, 2_000_000_000).Draw(t, "createdAtUnix")
		offsetSecs := rapid.Int64Range(0, 3600).Draw(t, "updatedAtOffset")
		createdAt := time.Unix(baseUnix, 0).UTC()
		updatedAt := time.Unix(baseUnix+offsetSecs, 0).UTC()

		original := &entity.TransactionEntity{
			ID:                id,
			AmountInCents:     amountInCents,
			Currency:          currency,
			PaymentMethod:     paymentMethod,
			CustomerID:        customerID,
			CustomerName:      customerName,
			CustomerEmail:     customerEmail,
			CustomerPhone:     customerPhone,
			CustomerIPAddress: customerIP,
			Status:            status,
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
		}

		// Save via mock repository
		if err := repo.Save(context.Background(), original); err != nil {
			t.Fatalf("failed to save transaction: %v", err)
		}

		// Retrieve via GetTransactionUseCase
		retrieved, err := uc.Execute(context.Background(), id)
		if err != nil {
			t.Fatalf("unexpected error retrieving transaction %s: %v", id, err)
		}

		// Assert: all fields match the originally saved transaction
		if retrieved.ID != original.ID {
			t.Fatalf("ID mismatch: got %q, want %q", retrieved.ID, original.ID)
		}
		if retrieved.AmountInCents != original.AmountInCents {
			t.Fatalf("AmountInCents mismatch: got %d, want %d", retrieved.AmountInCents, original.AmountInCents)
		}
		if retrieved.Currency != original.Currency {
			t.Fatalf("Currency mismatch: got %q, want %q", retrieved.Currency, original.Currency)
		}
		if retrieved.PaymentMethod != original.PaymentMethod {
			t.Fatalf("PaymentMethod mismatch: got %q, want %q", retrieved.PaymentMethod, original.PaymentMethod)
		}
		if retrieved.CustomerID != original.CustomerID {
			t.Fatalf("CustomerID mismatch: got %q, want %q", retrieved.CustomerID, original.CustomerID)
		}
		if retrieved.CustomerName != original.CustomerName {
			t.Fatalf("CustomerName mismatch: got %q, want %q", retrieved.CustomerName, original.CustomerName)
		}
		if retrieved.CustomerEmail != original.CustomerEmail {
			t.Fatalf("CustomerEmail mismatch: got %q, want %q", retrieved.CustomerEmail, original.CustomerEmail)
		}
		if retrieved.CustomerPhone != original.CustomerPhone {
			t.Fatalf("CustomerPhone mismatch: got %q, want %q", retrieved.CustomerPhone, original.CustomerPhone)
		}
		if retrieved.CustomerIPAddress != original.CustomerIPAddress {
			t.Fatalf("CustomerIPAddress mismatch: got %q, want %q", retrieved.CustomerIPAddress, original.CustomerIPAddress)
		}
		if retrieved.Status != original.Status {
			t.Fatalf("Status mismatch: got %q, want %q", retrieved.Status, original.Status)
		}
		if !retrieved.CreatedAt.Equal(original.CreatedAt) {
			t.Fatalf("CreatedAt mismatch: got %v, want %v", retrieved.CreatedAt, original.CreatedAt)
		}
		if !retrieved.UpdatedAt.Equal(original.UpdatedAt) {
			t.Fatalf("UpdatedAt mismatch: got %v, want %v", retrieved.UpdatedAt, original.UpdatedAt)
		}
	})
}
