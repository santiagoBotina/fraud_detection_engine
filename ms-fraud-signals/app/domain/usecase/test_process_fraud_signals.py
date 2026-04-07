"""Unit tests for ProcessFraudSignalsUseCase.

Tests orchestration with mocked pipeline, store, cache, publisher.
Tests resilience when DynamoDB or Redis fail.
Tests that signals_summary JSON is included in DynamoDB save call.

Validates: Requirements 5.1, 5.2, 5.3, 10.1, 10.2, 10.3
"""

from __future__ import annotations

import json
from datetime import datetime
from unittest.mock import MagicMock

from app.domain.entity.fraud_signal_request import FraudSignalRequest
from app.domain.entity.fraud_signal_result import FraudSignalResult
from app.domain.entity.signal_context import SignalContext
from app.domain.entity.signal_result import SignalResult
from app.domain.port.score_cache import ScoreCache
from app.domain.port.score_publisher import ScorePublisher
from app.domain.port.score_store import ScoreStore
from app.domain.service.signal_pipeline import SignalPipeline
from app.domain.usecase.process_fraud_signals import ProcessFraudSignalsUseCase


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_request() -> FraudSignalRequest:
    return FraudSignalRequest(
        transaction_id="txn-uc-001",
        amount_in_cents=50000,
        currency="USD",
        payment_method="CARD",
        customer_id="cust-1",
        customer_ip_address="10.0.0.1",
        timestamp="2025-01-15T10:30:00Z",
    )


def _fraud_score_result(value: float = 55.0) -> SignalResult:
    return SignalResult(
        signal_id="fraud-score",
        executed=True,
        value=value,
        metadata={},
        skip_reason=None,
        error=None,
    )


def _similarity_result(value: float = -5.0) -> SignalResult:
    return SignalResult(
        signal_id="similarity",
        executed=True,
        value=value,
        metadata={"matched_transactions": ["txn-a", "txn-b"]},
        skip_reason=None,
        error=None,
    )


def _similarity_skipped() -> SignalResult:
    return SignalResult(
        signal_id="similarity",
        executed=False,
        value=None,
        metadata={},
        skip_reason="score-outside-neutral-range",
        error=None,
    )


def _build_pipeline_mock(results: dict[str, SignalResult]) -> MagicMock:
    """Return a MagicMock for SignalPipeline whose run() populates context.results."""
    pipeline = MagicMock(spec=SignalPipeline)

    def _run(context: SignalContext) -> SignalContext:
        context.results = dict(results)
        return context

    pipeline.run.side_effect = _run
    return pipeline


class MockScoreStore(ScoreStore):
    def __init__(self) -> None:
        self.calls: list[tuple] = []

    def save(self, transaction_id: str, score: int, calculated_at: datetime, signals_summary: str = "") -> None:
        self.calls.append((transaction_id, score, calculated_at, signals_summary))

    def get(self, transaction_id: str) -> dict | None:
        return None


class MockScoreCache(ScoreCache):
    def __init__(self) -> None:
        self.calls: list[tuple[str, int]] = []

    def set(self, transaction_id: str, score: int) -> None:
        self.calls.append((transaction_id, score))

    def get(self, transaction_id: str) -> int | None:
        return None


class MockScorePublisher(ScorePublisher):
    def __init__(self) -> None:
        self.calls: list[FraudSignalResult] = []

    def publish(self, result: FraudSignalResult) -> None:
        self.calls.append(result)


class FailingScoreStore(ScoreStore):
    """ScoreStore that always raises on save(), simulating DynamoDB failure."""

    def save(self, transaction_id: str, score: int, calculated_at: datetime, signals_summary: str = "") -> None:
        raise RuntimeError("DynamoDB write failed")

    def get(self, transaction_id: str) -> dict | None:
        raise RuntimeError("DynamoDB read failed")


class FailingScoreCache(ScoreCache):
    """ScoreCache that always raises on set(), simulating Redis failure."""

    def set(self, transaction_id: str, score: int) -> None:
        raise ConnectionError("Redis unavailable")

    def get(self, transaction_id: str) -> int | None:
        raise ConnectionError("Redis unavailable")


# ---------------------------------------------------------------------------
# Test: Orchestration with mocked pipeline, store, cache, publisher
# Validates: Requirements 5.1, 5.2, 5.3
# ---------------------------------------------------------------------------


