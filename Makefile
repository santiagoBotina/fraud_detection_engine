include .env

run:
	docker compose up -d

setup:
	create-transactions-table create-rules-table

# === TRANSACTION EVALUATOR ===
create-transactions-table:
	docker run --rm \
	  --network fraud_detection_engine_local-network \
	  -e AWS_ACCESS_KEY_ID=dummy \
	  -e AWS_SECRET_ACCESS_KEY=dummy \
	  -e AWS_DEFAULT_REGION=us-east-1 \
	  amazon/aws-cli dynamodb create-table \
	  --table-name $(DYNAMO_DB_TRANSACTIONS_TABLE) \
	  --attribute-definitions \
	    AttributeName=id,AttributeType=S \
	  --key-schema \
	    AttributeName=id,KeyType=HASH \
	  --billing-mode PAY_PER_REQUEST \
	  --endpoint-url $(DYNAMO_DB_ENDPOINT) \
	  --region us-east-1

# === DECISION SERVICE ===
create-rules-table:
	docker run --rm \
	  --network fraud_detection_engine_local-network \
	  -e AWS_ACCESS_KEY_ID=dummy \
	  -e AWS_SECRET_ACCESS_KEY=dummy \
	  -e AWS_DEFAULT_REGION=us-east-1 \
	  amazon/aws-cli dynamodb create-table \
	  --table-name $(DYNAMO_DB_RULES_TABLE) \
	  --attribute-definitions \
	    AttributeName=rule_id,AttributeType=S \
	  --key-schema \
	    AttributeName=rule_id,KeyType=HASH \
	  --billing-mode PAY_PER_REQUEST \
	  --endpoint-url $(DYNAMO_DB_ENDPOINT) \
	  --region us-east-1
