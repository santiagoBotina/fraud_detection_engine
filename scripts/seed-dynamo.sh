#!/usr/bin/env bash
# Seed all DynamoDB tables for local testing.
# Requires: docker running, DynamoDB local on port 8000, tables already created (make setup).

set -euo pipefail

ENDPOINT="http://localhost:8000"
REGION="us-east-1"

aws="docker run --rm --network fraud_detection_engine_local-network \
  -e AWS_ACCESS_KEY_ID=dummy \
  -e AWS_SECRET_ACCESS_KEY=dummy \
  -e AWS_DEFAULT_REGION=${REGION} \
  amazon/aws-cli"

echo "=== Seeding ddb-rules ==="

# Rule 1 (Priority 1): CRYPTO payments → DECLINED
$aws dynamodb put-item \
  --table-name ddb-rules \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item '{
    "rule_id":            {"S": "rule-001"},
    "rule_name":          {"S": "Block CRYPTO payments"},
    "condition_field":    {"S": "payment_method"},
    "condition_operator": {"S": "EQUAL"},
    "condition_value":    {"S": "CRYPTO"},
    "result_status":      {"S": "DECLINED"},
    "priority":           {"N": "1"},
    "is_active":          {"BOOL": true}
  }'

# Rule 2 (Priority 2): Amount > 5,000,000 cents ($50,000) → DECLINED
$aws dynamodb put-item \
  --table-name ddb-rules \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item '{
    "rule_id":            {"S": "rule-002"},
    "rule_name":          {"S": "Decline high-value transactions"},
    "condition_field":    {"S": "amount_in_cents"},
    "condition_operator": {"S": "GREATER_THAN"},
    "condition_value":    {"S": "5000000"},
    "result_status":      {"S": "DECLINED"},
    "priority":           {"N": "2"},
    "is_active":          {"BOOL": true}
  }'

# Rule 3 (Priority 3): Amount > 500,000 cents ($5,000) → FRAUD_CHECK
$aws dynamodb put-item \
  --table-name ddb-rules \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item '{
    "rule_id":            {"S": "rule-003"},
    "rule_name":          {"S": "Fraud check medium-value transactions"},
    "condition_field":    {"S": "amount_in_cents"},
    "condition_operator": {"S": "GREATER_THAN"},
    "condition_value":    {"S": "500000"},
    "result_status":      {"S": "FRAUD_CHECK"},
    "priority":           {"N": "3"},
    "is_active":          {"BOOL": true}
  }'

# Rule 4 (Priority 4): COP currency → FRAUD_CHECK
$aws dynamodb put-item \
  --table-name ddb-rules \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item '{
    "rule_id":            {"S": "rule-004"},
    "rule_name":          {"S": "Fraud check COP transactions"},
    "condition_field":    {"S": "currency"},
    "condition_operator": {"S": "EQUAL"},
    "condition_value":    {"S": "COP"},
    "result_status":      {"S": "FRAUD_CHECK"},
    "priority":           {"N": "4"},
    "is_active":          {"BOOL": true}
  }'

# Rule 5 (Priority 10): Fraud score >= 80 → DECLINED
$aws dynamodb put-item \
  --table-name ddb-rules \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item '{
    "rule_id":            {"S": "rule-005"},
    "rule_name":          {"S": "Decline high fraud score"},
    "condition_field":    {"S": "fraud_score"},
    "condition_operator": {"S": "GREATER_THAN_OR_EQUAL"},
    "condition_value":    {"S": "80"},
    "result_status":      {"S": "DECLINED"},
    "priority":           {"N": "10"},
    "is_active":          {"BOOL": true}
  }'

# Rule 6 (Priority 11): Fraud score >= 50 → DECLINED
$aws dynamodb put-item \
  --table-name ddb-rules \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item '{
    "rule_id":            {"S": "rule-006"},
    "rule_name":          {"S": "Decline medium fraud score"},
    "condition_field":    {"S": "fraud_score"},
    "condition_operator": {"S": "GREATER_THAN_OR_EQUAL"},
    "condition_value":    {"S": "50"},
    "result_status":      {"S": "DECLINED"},
    "priority":           {"N": "11"},
    "is_active":          {"BOOL": true}
  }'

# Rule 7 (Priority 12): Fraud score < 50 → APPROVED
$aws dynamodb put-item \
  --table-name ddb-rules \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item '{
    "rule_id":            {"S": "rule-007"},
    "rule_name":          {"S": "Approve low fraud score"},
    "condition_field":    {"S": "fraud_score"},
    "condition_operator": {"S": "LESS_THAN"},
    "condition_value":    {"S": "50"},
    "result_status":      {"S": "APPROVED"},
    "priority":           {"N": "12"},
    "is_active":          {"BOOL": true}
  }'

