from __future__ import annotations

from abc import ABC, abstractmethod
from datetime import datetime


class ScoreStore(ABC):
    """Port for persisting fraud scores."""

    @abstractmethod
    def save(self, transaction_id: str, score: int, calculated_at: datetime) -> None: ...

    @abstractmethod
    def get(self, transaction_id: str) -> dict | None: ...
