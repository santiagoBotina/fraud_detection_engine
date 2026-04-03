package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"ms-transaction-evaluator/internal/domain/entity"
	"sort"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// Feature: fraud-analyst-dashboard, Property 1: Paginated list respects limit and ordering
// Validates: Requirements 1.1, 1.2

// paginatedMockRepo is a hand-written mock implementing TransactionRepository
// that stores a slice of transactions and simulates FindAllPaginated by sorting
// by created_at descending and slicing to the requested limit.
type paginatedMockRepo struct {
	transactions []entity.TransactionEntity
}

func (m *paginatedMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *paginatedMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus) error {
	return nil
}

func (m *paginatedMockRepo) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return nil, nil
}

func (m *paginatedMockRepo) FindAllPaginated(_ context.Context, limit int, _ string) ([]entity.TransactionEntity, string, error) {
	// Copy and sort by created_at descending (simulates real DB behavior)
	sorted := make([]entity.TransactionEntity, len(m.transactions))
	copy(sorted, m.transactions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	if limit > len(sorted) {
		limit = len(sorted)
	}

	result := sorted[:limit]
	nextCursor := ""
	if limit < len(sorted) {
		nextCursor = "has_more"
	}

	return result, nextCursor, nil
}

func TestProperty_PaginatedListRespectsLimitAndOrdering(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random number of transactions (0–50)
		numTxns := rapid.IntRange(0, 50).Draw(t, "numTransactions")

		// Generate random transactions with distinct created_at timestamps.
		// Use sequential offsets to guarantee uniqueness, then shuffle to
		// simulate unordered input to the repository.
		baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		txns := make([]entity.TransactionEntity, numTxns)
		for i := range numTxns {
			txns[i] = entity.TransactionEntity{
				ID:            rapid.StringMatching(`^txn_[a-z0-9]{8}$`).Draw(t, "txnID"),
				AmountInCents: rapid.Int64Range(1, 10_000_000).Draw(t, "amount"),
				Currency:      entity.USD,
				PaymentMethod: entity.CARD,
				CustomerID:    "cust_test",
				Status:        entity.PENDING,
				CreatedAt:     baseTime.Add(time.Duration(i+1) * time.Second),
				UpdatedAt:     baseTime.Add(time.Duration(i+1) * time.Second),
			}
		}

		// Shuffle using a rapid-drawn seed for reproducibility
		seed := rapid.Int64().Draw(t, "shuffleSeed")
		rng := rand.New(rand.NewSource(seed))
		rng.Shuffle(len(txns), func(i, j int) {
			txns[i], txns[j] = txns[j], txns[i]
		})

		// Generate a valid limit (1–100)
		limit := rapid.IntRange(1, 100).Draw(t, "limit")

		repo := &paginatedMockRepo{transactions: txns}
		uc := NewListTransactionsUseCase(repo)

		result, _, err := uc.Execute(context.Background(), limit, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert: result length ≤ limit
		if len(result) > limit {
			t.Fatalf("result length %d exceeds limit %d", len(result), limit)
		}

		// Assert: result length ≤ number of transactions
		if len(result) > numTxns {
			t.Fatalf("result length %d exceeds total transactions %d", len(result), numTxns)
		}

		// Assert: created_at values in strictly descending order
		for i := 1; i < len(result); i++ {
			if !result[i-1].CreatedAt.After(result[i].CreatedAt) {
				t.Fatalf("ordering violation at index %d: %v is not after %v",
					i, result[i-1].CreatedAt, result[i].CreatedAt)
			}
		}
	})
}

// Feature: fraud-analyst-dashboard, Property 2: Cursor pagination yields non-overlapping complete coverage
// Validates: Requirements 1.3

// cursorPaginatedMockRepo simulates cursor-based pagination with base64-encoded
// cursors containing id and created_at, matching the real DynamoDB adapter behavior.
type cursorPaginatedMockRepo struct {
	transactions []entity.TransactionEntity
}

func (m *cursorPaginatedMockRepo) Save(_ context.Context, _ *entity.TransactionEntity) error {
	return nil
}

func (m *cursorPaginatedMockRepo) UpdateStatus(_ context.Context, _ string, _ entity.TransactionStatus) error {
	return nil
}

func (m *cursorPaginatedMockRepo) FindByID(_ context.Context, _ string) (*entity.TransactionEntity, error) {
	return nil, nil
}

type testPaginationCursor struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

