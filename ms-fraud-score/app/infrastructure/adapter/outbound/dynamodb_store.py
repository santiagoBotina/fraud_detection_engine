from datetime import datetime

from app.domain.port.score_store import ScoreStore


class DynamoDBScoreStore(ScoreStore):
    """DynamoDB-backed implementation of the ScoreStore port."""

    def __init__(self, table) -> None:
        self._table = table

    def save(self, transaction_id: str, score: int, calculated_at: datetime) -> None:
        self._table.put_item(
            Item={
                "transaction_id": transaction_id,
                "fraud_score": score,
                "calculated_at": calculated_at.isoformat(),
            }
        )
