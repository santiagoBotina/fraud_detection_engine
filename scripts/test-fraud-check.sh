#!/usr/bin/env bash
# Test: Transaction that should trigger FRAUD_CHECK
# Medium-value USD CARD payment triggers rule-003 (Priority 3) → FRAUD_CHECK
set -euo pipefail

BASE_URL="${EVALUATOR_URL:-http://localhost:3000}"

echo "=== Testing FRAUD_CHECK transaction ==="
echo "Payload: \$10,000 USD via CARD (triggers rule-003: Amount > \$5,000)"
echo ""

curl -s -X POST "${BASE_URL}/evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_in_cents": 1000000,
    "currency": "USD",
    "payment_method": "CARD",
    "customer": {
      "customer_id": "cust_fraud_check_test",
      "name": "Dave Suspicious",
      "email": "dave@example.com",
      "phone": "+1555000444",
      "ip_address": "203.0.113.42"
    }
  }' | python3 -m json.tool

echo ""
echo "Expected: Transaction saved with status PENDING, then decision-service evaluates → FRAUD_CHECK"
echo "(Matches rule-003: amount_in_cents > 500,000 → FRAUD_CHECK)"
echo "The transaction is then published to FraudSignals.Request topic for fraud signal processing."
echo ""
echo "--- Alternative: COP currency ---"
echo "Payload: \$200 COP via CARD (triggers rule-004: COP currency)"
echo ""

curl -s -X POST "${BASE_URL}/evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_in_cents": 20000,
    "currency": "COP",
    "payment_method": "CARD",
    "customer": {
      "customer_id": "cust_fraud_check_test_2",
      "name": "Eva Colombian",
      "email": "eva@example.com",
      "phone": "+5712345678",
      "ip_address": "181.49.100.1"
    }
  }' | python3 -m json.tool

echo ""
echo "Expected: → FRAUD_CHECK (Matches rule-004: currency EQUAL COP → FRAUD_CHECK)"
