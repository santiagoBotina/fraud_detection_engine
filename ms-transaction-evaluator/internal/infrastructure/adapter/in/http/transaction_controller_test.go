package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ms-transaction-evaluator/internal/domain/usecase"

	"github.com/labstack/echo/v5"
)

func TestTransactionController_EvaluateTransaction(t *testing.T) {
	// Setup
	validateUseCase := usecase.NewValidateCreateTransactionPayloadUseCase()
	controller := NewTransactionController(validateUseCase)
	e := echo.New()

	t.Run("should return 200 with valid transaction request", func(t *testing.T) {
		requestBody := `{
			"amount_in_cents": 10000,
			"currency": "USD",
			"payment_method": "CARD",
			"customer": {
				"customer_id": "cust_123",
				"name": "John Doe",
				"email": "john@example.com",
				"phone": "+1234567890",
				"ip_address": "192.168.1.1"
			}
		}`

		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.EvaluateTransaction(c)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "Transaction validation successful" {
			t.Errorf("Expected success message, got: %v", response["message"])
		}
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		requestBody := `{invalid json}`

		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.EvaluateTransaction(c)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid request body" {
			t.Errorf("Expected 'Invalid request body' error, got: %v", response["error"])
		}
	})

	t.Run("should return 400 when amount is zero", func(t *testing.T) {
		requestBody := `{
			"amount_in_cents": 0,
			"currency": "USD",
			"payment_method": "CARD",
			"customer": {
				"customer_id": "cust_123",
				"name": "John Doe",
				"email": "john@example.com",
				"phone": "+1234567890",
				"ip_address": "192.168.1.1"
			}
		}`

		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.EvaluateTransaction(c)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Validation failed" {
			t.Errorf("Expected 'Validation failed' error, got: %v", response["error"])
		}
	})

	t.Run("should return 400 when currency is invalid", func(t *testing.T) {
		requestBody := `{
			"amount_in_cents": 10000,
			"currency": "INVALID",
			"payment_method": "CARD",
			"customer": {
				"customer_id": "cust_123",
				"name": "John Doe",
				"email": "john@example.com",
				"phone": "+1234567890",
				"ip_address": "192.168.1.1"
			}
		}`

		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.EvaluateTransaction(c)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should return 400 when email is invalid", func(t *testing.T) {
		requestBody := `{
			"amount_in_cents": 10000,
			"currency": "USD",
			"payment_method": "CARD",
			"customer": {
				"customer_id": "cust_123",
				"name": "John Doe",
				"email": "invalid-email",
				"phone": "+1234567890",
				"ip_address": "192.168.1.1"
			}
		}`

		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.EvaluateTransaction(c)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Validation failed" {
			t.Errorf("Expected 'Validation failed' error, got: %v", response["error"])
		}
	})

	t.Run("should return 400 when customer fields are missing", func(t *testing.T) {
		requestBody := `{
			"amount_in_cents": 10000,
			"currency": "USD",
			"payment_method": "CARD",
			"customer": {
				"customer_id": "",
				"name": "",
				"email": "",
				"phone": "",
				"ip_address": ""
			}
		}`

		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.EvaluateTransaction(c)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should accept all valid currencies", func(t *testing.T) {
		currencies := []string{"USD", "COP", "EUR"}

		for _, currency := range currencies {
			requestBody := `{
				"amount_in_cents": 10000,
				"currency": "` + currency + `",
				"payment_method": "CARD",
				"customer": {
					"customer_id": "cust_123",
					"name": "John Doe",
					"email": "john@example.com",
					"phone": "+1234567890",
					"ip_address": "192.168.1.1"
				}
			}`

			req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := controller.EvaluateTransaction(c)
			if err != nil {
				t.Fatalf("Expected no error for currency %s, got: %v", currency, err)
			}

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status code %d for currency %s, got %d", http.StatusOK, currency, rec.Code)
			}
		}
	})

	t.Run("should accept all valid payment methods", func(t *testing.T) {
		methods := []string{"CARD", "BANK_TRANSFER", "CRYPTO"}

		for _, method := range methods {
			requestBody := `{
				"amount_in_cents": 10000,
				"currency": "USD",
				"payment_method": "` + method + `",
				"customer": {
					"customer_id": "cust_123",
					"name": "John Doe",
					"email": "john@example.com",
					"phone": "+1234567890",
					"ip_address": "192.168.1.1"
				}
			}`

			req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := controller.EvaluateTransaction(c)
			if err != nil {
				t.Fatalf("Expected no error for payment method %s, got: %v", method, err)
			}

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status code %d for payment method %s, got %d", http.StatusOK, method, rec.Code)
			}
		}
	})
}
