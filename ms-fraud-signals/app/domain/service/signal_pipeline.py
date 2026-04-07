from __future__ import annotations

import logging

from app.domain.entity.signal_context import SignalContext
from app.domain.entity.signal_result import SignalResult
from app.domain.service.signal import Signal

logger = logging.getLogger(__name__)


class SignalPipeline:
    def __init__(self) -> None:
        self._signals: list[Signal] = []

    def register(self, signal: Signal) -> None:
        self._signals.append(signal)

    def run(self, context: SignalContext) -> SignalContext:
        for signal in self._signals:
            skip, reason = signal.should_skip(context)
            if skip:
                context.results[signal.signal_id] = SignalResult(
                    signal_id=signal.signal_id,
                    executed=False,
                    value=None,
                    metadata={},
                    skip_reason=reason,
                    error=None,
                )
                continue
            try:
                context = signal.execute(context)
            except Exception as exc:
                logger.error("Signal %s failed: %s", signal.signal_id, exc)
                context.results[signal.signal_id] = SignalResult(
                    signal_id=signal.signal_id,
                    executed=True,
                    value=None,
                    metadata={},
                    skip_reason=None,
                    error=str(exc),
                )
        return context
