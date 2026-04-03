from abc import ABC, abstractmethod

from app.domain.entity.fraud_score_result import FraudScoreResult


class ScorePublisher(ABC):
    """Port for publishing computed fraud score results."""

    @abstractmethod
    def publish(self, result: FraudScoreResult) -> None: ...
