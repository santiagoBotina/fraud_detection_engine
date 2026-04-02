.PHONY: run run-dev lint lint-fix test list-rules

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

test:
	go test ./...

list-rules:
	aws dynamodb scan \
	  --table-name $(DYNAMO_DB_RULES_TABLE) \
	  --endpoint-url $(DYNAMO_DB_ENDPOINT) \
	  --region us-east-1
