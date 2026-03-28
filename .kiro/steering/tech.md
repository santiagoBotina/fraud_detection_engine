# Tech Stack

## Language & Runtime
- Go 1.25+ (module: `ms-transaction-evaluator`)

## Frameworks & Libraries
- HTTP: Echo v5 (`github.com/labstack/echo/v5`)
- AWS SDK: `github.com/aws/aws-sdk-go-v2` (DynamoDB)
- UUID: `github.com/google/uuid`
- Env: `github.com/joho/godotenv`
- Swagger: `github.com/swaggo/swag` + `github.com/swaggo/echo-swagger`

## Infrastructure (Docker Compose)
- DynamoDB Local (in-memory, shared DB mode)
- Apache Kafka (Confluent cp-kafka 7.5.0) + Zookeeper
- Kafka UI (provectuslabs/kafka-ui)

## Dev Tools
- Air for hot-reload during development
- golangci-lint v2 with extensive linter configuration (see `.golangci.yml`)
- Swagger/OpenAPI doc generation via swag annotations

## Linting
golangci-lint is configured with ~50 linters enabled. Key thresholds:
- Cyclomatic complexity max: 25
- Cognitive complexity min: 20
- Function length: 80 lines / 60 statements
- Duplicate threshold: 150
- Nested if complexity: 4
- Formatters: gofmt, gofumpt, goimports
- `fmt.Print*` and `print/println` are forbidden — use structured logging

## Common Commands

All commands below run from `ms-transaction-evaluator/` unless noted.

| Action | Command | Notes |
|---|---|---|
| Run app | `make run` | Requires `.env` |
| Run with hot-reload | `make run-dev` | Uses Air |
| Run tests | `make test` | `go test ./...` |
| Lint | `make lint` | golangci-lint |
| Lint + autofix | `make lint-fix` | |
| Start infra | `make run` (root) | `docker compose up -d` |
| Create DynamoDB table | `make create_transactions_table` (root) | Runs via Docker |
