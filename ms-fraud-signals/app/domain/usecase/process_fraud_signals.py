"""Use case for processing fraud signals through the pipeline.

Replaces ComputeFraudScoreUseCase with pipeline-based signal processing.
Implements Requirements 1.4, 5.1, 5.2, 5.3, 9.2, 9.3, 10.1, 10.2, 10.3.
"""

from __future__ import annotations

import json
import logging
from datetime import datetime, timezone

from app.domain.entity.fraud_signal_request import FraudSignalRequest
from app.domain.entity.fraud_signal_result import FraudSignalResult
from app.domain.entity.signal_context import SignalContext
from app.domain.port.score_cache import ScoreCache
from app.domain.port.score_publisher import ScorePublisher
from app.domain.port.score_store import ScoreStore
from app.domain.service.signal_pipeline import SignalPipeline

logger = logging.getLogger(__name__)


def _clamp(value: int, lo: int, hi: int) -> int:
    return max(lo, min(hi, value))


class ProcessFraudSignalsUseCase:
    """Orchestrates fraud signal processing, persistence, caching, and publishing."""

    def __init__(
        self,
        pipeline: SignalPipeline,
        publisher: ScorePublisher,
        cache: ScoreCache,
        store: ScoreStore,
    ) -> None:
        self._pipeline = pipeline
        self._publisher = publisher
        self._cache = cache
        self._store = store

    def execute(self, request: FraudSignalRequest) -> FraudSignalResult:
        """Process fraud signals for the given request.

        Runs the signal pipeline, computes the final clamped score,
        persists/caches/publishes the result.
        """
        context = SignalContext(request=request)
        context = self._pipeline.run(context)

        base_score_result = context.results.get("fraud-score")
        base_score = int(base_score_result.value) if base_score_result and base_score_result.value is not None else 0

        similarity_result = context.results.get("similarity")
        if similarity_result and similarity_result.executed and similarity_result.value is not None:
            adjustment = int(similarity_result.value)
        else:
            adjustment = 0

        fraud_score = _clamp(base_score + adjustment, 0, 100)
        calculated_at = datetime.now(timezone.utc)

        result = FraudSignalResult(
            transaction_id=request.transaction_id,
            fraud_score=fraud_score,
            signal_results=list(context.results.values()),
            calculated_at=calculated_at,
        )

        signals_summary = json.dumps([
            {"signal_id": sr.signal_id, "executed": sr.executed, "value": sr.value}
            for sr in result.signal_results
        ])

        # Persist to DynamoDB — failure is non-fatal (Req 10.1)
        try:
            self._store.save(request.transaction_id, fraud_score, calculated_at, signals_summary)
            logger.info("Persisted score to DynamoDB for transaction %s", request.transaction_id)
        except Exception:
            logger.exception("DynamoDB store failed for transaction %s", request.transaction_id)

        # Cache in Redis — failure is non-fatal (Req 10.2)
        try:
            self._cache.set(request.transaction_id, fraud_score)
            logger.info("Cached score in Redis for transaction %s", request.transaction_id)
        except Exception:
            logger.exception("Redis cache failed for transaction %s", request.transaction_id)

        # Publish to Kafka — errors propagate (Req 9.2)
        self._publisher.publish(result)

        return result
