.PHONY: coverage

ifneq (,$(wildcard .env))
    include .env
    export
endif

run:
	go run ./...

run-dev:
	air ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run ./... --fix

install:
	go install ./...

test:
	go test ./...

migrate:
	go run cmd/migrate/main.go

generate-migration-ddl:
	migrate create -ext sql -dir internal/infrastructure/adapter/out/postgres/migrations/ddl $(name)

generate-migration-dml:
	migrate create -ext sql -dir internal/infrastructure/adapter/out/postgres/migrations/dml $(name)

run-migrations-ddl:
	migrate -path internal/infrastructure/adapter/out/postgres/migrations/ddl -database "postgres://$(DB_WRITER_USER):$(DB_WRITER_PASSWORD)@$(DB_WRITER_HOST)/$(DB_DATABASE_NAME)?sslmode=disable" up

run-migrations-dml:
	migrate -path internal/infrastructure/adapter/out/postgres/migrations/dml -database "postgres://$(DB_WRITER_USER):$(DB_WRITER_PASSWORD)@$(DB_WRITER_HOST)/$(DB_DATABASE_NAME)?sslmode=disable" up
