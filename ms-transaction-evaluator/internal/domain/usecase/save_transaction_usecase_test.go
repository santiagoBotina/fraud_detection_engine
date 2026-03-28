package usecase

import (
	"context"
	"errors"
	"testing"

	"ms-transaction-evaluator/internal/domain/entity"
)

type mockTransactionRepository struct {
	saveFunc func(ctx context.Context, transaction *entity.TransactionEntity) error
}

func (m *mockTransactionRepository) Save(ctx context.Context, transaction *entity.TransactionEntity) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, transaction)
	}
	return nil
}

func TestSaveTransactionUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     *entity.EvaluateTransactionRequest
		setupMock   func(*mockTransactionRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful save",
			request: &entity.EvaluateTransactionRequest{
				AmountInCents: 10000,
				Currency:      entity.USD,
				PaymentMethod: entity.CARD,
				CustomerInfo: entity.CustomerInfo{
					CustomerID: "cust_123",
					Name:       "John Doe",
					Email:      "john@example.com",
					Phone:      "+1234567890",
					IpAddress:  "192.168.1.1",
				},
			},
			setupMock: func(m *mockTransactionRepository) {
				m.saveFunc = func(ctx context.Context, transaction *entity.TransactionEntity) error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name:        "nil request",
			request:     nil,
			setupMock:   func(m *mockTransactionRepository) {},
			expectError: true,
			errorMsg:    "request is nil",
		},
		{
			name: "repository save error",
			request: &entity.EvaluateTransactionRequest{
				AmountInCents: 10000,
				Currency:      entity.USD,
				PaymentMethod: entity.CARD,
				CustomerInfo: entity.CustomerInfo{
					CustomerID: "cust_123",
					Name:       "John Doe",
					Email:      "john@example.com",
					Phone:      "+1234567890",
					IpAddress:  "192.168.1.1",
				},
			},
			setupMock: func(m *mockTransactionRepository) {
				m.saveFunc = func(ctx context.Context, transaction *entity.TransactionEntity) error {
					return errors.New("database error")
				}
			},
			expectError: true,
			errorMsg:    "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTransactionRepository{}
			tt.setupMock(mockRepo)
			useCase := NewSaveTransactionUseCase(mockRepo)

			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s' but got '%s'", tt.errorMsg, err.Error())
				}
				if result != nil {
					t.Errorf("expected nil result but got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				} else {
					if result.ID == "" {
						t.Errorf("expected ID to be generated")
					}
					if result.AmountInCents != tt.request.AmountInCents {
						t.Errorf("expected amount %d but got %d", tt.request.AmountInCents, result.AmountInCents)
					}
					if result.Currency != tt.request.Currency {
						t.Errorf("expected currency %s but got %s", tt.request.Currency, result.Currency)
					}
					if result.PaymentMethod != tt.request.PaymentMethod {
						t.Errorf("expected payment method %s but got %s", tt.request.PaymentMethod, result.PaymentMethod)
					}
					if result.CustomerID != tt.request.CustomerInfo.CustomerID {
						t.Errorf("expected customer ID %s but got %s", tt.request.CustomerInfo.CustomerID, result.CustomerID)
					}
					if result.CustomerName != tt.request.CustomerInfo.Name {
						t.Errorf("expected customer name %s but got %s", tt.request.CustomerInfo.Name, result.CustomerName)
					}
					if result.CustomerEmail != tt.request.CustomerInfo.Email {
						t.Errorf("expected customer email %s but got %s", tt.request.CustomerInfo.Email, result.CustomerEmail)
					}
					if result.CustomerPhone != tt.request.CustomerInfo.Phone {
						t.Errorf("expected customer phone %s but got %s", tt.request.CustomerInfo.Phone, result.CustomerPhone)
					}
					if result.CustomerIPAddress != tt.request.CustomerInfo.IpAddress {
						t.Errorf("expected customer IP %s but got %s", tt.request.CustomerInfo.IpAddress, result.CustomerIPAddress)
					}
					if result.Status != entity.PENDING {
						t.Errorf("expected status PENDING but got %s", result.Status)
					}
					if result.CreatedAt.IsZero() {
						t.Errorf("expected CreatedAt to be set")
					}
					if result.UpdatedAt.IsZero() {
						t.Errorf("expected UpdatedAt to be set")
					}
				}
			}
		})
	}
}
