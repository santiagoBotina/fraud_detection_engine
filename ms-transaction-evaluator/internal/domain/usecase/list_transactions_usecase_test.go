package usecase

import (
	"context"
	"errors"
	"ms-transaction-evaluator/internal/domain/entity"
	"testing"
	"time"
)

// listTransactionsMockRepo is a hand-written mock implementing TransactionRepository
// for ListTransactionsUseCase tests.
type listTransactionsMockRepo struct {
	findAllPaginatedFunc func(ctx context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error)
}

func (m *listTransactionsMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *listTransactionsMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus) error {
	return nil
}

func (m *listTransactionsMockRepo) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return nil, nil
}

func (m *listTransactionsMockRepo) FindAllPaginated(ctx context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error) {
	if m.findAllPaginatedFunc != nil {
		return m.findAllPaginatedFunc(ctx, limit, cursor)
	}
	return nil, "", nil
}

func TestListTransactionsUseCase_Execute(t *testing.T) {
	now := time.Now().UTC()
	sampleTxns := []entity.TransactionEntity{
		{
			ID:            "txn_001",
			AmountInCents: 15000,
			Currency:      entity.USD,
			PaymentMethod: entity.CARD,
			CustomerID:    "cust_123",
			Status:        entity.APPROVED,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "txn_002",
			AmountInCents: 5000,
			Currency:      entity.EUR,
			PaymentMethod: entity.BANK_TRANSFER,
			CustomerID:    "cust_456",
			Status:        entity.PENDING,
			CreatedAt:     now.Add(-time.Hour),
			UpdatedAt:     now.Add(-time.Hour),
		},
	}

	tests := []struct {
		name                 string
		limit                int
		cursor               string
		findAllPaginatedFunc func(ctx context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error)
		expectError          bool
		checkSentinel        error
		expectCount          int
		expectNextCursor     string
	}{
		{
			name:   "valid limit returns transactions",
			limit:  20,
			cursor: "",
			findAllPaginatedFunc: func(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
				return sampleTxns, "next_cursor_abc", nil
			},
			expectError:      false,
			expectCount:      2,
			expectNextCursor: "next_cursor_abc",
		},
		{
			name:   "limit at minimum boundary (1)",
			limit:  1,
			cursor: "",
			findAllPaginatedFunc: func(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
				return sampleTxns[:1], "", nil
			},
			expectError:      false,
			expectCount:      1,
			expectNextCursor: "",
		},
		{
			name:   "limit at maximum boundary (100)",
			limit:  100,
			cursor: "",
			findAllPaginatedFunc: func(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
				return sampleTxns, "", nil
			},
			expectError:      false,
			expectCount:      2,
			expectNextCursor: "",
		},
		{
			name:          "invalid limit zero",
			limit:         0,
			cursor:        "",
			expectError:   true,
			checkSentinel: ErrInvalidLimit,
		},
		{
			name:          "invalid limit negative",
			limit:         -1,
			cursor:        "",
			expectError:   true,
			checkSentinel: ErrInvalidLimit,
		},
		{
			name:          "invalid limit exceeds max (101)",
			limit:         101,
			cursor:        "",
			expectError:   true,
			checkSentinel: ErrInvalidLimit,
		},
		{
			name:   "empty database returns empty slice",
			limit:  20,
			cursor: "",
			findAllPaginatedFunc: func(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
				return []entity.TransactionEntity{}, "", nil
			},
			expectError:      false,
			expectCount:      0,
			expectNextCursor: "",
		},
		{
			name:   "malformed cursor propagates repository error",
			limit:  20,
			cursor: "not-valid-base64!@#$",
			findAllPaginatedFunc: func(_ context.Context, _ int, cursor string) ([]entity.TransactionEntity, string, error) {
				return nil, "", ErrInvalidCursor
			},
			expectError:   true,
			checkSentinel: ErrInvalidCursor,
		},
		{
			name:   "repository error is propagated",
			limit:  20,
			cursor: "",
			findAllPaginatedFunc: func(_ context.Context, _ int, _ string) ([]entity.TransactionEntity, string, error) {
				return nil, "", errors.New("dynamodb connection failed")
			},
			expectError: true,
		},
		{
			name:   "valid cursor passes through to repository",
			limit:  10,
			cursor: "eyJpZCI6InR4bl8wMDEifQ==",
			findAllPaginatedFunc: func(_ context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error) {
				if limit != 10 {
					return nil, "", errors.New("unexpected limit")
				}
				if cursor != "eyJpZCI6InR4bl8wMDEifQ==" {
					return nil, "", errors.New("unexpected cursor")
				}
				return sampleTxns[1:], "", nil
			},
			expectError:      false,
			expectCount:      1,
			expectNextCursor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &listTransactionsMockRepo{findAllPaginatedFunc: tt.findAllPaginatedFunc}
			uc := NewListTransactionsUseCase(repo)

			txns, nextCursor, err := uc.Execute(context.Background(), tt.limit, tt.cursor)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.checkSentinel != nil && !errors.Is(err, tt.checkSentinel) {
					t.Errorf("expected error to wrap %v, got: %v", tt.checkSentinel, err)
				}
				if txns != nil {
					t.Errorf("expected nil transactions on error, got %v", txns)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
				if len(txns) != tt.expectCount {
					t.Errorf("expected %d transactions, got %d", tt.expectCount, len(txns))
				}
				if nextCursor != tt.expectNextCursor {
					t.Errorf("expected next_cursor %q, got %q", tt.expectNextCursor, nextCursor)
				}
			}
		})
	}
}
