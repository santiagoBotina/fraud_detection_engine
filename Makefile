include .env

run:
	docker compose up -d

# === TRANSACTION EVALUATOR ===
create_transactions_table:
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
