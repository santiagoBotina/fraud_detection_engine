"""Use case for computing fraud scores.

Implements Requirements 2.1, 3.1, 3.2, 3.3, 3.4, 4.1.
"""

import logging
from datetime import datetime, timezone

from app.domain.entity.fraud_score_request import FraudScoreRequest
from app.domain.entity.fraud_score_result import FraudScoreResult
from app.domain.port.score_cache import ScoreCache
from app.domain.port.score_publisher import ScorePublisher
from app.domain.port.score_store import ScoreStore
from app.domain.service.fuzzy_logic_scorer import FuzzyLogicScorer

logger = logging.getLogger(__name__)


class ComputeFraudScoreUseCase:
    """Orchestrates fraud score computation, persistence, caching, and publishing."""

    def __init__(
        self,
        scorer: FuzzyLogicScorer,
        publisher: ScorePublisher,
        cache: ScoreCache,
        store: ScoreStore,
    ) -> None:
        self._scorer = scorer
        self._publisher = publisher
        self._cache = cache
        self._store = store

    def execute(self, request: FraudScoreRequest) -> FraudScoreResult:
        """Compute a fraud score for the given request and persist/publish the result.

        Args:
            request: The fraud score request containing transaction attributes.

        Returns:
            The computed FraudScoreResult.
        """
        score = self._scorer.compute(
            request.amount_in_cents,
            request.payment_method,
            request.customer_ip_address,
        )

        logger.info(
            "Computed fraud score for transaction %s: score=%d (amount=%d, method=%s, ip=%s)",
            request.transaction_id,
            score,
            request.amount_in_cents,
            request.payment_method,
            request.customer_ip_address,
        )

        result = FraudScoreResult(
            transaction_id=request.transaction_id,
            fraud_score=score,
            calculated_at=datetime.now(timezone.utc),
        )

        # Persist to DynamoDB — failure is non-fatal (Req 3.4)
        try:
            self._store.save(result.transaction_id, result.fraud_score, result.calculated_at)
            logger.info("Persisted score to DynamoDB for transaction %s", result.transaction_id)
        except Exception as e:
            logger.error("DynamoDB store failed: %s", e)

        # Cache in Redis — failure is non-fatal (Req 3.3)
        try:
            self._cache.set(result.transaction_id, result.fraud_score)
            logger.info("Cached score in Redis for transaction %s", result.transaction_id)
        except Exception as e:
            logger.error("Redis cache failed: %s", e)

        # Publish to Kafka (retries handled in adapter)
        self._publisher.publish(result)

        return result