func (m *cursorPaginatedMockRepo) FindAllPaginated(_ context.Context, limit int, cursor string) ([]entity.TransactionEntity, string, error) {
	// Sort all transactions by created_at descending (matches real adapter)
	sorted := make([]entity.TransactionEntity, len(m.transactions))
	copy(sorted, m.transactions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	// Determine start index based on cursor (ExclusiveStartKey semantics)
	startIdx := 0
	if cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}

		var cur testPaginationCursor
		if err := json.Unmarshal(decoded, &cur); err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}

		cursorTime, err := time.Parse("2006-01-02T15:04:05Z07:00", cur.CreatedAt)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor time: %w", err)
		}

		// Find the position after the cursor item in the sorted list
		found := false
		for i, txn := range sorted {
			if txn.ID == cur.ID && txn.CreatedAt.Equal(cursorTime) {
				startIdx = i + 1
				found = true
				break
			}
		}
		if !found {
			return nil, "", fmt.Errorf("cursor item not found")
		}
	}

	// Slice from startIdx
	remaining := sorted[startIdx:]
	end := limit
	if end > len(remaining) {
		end = len(remaining)
	}
	page := remaining[:end]

	// Build next cursor if there are more items
	nextCursor := ""
	if end < len(remaining) {
		lastItem := page[len(page)-1]
		cur := testPaginationCursor{
			ID:        lastItem.ID,
			CreatedAt: lastItem.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		curJSON, _ := json.Marshal(cur)
		nextCursor = base64.StdEncoding.EncodeToString(curJSON)
	}

	return page, nextCursor, nil
}

func TestProperty_CursorPaginationYieldsCompleteCoverage(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random number of transactions (0–50)
		numTxns := rapid.IntRange(0, 50).Draw(t, "numTransactions")

		// Generate transactions with distinct created_at timestamps
		baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		txns := make([]entity.TransactionEntity, numTxns)
		for i := range numTxns {
			txns[i] = entity.TransactionEntity{
				ID:            fmt.Sprintf("txn_%08d", i),
				AmountInCents: rapid.Int64Range(1, 10_000_000).Draw(t, "amount"),
				Currency:      entity.USD,
				PaymentMethod: entity.CARD,
				CustomerID:    "cust_test",
				Status:        entity.PENDING,
				CreatedAt:     baseTime.Add(time.Duration(i+1) * time.Second),
				UpdatedAt:     baseTime.Add(time.Duration(i+1) * time.Second),
			}
		}

		// Shuffle to simulate unordered storage
		seed := rapid.Int64().Draw(t, "shuffleSeed")
		rng := rand.New(rand.NewSource(seed))
		rng.Shuffle(len(txns), func(i, j int) {
			txns[i], txns[j] = txns[j], txns[i]
		})

		// Use a page size between 1 and 15 to force multiple pages
		pageSize := rapid.IntRange(1, 15).Draw(t, "pageSize")

		repo := &cursorPaginatedMockRepo{transactions: txns}
		uc := NewListTransactionsUseCase(repo)

		// Iterate through all pages collecting every transaction
		var allCollected []entity.TransactionEntity
		cursor := ""
		maxPages := numTxns + 2 // safety bound to prevent infinite loops

		for page := 0; page < maxPages; page++ {
			result, nextCursor, err := uc.Execute(context.Background(), pageSize, cursor)
			if err != nil {
				t.Fatalf("unexpected error on page %d: %v", page, err)
			}

			allCollected = append(allCollected, result...)

			if nextCursor == "" {
				break
			}
			cursor = nextCursor
		}

		// Assert: total collected equals total transactions (no omissions)
		if len(allCollected) != numTxns {
			t.Fatalf("collected %d transactions, expected %d", len(allCollected), numTxns)
		}

		// Assert: no duplicates — all IDs are unique
		seen := make(map[string]bool, len(allCollected))
		for _, txn := range allCollected {
			if seen[txn.ID] {
				t.Fatalf("duplicate transaction found: %s", txn.ID)
			}
			seen[txn.ID] = true
		}

		// Assert: union of collected IDs matches original set exactly
		originalIDs := make(map[string]bool, numTxns)
		for _, txn := range txns {
			originalIDs[txn.ID] = true
		}
		for _, txn := range allCollected {
			if !originalIDs[txn.ID] {
				t.Fatalf("collected unknown transaction: %s", txn.ID)
			}
		}
	})
}

// Feature: fraud-analyst-dashboard, Property 3: Invalid limit values produce 400 errors
// Validates: Requirements 1.5

func TestProperty_InvalidLimitValuesProduceErrors(t *testing.T) {
	// Use a minimal mock — the repo should never be called for invalid limits
	repo := &paginatedMockRepo{}
	uc := NewListTransactionsUseCase(repo)

	t.Run("limit_leq_zero", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate limit values ≤ 0 (includes zero and negative numbers)
			limit := rapid.IntRange(-1000, 0).Draw(t, "invalidLimit")

			_, _, err := uc.Execute(context.Background(), limit, "")
			if err == nil {
				t.Fatalf("expected error for limit %d, got nil", limit)
			}
			if !errors.Is(err, ErrInvalidLimit) {
				t.Fatalf("expected ErrInvalidLimit for limit %d, got: %v", limit, err)
			}
		})
	})

	t.Run("limit_gt_100", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate limit values > 100
			limit := rapid.IntRange(101, 10000).Draw(t, "invalidLimit")

			_, _, err := uc.Execute(context.Background(), limit, "")
			if err == nil {
				t.Fatalf("expected error for limit %d, got nil", limit)
			}
			if !errors.Is(err, ErrInvalidLimit) {
				t.Fatalf("expected ErrInvalidLimit for limit %d, got: %v", limit, err)
			}
		})
	})
}
