"""FastAPI application entrypoint with Kafka consumer lifecycle management.

Wires all adapters (Kafka, Redis, DynamoDB, Qdrant) and domain services
following hexagonal architecture. Starts the Kafka consumer in a background
thread on application startup and stops it on shutdown.

Implements Requirements 1.1, 1.2, 5.1, 5.2, 5.3, 6.2, 9.1, 9.2, 11.3.
"""

from __future__ import annotations

import logging
import threading
from contextlib import asynccontextmanager

import boto3
import redis
from confluent_kafka import Consumer, Producer
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware

from app.config import config
from app.domain.port.score_cache import ScoreCache
from app.domain.port.score_store import ScoreStore
from app.domain.service.fraud_score_signal import FraudScoreSignal
from app.domain.service.fuzzy_logic_scorer import FuzzyLogicScorer
from app.domain.service.signal_pipeline import SignalPipeline
from app.domain.service.similarity_signal import SimilaritySignal
from app.domain.usecase.process_fraud_signals import ProcessFraudSignalsUseCase
from app.infrastructure.adapter.inbound.kafka_consumer import FraudScoreRequestConsumer
from app.infrastructure.adapter.outbound.dynamodb_store import DynamoDBScoreStore
from app.infrastructure.adapter.outbound.kafka_publisher import KafkaScorePublisher
from app.infrastructure.adapter.outbound.qdrant_search import QdrantVectorSearch
from app.infrastructure.adapter.outbound.redis_cache import RedisScoreCache
from app.infrastructure.health import health_router

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

_consumer_instance: FraudScoreRequestConsumer | None = None
_consumer_thread: threading.Thread | None = None
_score_cache: ScoreCache | None = None
_score_store: ScoreStore | None = None


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


def _wire_consumer() -> tuple[FraudScoreRequestConsumer, RedisScoreCache, DynamoDBScoreStore]:
    """Create and wire all dependencies, returning the Kafka consumer adapter and data adapters."""
    producer = _build_kafka_producer()
    consumer = _build_kafka_consumer()
    redis_client = _build_redis_client()
    dynamo_table = _build_dynamodb_table()

    scorer = FuzzyLogicScorer()
    publisher = KafkaScorePublisher(producer, config.KAFKA_FRAUD_SIGNALS_CALCULATED_TOPIC)
    cache = RedisScoreCache(redis_client)
    store = DynamoDBScoreStore(dynamo_table)

    # Build signal pipeline with registered signals
    pipeline = SignalPipeline()
    pipeline.register(FraudScoreSignal(scorer))

    vector_search = QdrantVectorSearch(host=config.QDRANT_HOST, port=config.QDRANT_PORT)
    pipeline.register(SimilaritySignal(vector_search))

    use_case = ProcessFraudSignalsUseCase(
        pipeline=pipeline,
        publisher=publisher,
        cache=cache,
        store=store,
    )

    kafka_consumer = FraudScoreRequestConsumer(
        consumer=consumer,
        topic=config.KAFKA_FRAUD_SIGNALS_REQUEST_TOPIC,
        use_case=use_case,
    )

    return kafka_consumer, cache, store


@asynccontextmanager
async def lifespan(application: FastAPI):
    """Manage Kafka consumer lifecycle: start on startup, stop on shutdown."""
    global _consumer_instance, _consumer_thread, _score_cache, _score_store

    logger.info("Starting Kafka consumer for topic %s", config.KAFKA_FRAUD_SIGNALS_REQUEST_TOPIC)
    _consumer_instance, _score_cache, _score_store = _wire_consumer()
    _consumer_thread = threading.Thread(target=_consumer_instance.start, daemon=True)
    _consumer_thread.start()

    yield

    logger.info("Stopping Kafka consumer")
    if _consumer_instance is not None:
        _consumer_instance.stop()
    if _consumer_thread is not None:
        _consumer_thread.join(timeout=5.0)


app = FastAPI(title="Fraud Score Service", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(health_router)


@app.get("/scores/{transaction_id}")
async def get_score(transaction_id: str):
    """Retrieve fraud score for a transaction.

    Checks Redis cache first, falls back to DynamoDB.
    Returns 404 if not found in either.
    """
    # Cache-first lookup
    if _score_cache is not None:
        try:
            cached_score = _score_cache.get(transaction_id)
            if cached_score is not None:
                return {
                    "transaction_id": transaction_id,
                    "fraud_score": cached_score,
                    "calculated_at": None,
                }
        except Exception:
            logger.warning("Redis lookup failed for %s, falling back to DynamoDB", transaction_id)

    # Fallback to DynamoDB
    if _score_store is not None:
        record = _score_store.get(transaction_id)
        if record is not None:
            response = {
                "transaction_id": record["transaction_id"],
                "fraud_score": record["fraud_score"],
                "calculated_at": record["calculated_at"],
            }
            if "signals_summary" in record:
                response["signals_summary"] = record["signals_summary"]
            return response

    raise HTTPException(
        status_code=404,
        detail=f"Fraud score not found for transaction_id: {transaction_id}",
    )
