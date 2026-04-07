"""Property-based tests for FraudSignalResult serialization.

Feature: fraud-signals-pipeline, Property 8: FraudSignalResult serialization preserves backward compatibility

Validates: Requirements 9.3, 10.3
"""

from datetime import datetime, timezone

from hypothesis import given, settings
from hypothesis import strategies as st

from app.domain.entity.fraud_signal_result import FraudSignalResult
from app.domain.entity.signal_result import SignalResult

# --- Strategies ---

signal_result_strategy = st.builds(
    SignalResult,
    signal_id=st.text(min_size=1, max_size=30),
    executed=st.booleans(),
    value=st.one_of(st.none(), st.floats(allow_nan=False, allow_infinity=False)),
    metadata=st.just({}),
)

fraud_signal_result_strategy = st.builds(
    FraudSignalResult,
    transaction_id=st.text(min_size=1, max_size=50),
    fraud_score=st.integers(min_value=0, max_value=100),
    signal_results=st.lists(signal_result_strategy, min_size=0, max_size=5),
    calculated_at=st.datetimes(timezones=st.just(timezone.utc)),
)


# --- Property Test ---


@settings(max_examples=100)
@given(result=fraud_signal_result_strategy)
def test_property8_serialization_preserves_backward_compatibility(
    result: FraudSignalResult,
) -> None:
    """Property 8: FraudSignalResult serialization preserves backward compatibility.

    **Validates: Requirements 9.3, 10.3**

    For any valid FraudSignalResult, serializing via to_dict() SHALL produce a
    dict containing top-level keys 'transaction_id', 'fraud_score',
    'calculated_at', and 'signals', and the 'fraud_score' value SHALL equal
    the fraud_score attribute.
    """
    d = result.to_dict()

    # Required top-level keys are present
    assert "transaction_id" in d
    assert "fraud_score" in d
    assert "calculated_at" in d
    assert "signals" in d

    # fraud_score in serialized output matches the entity attribute
    assert d["fraud_score"] == result.fraud_score

    # signals list length matches signal_results
    assert len(d["signals"]) == len(result.signal_results)
