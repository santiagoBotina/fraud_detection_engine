"""Unit tests for FraudSignalResult entity.

Validates: Requirements 5.1, 9.3, 10.3
"""

from datetime import datetime, timezone

from app.domain.entity.fraud_signal_result import FraudSignalResult
from app.domain.entity.signal_result import SignalResult


def test_to_dict_backward_compatible_keys() -> None:
    """to_dict() output contains top-level transaction_id, fraud_score, calculated_at, and signals."""
    result = FraudSignalResult(
        transaction_id="txn-001",
        fraud_score=45,
        signal_results=[
            SignalResult(signal_id="fraud-score", executed=True, value=50, metadata={}),
            SignalResult(signal_id="similarity", executed=True, value=-5, metadata={}),
        ],
        calculated_at=datetime(2024, 1, 15, 10, 30, 1, tzinfo=timezone.utc),
    )

    d = result.to_dict()

    assert d["transaction_id"] == "txn-001"
    assert d["fraud_score"] == 45
    assert d["calculated_at"] == "2024-01-15T10:30:01+00:00"
    assert len(d["signals"]) == 2
    assert d["signals"][0] == {"signal_id": "fraud-score", "executed": True, "value": 50}
    assert d["signals"][1] == {"signal_id": "similarity", "executed": True, "value": -5}


def test_to_dict_empty_signals() -> None:
    """to_dict() works with an empty signal_results list."""
    result = FraudSignalResult(
        transaction_id="txn-002",
        fraud_score=80,
        signal_results=[],
        calculated_at=datetime(2024, 6, 1, 12, 0, 0, tzinfo=timezone.utc),
    )

    d = result.to_dict()

    assert d["transaction_id"] == "txn-002"
    assert d["fraud_score"] == 80
    assert d["signals"] == []


def test_to_dict_fraud_score_matches_attribute() -> None:
    """The fraud_score in to_dict() output matches the entity attribute."""
    result = FraudSignalResult(
        transaction_id="txn-003",
        fraud_score=0,
        signal_results=[
            SignalResult(signal_id="fraud-score", executed=True, value=0, metadata={}),
        ],
        calculated_at=datetime(2024, 1, 1, 0, 0, 0, tzinfo=timezone.utc),
    )

    assert result.to_dict()["fraud_score"] == result.fraud_score


def test_frozen_dataclass() -> None:
    """FraudSignalResult is immutable."""
    result = FraudSignalResult(
        transaction_id="txn-004",
        fraud_score=50,
        signal_results=[],
        calculated_at=datetime(2024, 1, 1, tzinfo=timezone.utc),
    )

    try:
        result.fraud_score = 99  # type: ignore[misc]
        assert False, "Should have raised FrozenInstanceError"
    except AttributeError:
        pass
