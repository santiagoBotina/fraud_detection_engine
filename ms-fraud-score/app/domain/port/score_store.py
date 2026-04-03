from abc import ABC, abstractmethod
from datetime import datetime


class ScoreStore(ABC):
    """Port for persisting fraud scores."""

    @abstractmethod
    def save(self, transaction_id: str, score: int, calculated_at: datetime) -> None: ...
