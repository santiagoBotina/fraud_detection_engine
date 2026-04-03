"""Property-based test for ComputeFraudScoreUseCase.

Feature: fraud-score-service, Property 4: Compute use case persists, caches, and publishes for every score
Validates: Requirements 3.1, 3.2, 4.1
"""

from __future__ import annotations

from datetime import datetime

from hypothesis import given, settings
from hypothesis import strategies as st

from app.domain.entity.fraud_score_request import FraudScoreRequest
from app.domain.entity.fraud_score_result import FraudScoreResult
from app.domain.port.score_cache import ScoreCache
from app.domain.port.score_publisher import ScorePublisher
from app.domain.port.score_store import ScoreStore
from app.domain.service.fuzzy_logic_scorer import FuzzyLogicScorer
from app.domain.usecase.compute_fraud_score import ComputeFraudScoreUseCase


class MockScoreStore(ScoreStore):
    def __init__(self) -> None:
        self.calls: list[tuple[str, int, datetime]] = []

    def save(self, transaction_id: str, score: int, calculated_at: datetime) -> None:
        self.calls.append((transaction_id, score, calculated_at))

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
        self.calls: list[FraudScoreResult] = []

    def publish(self, result: FraudScoreResult) -> None:
        self.calls.append(result)


fraud_score_request_strategy = st.builds(
    FraudScoreRequest,
    transaction_id=st.text(min_size=1),
    amount_in_cents=st.integers(min_value=0, max_value=10_000_000),
    currency=st.sampled_from(["USD", "COP", "EUR"]),
    payment_method=st.sampled_from(["CARD", "BANK_TRANSFER", "CRYPTO"]),
    customer_id=st.text(min_size=1),
    customer_ip_address=st.text(min_size=1),
    timestamp=st.datetimes().map(lambda dt: dt.isoformat() + "Z"),
)


@given(request=fraud_score_request_strategy)
@settings(max_examples=100)
def test_compute_use_case_persists_caches_and_publishes(request: FraudScoreRequest) -> None:
    """For any valid FraudScoreRequest, the use case stores, caches, and publishes
    the computed fraud score with correct arguments.

    Feature: fraud-score-service, Property 4: Compute use case persists, caches, and publishes for every score
    Validates: Requirements 3.1, 3.2, 4.1
    """
    store = MockScoreStore()
    cache = MockScoreCache()
    publisher = MockScorePublisher()
    scorer = FuzzyLogicScorer()

    use_case = ComputeFraudScoreUseCase(
        scorer=scorer, publisher=publisher, cache=cache, store=store
    )

    result = use_case.execute(request)

    # store.save called with (transaction_id, score, calculated_at) — Req 3.1
    assert len(store.calls) == 1
    stored_txn_id, stored_score, stored_at = store.calls[0]
    assert stored_txn_id == request.transaction_id
    assert stored_score == result.fraud_score
    assert stored_at == result.calculated_at

    # cache.set called with (transaction_id, score) — Req 3.2
    assert len(cache.calls) == 1
    cached_txn_id, cached_score = cache.calls[0]
    assert cached_txn_id == request.transaction_id
    assert cached_score == result.fraud_score

    # publisher.publish called with matching FraudScoreResult — Req 4.1
    assert len(publisher.calls) == 1
    published = publisher.calls[0]
    assert published.transaction_id == request.transaction_id
    assert published.fraud_score == result.fraud_score
    assert published.calculated_at == result.calculated_at


# ---------------------------------------------------------------------------
# Unit tests for ComputeFraudScoreUseCase error handling
# Task 4.3 — Validates: Requirements 3.3, 3.4
# ---------------------------------------------------------------------------


class FailingScoreCache(ScoreCache):
    """A ScoreCache that always raises on set(), simulating Redis unavailability."""

    def set(self, transaction_id: str, score: int) -> None:
        raise ConnectionError("Redis unavailable")

    def get(self, transaction_id: str) -> int | None:
        raise ConnectionError("Redis unavailable")


class FailingScoreStore(ScoreStore):
    """A ScoreStore that always raises on save(), simulating DynamoDB write failure."""

    def save(self, transaction_id: str, score: int, calculated_at: datetime) -> None:
        raise RuntimeError("DynamoDB write failed")

    def get(self, transaction_id: str) -> dict | None:
        raise RuntimeError("DynamoDB read failed")


def _make_request() -> FraudScoreRequest:
    return FraudScoreRequest(
        transaction_id="txn-err-001",
        amount_in_cents=50000,
        currency="USD",
        payment_method="CARD",
        customer_id="cust-1",
        customer_ip_address="10.0.0.1",
        timestamp="2025-01-15T10:30:00Z",
    )


def test_redis_unavailable_still_stores_and_publishes() -> None:
    """When Redis cache fails, the use case still persists to DynamoDB and publishes.

    Validates: Requirement 3.3
    """
    store = MockScoreStore()
    cache = FailingScoreCache()
    publisher = MockScorePublisher()
    scorer = FuzzyLogicScorer()

    use_case = ComputeFraudScoreUseCase(
        scorer=scorer, publisher=publisher, cache=cache, store=store
    )

    result = use_case.execute(_make_request())

    # DynamoDB store was called
    assert len(store.calls) == 1
    assert store.calls[0][0] == "txn-err-001"
    assert store.calls[0][1] == result.fraud_score

    # Publisher was called
    assert len(publisher.calls) == 1
    assert publisher.calls[0].transaction_id == "txn-err-001"
    assert publisher.calls[0].fraud_score == result.fraud_score


def test_dynamodb_failure_still_caches_and_publishes() -> None:
    """When DynamoDB store fails, the use case still caches to Redis and publishes.

    Validates: Requirement 3.4
    """
    store = FailingScoreStore()
    cache = MockScoreCache()
    publisher = MockScorePublisher()
    scorer = FuzzyLogicScorer()

    use_case = ComputeFraudScoreUseCase(
        scorer=scorer, publisher=publisher, cache=cache, store=store
    )

    result = use_case.execute(_make_request())

    # Redis cache was called
    assert len(cache.calls) == 1
    assert cache.calls[0][0] == "txn-err-001"
    assert cache.calls[0][1] == result.fraud_score

    # Publisher was called
    assert len(publisher.calls) == 1
    assert publisher.calls[0].transaction_id == "txn-err-001"
    assert publisher.calls[0].fraud_score == result.fraud_score
