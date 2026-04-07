"""Seed script for Qdrant transaction-embeddings collection.

Creates the 'transaction-embeddings' collection and inserts sample
transaction embeddings for local development.

Implements Requirements 8.1, 8.2, 8.3.

Usage:
    python scripts/seed-qdrant.py
"""

import os

from qdrant_client import QdrantClient
from qdrant_client.models import Distance, PointStruct, VectorParams

QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", "6333"))
COLLECTION_NAME = "transaction-embeddings"
VECTOR_DIMENSION = 3

# Sample transactions: [amount_normalized, payment_risk_norm, ip_risk_norm]
# amount_normalized = min(amount_in_cents / 1_000_000, 1.0)
# payment_risk_norm = PAYMENT_METHOD_RISK / 10.0  (BANK_TRANSFER=0.1, CARD=0.3, CRYPTO=0.8)
# ip_risk_norm = _ip_risk_score(ip) / 10.0  (hash-based, 0.0–1.0)
SAMPLE_TRANSACTIONS = [
    {
        "id": 1,
        "vector": [0.005, 0.1, 0.2],
        "payload": {
            "transaction_id": "txn-hist-001",
            "status": "approved",
            "amount_in_cents": 5000,
            "payment_method": "BANK_TRANSFER",
            "customer_id": "cust_201",
        },
    },
    {
        "id": 2,
        "vector": [0.015, 0.3, 0.3],
        "payload": {
            "transaction_id": "txn-hist-002",
            "status": "approved",
            "amount_in_cents": 15000,
            "payment_method": "CARD",
            "customer_id": "cust_202",
        },
    },
    {
        "id": 3,
        "vector": [0.95, 0.8, 0.9],
        "payload": {
            "transaction_id": "txn-hist-003",
            "status": "declined",
            "amount_in_cents": 950000,
            "payment_method": "CRYPTO",
            "customer_id": "cust_203",
        },
    },
    {
        "id": 4,
        "vector": [0.75, 0.8, 0.7],
        "payload": {
            "transaction_id": "txn-hist-004",
            "status": "declined",
            "amount_in_cents": 750000,
            "payment_method": "CRYPTO",
            "customer_id": "cust_204",
        },
    },
    {
        "id": 5,
        "vector": [0.05, 0.3, 0.4],
        "payload": {
            "transaction_id": "txn-hist-005",
            "status": "approved",
            "amount_in_cents": 50000,
            "payment_method": "CARD",
            "customer_id": "cust_205",
        },
    },
    {
        "id": 6,
        "vector": [0.2, 0.1, 0.1],
        "payload": {
            "transaction_id": "txn-hist-006",
            "status": "approved",
            "amount_in_cents": 200000,
            "payment_method": "BANK_TRANSFER",
            "customer_id": "cust_206",
        },
    },
    {
        "id": 7,
        "vector": [0.5, 0.3, 0.6],
        "payload": {
            "transaction_id": "txn-hist-007",
            "status": "declined",
            "amount_in_cents": 500000,
            "payment_method": "CARD",
            "customer_id": "cust_207",
        },
    },
    {
        "id": 8,
        "vector": [1.0, 0.8, 0.8],
        "payload": {
            "transaction_id": "txn-hist-008",
            "status": "declined",
            "amount_in_cents": 1000000,
            "payment_method": "CRYPTO",
            "customer_id": "cust_208",
        },
    },
    {
        "id": 9,
        "vector": [0.01, 0.3, 0.5],
        "payload": {
            "transaction_id": "txn-hist-009",
            "status": "approved",
            "amount_in_cents": 10000,
            "payment_method": "CARD",
            "customer_id": "cust_209",
        },
    },
    {
        "id": 10,
        "vector": [0.3, 0.1, 0.3],
        "payload": {
            "transaction_id": "txn-hist-010",
            "status": "approved",
            "amount_in_cents": 300000,
            "payment_method": "BANK_TRANSFER",
            "customer_id": "cust_210",
        },
    },
    {
        "id": 11,
        "vector": [0.6, 0.8, 0.5],
        "payload": {
            "transaction_id": "txn-hist-011",
            "status": "declined",
            "amount_in_cents": 600000,
            "payment_method": "CRYPTO",
            "customer_id": "cust_211",
        },
    },
    {
        "id": 12,
        "vector": [0.1, 0.3, 0.2],
        "payload": {
            "transaction_id": "txn-hist-012",
            "status": "approved",
            "amount_in_cents": 100000,
            "payment_method": "CARD",
            "customer_id": "cust_212",
        },
    },
]


def main() -> None:
    print(f"Connecting to Qdrant at {QDRANT_HOST}:{QDRANT_PORT}...")
    client = QdrantClient(host=QDRANT_HOST, port=QDRANT_PORT)

    # Recreate collection (drop if exists)
    collections = [c.name for c in client.get_collections().collections]
    if COLLECTION_NAME in collections:
        print(f"Dropping existing collection '{COLLECTION_NAME}'...")
        client.delete_collection(COLLECTION_NAME)

    print(f"Creating collection '{COLLECTION_NAME}' (dim={VECTOR_DIMENSION}, cosine)...")
    client.create_collection(
        collection_name=COLLECTION_NAME,
        vectors_config=VectorParams(size=VECTOR_DIMENSION, distance=Distance.COSINE),
    )

    # Insert sample points
    points = [
        PointStruct(id=t["id"], vector=t["vector"], payload=t["payload"])
        for t in SAMPLE_TRANSACTIONS
    ]
    client.upsert(collection_name=COLLECTION_NAME, points=points)
    print(f"Inserted {len(points)} sample transaction embeddings.")
    print("Qdrant seed complete.")


if __name__ == "__main__":
    main()
