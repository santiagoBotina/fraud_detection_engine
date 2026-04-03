from __future__ import annotations

import json
import logging

from confluent_kafka import KafkaException, Producer

from app.domain.entity.fraud_score_result import FraudScoreResult
from app.domain.port.score_publisher import ScorePublisher

logger = logging.getLogger(__name__)

MAX_RETRIES = 3


class KafkaScorePublisher(ScorePublisher):
    """Publishes FraudScoreResult to the FraudScore.Calculated Kafka topic."""

    def __init__(self, producer: Producer, topic: str) -> None:
        self._producer = producer
        self._topic = topic

    def publish(self, result: FraudScoreResult) -> None:
        key = result.transaction_id.encode("utf-8")
        value = json.dumps(result.to_dict()).encode("utf-8")

        last_error: Exception | None = None
        for attempt in range(1, MAX_RETRIES + 1):
            try:
                self._producer.produce(
                    topic=self._topic,
                    key=key,
                    value=value,
                )
                self._producer.flush()
                return
            except (KafkaException, BufferError) as exc:
                last_error = exc
                logger.error(
                    "Kafka publish attempt %d/%d failed for transaction %s: %s",
                    attempt,
                    MAX_RETRIES,
                    result.transaction_id,
                    exc,
                )

        logger.error(
            "Failed to publish score for transaction %s after %d retries",
            result.transaction_id,
            MAX_RETRIES,
        )
        raise last_error  # type: ignore[misc]