class TestOrchestration:
    """Verify the use case orchestrates pipeline → store → cache → publish."""

    def test_happy_path_with_similarity_executed(self) -> None:
        """When both signals execute, final score = clamp(base + adjustment, 0, 100).

        Validates: Requirements 5.1, 5.2
        """
        results = {
            "fraud-score": _fraud_score_result(55.0),
            "similarity": _similarity_result(-5.0),
        }
        pipeline = _build_pipeline_mock(results)
        store = MockScoreStore()
        cache = MockScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        result = uc.execute(_make_request())

        assert result.fraud_score == 50  # 55 + (-5) = 50
        assert result.transaction_id == "txn-uc-001"

        # Store, cache, publisher all called once
        assert len(store.calls) == 1
        assert len(cache.calls) == 1
        assert len(publisher.calls) == 1

    def test_happy_path_with_similarity_skipped(self) -> None:
        """When similarity is skipped, final score = base score.

        Validates: Requirements 5.3
        """
        results = {
            "fraud-score": _fraud_score_result(80.0),
            "similarity": _similarity_skipped(),
        }
        pipeline = _build_pipeline_mock(results)
        store = MockScoreStore()
        cache = MockScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        result = uc.execute(_make_request())

        assert result.fraud_score == 80  # base only, no adjustment

    def test_final_score_clamped_to_100(self) -> None:
        """Final score cannot exceed 100.

        Validates: Requirement 5.2
        """
        results = {
            "fraud-score": _fraud_score_result(95.0),
            "similarity": _similarity_result(20.0),
        }
        pipeline = _build_pipeline_mock(results)
        store = MockScoreStore()
        cache = MockScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        result = uc.execute(_make_request())

        assert result.fraud_score == 100  # 95 + 20 = 115 → clamped to 100

    def test_final_score_clamped_to_0(self) -> None:
        """Final score cannot go below 0.

        Validates: Requirement 5.2
        """
        results = {
            "fraud-score": _fraud_score_result(5.0),
            "similarity": _similarity_result(-20.0),
        }
        pipeline = _build_pipeline_mock(results)
        store = MockScoreStore()
        cache = MockScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        result = uc.execute(_make_request())

        assert result.fraud_score == 0  # 5 + (-20) = -15 → clamped to 0

    def test_published_result_matches_returned_result(self) -> None:
        """The result published to Kafka matches what execute() returns."""
        results = {
            "fraud-score": _fraud_score_result(60.0),
            "similarity": _similarity_result(10.0),
        }
        pipeline = _build_pipeline_mock(results)
        store = MockScoreStore()
        cache = MockScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        result = uc.execute(_make_request())

        published = publisher.calls[0]
        assert published.transaction_id == result.transaction_id
        assert published.fraud_score == result.fraud_score
        assert published.calculated_at == result.calculated_at


# ---------------------------------------------------------------------------
# Test: DynamoDB failure does not prevent caching and publishing
# Validates: Requirement 10.1
# ---------------------------------------------------------------------------


class TestDynamoDBFailureResilience:
    """DynamoDB store failure is non-fatal; caching and publishing still happen."""

    def test_dynamodb_failure_still_caches_and_publishes(self) -> None:
        """Validates: Requirement 10.1"""
        results = {
            "fraud-score": _fraud_score_result(50.0),
            "similarity": _similarity_result(5.0),
        }
        pipeline = _build_pipeline_mock(results)
        store = FailingScoreStore()
        cache = MockScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        result = uc.execute(_make_request())

        # Cache was still called
        assert len(cache.calls) == 1
        assert cache.calls[0] == ("txn-uc-001", result.fraud_score)

        # Publisher was still called
        assert len(publisher.calls) == 1
        assert publisher.calls[0].fraud_score == result.fraud_score


# ---------------------------------------------------------------------------
# Test: Redis failure does not prevent storing and publishing
# Validates: Requirement 10.2
# ---------------------------------------------------------------------------


class TestRedisFailureResilience:
    """Redis cache failure is non-fatal; storing and publishing still happen."""

    def test_redis_failure_still_stores_and_publishes(self) -> None:
        """Validates: Requirement 10.2"""
        results = {
            "fraud-score": _fraud_score_result(45.0),
            "similarity": _similarity_result(-10.0),
        }
        pipeline = _build_pipeline_mock(results)
        store = MockScoreStore()
        cache = FailingScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        result = uc.execute(_make_request())

        # Store was still called
        assert len(store.calls) == 1
        assert store.calls[0][0] == "txn-uc-001"
        assert store.calls[0][1] == result.fraud_score

        # Publisher was still called
        assert len(publisher.calls) == 1
        assert publisher.calls[0].fraud_score == result.fraud_score


# ---------------------------------------------------------------------------
# Test: signals_summary JSON is included in DynamoDB save call
# Validates: Requirement 10.3
# ---------------------------------------------------------------------------


class TestSignalsSummaryPersistence:
    """Verify that signals_summary JSON is passed to store.save()."""

    def test_signals_summary_included_in_save(self) -> None:
        """Validates: Requirement 10.3"""
        results = {
            "fraud-score": _fraud_score_result(60.0),
            "similarity": _similarity_result(-8.0),
        }
        pipeline = _build_pipeline_mock(results)
        store = MockScoreStore()
        cache = MockScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        uc.execute(_make_request())

        # Fourth argument to store.save is signals_summary
        signals_summary_json = store.calls[0][3]
        summary = json.loads(signals_summary_json)

        assert isinstance(summary, list)
        assert len(summary) == 2

        signal_ids = [s["signal_id"] for s in summary]
        assert "fraud-score" in signal_ids
        assert "similarity" in signal_ids

        for entry in summary:
            assert "signal_id" in entry
            assert "executed" in entry
            assert "value" in entry

    def test_signals_summary_with_skipped_signal(self) -> None:
        """When similarity is skipped, signals_summary still includes it with executed=False."""
        results = {
            "fraud-score": _fraud_score_result(85.0),
            "similarity": _similarity_skipped(),
        }
        pipeline = _build_pipeline_mock(results)
        store = MockScoreStore()
        cache = MockScoreCache()
        publisher = MockScorePublisher()

        uc = ProcessFraudSignalsUseCase(pipeline=pipeline, publisher=publisher, cache=cache, store=store)
        uc.execute(_make_request())

        signals_summary_json = store.calls[0][3]
        summary = json.loads(signals_summary_json)

        similarity_entry = next(s for s in summary if s["signal_id"] == "similarity")
        assert similarity_entry["executed"] is False
        assert similarity_entry["value"] is None

        fraud_entry = next(s for s in summary if s["signal_id"] == "fraud-score")
        assert fraud_entry["executed"] is True
        assert fraud_entry["value"] == 85.0
