.PHONY: run test test-v lint install clean list-scores

ifneq (,$(wildcard .env))
    include .env
    export
endif

install:
	pip install -r requirements.txt

run:
	uvicorn app.main:app --host 0.0.0.0 --port $(FRAUD_SCORE_APP_PORT) --reload

test:
	python -m pytest

test-v:
	python -m pytest -v

lint:
	python -m ruff check app/

clean:
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type d -name .pytest_cache -exec rm -rf {} +
	find . -type d -name .hypothesis -exec rm -rf {} +

list-scores:
	aws dynamodb scan \
	  --table-name $(DYNAMO_DB_FRAUD_SCORES_TABLE) \
	  --endpoint-url $(DYNAMO_DB_ENDPOINT) \
	  --region $(AWS_REGION)
