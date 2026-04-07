"""Unit tests for GET /scores/{transaction_id} endpoint.

Tests cache-first lookup, DynamoDB fallback, 404 handling,
Redis failure resilience, and CORS headers.

Validates: Requirements 5.1, 5.2, 5.3, 12.3
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

from fastapi.testclient import TestClient

from app.main import app


@patch("app.main._score_cache")
@patch("app.main._score_store")
def test_cache_hit_returns_score_without_dynamodb_call(
    mock_store: MagicMock,
    mock_cache: MagicMock,
) -> None:
    """When the score is in Redis cache, return it without querying DynamoDB."""
    mock_cache.get.return_value = 75

    client = TestClient(app, raise_server_exceptions=False)
    response = client.get("/scores/txn-001")

    assert response.status_code == 200
    body = response.json()
    assert body["transaction_id"] == "txn-001"
    assert body["fraud_score"] == 75
    assert body["calculated_at"] is None

    mock_cache.get.assert_called_once_with("txn-001")
    mock_store.get.assert_not_called()


@patch("app.main._score_cache")
@patch("app.main._score_store")
def test_cache_miss_falls_back_to_dynamodb(
    mock_store: MagicMock,
    mock_cache: MagicMock,
) -> None:
    """When Redis cache returns None, fall back to DynamoDB."""
    mock_cache.get.return_value = None
    mock_store.get.return_value = {
        "transaction_id": "txn-002",
        "fraud_score": 42,
        "calculated_at": "2025-01-15T10:30:02Z",
    }

    client = TestClient(app, raise_server_exceptions=False)
    response = client.get("/scores/txn-002")

    assert response.status_code == 200
    body = response.json()
    assert body["transaction_id"] == "txn-002"
    assert body["fraud_score"] == 42
    assert body["calculated_at"] == "2025-01-15T10:30:02Z"

    mock_cache.get.assert_called_once_with("txn-002")
    mock_store.get.assert_called_once_with("txn-002")


@patch("app.main._score_cache")
@patch("app.main._score_store")
def test_not_found_in_cache_and_dynamodb_returns_404(
    mock_store: MagicMock,
    mock_cache: MagicMock,
) -> None:
    """When score is not in Redis or DynamoDB, return 404."""
    mock_cache.get.return_value = None
    mock_store.get.return_value = None

    client = TestClient(app, raise_server_exceptions=False)
    response = client.get("/scores/txn-missing")

    assert response.status_code == 404
    body = response.json()
    assert "txn-missing" in body["detail"]


@patch("app.main._score_cache")
@patch("app.main._score_store")
def test_redis_failure_falls_back_to_dynamodb_silently(
    mock_store: MagicMock,
    mock_cache: MagicMock,
) -> None:
    """When Redis raises an exception, silently fall back to DynamoDB."""
    mock_cache.get.side_effect = ConnectionError("Redis unavailable")
    mock_store.get.return_value = {
        "transaction_id": "txn-003",
        "fraud_score": 88,
        "calculated_at": "2025-01-15T11:00:00Z",
    }

    client = TestClient(app, raise_server_exceptions=False)
    response = client.get("/scores/txn-003")

    assert response.status_code == 200
    body = response.json()
    assert body["transaction_id"] == "txn-003"
    assert body["fraud_score"] == 88
    assert body["calculated_at"] == "2025-01-15T11:00:00Z"

    mock_store.get.assert_called_once_with("txn-003")


def test_cors_headers_present() -> None:
    """CORS headers allow the dashboard origin."""
    client = TestClient(app, raise_server_exceptions=False)
    response = client.options(
        "/scores/txn-any",
        headers={
            "Origin": "http://localhost:5173",
            "Access-Control-Request-Method": "GET",
        },
    )

    assert response.headers.get("access-control-allow-origin") == "http://localhost:5173"
