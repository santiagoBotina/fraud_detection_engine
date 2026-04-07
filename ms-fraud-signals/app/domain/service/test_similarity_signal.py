"""Property-based and unit tests for SimilaritySignal.

Feature: fraud-signals-pipeline
Property 5: Neutral score boundary determines similarity execution.
Property 6: Similarity adjustment is clamped to [-20, +20].
Unit tests: skip logic, adjustment computation, error handling.
"""

from __future__ import annotations

from hypothesis import given, settings, strategies as st

from app.domain.entity.fraud_signal_request import FraudSignalRequest
from app.domain.entity.signal_context import SignalContext
from app.domain.entity.signal_result import SignalResult
from app.domain.port.vector_search import SimilarityMatch, VectorSearch
from app.domain.service.similarity_signal import SimilaritySignal


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


class MockVectorSearch(VectorSearch):
    def __init__(self, matches=None, error=None):
        self._matches = matches or []
        self._error = error

    def find_similar(self, embedding, top_k):
        if self._error:
            raise self._error
        return self._matches


def _make_request() -> FraudSignalRequest:
    return FraudSignalRequest(
        transaction_id="txn-test",
        amount_in_cents=50000,
        currency="USD",
        payment_method="CARD",
        customer_id="cust-1",
        customer_ip_address="10.0.0.1",
        timestamp="2024-01-01T00:00:00Z",
    )


def _make_context_with_fraud_score(score: float) -> SignalContext:
    ctx = SignalContext(request=_make_request())
    ctx.results["fraud-score"] = SignalResult(
        signal_id="fraud-score",
        executed=True,
        value=score,
        metadata={},
        skip_reason=None,
        error=None,
    )
    return ctx


# ---------------------------------------------------------------------------
# Strategies
# ---------------------------------------------------------------------------

_score_strategy = st.integers(min_value=0, max_value=100)

_similarity_match_strategy = st.builds(
    SimilarityMatch,
    transaction_id=st.text(min_size=1, max_size=20),
    score=st.floats(min_value=0.0, max_value=1.0),
    status=st.sampled_from(["approved", "declined"]),
)

_similarity_match_list_strategy = st.lists(
    _similarity_match_strategy,
    min_size=0,
    max_size=10,
)


# ---------------------------------------------------------------------------
# Property 5: Neutral score boundary determines similarity execution
# ---------------------------------------------------------------------------


@settings(max_examples=100)
@given(score=_score_strategy)
def test_neutral_score_boundary_determines_skip(score: int) -> None:
    """**Validates: Requirements 3.1, 3.2, 3.3**

    For any fraud score value, the SimilaritySignal's should_skip SHALL
    return (False, None) if and only if the score is in [30, 70] inclusive,
    and SHALL return (True, "score-outside-neutral-range") otherwise.
    """
    signal = SimilaritySignal(vector_search=MockVectorSearch())
    context = _make_context_with_fraud_score(float(score))

    skip, reason = signal.should_skip(context)

    if 30 <= score <= 70:
        assert skip is False, f"Score {score} is neutral, should_skip must be False"
        assert reason is None
    else:
        assert skip is True, f"Score {score} is outside neutral range, should_skip must be True"
        assert reason == "score-outside-neutral-range"


# ---------------------------------------------------------------------------
# Property 6: Similarity adjustment is clamped to [-20, +20]
# ---------------------------------------------------------------------------


@settings(max_examples=100)
@given(matches=_similarity_match_list_strategy)
def test_similarity_adjustment_clamped(matches: list[SimilarityMatch]) -> None:
    """**Validates: Requirements 4.3**

    For any list of SimilarityMatch results returned by VectorSearch, the
    SimilaritySignal SHALL compute an adjustment value in the range
    [-20, +20] inclusive.
    """
    adjustment = SimilaritySignal._compute_adjustment(matches)

    assert -20 <= adjustment <= 20, (
        f"Adjustment {adjustment} is outside [-20, +20] for matches: {matches}"
    )


# ---------------------------------------------------------------------------
# Unit tests for SimilaritySignal
# ---------------------------------------------------------------------------


