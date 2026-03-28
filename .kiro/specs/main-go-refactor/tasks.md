# Implementation Plan: main-go-refactor

## Overview

Remove the inline root `/` route handler from `cmd/api/main.go` and verify the refactored entrypoint meets all requirements. The existing DynamoDB initialization, dependency wiring, controller registration, Swagger route, and server startup are already correct and remain unchanged.

## Tasks

- [x] 1. Remove the root `/` route handler from main.go
  - Delete the `e.GET("/", func(c *echo.Context) error { ... })` block from `ms-transaction-evaluator/cmd/api/main.go`
  - Remove the `"net/http"` import if it becomes unused after the deletion
  - Confirm no other inline handler definitions remain for application endpoints
  - _Requirements: 4.1, 4.2, 3.2_

- [x] 2. Verify refactored main.go structure
  - [x] 2.1 Validate that DynamoDB client initialization is intact
    - Confirm AWS SDK config loading with `AWS_REGION` env var and default `us-east-1`
    - Confirm custom endpoint logic for `DYNAMO_DB_ENDPOINT`
    - Confirm fatal log on SDK config failure
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [x] 2.2 Validate dependency wiring is intact
    - Confirm repository, use case, and controller instantiation order
    - Confirm manual dependency injection with no framework
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [x] 2.3 Validate controller route registration and Swagger
    - Confirm `transactionController.RegisterRoutes(e)` call exists
    - Confirm Swagger route at `/swagger/*` is registered
    - Confirm swag annotation comments are preserved
    - _Requirements: 3.1, 5.1, 5.2, 5.3_

  - [x] 2.4 Validate server startup
    - Confirm `EVALUATOR_APP_PORT` is read and passed to `e.Start()`
    - Confirm `middleware.RequestLogger()` is applied
    - Confirm error logging on startup failure
    - _Requirements: 6.1, 6.2, 6.3_

- [x] 3. Checkpoint - Ensure the build passes
  - Run `make lint` and `make test` from `ms-transaction-evaluator/`
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Write unit test for root route removal
  - [x] 4.1 Write test verifying GET `/` returns 404
    - Create a test in `ms-transaction-evaluator/cmd/api/` or appropriate location
    - Use `httptest.NewRequest` / `httptest.NewRecorder` with Echo context
    - Assert that `GET /` returns HTTP 404 (Echo default not-found behavior)
    - _Requirements: 4.1, 4.2_

- [x] 5. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- The refactoring is minimal — only the root `/` handler and its unused import are removed
- All other code in `main.go` is already in the target state per the design document
