import redis

from app.domain.port.score_cache import ScoreCache

_TTL_SECONDS = 3600


class RedisScoreCache(ScoreCache):
    """Redis-backed implementation of the ScoreCache port."""

    def __init__(self, client: redis.Redis) -> None:
        self._client = client

    def set(self, transaction_id: str, score: int) -> None:
        key = f"fraud_score:{transaction_id}"
        self._client.set(key, score, ex=_TTL_SECONDS)
