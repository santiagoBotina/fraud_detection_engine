"""Property-based and unit tests for FraudScoreSignal.

Feature: fraud-signals-pipeline
Property 4: FraudScoreSignal produces a score in [0, 100].
Unit tests: delegation, signal_id, should_skip.
"""

from __future__ import annotations

import pytest
from hypothesis import given, settings, strategies as st

from app.domain.entity.fraud_signal_request import FraudSignalRequest
from app.domain.entity.signal_context import SignalContext
from app.domain.service.fraud_score_signal import FraudScoreSignal
from app.domain.service.fuzzy_logic_scorer import FuzzyLogicScorer


# ---------------------------------------------------------------------------
# Shared fixtures
# ---------------------------------------------------------------------------

_scorer = FuzzyLogicScorer()
_signal = FraudScoreSignal(scorer=_scorer)


def _make_context(request: FraudSignalRequest) -> SignalContext:
    return SignalContext(request=request)


# ---------------------------------------------------------------------------
# Hypothesis strategy for FraudSignalRequest
# ---------------------------------------------------------------------------

_fraud_signal_request_strategy = st.builds(
    FraudSignalRequest,
    transaction_id=st.text(min_size=1, max_size=20),
    amount_in_cents=st.integers(min_value=0, max_value=10_000_000),
    currency=st.sampled_from(["USD", "COP", "EUR"]),
    payment_method=st.sampled_from(["CARD", "BANK_TRANSFER", "CRYPTO"]),
    customer_id=st.text(min_size=1, max_size=20),
    customer_ip_address=st.text(min_size=1, max_size=40),
    timestamp=st.text(min_size=1, max_size=30),
)


# ---------------------------------------------------------------------------
# Property 4: FraudScoreSignal produces a score in [0, 100]
# ---------------------------------------------------------------------------


@settings(max_examples=100)
@given(request=_fraud_signal_request_strategy)
def test_fraud_score_signal_produces_score_in_valid_range(
    request: FraudSignalRequest,
) -> None:
    """**Validates: Requirements 2.1, 2.2**

    For any valid FraudSignalRequest, the FraudScoreSignal SHALL produce a
    SignalResult with signal_id="fraud-score", executed=True, and value in
    the integer range [0, 100].
    """
    scorer = FuzzyLogicScorer()
    signal = FraudScoreSignal(scorer=scorer)
    context = _make_context(request)

    result_context = signal.execute(context)

    signal_result = result_context.results["fraud-score"]
    assert signal_result.signal_id == "fraud-score"
    assert signal_result.executed is True
    assert signal_result.value is not None
    assert 0 <= signal_result.value <= 100


# ---------------------------------------------------------------------------
# Unit tests for FraudScoreSignal
# ---------------------------------------------------------------------------


class TestFraudScoreSignalId:
    """Test that signal_id is 'fraud-score'."""

    def test_signal_id_is_fraud_score(self) -> None:
        assert _signal.signal_id == "fraud-score"


class TestFraudScoreSignalShouldSkip:
    """Test that should_skip always returns (False, None)."""

    def test_should_skip_returns_false_none(self) -> None:
        request = FraudSignalRequest(
            transaction_id="txn-001",
            amount_in_cents=50000,
            currency="USD",
            payment_method="CARD",
            customer_id="cust-1",
            customer_ip_address="10.0.0.1",
            timestamp="2024-01-15T10:30:00Z",
        )
        context = _make_context(request)
        skip, reason = _signal.should_skip(context)
        assert skip is False
        assert reason is None


class TestFraudScoreSignalDelegation:
    """Test that FraudScoreSignal delegates to FuzzyLogicScorer with correct arguments."""

    def test_delegates_correct_arguments_to_scorer(self) -> None:
        """The signal passes amount_in_cents, payment_method, and
        customer_ip_address from the request to FuzzyLogicScorer.compute()."""
        request = FraudSignalRequest(
            transaction_id="txn-delegate",
            amount_in_cents=150000,
            currency="USD",
            payment_method="CRYPTO",
            customer_id="cust-100",
            customer_ip_address="192.168.1.10",
            timestamp="2024-01-15T10:30:00Z",
        )
        scorer = FuzzyLogicScorer()
        signal = FraudScoreSignal(scorer=scorer)
        context = _make_context(request)

        # Compute expected score directly from the scorer
        expected_score = scorer.compute(
            request.amount_in_cents,
            request.payment_method,
            request.customer_ip_address,
        )

        result_context = signal.execute(context)
        signal_result = result_context.results["fraud-score"]

        assert signal_result.value == float(expected_score)
        assert signal_result.executed is True
        assert signal_result.skip_reason is None
        assert signal_result.error is None
        assert signal_result.metadata == {}

    def test_result_written_to_context(self) -> None:
        """The signal writes its result into context.results under key 'fraud-score'."""
        request = FraudSignalRequest(
            transaction_id="txn-ctx",
            amount_in_cents=1000,
            currency="EUR",
            payment_method="BANK_TRANSFER",
            customer_id="cust-2",
            customer_ip_address="8.8.8.8",
            timestamp="2024-06-01T00:00:00Z",
        )
        context = _make_context(request)
        assert "fraud-score" not in context.results

        result_context = _signal.execute(context)
        assert "fraud-score" in result_context.results

        sr = result_context.results["fraud-score"]
        assert sr.signal_id == "fraud-score"
        assert sr.executed is True
        assert sr.value is not None
        assert 0 <= sr.value <= 100
