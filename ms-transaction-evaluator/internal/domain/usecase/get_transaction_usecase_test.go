package usecase

import (
	"context"
	"errors"
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"
)

// getTransactionMockRepo is a hand-written mock implementing TransactionRepository
// for GetTransactionUseCase tests.
type getTransactionMockRepo struct {
	findByIDFunc func(ctx context.Context, id string) (*entity.TransactionEntity, error)
}

func (m *getTransactionMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *getTransactionMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus) error {
	return nil
}

func (m *getTransactionMockRepo) FindByID(ctx context.Context, id string) (*entity.TransactionEntity, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *getTransactionMockRepo) FindAllPaginated(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
	return nil, "", nil
}

func TestGetTransactionUseCase_Execute(t *testing.T) {
	now := time.Now().UTC()
	sampleTxn := &entity.TransactionEntity{
		ID:                "txn_001",
		AmountInCents:     15000,
		Currency:          entity.USD,
		PaymentMethod:     entity.CARD,
		CustomerID:        "cust_123",
		CustomerName:      "John Doe",
		CustomerEmail:     "john@example.com",
		CustomerPhone:     "+1234567890",
		CustomerIPAddress: "192.168.1.1",
		Status:            entity.APPROVED,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	tests := []struct {
		name          string
		id            string
		findByIDFunc  func(ctx context.Context, id string) (*entity.TransactionEntity, error)
		expectError   bool
		checkSentinel error
		expectTxn     *entity.TransactionEntity
	}{
		{
			name: "successful retrieval",
			id:   "txn_001",
			findByIDFunc: func(_ context.Context, id string) (*entity.TransactionEntity, error) {
				if id == "txn_001" {
					return sampleTxn, nil
				}
				return nil, nil
			},
			expectError: false,
			expectTxn:   sampleTxn,
		},
		{
			name: "not found returns nil",
			id:   "txn_nonexistent",
			findByIDFunc: func(_ context.Context, _ string) (*entity.TransactionEntity, error) {
				return nil, nil
			},
			expectError:   true,
			checkSentinel: ErrTransactionNotFound,
		},
		{
			name: "repository error wraps ErrTransactionNotFound",
			id:   "txn_err",
			findByIDFunc: func(_ context.Context, _ string) (*entity.TransactionEntity, error) {
				return nil, errors.New("dynamodb connection failed")
			},
			expectError:   true,
			checkSentinel: ErrTransactionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &getTransactionMockRepo{findByIDFunc: tt.findByIDFunc}
			uc := NewGetTransactionUseCase(repo)

			result, err := uc.Execute(context.Background(), tt.id)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.checkSentinel != nil && !errors.Is(err, tt.checkSentinel) {
					t.Errorf("expected error to wrap %v, got: %v", tt.checkSentinel, err)
				}
				if result != nil {
					t.Errorf("expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
				if result == nil {
					t.Fatal("expected result but got nil")
				}
				if result.ID != tt.expectTxn.ID {
					t.Errorf("expected ID %s, got %s", tt.expectTxn.ID, result.ID)
				}
				if result.AmountInCents != tt.expectTxn.AmountInCents {
					t.Errorf("expected AmountInCents %d, got %d", tt.expectTxn.AmountInCents, result.AmountInCents)
				}
				if result.Currency != tt.expectTxn.Currency {
					t.Errorf("expected Currency %s, got %s", tt.expectTxn.Currency, result.Currency)
				}
				if result.Status != tt.expectTxn.Status {
					t.Errorf("expected Status %s, got %s", tt.expectTxn.Status, result.Status)
				}
			}
		})
	}
}
