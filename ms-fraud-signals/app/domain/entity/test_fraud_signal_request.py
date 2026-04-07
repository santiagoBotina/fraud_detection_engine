"""Unit and property-based tests for FraudSignalRequest.

Feature: fraud-signals-pipeline
Validates: Requirements 1.1, 9.1
"""

from hypothesis import given, settings
from hypothesis import strategies as st

from app.domain.entity.fraud_signal_request import FraudSignalRequest

fraud_signal_request_strategy = st.builds(
    FraudSignalRequest,
    transaction_id=st.text(min_size=1),
    amount_in_cents=st.integers(min_value=0),
    currency=st.sampled_from(["USD", "COP", "EUR"]),
    payment_method=st.sampled_from(["CARD", "BANK_TRANSFER", "CRYPTO"]),
    customer_id=st.text(min_size=1),
    customer_ip_address=st.text(min_size=1),
    timestamp=st.datetimes().map(lambda dt: dt.isoformat() + "Z"),
)


@given(request=fraud_signal_request_strategy)
@settings(max_examples=100)
def test_fraud_signal_request_json_round_trip(request: FraudSignalRequest) -> None:
    """For any valid FraudSignalRequest, serializing to dict and deserializing back
    produces an equivalent object.

    Feature: fraud-signals-pipeline
    Validates: Requirements 1.1, 9.1
    """
    assert FraudSignalRequest.from_dict(request.to_dict()) == request


def test_fraud_signal_request_is_frozen() -> None:
    """FraudSignalRequest should be immutable (frozen dataclass)."""
    request = FraudSignalRequest(
        transaction_id="txn-001",
        amount_in_cents=150000,
        currency="USD",
        payment_method="CARD",
        customer_id="cust_100",
        customer_ip_address="192.168.1.10",
        timestamp="2024-01-15T10:30:00Z",
    )
    try:
        request.transaction_id = "txn-002"  # type: ignore[misc]
        assert False, "Should have raised FrozenInstanceError"
    except AttributeError:
        pass


def test_fraud_signal_request_to_dict() -> None:
    """to_dict() should return all fields as a plain dict."""
    request = FraudSignalRequest(
        transaction_id="txn-001",
        amount_in_cents=150000,
        currency="USD",
        payment_method="CARD",
        customer_id="cust_100",
        customer_ip_address="192.168.1.10",
        timestamp="2024-01-15T10:30:00Z",
    )
    result = request.to_dict()
    assert result == {
        "transaction_id": "txn-001",
        "amount_in_cents": 150000,
        "currency": "USD",
        "payment_method": "CARD",
        "customer_id": "cust_100",
        "customer_ip_address": "192.168.1.10",
        "timestamp": "2024-01-15T10:30:00Z",
    }


def test_fraud_signal_request_from_dict() -> None:
    """from_dict() should construct a FraudSignalRequest from a dict."""
    data = {
        "transaction_id": "txn-002",
        "amount_in_cents": 50000,
        "currency": "EUR",
        "payment_method": "CRYPTO",
        "customer_id": "cust_200",
        "customer_ip_address": "10.0.0.1",
        "timestamp": "2024-06-01T12:00:00Z",
    }
    request = FraudSignalRequest.from_dict(data)
    assert request.transaction_id == "txn-002"
    assert request.amount_in_cents == 50000
    assert request.currency == "EUR"
    assert request.payment_method == "CRYPTO"
    assert request.customer_id == "cust_200"
    assert request.customer_ip_address == "10.0.0.1"
    assert request.timestamp == "2024-06-01T12:00:00Z"
