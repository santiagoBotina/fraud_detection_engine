"""Property-based tests for fraud score retrieval endpoint.

Uses hypothesis to verify correctness properties of the
GET /scores/{transaction_id} endpoint.

Validates: Requirements 5.1, 5.3
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

from fastapi.testclient import TestClient
from hypothesis import given, settings
from hypothesis import strategies as st

from app.main import app

# Strategy for transaction IDs: non-empty alphanumeric strings with hyphens/underscores
transaction_id_strategy = st.from_regex(r"txn-[a-z0-9]{1,20}", fullmatch=True)

# Strategy for fraud scores: integers in [0, 100]
fraud_score_strategy = st.integers(min_value=0, max_value=100)

# Strategy for ISO 8601 timestamps
calculated_at_strategy = st.datetimes().map(lambda dt: dt.isoformat() + "Z")


# Feature: fraud-analyst-dashboard, Property 7: Fraud score retrieval round-trip
class TestFraudScoreRetrievalRoundTrip:
    """**Validates: Requirements 5.1**"""

    @settings(max_examples=100)
    @given(
        txn_id=transaction_id_strategy,
        score=fraud_score_strategy,
        calculated_at=calculated_at_strategy,
    )
    @patch("app.main._score_cache")
    @patch("app.main._score_store")
    def test_stored_score_matches_retrieved_score(
        self,
        mock_store: MagicMock,
        mock_cache: MagicMock,
        txn_id: str,
        score: int,
        calculated_at: str,
    ) -> None:
        """For any fraud score persisted in DynamoDB, a GET request to
        /scores/{transaction_id} returns a record whose transaction_id,
        fraud_score, and calculated_at fields match the stored values."""
        # Force DynamoDB path: cache returns None
        mock_cache.get.return_value = None
        mock_store.get.return_value = {
            "transaction_id": txn_id,
            "fraud_score": score,
            "calculated_at": calculated_at,
        }

        client = TestClient(app, raise_server_exceptions=False)
        response = client.get(f"/scores/{txn_id}")

        assert response.status_code == 200
        body = response.json()
        assert body["transaction_id"] == txn_id
        assert body["fraud_score"] == score
        assert body["calculated_at"] == calculated_at


# Feature: fraud-analyst-dashboard, Property 8: Cache-first score lookup
class TestCacheFirstScoreLookup:
    """**Validates: Requirements 5.3**"""

    @settings(max_examples=100)
    @given(
        txn_id=transaction_id_strategy,
        score=fraud_score_strategy,
    )
    @patch("app.main._score_cache")
    @patch("app.main._score_store")
    def test_dynamodb_not_called_when_cache_returns_value(
        self,
        mock_store: MagicMock,
        mock_cache: MagicMock,
        txn_id: str,
        score: int,
    ) -> None:
        """For any transaction ID whose fraud score exists in the Redis cache,
        a GET request to /scores/{transaction_id} returns the cached score
        without querying DynamoDB."""
        mock_cache.get.return_value = score

        client = TestClient(app, raise_server_exceptions=False)
        response = client.get(f"/scores/{txn_id}")

        assert response.status_code == 200
        body = response.json()
        assert body["transaction_id"] == txn_id
        assert body["fraud_score"] == score

        # DynamoDB must never be called when cache has the value
        mock_store.get.assert_not_called()
