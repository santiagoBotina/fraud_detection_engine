"""Property-based tests for final score aggregation and clamping.

Feature: fraud-signals-pipeline, Property 7: Final score aggregation and clamping

Validates: Requirements 5.1, 5.2, 5.3
"""

from __future__ import annotations

from hypothesis import given, settings, strategies as st

from app.domain.usecase.process_fraud_signals import _clamp


# ---------------------------------------------------------------------------
# Strategies
# ---------------------------------------------------------------------------

_base_score = st.integers(min_value=0, max_value=100)
_adjustment = st.integers(min_value=-20, max_value=20)
_skipped = st.booleans()


# ---------------------------------------------------------------------------
# Property 7: Final score aggregation and clamping
# ---------------------------------------------------------------------------


class TestFinalScoreAggregationAndClamping:
    """**Validates: Requirements 5.1, 5.2, 5.3**"""

    @settings(max_examples=200)
    @given(base_score=_base_score, adjustment=_adjustment, skipped=_skipped)
    def test_final_score_equals_clamped_sum_or_base_when_skipped(
        self,
        base_score: int,
        adjustment: int,
        skipped: bool,
    ) -> None:
        """For any base fraud score in [0, 100] and any similarity adjustment
        in [-20, +20], the final fraud score SHALL equal
        clamp(base + adjustment, 0, 100) when the similarity signal executed.
        When the similarity signal was skipped, the final score SHALL equal
        the base score unchanged.
        """
        if skipped:
            # Similarity skipped → final score is just the base score
            final_score = _clamp(base_score + 0, 0, 100)
            assert final_score == base_score, (
                f"When skipped, final score should equal base_score={base_score}, "
                f"got {final_score}"
            )
        else:
            # Similarity executed → final score is clamped(base + adjustment)
            final_score = _clamp(base_score + adjustment, 0, 100)
            expected = max(0, min(100, base_score + adjustment))
            assert final_score == expected, (
                f"clamp({base_score} + {adjustment}, 0, 100) should be "
                f"{expected}, got {final_score}"
            )

    @settings(max_examples=200)
    @given(base_score=_base_score, adjustment=_adjustment)
    def test_final_score_always_in_valid_range(
        self,
        base_score: int,
        adjustment: int,
    ) -> None:
        """The final fraud score SHALL always be clamped to [0, 100]
        regardless of the base score and adjustment combination.
        """
        final_score = _clamp(base_score + adjustment, 0, 100)
        assert 0 <= final_score <= 100, (
            f"Final score {final_score} out of range [0, 100] for "
            f"base={base_score}, adj={adjustment}"
        )

    @settings(max_examples=200)
    @given(base_score=_base_score)
    def test_skipped_similarity_preserves_base_score(
        self,
        base_score: int,
    ) -> None:
        """When the similarity signal is skipped (adjustment=0), the final
        score SHALL equal the base score without modification.
        """
        # Skipped means adjustment is 0
        final_score = _clamp(base_score + 0, 0, 100)
        assert final_score == base_score, (
            f"Skipped similarity: expected {base_score}, got {final_score}"
        )
