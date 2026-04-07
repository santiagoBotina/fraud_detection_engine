from abc import ABC, abstractmethod

from app.domain.entity.fraud_signal_result import FraudSignalResult


class ScorePublisher(ABC):
    """Port for publishing computed fraud signal results."""

    @abstractmethod
    def publish(self, result: FraudSignalResult) -> None: ...
