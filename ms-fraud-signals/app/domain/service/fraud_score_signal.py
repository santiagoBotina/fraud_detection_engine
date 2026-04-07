from __future__ import annotations

from app.domain.entity.signal_context import SignalContext
from app.domain.entity.signal_result import SignalResult
from app.domain.service.fuzzy_logic_scorer import FuzzyLogicScorer
from app.domain.service.signal import Signal


class FraudScoreSignal(Signal):
    def __init__(self, scorer: FuzzyLogicScorer) -> None:
        self._scorer = scorer

    @property
    def signal_id(self) -> str:
        return "fraud-score"

    def should_skip(self, context: SignalContext) -> tuple[bool, str | None]:
        return (False, None)

    def execute(self, context: SignalContext) -> SignalContext:
        score = self._scorer.compute(
            context.request.amount_in_cents,
            context.request.payment_method,
            context.request.customer_ip_address,
        )
        context.results[self.signal_id] = SignalResult(
            signal_id=self.signal_id,
            executed=True,
            value=float(score),
            metadata={},
            skip_reason=None,
            error=None,
        )
        return context
