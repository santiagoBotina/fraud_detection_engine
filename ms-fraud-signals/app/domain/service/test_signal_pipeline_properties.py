"""Property-based tests for SignalPipeline.

Feature: fraud-signals-pipeline
Properties 1–3: Pipeline completeness, skip recording, error resilience.
"""

from __future__ import annotations

from hypothesis import given, settings, strategies as st

from app.domain.entity.fraud_signal_request import FraudSignalRequest
from app.domain.entity.signal_context import SignalContext
from app.domain.entity.signal_result import SignalResult
from app.domain.service.signal import Signal
from app.domain.service.signal_pipeline import SignalPipeline


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


class MockSignal(Signal):
    def __init__(
        self,
        sid: str,
        skip: bool = False,
        skip_reason: str | None = None,
        fail: bool = False,
    ):
        self._signal_id = sid
        self._skip = skip
        self._skip_reason = skip_reason
        self._fail = fail
        self.execute_called = False

    @property
    def signal_id(self) -> str:
        return self._signal_id

    def should_skip(self, context: SignalContext) -> tuple[bool, str | None]:
        return (self._skip, self._skip_reason)

    def execute(self, context: SignalContext) -> SignalContext:
        self.execute_called = True
        if self._fail:
            raise RuntimeError(f"Signal {self._signal_id} failed")
        context.results[self.signal_id] = SignalResult(
            signal_id=self.signal_id,
            executed=True,
            value=42.0,
            metadata={},
            skip_reason=None,
            error=None,
        )
        return context


def _make_context() -> SignalContext:
    return SignalContext(
        request=FraudSignalRequest(
            transaction_id="txn-test",
            amount_in_cents=50000,
            currency="USD",
            payment_method="CARD",
            customer_id="cust-1",
            customer_ip_address="10.0.0.1",
            timestamp="2024-01-01T00:00:00Z",
        )
    )


# ---------------------------------------------------------------------------
# Strategies
# ---------------------------------------------------------------------------

# Unique signal IDs: lists of 1–10 unique lowercase alpha strings
_unique_signal_ids = st.lists(
    st.text(alphabet="abcdefghijklmnopqrstuvwxyz", min_size=1, max_size=12),
    min_size=1,
    max_size=10,
    unique=True,
)

# (signal_id, should_skip) tuples – at least one entry
_skip_configs = st.lists(
    st.tuples(
        st.text(alphabet="abcdefghijklmnopqrstuvwxyz", min_size=1, max_size=12),
        st.booleans(),
    ),
    min_size=1,
    max_size=10,
    unique_by=lambda t: t[0],
)

# (signal_id, should_fail) tuples – at least one entry
_fail_configs = st.lists(
    st.tuples(
        st.text(alphabet="abcdefghijklmnopqrstuvwxyz", min_size=1, max_size=12),
        st.booleans(),
    ),
    min_size=1,
    max_size=10,
    unique_by=lambda t: t[0],
)


# ---------------------------------------------------------------------------
# Property 1: Pipeline executes all registered signals and produces a
#              complete context
# ---------------------------------------------------------------------------


@settings(max_examples=100)
@given(signal_ids=_unique_signal_ids)
def test_pipeline_produces_complete_context(signal_ids: list[str]) -> None:
    """**Validates: Requirements 1.1, 1.2, 1.4**

    For any ordered list of registered Signal instances and any valid
    FraudSignalRequest, running the SignalPipeline SHALL produce a
    SignalContext containing exactly one SignalResult per registered Signal,
    and the results SHALL appear in registration order.
    """
    signals = [MockSignal(sid=sid) for sid in signal_ids]
    pipeline = SignalPipeline()
    for s in signals:
        pipeline.register(s)

    context = pipeline.run(_make_context())

    # One result per signal
    assert len(context.results) == len(signal_ids)

    # Results appear in registration order
    result_keys = list(context.results.keys())
    assert result_keys == signal_ids


# ---------------------------------------------------------------------------
# Property 2: Pipeline records skipped signals correctly
# ---------------------------------------------------------------------------


@settings(max_examples=100)
@given(configs=_skip_configs)
def test_pipeline_records_skipped_signals(configs: list[tuple[str, bool]]) -> None:
    """**Validates: Requirements 1.3**

    For any Signal whose should_skip returns (True, reason), the
    SignalPipeline SHALL not call execute on that Signal and SHALL record a
    SignalResult with executed=False and the provided skip reason.
    """
    signals = [
        MockSignal(
            sid=sid,
            skip=should_skip,
            skip_reason=f"reason-{sid}" if should_skip else None,
        )
        for sid, should_skip in configs
    ]
    pipeline = SignalPipeline()
    for s in signals:
        pipeline.register(s)

    context = pipeline.run(_make_context())

    for sig, (sid, should_skip) in zip(signals, configs):
        result = context.results[sid]
        if should_skip:
            assert result.executed is False, f"Signal {sid} should not have executed"
            assert result.skip_reason == f"reason-{sid}"
            assert sig.execute_called is False, f"execute() should not be called for skipped signal {sid}"
        else:
            assert result.executed is True, f"Signal {sid} should have executed"
            assert result.skip_reason is None
            assert sig.execute_called is True


# ---------------------------------------------------------------------------
# Property 3: Pipeline continues after signal failure
# ---------------------------------------------------------------------------


@settings(max_examples=100)
@given(configs=_fail_configs)
def test_pipeline_continues_after_failure(configs: list[tuple[str, bool]]) -> None:
    """**Validates: Requirements 1.5**

    For any pipeline containing a Signal that raises an exception during
    execute, the SignalPipeline SHALL record a failed SignalResult (with
    error set) for that Signal and SHALL continue executing all subsequent
    Signals.
    """
    signals = [MockSignal(sid=sid, fail=should_fail) for sid, should_fail in configs]
    pipeline = SignalPipeline()
    for s in signals:
        pipeline.register(s)

    context = pipeline.run(_make_context())

    # Every signal must have a result regardless of failure
    assert len(context.results) == len(configs)

    for sig, (sid, should_fail) in zip(signals, configs):
        result = context.results[sid]
        if should_fail:
            assert result.error is not None, f"Failed signal {sid} should have error set"
            assert sig.execute_called is True, f"execute() should have been called on failing signal {sid}"
        else:
            assert result.error is None, f"Successful signal {sid} should not have error"
            assert result.executed is True
            assert sig.execute_called is True

    # All signals after a failing one must still have been called
    for i, (sig, (sid, should_fail)) in enumerate(zip(signals, configs)):
        if should_fail:
            # Every subsequent signal must have execute_called == True
            for later_sig in signals[i + 1 :]:
                assert later_sig.execute_called is True, (
                    f"Signal {later_sig.signal_id} after failing {sid} should still execute"
                )
