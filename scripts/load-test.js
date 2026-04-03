import http from "k6/http";
import { check, sleep } from "k6";

function randomIntBetween(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

// Configure stages: ramp up, sustain, ramp down
export const options = {
  stages: [
    { duration: "30s", target: 10 },  // ramp up to 10 VUs
    { duration: "1m", target: 10 },   // hold at 10 VUs
    { duration: "30s", target: 30 },  // ramp up to 30 VUs
    { duration: "1m", target: 30 },   // hold at 30 VUs
    { duration: "30s", target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<2000"], // 95% of requests under 2s
    http_req_failed: ["rate<0.05"],    // less than 5% errors
  },
};

const BASE_URL = __ENV.BASE_URL || "http://ms-transaction-evaluator:3000";

const currencies = ["USD", "COP", "EUR"];
const paymentMethods = ["CARD", "BANK_TRANSFER", "CRYPTO"];

function randomPayload() {
  return JSON.stringify({
    amount_in_cents: randomIntBetween(100, 100000),
    currency: currencies[randomIntBetween(0, currencies.length - 1)],
    payment_method:
      paymentMethods[randomIntBetween(0, paymentMethods.length - 1)],
    customer: {
      customer_id: `cust_${randomIntBetween(1, 1000)}`,
      name: `User ${randomIntBetween(1, 1000)}`,
      email: `user${randomIntBetween(1, 1000)}@example.com`,
      phone: `+1${randomIntBetween(1000000000, 9999999999)}`,
      ip_address: `${randomIntBetween(1, 255)}.${randomIntBetween(0, 255)}.${randomIntBetween(0, 255)}.${randomIntBetween(1, 255)}`,
    },
  });
}

export default function () {
  // POST /evaluate — submit a transaction
  const evalRes = http.post(`${BASE_URL}/evaluate`, randomPayload(), {
    headers: { "Content-Type": "application/json" },
    tags: { name: "POST /evaluate" },
  });

  check(evalRes, {
    "evaluate status is 200": (r) => r.status === 200,
    "evaluate has transaction id": (r) => {
      try {
        return JSON.parse(r.body).data.id !== undefined;
      } catch {
        return false;
      }
    },
  });

  // GET /transactions — list transactions
  const listRes = http.get(`${BASE_URL}/transactions?limit=10`, {
    tags: { name: "GET /transactions" },
  });

  check(listRes, {
    "list status is 200": (r) => r.status === 200,
  });

  // GET /transactions/:id — get a specific transaction (if evaluate succeeded)
  if (evalRes.status === 200) {
    try {
      const txnId = JSON.parse(evalRes.body).data.id;
      const detailRes = http.get(`${BASE_URL}/transactions/${txnId}`, {
        tags: { name: "GET /transactions/:id" },
      });

      check(detailRes, {
        "detail status is 200": (r) => r.status === 200,
      });
    } catch (_) {
      // skip if parse fails
    }
  }

  sleep(randomIntBetween(1, 3));
}
