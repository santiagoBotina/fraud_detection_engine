# Project Structure

Monorepo with microservice(s) and shared infrastructure config at the root.

## Architecture Pattern
Hexagonal Architecture (Ports & Adapters) with clean domain separation.

```
.
├── docker-compose.yml          # Shared infra (DynamoDB, Kafka, Zookeeper)
├── Makefile                    # Root-level infra commands
├── .env                        # Shared env vars for Docker services
│
└── ms-transaction-evaluator/   # Transaction Evaluator microservice
    ├── cmd/api/main.go         # Entrypoint — wires dependencies, starts Echo server
    ├── internal/
    │   ├── domain/
    │   │   ├── entity/         # Domain models and value types (Currency, PaymentMethod, etc.)
    │   │   ├── repository/     # Repository interfaces (ports)
    │   │   └── usecase/        # Business logic use cases
    │   └── infrastructure/
    │       └── adapter/
    │           ├── in/http/    # Inbound HTTP adapter (Echo controllers, response models)
    │           └── out/aws/    # Outbound adapters (DynamoDB repository implementation)
    ├── docs/                   # Swagger/OpenAPI generated docs
    ├── Makefile                # Service-level commands (run, test, lint)
    └── .env                    # Service-specific env vars
```

## Conventions
- Domain layer has zero infrastructure dependencies — only stdlib and domain types
- Repository interfaces live in `domain/repository/`; implementations in `infrastructure/adapter/out/`
- Use cases are concrete structs with an `Execute` method, instantiated via `New*` constructors
- Controllers live in `infrastructure/adapter/in/http/` and register routes via `RegisterRoutes`
- Dependency injection is manual, wired in `cmd/api/main.go`
- JSON field names use `snake_case` (via struct tags)
- Errors are package-level `var` sentinels (e.g., `ErrAmountRequired`)

## Testing Conventions
- Tests use Go's standard `testing` package — no third-party assertion libraries
- Table-driven tests with `t.Run` subtests
- Mocks are hand-written structs implementing domain interfaces (no codegen mocking)
- Test files are co-located with the code they test (`_test.go` suffix)
- HTTP tests use `httptest.NewRequest` / `httptest.NewRecorder` with Echo's `NewContext`
