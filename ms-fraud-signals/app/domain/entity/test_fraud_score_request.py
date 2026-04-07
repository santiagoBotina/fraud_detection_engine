"""Property-based test for FraudScoreRequest JSON round-trip.

Feature: fraud-score-service, Property 5: FraudScoreRequest JSON round-trip
Validates: Requirements 7.1, 7.3, 7.5
"""

from hypothesis import given, settings
from hypothesis import strategies as st

from app.domain.entity.fraud_score_request import FraudScoreRequest

fraud_score_request_strategy = st.builds(
    FraudScoreRequest,
    transaction_id=st.text(min_size=1),
    amount_in_cents=st.integers(min_value=0),
    currency=st.sampled_from(["USD", "COP", "EUR"]),
    payment_method=st.sampled_from(["CARD", "BANK_TRANSFER", "CRYPTO"]),
    customer_id=st.text(min_size=1),
    customer_ip_address=st.text(min_size=1),
    timestamp=st.datetimes().map(lambda dt: dt.isoformat() + "Z"),
)


@given(request=fraud_score_request_strategy)
@settings(max_examples=100)
def test_fraud_score_request_json_round_trip(request: FraudScoreRequest) -> None:
    """For any valid FraudScoreRequest, serializing to dict and deserializing back
    produces an equivalent object.

    Feature: fraud-score-service, Property 5: FraudScoreRequest JSON round-trip
    Validates: Requirements 7.1, 7.3, 7.5
    """
    assert FraudScoreRequest.from_dict(request.to_dict()) == request
