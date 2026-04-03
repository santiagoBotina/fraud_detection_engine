"""Kafka consumer adapter for FraudScore.Request topic.

Consumes fraud score request messages, deserializes them, and delegates
to ComputeFraudScoreUseCase for processing.

Implements Requirements 2.1, 7.3.
"""

from __future__ import annotations

import json
import logging

from confluent_kafka import Consumer, KafkaError, Message

from app.domain.entity.fraud_score_request import FraudScoreRequest
from app.domain.usecase.compute_fraud_score import ComputeFraudScoreUseCase

logger = logging.getLogger(__name__)


class FraudScoreRequestConsumer:
    """Inbound Kafka adapter that consumes from the FraudScore.Request topic."""

    def __init__(
        self,
        consumer: Consumer,
        topic: str,
        use_case: ComputeFraudScoreUseCase,
    ) -> None:
        self._consumer = consumer
        self._topic = topic
        self._use_case = use_case
        self._running = False

    def start(self) -> None:
        """Subscribe to the topic and poll for messages in a loop.

        Runs until ``_running`` is set to False. Malformed or invalid
        messages are logged and skipped without crashing.
        """
        self._running = True
        self._consumer.subscribe([self._topic])
        logger.info("Subscribed to topic %s", self._topic)

        try:
            while self._running:
                msg: Message | None = self._consumer.poll(timeout=1.0)
                if msg is None:
                    continue

                error = msg.error()
                if error is not None:
                    if error.code() == KafkaError._PARTITION_EOF:
                        continue
                    logger.error("Kafka consumer error: %s", error)
                    continue

                self._handle_message(msg)
        finally:
            self._consumer.close()

    def stop(self) -> None:
        """Signal the consumer loop to stop."""
        self._running = False

    def _handle_message(self, msg: Message) -> None:
        """Deserialize a single message and invoke the use case."""
        raw = msg.value()
        logger.info(
            "Message received on %s [partition=%s offset=%s key=%s]",
            msg.topic(),
            msg.partition(),
            msg.offset(),
            msg.key(),
        )
        try:
            data = json.loads(raw)
            request = FraudScoreRequest.from_dict(data)
        except (json.JSONDecodeError, KeyError, TypeError) as exc:
            logger.error(
                "Malformed message on %s [partition=%s offset=%s]: %s – raw: %s",
                msg.topic(),
                msg.partition(),
                msg.offset(),
                exc,
                raw,
            )
            return

        try:
            self._use_case.execute(request)
            logger.info(
                "Successfully processed fraud score request for transaction %s",
                request.transaction_id,
            )
        except Exception:
            logger.exception(
                "Failed to process message for transaction %s",
                request.transaction_id,
            )
