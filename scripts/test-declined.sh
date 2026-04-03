#!/usr/bin/env bash
# Test: Transaction that should be DECLINED
# CRYPTO payment triggers rule-001 (Priority 1) → DECLINED
set -euo pipefail

BASE_URL="${EVALUATOR_URL:-http://localhost:3000}"

echo "=== Testing DECLINED transaction ==="
echo "Payload: \$500 USD via CRYPTO (triggers rule-001: Block CRYPTO)"
echo ""

curl -s -X POST "${BASE_URL}/evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_in_cents": 50000,
    "currency": "USD",
    "payment_method": "CRYPTO",
    "customer": {
      "customer_id": "cust_declined_test",
      "name": "Bob Risky",
      "email": "bob@example.com",
      "phone": "+1555000222",
      "ip_address": "10.0.0.99"
    }
  }' | python3 -m json.tool

echo ""
echo "Expected: Transaction saved with status PENDING, then decision-service evaluates → DECLINED"
echo "(Matches rule-001: payment_method EQUAL CRYPTO → DECLINED)"
echo ""
echo "--- Alternative: High-value transaction ---"
echo "Payload: \$75,000 USD via BANK_TRANSFER (triggers rule-002: Amount > \$50,000)"
echo ""

curl -s -X POST "${BASE_URL}/evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_in_cents": 7500000,
    "currency": "USD",
    "payment_method": "BANK_TRANSFER",
    "customer": {
      "customer_id": "cust_declined_test_2",
      "name": "Charlie Whale",
      "email": "charlie@example.com",
      "phone": "+1555000333",
      "ip_address": "172.16.0.1"
    }
  }' | python3 -m json.tool

echo ""
echo "Expected: → DECLINED (Matches rule-002: amount_in_cents > 5,000,000)"
