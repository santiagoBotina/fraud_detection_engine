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

list-records:
	aws dynamodb scan \
	  --table-name $(DYNAMO_DB_TRANSACTIONS_TABLE) \
	  --endpoint-url $(DYNAMO_DB_ENDPOINT) \
	  --region us-east-1
