# Requirements Document

## Introduction

Refactor `ms-transaction-evaluator/cmd/api/main.go` to reduce its responsibilities to three focused concerns: loading infrastructure clients (DynamoDB), wiring controllers via manual dependency injection, and registering routes through controllers' `RegisterRoutes` methods. The root `/` route is removed, and Swagger documentation is preserved. The goal is a cleaner, more maintainable entrypoint that aligns with the project's hexagonal architecture conventions.

## Glossary

- **Main_Entrypoint**: The `main()` function in `ms-transaction-evaluator/cmd/api/main.go` that bootstraps the application
- **Controller**: An HTTP adapter struct in `infrastructure/adapter/in/http/` that handles requests and registers its own routes via `RegisterRoutes`
- **DynamoDB_Client**: The AWS DynamoDB client initialized with SDK v2, used by outbound adapters
- **Echo_Server**: The Echo v5 HTTP server instance created and started in the Main_Entrypoint
- **Swagger_UI**: The Swagger documentation endpoint served at `/swagger/*` via echo-swagger
- **Route_Registration**: The process by which a Controller registers its HTTP routes on the Echo_Server via its `RegisterRoutes` method

## Requirements

### Requirement 1: DynamoDB Client Initialization

**User Story:** As a developer, I want the Main_Entrypoint to initialize the DynamoDB_Client, so that outbound adapters can persist data.

#### Acceptance Criteria

1. WHEN the application starts, THE Main_Entrypoint SHALL load the AWS SDK configuration using environment variables for region
2. WHEN the `DYNAMO_DB_ENDPOINT` environment variable is set, THE Main_Entrypoint SHALL create the DynamoDB_Client with the custom endpoint
3. WHEN the `DYNAMO_DB_ENDPOINT` environment variable is empty, THE Main_Entrypoint SHALL create the DynamoDB_Client with the default AWS endpoint
4. IF the AWS SDK configuration fails to load, THEN THE Main_Entrypoint SHALL terminate the application with a fatal log message

### Requirement 2: Manual Dependency Injection

**User Story:** As a developer, I want the Main_Entrypoint to wire all dependencies manually, so that the application follows the hexagonal architecture convention of composing adapters and use cases at the entrypoint.

#### Acceptance Criteria

1. THE Main_Entrypoint SHALL instantiate repository implementations using the DynamoDB_Client
2. THE Main_Entrypoint SHALL instantiate use cases by injecting repository implementations
3. THE Main_Entrypoint SHALL instantiate each Controller by injecting the required use cases
4. THE Main_Entrypoint SHALL contain only dependency wiring logic, infrastructure client setup, and server startup

### Requirement 3: Controller Route Registration

**User Story:** As a developer, I want each Controller to register its own routes on the Echo_Server, so that the Main_Entrypoint delegates routing concerns to controllers.

#### Acceptance Criteria

1. THE Main_Entrypoint SHALL call `RegisterRoutes` on each Controller to register HTTP routes on the Echo_Server
2. THE Main_Entrypoint SHALL contain no inline route handler definitions for application endpoints

### Requirement 4: Remove Root Route

**User Story:** As a developer, I want the root `/` route removed from the Main_Entrypoint, so that the server does not expose an unnecessary welcome endpoint.

#### Acceptance Criteria

1. THE Main_Entrypoint SHALL not define a handler for the `/` path
2. WHEN a request is made to `/`, THE Echo_Server SHALL respond with its default 404 behavior

### Requirement 5: Preserve Swagger Documentation

**User Story:** As a developer, I want the Swagger_UI to remain accessible, so that API consumers can browse the API documentation.

#### Acceptance Criteria

1. THE Main_Entrypoint SHALL register the Swagger_UI route at `/swagger/*`
2. WHEN a request is made to `/swagger/index.html`, THE Echo_Server SHALL serve the Swagger documentation page
3. THE Main_Entrypoint SHALL retain the swag annotation comments for API metadata generation

### Requirement 6: Server Startup

**User Story:** As a developer, I want the Echo_Server to start on the configured port, so that the application is reachable.

#### Acceptance Criteria

1. THE Main_Entrypoint SHALL read the server port from the `EVALUATOR_APP_PORT` environment variable
2. WHEN the Echo_Server fails to start, THE Main_Entrypoint SHALL log the error
3. THE Main_Entrypoint SHALL apply the request logger middleware to the Echo_Server
