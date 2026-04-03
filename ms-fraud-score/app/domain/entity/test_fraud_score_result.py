"""Property-based test for FraudScoreResult JSON round-trip.

Feature: fraud-score-service, Property 6: FraudScoreResult JSON round-trip
Validates: Requirements 7.2, 7.4, 7.6
"""

from hypothesis import given, settings
from hypothesis import strategies as st

from app.domain.entity.fraud_score_result import FraudScoreResult

fraud_score_result_strategy = st.builds(
    FraudScoreResult,
    transaction_id=st.text(min_size=1),
    fraud_score=st.integers(min_value=0, max_value=100),
    calculated_at=st.datetimes(),
)


@given(result=fraud_score_result_strategy)
@settings(max_examples=100)
def test_fraud_score_result_json_round_trip(result: FraudScoreResult) -> None:
    """For any valid FraudScoreResult, serializing to dict and deserializing back
    produces an equivalent object.

    Feature: fraud-score-service, Property 6: FraudScoreResult JSON round-trip
    Validates: Requirements 7.2, 7.4, 7.6
    """
    assert FraudScoreResult.from_dict(result.to_dict()) == result
