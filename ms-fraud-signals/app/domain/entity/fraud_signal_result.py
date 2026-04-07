from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime

from app.domain.entity.signal_result import SignalResult


@dataclass(frozen=True)
class FraudSignalResult:
    """Aggregated output of the Signal Pipeline.

    Published to Kafka (FraudScore.Calculated) and persisted to DynamoDB.
    fraud_score is the final clamped score in [0, 100].
    """

    transaction_id: str
    fraud_score: int
    signal_results: list[SignalResult]
    calculated_at: datetime

    def to_dict(self) -> dict:
        """Serialize to a backward-compatible dictionary.

        Top-level keys preserve the existing FraudScore.Calculated schema.
        The ``signals`` list is additive.
        """
        return {
            "transaction_id": self.transaction_id,
            "fraud_score": self.fraud_score,
            "calculated_at": self.calculated_at.isoformat(),
            "signals": [
                {
                    "signal_id": sr.signal_id,
                    "executed": sr.executed,
                    "value": sr.value,
                }
                for sr in self.signal_results
            ],
        }
