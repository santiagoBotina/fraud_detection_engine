"""Similarity signal — vector-search-based fraud adjustment.

Implements Requirements 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 4.4.
"""

from __future__ import annotations

import logging

from app.domain.entity.signal_context import SignalContext
from app.domain.entity.signal_result import SignalResult
from app.domain.port.vector_search import SimilarityMatch, VectorSearch
from app.domain.service.embedding import build_embedding
from app.domain.service.signal import Signal

logger = logging.getLogger(__name__)

_NEUTRAL_LOW = 30
_NEUTRAL_HIGH = 70
_ADJUSTMENT_CLAMP = 20


class SimilaritySignal(Signal):
    """Queries historical transactions for similarity-based fraud adjustment."""

    def __init__(self, vector_search: VectorSearch) -> None:
        self._vector_search = vector_search

    @property
    def signal_id(self) -> str:
        return "similarity"

    def should_skip(self, context: SignalContext) -> tuple[bool, str | None]:
        fraud_score_result = context.results.get("fraud-score")
        if fraud_score_result is None or fraud_score_result.value is None:
            return (True, "score-outside-neutral-range")
        score = fraud_score_result.value
        if score < _NEUTRAL_LOW or score > _NEUTRAL_HIGH:
            return (True, "score-outside-neutral-range")
        return (False, None)

    def execute(self, context: SignalContext) -> SignalContext:
        try:
            embedding = build_embedding(
                context.request.amount_in_cents,
                context.request.payment_method,
                context.request.customer_ip_address,
            )
            matches = self._vector_search.find_similar(embedding, top_k=5)
            adjustment = self._compute_adjustment(matches)
        except Exception as exc:
            logger.error("SimilaritySignal error: %s", exc)
            context.results[self.signal_id] = SignalResult(
                signal_id=self.signal_id,
                executed=True,
                value=0.0,
                metadata={},
                skip_reason=None,
                error=str(exc),
            )
            return context

        context.results[self.signal_id] = SignalResult(
            signal_id=self.signal_id,
            executed=True,
            value=float(adjustment),
            metadata={"matched_transactions": [m.transaction_id for m in matches]},
            skip_reason=None,
            error=None,
        )
        return context

    @staticmethod
    def _compute_adjustment(matches: list[SimilarityMatch]) -> float:
        raw = 0.0
        for match in matches:
            if match.status == "declined":
                raw += match.score * 10.0
            elif match.status == "approved":
                raw -= match.score * 10.0
        return max(-_ADJUSTMENT_CLAMP, min(_ADJUSTMENT_CLAMP, raw))
