include .env

run:
	docker compose up -d

setup:
	create-transactions-table create-rules-table create-fraud-scores-table

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

# === FRAUD SCORE SERVICE ===
create-fraud-scores-table:
	docker run --rm \
	  --network fraud_detection_engine_local-network \
	  -e AWS_ACCESS_KEY_ID=dummy \
	  -e AWS_SECRET_ACCESS_KEY=dummy \
	  -e AWS_DEFAULT_REGION=us-east-1 \
	  amazon/aws-cli dynamodb create-table \
	  --table-name $(DYNAMO_DB_FRAUD_SCORES_TABLE) \
	  --attribute-definitions \
	    AttributeName=transaction_id,AttributeType=S \
	  --key-schema \
	    AttributeName=transaction_id,KeyType=HASH \
	  --billing-mode PAY_PER_REQUEST \
	  --endpoint-url $(DYNAMO_DB_ENDPOINT) \
	  --region us-east-1


create-fraud-score-topics:
	docker exec $(KAFKA_CONTAINER_NAME) \
	  kafka-topics --create\
	  --topic FraudScore.Request \
	  --bootstrap-server localhost:$(KAFKA_PORT) \
	  --partitions 1 \
	  --replication-factor 1 && \
	docker exec $(KAFKA_CONTAINER_NAME) \
	  kafka-topics --create\
	  --topic FraudScore.Calculated \
	  --bootstrap-server localhost:$(KAFKA_PORT) \
	  --partitions 1 \
	  --replication-factor 1

