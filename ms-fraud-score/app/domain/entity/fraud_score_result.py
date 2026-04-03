from dataclasses import dataclass
from datetime import datetime


@dataclass(frozen=True)
class FraudScoreResult:
    """Domain entity representing a computed fraud score for the FraudScore.Calculated Kafka topic.

    Maps to the FraudScore.Calculated Kafka message schema.
    fraud_score is an integer in [0, 100] inclusive.
    calculated_at serializes to/from ISO 8601 string format.
    """

    transaction_id: str
    fraud_score: int
    calculated_at: datetime

    def to_dict(self) -> dict:
        return {
            "transaction_id": self.transaction_id,
            "fraud_score": self.fraud_score,
            "calculated_at": self.calculated_at.isoformat(),
        }

    @classmethod
    def from_dict(cls, data: dict) -> "FraudScoreResult":
        calculated_at = data["calculated_at"]
        if isinstance(calculated_at, str):
            calculated_at = datetime.fromisoformat(calculated_at)
        return cls(
            transaction_id=data["transaction_id"],
            fraud_score=data["fraud_score"],
            calculated_at=calculated_at,
        )
