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

get-transaction:
	@test -n "$(ID)" || (echo "Usage: make get-transaction ID=<transaction-id>" && exit 1)
	aws dynamodb get-item \
	  --table-name $(DYNAMO_DB_TRANSACTIONS_TABLE) \
	  --key '{"id": {"S": "$(ID)"}}' \
	  --endpoint-url $(DYNAMO_DB_ENDPOINT) \
	  --region us-east-1
