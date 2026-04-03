"""Property-based test for Redis cache round-trip.

Feature: fraud-score-service, Property 9: Redis cache round-trip
Validates: Requirements 8.3
"""

import fakeredis
from hypothesis import given, settings
from hypothesis import strategies as st

from app.infrastructure.adapter.outbound.redis_cache import RedisScoreCache


@given(
    transaction_id=st.text(min_size=1),
    score=st.integers(min_value=0, max_value=100),
)
@settings(max_examples=100)
def test_redis_cache_round_trip(transaction_id: str, score: int) -> None:
    """For any transaction ID and fraud score (0-100), storing the score
    in the Redis cache and then retrieving it by transaction ID returns
    the same score.

    Feature: fraud-score-service, Property 9: Redis cache round-trip
    Validates: Requirements 8.3
    """
    client = fakeredis.FakeRedis()
    cache = RedisScoreCache(client)

    cache.set(transaction_id, score)

    raw = client.get(f"fraud_score:{transaction_id}")
    assert raw is not None
    assert int(raw) == score
