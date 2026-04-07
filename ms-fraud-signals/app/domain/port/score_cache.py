from __future__ import annotations

from abc import ABC, abstractmethod


class ScoreCache(ABC):
    """Port for caching fraud scores."""

    @abstractmethod
    def set(self, transaction_id: str, score: int) -> None: ...

    @abstractmethod
    def get(self, transaction_id: str) -> int | None: ...
