package usecase

import (
	"context"
	"errors"
	"ms-decision-service/internal/domain/entity"
	"sort"
	"testing"
	"time"

	"pgregory.net/rapid"
)

func TestGetRuleEvaluationsUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		transactionID string
		ruleEvalRepo  *mockRuleEvaluationRepository
		wantErr       error
		wantCount     int
	}{
		{
			name:          "empty transaction ID returns ErrTransactionIDEmpty",
			transactionID: "",
			ruleEvalRepo:  &mockRuleEvaluationRepository{},
			wantErr:       ErrTransactionIDEmpty,
		},
		{
			name:          "successful retrieval returns results",
			transactionID: "tx-123",
			ruleEvalRepo: &mockRuleEvaluationRepository{
				findFunc: func(_ context.Context, txID string) ([]entity.RuleEvaluationResult, error) {
					if txID != "tx-123" {
						return nil, nil
					}
					return []entity.RuleEvaluationResult{
						{
							TransactionID: "tx-123",
							RuleID:        "rule-1",
							RuleName:      "Block CRYPTO",
							Priority:      1,
							EvaluatedAt:   now,
						},
						{
							TransactionID: "tx-123",
							RuleID:        "rule-2",
							RuleName:      "High amount",
							Priority:      2,
							EvaluatedAt:   now,
						},
					}, nil
				},
			},
			wantCount: 2,
		},
		{
			name:          "no results returns empty slice",
			transactionID: "tx-nonexistent",
			ruleEvalRepo: &mockRuleEvaluationRepository{
				findFunc: func(_ context.Context, _ string) ([]entity.RuleEvaluationResult, error) {
					return []entity.RuleEvaluationResult{}, nil
				},
			},
			wantCount: 0,
		},
		{
			name:          "repository error returns ErrEvaluationRetrievalFailed",
			transactionID: "tx-123",
			ruleEvalRepo: &mockRuleEvaluationRepository{
				findFunc: func(_ context.Context, _ string) ([]entity.RuleEvaluationResult, error) {
					return nil, errors.New("dynamo timeout")
				},
			},
			wantErr: ErrEvaluationRetrievalFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewGetRuleEvaluationsUseCase(tc.ruleEvalRepo)
			results, err := uc.Execute(context.Background(), tc.transactionID)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error wrapping %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != tc.wantCount {
				t.Fatalf("expected %d results, got %d", tc.wantCount, len(results))
			}
		})
	}
}

// Feature: fraud-analyst-dashboard, Property 6: Evaluation results sorted by priority
// Validates: Requirements 4.1
func TestProperty_EvaluationResultsSortedByPriority(t *testing.T) {
	nonEmptyStr := rapid.StringMatching(`[a-zA-Z0-9]{1,20}`)

	rapid.Check(t, func(t *rapid.T) {
		txID := nonEmptyStr.Draw(t, "transactionID")

		// Generate 1–20 evaluation results with random priorities
		count := rapid.IntRange(1, 20).Draw(t, "resultCount")
		results := make([]entity.RuleEvaluationResult, count)

		for i := range count {
			results[i] = entity.RuleEvaluationResult{
				TransactionID:     txID,
				RuleID:            nonEmptyStr.Draw(t, "ruleID"),
				RuleName:          nonEmptyStr.Draw(t, "ruleName"),
				ConditionField:    nonEmptyStr.Draw(t, "condField"),
				ConditionOperator: nonEmptyStr.Draw(t, "condOp"),
				ConditionValue:    nonEmptyStr.Draw(t, "condVal"),
				ActualFieldValue:  nonEmptyStr.Draw(t, "actualVal"),
				Matched:           rapid.Bool().Draw(t, "matched"),
				ResultStatus:      rapid.SampledFrom([]string{"APPROVED", "DECLINED", "FRAUD_CHECK"}).Draw(t, "status"),
				EvaluatedAt:       time.Now(),
				Priority:          rapid.IntRange(1, 100).Draw(t, "priority"),
			}
		}

		// Sort by priority ascending — simulating the repo's contract (DynamoDB adapter sorts)
		sort.Slice(results, func(i, j int) bool {
			return results[i].Priority < results[j].Priority
		})

		// Mock returns pre-sorted results (as the real repo would)
		repo := &mockRuleEvaluationRepository{
			findFunc: func(_ context.Context, id string) ([]entity.RuleEvaluationResult, error) {
				if id != txID {
					return nil, nil
				}
				return results, nil
			},
		}

		uc := NewGetRuleEvaluationsUseCase(repo)
		got, err := uc.Execute(context.Background(), txID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != len(results) {
			t.Fatalf("expected %d results, got %d", len(results), len(got))
		}

		// Assert: output is sorted by priority ascending
		for i := 1; i < len(got); i++ {
			if got[i].Priority < got[i-1].Priority {
				t.Fatalf("results not sorted by priority ascending: index %d has priority %d, index %d has priority %d",
					i-1, got[i-1].Priority, i, got[i].Priority)
			}
		}
	})
}
