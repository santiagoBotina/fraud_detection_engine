from __future__ import annotations

from datetime import datetime

from app.domain.port.score_store import ScoreStore


class DynamoDBScoreStore(ScoreStore):
    """DynamoDB-backed implementation of the ScoreStore port."""

    def __init__(self, table) -> None:
        self._table = table

    def save(self, transaction_id: str, score: int, calculated_at: datetime, signals_summary: str = "") -> None:
        item: dict = {
            "transaction_id": transaction_id,
            "fraud_score": score,
            "calculated_at": calculated_at.isoformat(),
        }
        if signals_summary:
            item["signals_summary"] = signals_summary
        self._table.put_item(Item=item)

    def get(self, transaction_id: str) -> dict | None:
        response = self._table.get_item(Key={"transaction_id": transaction_id})
        item = response.get("Item")
        if item is None:
            return None
        result: dict = {
            "transaction_id": item["transaction_id"],
            "fraud_score": int(item["fraud_score"]),
            "calculated_at": item["calculated_at"],
        }
        if "signals_summary" in item:
            result["signals_summary"] = item["signals_summary"]
        return result