class TestSimilaritySignalSkipLogic:
    """Test should_skip with scores below 30, above 70, and within [30, 70]."""

    def test_skip_when_score_below_30(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        context = _make_context_with_fraud_score(15.0)
        skip, reason = signal.should_skip(context)
        assert skip is True
        assert reason == "score-outside-neutral-range"

    def test_skip_when_score_above_70(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        context = _make_context_with_fraud_score(85.0)
        skip, reason = signal.should_skip(context)
        assert skip is True
        assert reason == "score-outside-neutral-range"

    def test_no_skip_when_score_in_neutral_range(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        context = _make_context_with_fraud_score(50.0)
        skip, reason = signal.should_skip(context)
        assert skip is False
        assert reason is None

    def test_no_skip_at_lower_boundary(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        context = _make_context_with_fraud_score(30.0)
        skip, reason = signal.should_skip(context)
        assert skip is False
        assert reason is None

    def test_no_skip_at_upper_boundary(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        context = _make_context_with_fraud_score(70.0)
        skip, reason = signal.should_skip(context)
        assert skip is False
        assert reason is None

    def test_skip_at_boundary_minus_one(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        context = _make_context_with_fraud_score(29.0)
        skip, reason = signal.should_skip(context)
        assert skip is True
        assert reason == "score-outside-neutral-range"

    def test_skip_at_boundary_plus_one(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        context = _make_context_with_fraud_score(71.0)
        skip, reason = signal.should_skip(context)
        assert skip is True
        assert reason == "score-outside-neutral-range"

    def test_skip_when_no_fraud_score_result(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        context = SignalContext(request=_make_request())
        skip, reason = signal.should_skip(context)
        assert skip is True
        assert reason == "score-outside-neutral-range"


class TestSimilaritySignalAdjustment:
    """Test adjustment computation with mocked VectorSearch returning known matches."""

    def test_all_declined_matches_positive_adjustment(self) -> None:
        matches = [
            SimilarityMatch(transaction_id="txn-1", score=0.9, status="declined"),
            SimilarityMatch(transaction_id="txn-2", score=0.8, status="declined"),
        ]
        signal = SimilaritySignal(vector_search=MockVectorSearch(matches=matches))
        context = _make_context_with_fraud_score(50.0)

        result_context = signal.execute(context)
        result = result_context.results["similarity"]

        assert result.executed is True
        assert result.value is not None
        assert result.value > 0  # declined → positive adjustment
        assert result.error is None

    def test_all_approved_matches_negative_adjustment(self) -> None:
        matches = [
            SimilarityMatch(transaction_id="txn-1", score=0.9, status="approved"),
            SimilarityMatch(transaction_id="txn-2", score=0.8, status="approved"),
        ]
        signal = SimilaritySignal(vector_search=MockVectorSearch(matches=matches))
        context = _make_context_with_fraud_score(50.0)

        result_context = signal.execute(context)
        result = result_context.results["similarity"]

        assert result.executed is True
        assert result.value is not None
        assert result.value < 0  # approved → negative adjustment
        assert result.error is None

    def test_empty_matches_zero_adjustment(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch(matches=[]))
        context = _make_context_with_fraud_score(50.0)

        result_context = signal.execute(context)
        result = result_context.results["similarity"]

        assert result.executed is True
        assert result.value == 0.0
        assert result.error is None

    def test_matched_transaction_ids_in_metadata(self) -> None:
        matches = [
            SimilarityMatch(transaction_id="txn-a", score=0.5, status="approved"),
            SimilarityMatch(transaction_id="txn-b", score=0.6, status="declined"),
        ]
        signal = SimilaritySignal(vector_search=MockVectorSearch(matches=matches))
        context = _make_context_with_fraud_score(50.0)

        result_context = signal.execute(context)
        result = result_context.results["similarity"]

        assert result.metadata["matched_transactions"] == ["txn-a", "txn-b"]

    def test_adjustment_clamped_to_max_20(self) -> None:
        # 5 declined matches with perfect similarity → raw = 5 * 1.0 * 10 = 50
        matches = [
            SimilarityMatch(transaction_id=f"txn-{i}", score=1.0, status="declined")
            for i in range(5)
        ]
        signal = SimilaritySignal(vector_search=MockVectorSearch(matches=matches))
        context = _make_context_with_fraud_score(50.0)

        result_context = signal.execute(context)
        result = result_context.results["similarity"]

        assert result.value == 20.0

    def test_adjustment_clamped_to_min_negative_20(self) -> None:
        # 5 approved matches with perfect similarity → raw = -5 * 1.0 * 10 = -50
        matches = [
            SimilarityMatch(transaction_id=f"txn-{i}", score=1.0, status="approved")
            for i in range(5)
        ]
        signal = SimilaritySignal(vector_search=MockVectorSearch(matches=matches))
        context = _make_context_with_fraud_score(50.0)

        result_context = signal.execute(context)
        result = result_context.results["similarity"]

        assert result.value == -20.0


class TestSimilaritySignalErrorHandling:
    """Test error handling when VectorSearch raises an exception."""

    def test_vector_search_error_returns_zero_adjustment(self) -> None:
        error = ConnectionError("Qdrant unreachable")
        signal = SimilaritySignal(vector_search=MockVectorSearch(error=error))
        context = _make_context_with_fraud_score(50.0)

        result_context = signal.execute(context)
        result = result_context.results["similarity"]

        assert result.executed is True
        assert result.value == 0.0
        assert result.error == "Qdrant unreachable"
        assert result.skip_reason is None

    def test_vector_search_runtime_error(self) -> None:
        error = RuntimeError("Unexpected failure")
        signal = SimilaritySignal(vector_search=MockVectorSearch(error=error))
        context = _make_context_with_fraud_score(50.0)

        result_context = signal.execute(context)
        result = result_context.results["similarity"]

        assert result.executed is True
        assert result.value == 0.0
        assert result.error == "Unexpected failure"

    def test_signal_id_is_similarity(self) -> None:
        signal = SimilaritySignal(vector_search=MockVectorSearch())
        assert signal.signal_id == "similarity"
