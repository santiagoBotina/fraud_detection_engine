# Product Overview

Fraud Detection Engine — a backend system for evaluating financial transactions and detecting potential fraud.

The system currently exposes a Transaction Evaluator microservice (`ms-transaction-evaluator`) that:
- Accepts transaction evaluation requests via a REST API
- Validates transaction payloads (amount, currency, payment method, customer info)
- Persists transactions to DynamoDB
- Supports currencies: USD, COP, EUR
- Supports payment methods: CARD, BANK_TRANSFER, CRYPTO

The infrastructure includes Kafka (with Zookeeper) for event streaming, though it is not yet wired into the application code. This suggests future plans for event-driven fraud detection pipelines.
