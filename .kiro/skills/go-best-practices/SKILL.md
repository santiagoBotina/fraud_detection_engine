---
name: go-best-practices
description: Best practices considering this project's structure and architechture
---

# Go Best Practices & Project Architecture Guide

## Hexagonal Architecture

This project follows Hexagonal Architecture (Ports & Adapters). Respect these boundaries:

- `domain/` is the core — it MUST NOT import anything from `infrastructure/`. Only stdlib and domain types allowed.
- `domain/entity/` — domain models, value types, and typed constants. No business logic here.
- `domain/repository/` — port interfaces. Define what the domain needs, not how it's fulfilled.
- `domain/usecase/` — business logic. Each use case is a struct with an `Execute` method and a `New*` constructor.
- `infrastructure/adapter/in/` — inbound adapters (HTTP controllers). They translate HTTP into domain calls.
- `infrastructure/adapter/out/` — outbound adapters (DynamoDB, Kafka, etc.). They implement repository interfaces.

When adding a new feature, work from the inside out:
1. Define or extend entities in `domain/entity/`
2. Define or extend repository interfaces in `domain/repository/`
3. Implement use case logic in `domain/usecase/`
4. Add inbound adapter (controller) in `infrastructure/adapter/in/http/`
5. Add outbound adapter (repository impl) in `infrastructure/adapter/out/`
6. Wire everything in `cmd/api/main.go` via manual dependency injection

## Dependency Injection

- All DI is manual, wired in `cmd/api/main.go`. No DI frameworks.
- Constructors follow the pattern `NewXxx(deps...) *Xxx`.
- Controllers receive use cases as concrete struct pointers.
- Use cases receive repository interfaces (not concrete implementations).

```go
// Good: use case depends on interface
type SaveTransactionUseCase struct {
    transactionRepo repository.TransactionRepository
}

// Good: controller depends on concrete use case
type TransactionController struct {
    validateUseCase *usecase.ValidateCreateTransactionPayloadUseCase
    saveUseCase     *usecase.SaveTransactionUseCase
}
```

## Error Handling

- Define domain errors as package-level sentinel variables using `errors.New`.
- Name them with the `Err` prefix: `ErrAmountRequired`, `ErrCurrencyInvalid`.
- Wrap infrastructure errors with `fmt.Errorf("context: %w", err)` to preserve the error chain.
- Never use `fmt.Print*`, `print`, or `println` — they are forbidden by the linter. Use structured logging.

```go
var (
    ErrAmountRequired = errors.New("amount_in_cents is required")
    ErrCurrencyInvalid = errors.New("currency is invalid")
)

// Infrastructure wrapping
return fmt.Errorf("failed to marshal transaction: %w", err)
```

## Naming & Style Conventions

- JSON field names use `snake_case` via struct tags: `json:"amount_in_cents"`.
- DynamoDB attribute names also use `snake_case` via `dynamodbav` tags.
- Typed string constants for enums (Currency, PaymentMethod, TransactionStatus) — use `UPPER_CASE` values.
- Struct fields use `PascalCase` (exported). Unexported fields use `camelCase`.
- Package names are lowercase, single-word when possible. Use import aliases to avoid collisions:

```go
import (
    httpAdapter "ms-transaction-evaluator/internal/infrastructure/adapter/in/http"
    dynamodbAdapter "ms-transaction-evaluator/internal/infrastructure/adapter/out/aws/dynamodb"
)
```

## Use Case Pattern

Every use case follows this structure:

```go
type MyUseCase struct {
    // dependencies (repository interfaces, other use cases)
}

func NewMyUseCase(deps...) *MyUseCase {
    return &MyUseCase{...}
}

func (uc *MyUseCase) Execute(ctx context.Context, input *SomeInput) (*SomeOutput, error) {
    // 1. Validate input (nil check)
    // 2. Business logic
    // 3. Call repository if needed
    // 4. Return result or error
}
```

- Always nil-check the input at the top of `Execute`.
- Use cases that don't need persistence (e.g., validation-only) can omit `context.Context`.
- Return domain entities, not DTOs. Let the adapter layer handle transformation.

## Controller Pattern

```go
type MyController struct {
    someUseCase *usecase.SomeUseCase
}

func NewMyController(deps...) *MyController {
    return &MyController{...}
}

func (c *MyController) HandleSomething(ctx *echo.Context) error {
    // 1. Bind request body
    // 2. Call validation use case
    // 3. Call business use case
    // 4. Return JSON response with appropriate status code
}

func (c *MyController) RegisterRoutes(e *echo.Echo) {
    e.POST("/path", c.HandleSomething)
}
```

- Use `SuccessResponse` and `ErrorResponse` from `response_models.go` for consistent API responses.
- Add Swagger annotations (`// @Summary`, `// @Param`, etc.) to every handler.

## Testing

- Use Go's standard `testing` package only — no testify, no gomock.
- Table-driven tests with `t.Run` subtests for comprehensive coverage.
- Hand-written mocks that implement domain interfaces:

```go
type mockTransactionRepository struct {
    saveFunc func(ctx context.Context, transaction *entity.TransactionEntity) error
}

func (m *mockTransactionRepository) Save(ctx context.Context, transaction *entity.TransactionEntity) error {
    if m.saveFunc != nil {
        return m.saveFunc(ctx, transaction)
    }
    return nil
}
```

- Test files are co-located with source files (`_test.go` suffix, same package).
- Use helper functions like `createValidRequest()` to reduce test boilerplate.
- HTTP tests use `httptest.NewRequest`, `httptest.NewRecorder`, and `echo.NewContext`.
- Compare sentinel errors directly (`if err != ErrAmountRequired`) rather than string matching.

## Repository Implementation

- Outbound adapters live under `infrastructure/adapter/out/`.
- They implement interfaces from `domain/repository/`.
- Use a private mapping struct (e.g., `transactionItem`) for storage-specific serialization — don't leak storage tags into domain entities.
- Convert between domain entities and storage items explicitly in the adapter.

```go
// Private struct for DynamoDB serialization
type transactionItem struct {
    ID        string `dynamodbav:"id"`
    CreatedAt string `dynamodbav:"created_at"` // stored as ISO string
}
```

## Linting Awareness

The project uses golangci-lint v2 with strict settings. Keep in mind:
- Max cyclomatic complexity: 25
- Max cognitive complexity: 20
- Max function length: 80 lines / 60 statements
- No `fmt.Print*` or bare `print/println`
- Run `make lint` from `ms-transaction-evaluator/` before committing
