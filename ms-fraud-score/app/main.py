"""FastAPI application entrypoint with Kafka consumer lifecycle management.

Wires all adapters (Kafka, Redis, DynamoDB) and domain services following
hexagonal architecture. Starts the Kafka consumer in a background thread
on application startup and stops it on shutdown.

Implements Requirements 6.1, 6.2, 6.5.
"""

from __future__ import annotations

import logging
import threading
from contextlib import asynccontextmanager

import boto3
import redis
from confluent_kafka import Consumer, Producer
from fastapi import FastAPI

from app.config import config
from app.domain.service.fuzzy_logic_scorer import FuzzyLogicScorer
from app.domain.usecase.compute_fraud_score import ComputeFraudScoreUseCase
from app.infrastructure.adapter.inbound.kafka_consumer import FraudScoreRequestConsumer
from app.infrastructure.adapter.outbound.dynamodb_store import DynamoDBScoreStore
from app.infrastructure.adapter.outbound.kafka_publisher import KafkaScorePublisher
from app.infrastructure.adapter.outbound.redis_cache import RedisScoreCache
from app.infrastructure.health import health_router

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

_consumer_instance: FraudScoreRequestConsumer | None = None
_consumer_thread: threading.Thread | None = None


def _build_kafka_producer() -> Producer:
    return Producer({"bootstrap.servers": config.KAFKA_BROKER_ADDRESS})


def _build_kafka_consumer() -> Consumer:
    return Consumer(
        {
            "bootstrap.servers": config.KAFKA_BROKER_ADDRESS,
            "group.id": config.KAFKA_CONSUMER_GROUP,
            "auto.offset.reset": "earliest",
        }
    )


def _build_redis_client() -> redis.Redis:
    return redis.Redis(host=config.REDIS_HOST, port=config.REDIS_PORT, decode_responses=True)


def _build_dynamodb_table():
    dynamodb = boto3.resource(
        "dynamodb",
        endpoint_url=config.DYNAMO_DB_ENDPOINT,
        region_name=config.AWS_REGION,
        aws_access_key_id=config.AWS_ACCESS_KEY_ID,
        aws_secret_access_key=config.AWS_SECRET_ACCESS_KEY,
    )
    return dynamodb.Table(config.DYNAMO_DB_FRAUD_SCORES_TABLE)


def _wire_consumer() -> FraudScoreRequestConsumer:
    """Create and wire all dependencies, returning the Kafka consumer adapter."""
    producer = _build_kafka_producer()
    consumer = _build_kafka_consumer()
    redis_client = _build_redis_client()
    dynamo_table = _build_dynamodb_table()

    scorer = FuzzyLogicScorer()
    publisher = KafkaScorePublisher(producer, config.KAFKA_FRAUD_SCORE_CALCULATED_TOPIC)
    cache = RedisScoreCache(redis_client)
    store = DynamoDBScoreStore(dynamo_table)

    use_case = ComputeFraudScoreUseCase(
        scorer=scorer,
        publisher=publisher,
        cache=cache,
        store=store,
    )

    return FraudScoreRequestConsumer(
        consumer=consumer,
        topic=config.KAFKA_FRAUD_SCORE_REQUEST_TOPIC,
        use_case=use_case,
    )


@asynccontextmanager
async def lifespan(application: FastAPI):
    """Manage Kafka consumer lifecycle: start on startup, stop on shutdown."""
    global _consumer_instance, _consumer_thread

    logger.info("Starting Kafka consumer for topic %s", config.KAFKA_FRAUD_SCORE_REQUEST_TOPIC)
    _consumer_instance = _wire_consumer()
    _consumer_thread = threading.Thread(target=_consumer_instance.start, daemon=True)
    _consumer_thread.start()

    yield

    logger.info("Stopping Kafka consumer")
    if _consumer_instance is not None:
        _consumer_instance.stop()
    if _consumer_thread is not None:
        _consumer_thread.join(timeout=5.0)


app = FastAPI(title="Fraud Score Service", lifespan=lifespan)
app.include_router(health_router)
