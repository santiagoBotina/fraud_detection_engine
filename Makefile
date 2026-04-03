include .env

setup: start wait-for-infra create-transactions-table create-rules-table create-rule-evaluations-table create-fraud-scores-table seed create-topics

start:
	docker compose up -d --build

wait-for-infra:
	@echo "Waiting for DynamoDB to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  docker run --rm --network fraud_detection_engine_local-network \
	    -e AWS_ACCESS_KEY_ID=dummy -e AWS_SECRET_ACCESS_KEY=dummy -e AWS_DEFAULT_REGION=us-east-1 \
	    amazon/aws-cli dynamodb list-tables --endpoint-url $(DYNAMO_DB_ENDPOINT) --region us-east-1 > /dev/null 2>&1 \
	    && echo "DynamoDB is ready." && break \
	    || (echo "  attempt $$i/10..." && sleep 2); \
	done
	@echo "Waiting for Kafka to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  docker exec $(KAFKA_CONTAINER_NAME) kafka-topics --list --bootstrap-server localhost:$(KAFKA_PORT) > /dev/null 2>&1 \
	    && echo "Kafka is ready." && break \
	    || (echo "  attempt $$i/10..." && sleep 2); \
	done

seed:
	bash scripts/seed-dynamo.sh

create-topics: create-transactions-evaluator-topic create-decision-topic create-fraud-score-topics


# === SCRIPTS TO TEST SCENARIOS ===
test-approved:
	bash scripts/test-approved.sh

test-declined:
	bash scripts/test-declined.sh

test-fraud-check:
	bash scripts/test-fraud-check.sh


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

create-transactions-evaluator-topic:
	docker exec $(KAFKA_CONTAINER_NAME) \
	  kafka-topics --create \
	  --topic Transaction.Created \
	  --bootstrap-server localhost:$(KAFKA_PORT) \
	  --partitions 1 \
	  --replication-factor 1


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

create-decision-topic:
	docker exec $(KAFKA_CONTAINER_NAME) \
	  kafka-topics --create \
	  --topic Decision.Calculated \
	  --bootstrap-server localhost:$(KAFKA_PORT) \
	  --partitions 1 \
	  --replication-factor 1

create-rule-evaluations-table:
	docker run --rm \
	  --network fraud_detection_engine_local-network \
	  -e AWS_ACCESS_KEY_ID=dummy \
	  -e AWS_SECRET_ACCESS_KEY=dummy \
	  -e AWS_DEFAULT_REGION=us-east-1 \
	  amazon/aws-cli dynamodb create-table \
	  --table-name $(DYNAMO_DB_RULE_EVALUATIONS_TABLE) \
	  --attribute-definitions \
	    AttributeName=transaction_id,AttributeType=S \
	    AttributeName=rule_id,AttributeType=S \
	  --key-schema \
	    AttributeName=transaction_id,KeyType=HASH \
	    AttributeName=rule_id,KeyType=RANGE \
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
	  kafka-topics --create \
	  --topic FraudScore.Request \
	  --bootstrap-server localhost:$(KAFKA_PORT) \
	  --partitions 1 \
	  --replication-factor 1
	docker exec $(KAFKA_CONTAINER_NAME) \
	  kafka-topics --create \
	  --topic FraudScore.Calculated \
	  --bootstrap-server localhost:$(KAFKA_PORT) \
	  --partitions 1 \
	  --replication-factor 1