# Inactive rule (should be ignored by the engine)
$aws dynamodb put-item \
  --table-name ddb-rules \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item '{
    "rule_id":            {"S": "rule-008"},
    "rule_name":          {"S": "Block BANK_TRANSFER (disabled)"},
    "condition_field":    {"S": "payment_method"},
    "condition_operator": {"S": "EQUAL"},
    "condition_value":    {"S": "BANK_TRANSFER"},
    "result_status":      {"S": "DECLINED"},
    "priority":           {"N": "0"},
    "is_active":          {"BOOL": false}
  }'

echo "  ✓ 8 rules seeded (7 active, 1 inactive)"

echo ""
echo "=== Seeding ddb-transactions (sample transactions) ==="

NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# A small approved transaction
$aws dynamodb put-item \
  --table-name ddb-transactions \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item "{
    \"id\":                {\"S\": \"txn-seed-001\"},
    \"amount_in_cents\":   {\"N\": \"15000\"},
    \"currency\":          {\"S\": \"USD\"},
    \"payment_method\":    {\"S\": \"CARD\"},
    \"customer_id\":       {\"S\": \"cust_100\"},
    \"customer_name\":     {\"S\": \"Alice Johnson\"},
    \"customer_email\":    {\"S\": \"alice@example.com\"},
    \"customer_phone\":    {\"S\": \"+1234567890\"},
    \"customer_ip_address\":{\"S\": \"192.168.1.10\"},
    \"status\":            {\"S\": \"DECLINED\"},
    \"created_at\":        {\"S\": \"$NOW\"},
    \"updated_at\":        {\"S\": \"$NOW\"}
  }"

# A high-value declined transaction
$aws dynamodb put-item \
  --table-name ddb-transactions \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item "{
    \"id\":                {\"S\": \"txn-seed-002\"},
    \"amount_in_cents\":   {\"N\": \"7500000\"},
    \"currency\":          {\"S\": \"USD\"},
    \"payment_method\":    {\"S\": \"BANK_TRANSFER\"},
    \"customer_id\":       {\"S\": \"cust_200\"},
    \"customer_name\":     {\"S\": \"Bob Smith\"},
    \"customer_email\":    {\"S\": \"bob@example.com\"},
    \"customer_phone\":    {\"S\": \"+0987654321\"},
    \"customer_ip_address\":{\"S\": \"10.0.0.5\"},
    \"status\":            {\"S\": \"APPROVED\"},
    \"created_at\":        {\"S\": \"$NOW\"},
    \"updated_at\":        {\"S\": \"$NOW\"}
  }"

echo "  ✓ 2 sample transactions seeded"

echo ""
echo "=== Seeding ddb-fraud-scores (mock fraud score results) ==="

# Low fraud score (would be APPROVED)
$aws dynamodb put-item \
  --table-name ddb-fraud-scores \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item "{
    \"transaction_id\":  {\"S\": \"txn-seed-fs-001\"},
    \"fraud_score\":     {\"N\": \"15\"},
    \"calculated_at\":   {\"S\": \"$NOW\"}
  }"

# Medium fraud score (would trigger FRAUD_CHECK)
$aws dynamodb put-item \
  --table-name ddb-fraud-scores \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item "{
    \"transaction_id\":  {\"S\": \"txn-seed-fs-002\"},
    \"fraud_score\":     {\"N\": \"65\"},
    \"calculated_at\":   {\"S\": \"$NOW\"}
  }"

# High fraud score (would be DECLINED)
$aws dynamodb put-item \
  --table-name ddb-fraud-scores \
  --endpoint-url http://dynamodb:8000 \
  --region $REGION \
  --item "{
    \"transaction_id\":  {\"S\": \"txn-seed-fs-003\"},
    \"fraud_score\":     {\"N\": \"92\"},
    \"calculated_at\":   {\"S\": \"$NOW\"}
  }"

echo "  ✓ 3 fraud score records seeded"

echo ""
echo "=== Seed complete ==="
echo ""
echo "Rules summary (by priority):"
echo "  P1  rule-001  CRYPTO payment        → DECLINED"
echo "  P2  rule-002  Amount > \$50,000      → DECLINED"
echo "  P3  rule-003  Amount > \$5,000       → FRAUD_CHECK"
echo "  P4  rule-004  COP currency           → FRAUD_CHECK"
echo "  P10 rule-005  Fraud score >= 80      → DECLINED"
echo "  P11 rule-006  Fraud score >= 50      → DECLINED"
echo "  P12 rule-007  Fraud score < 50       → APPROVED"
echo "  --  rule-008  BANK_TRANSFER (inactive)"
