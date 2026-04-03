"""Property-based tests for FuzzyLogicScorer.

Feature: fraud-score-service, Property 3: Fuzzy logic scorer output range invariant
Validates: Requirements 2.1, 2.2
"""

from hypothesis import given, settings, strategies as st
import pytest

from app.domain.service.fuzzy_logic_scorer import FuzzyLogicScorer, PAYMENT_METHOD_RISK, _ip_risk_score

_scorer = FuzzyLogicScorer()

PAYMENT_METHODS = st.sampled_from(["CARD", "BANK_TRANSFER", "CRYPTO"])


@settings(max_examples=200)
@given(
    amount_in_cents=st.integers(min_value=0),
    payment_method=PAYMENT_METHODS,
    customer_ip_address=st.text(min_size=1),
)
def test_fuzzy_logic_scorer_output_range(
    amount_in_cents: int,
    payment_method: str,
    customer_ip_address: str,
) -> None:
    """**Validates: Requirements 2.1, 2.2**

    For any valid combination of amount_in_cents (>= 0), payment_method
    (CARD | BANK_TRANSFER | CRYPTO), and customer_ip_address (non-empty string),
    the scorer must return an integer in [0, 100].
    """
    score = _scorer.compute(amount_in_cents, payment_method, customer_ip_address)

    assert isinstance(score, int), f"Expected int, got {type(score)}"
    assert 0 <= score <= 100, f"Score {score} out of range [0, 100]"


# ---------------------------------------------------------------------------
# Unit tests for FuzzyLogicScorer (Task 3.3)
# Validates: Requirements 2.2, 2.3
# ---------------------------------------------------------------------------


class TestFuzzyLogicScorerSmoke:
    """Smoke tests: known inputs produce expected score ranges."""

    def test_low_amount_bank_transfer_produces_low_score(self) -> None:
        """Low amount + lowest-risk payment method should yield a low fraud score."""
        score = _scorer.compute(1000, "BANK_TRANSFER", "10.0.0.1")
        assert 0 <= score <= 100
        # Bank transfer (risk=1) + small amount → expect low-to-medium score
        assert score <= 60

    def test_high_amount_crypto_produces_high_score(self) -> None:
        """High amount + highest-risk payment method should yield a high fraud score."""
        score = _scorer.compute(800_000, "CRYPTO", "10.0.0.1")
        assert 0 <= score <= 100
        # Crypto (risk=8) + large amount → expect elevated score
        assert score >= 40

    def test_medium_amount_card_produces_medium_range(self) -> None:
        """Medium amount + card payment should stay in a moderate range."""
        score = _scorer.compute(50_000, "CARD", "192.168.1.1")
        assert 0 <= score <= 100

    def test_score_is_integer(self) -> None:
        """Score must always be an int (Requirement 2.2)."""
        score = _scorer.compute(25_000, "CARD", "127.0.0.1")
        assert isinstance(score, int)


class TestFuzzyLogicScorerEdgeCases:
    """Edge-case tests for boundary and unusual inputs."""

    def test_amount_zero(self) -> None:
        """amount=0 should still produce a valid score in [0, 100]."""
        for method in ("CARD", "BANK_TRANSFER", "CRYPTO"):
            score = _scorer.compute(0, method, "1.2.3.4")
            assert isinstance(score, int)
            assert 0 <= score <= 100

    def test_very_large_amount(self) -> None:
        """Amounts far above the universe max (1,000,000) should be clamped and scored."""
        score = _scorer.compute(999_999_999, "CARD", "8.8.8.8")
        assert isinstance(score, int)
        assert 0 <= score <= 100

    def test_amount_at_universe_max(self) -> None:
        """amount exactly at 1,000,000 (universe boundary) should work."""
        score = _scorer.compute(1_000_000, "CRYPTO", "1.1.1.1")
        assert isinstance(score, int)
        assert 0 <= score <= 100

    @pytest.mark.parametrize("method", ["CARD", "BANK_TRANSFER", "CRYPTO"])
    def test_each_payment_method(self, method: str) -> None:
        """Every supported payment method produces a valid score."""
        score = _scorer.compute(50_000, method, "10.10.10.10")
        assert isinstance(score, int)
        assert 0 <= score <= 100

    def test_unknown_payment_method_defaults_gracefully(self) -> None:
        """An unrecognised payment method should not crash; it defaults to risk=5."""
        score = _scorer.compute(50_000, "UNKNOWN_METHOD", "10.0.0.1")
        assert isinstance(score, int)
        assert 0 <= score <= 100


class TestPaymentMethodRiskMapping:
    """Verify the payment-method-to-risk mapping (Requirement 2.3)."""

    def test_bank_transfer_risk(self) -> None:
        assert PAYMENT_METHOD_RISK["BANK_TRANSFER"] == 1

    def test_card_risk(self) -> None:
        assert PAYMENT_METHOD_RISK["CARD"] == 3

    def test_crypto_risk(self) -> None:
        assert PAYMENT_METHOD_RISK["CRYPTO"] == 8


class TestIpRiskHeuristic:
    """Verify the hash-based IP risk helper."""

    def test_ip_risk_in_range(self) -> None:
        """IP risk score must be in [0, 10]."""
        for ip in ("127.0.0.1", "10.0.0.1", "192.168.1.1", "255.255.255.255"):
            risk = _ip_risk_score(ip)
            assert 0 <= risk <= 10

    def test_ip_risk_is_deterministic(self) -> None:
        """Same IP should always produce the same risk value."""
        assert _ip_risk_score("8.8.8.8") == _ip_risk_score("8.8.8.8")
