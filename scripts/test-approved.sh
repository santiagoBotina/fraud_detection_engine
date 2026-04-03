#!/usr/bin/env bash
# Test: Transaction that should be APPROVED
# Small USD CARD payment — no rule matches, so fail-open → APPROVED
set -euo pipefail

BASE_URL="${EVALUATOR_URL:-http://localhost:3000}"

echo "=== Testing APPROVED transaction ==="
echo "Payload: \$150 USD via CARD (no rule triggers)"
echo ""

curl -s -X POST "${BASE_URL}/evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_in_cents": 15000,
    "currency": "USD",
    "payment_method": "CARD",
    "customer": {
      "customer_id": "cust_approved_test",
      "name": "Jane Doe",
      "email": "jane@example.com",
      "phone": "+1555000111",
      "ip_address": "192.168.1.50"
    }
  }' | python3 -m json.tool

echo ""
echo "Expected: Transaction saved with status PENDING, then decision-service evaluates → APPROVED"
echo "(No rule matches: amount \$150 < \$5,000 threshold, USD currency, CARD payment)"
