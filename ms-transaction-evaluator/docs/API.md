# Transaction Evaluator API

## Endpoint: POST /evaluate

### Description
Validates a transaction request and evaluates it for potential fraud detection.

### Request Body

```json
{
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
}
```

### Validation Rules

#### Required Fields
All fields are required and cannot be empty.

#### Amount
- `amount_in_cents` (int64): Must be a positive number greater than 0

#### Currency
- `currency` (string): Must be one of:
  - `USD` - US Dollar
  - `COP` - Colombian Peso
  - `EUR` - Euro

#### Payment Method
- `payment_method` (string): Must be one of:
  - `CARD` - Credit/Debit Card
  - `BANK_TRANSFER` - Bank Transfer
  - `CRYPTO` - Cryptocurrency

#### Customer Information
- `customer_id` (string): Required, cannot be empty or whitespace only
- `name` (string): Required, cannot be empty or whitespace only
- `email` (string): Required, must be a valid email format
- `phone` (string): Required, cannot be empty or whitespace only
- `ip_address` (string): Required, cannot be empty or whitespace only

### Response

#### Success Response (200 OK)
```json
{
  "message": "Transaction validation successful",
  "data": {
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
  }
}
```

#### Validation Error (400 Bad Request)
```json
{
  "error": "Validation failed",
  "details": "customer email is invalid"
}
```

#### Invalid Request Body (400 Bad Request)
```json
{
  "error": "Invalid request body",
  "details": "error message here"
}
```

### Example cURL Commands

#### Valid Request
```bash
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

#### Invalid Amount (Zero)
```bash
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

#### Invalid Currency
```bash
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "amount_in_cents": 10000,
    "currency": "GBP",
    "payment_method": "CARD",
    "customer": {
      "customer_id": "cust_123",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+1234567890",
      "ip_address": "192.168.1.1"
    }
  }'
```

#### Invalid Email
```bash
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

### Testing with Different Payment Methods

#### BANK_TRANSFER
```bash
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "amount_in_cents": 50000,
    "currency": "EUR",
    "payment_method": "BANK_TRANSFER",
    "customer": {
      "customer_id": "cust_456",
      "name": "Jane Smith",
      "email": "jane@example.com",
      "phone": "+441234567890",
      "ip_address": "10.0.0.1"
    }
  }'
```

#### CRYPTO
```bash
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "amount_in_cents": 100000,
    "currency": "COP",
    "payment_method": "CRYPTO",
    "customer": {
      "customer_id": "cust_789",
      "name": "Carlos Rodriguez",
      "email": "carlos@example.com",
      "phone": "+573001234567",
      "ip_address": "200.50.100.25"
    }
  }'
```

