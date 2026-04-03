from __future__ import annotations

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

    def get(self, transaction_id: str) -> dict | None:
        response = self._table.get_item(Key={"transaction_id": transaction_id})
        item = response.get("Item")
        if item is None:
            return None
        return {
            "transaction_id": item["transaction_id"],
            "fraud_score": int(item["fraud_score"]),
            "calculated_at": item["calculated_at"],
        }
