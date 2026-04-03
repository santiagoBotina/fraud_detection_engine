from abc import ABC, abstractmethod


class ScoreCache(ABC):
    """Port for caching fraud scores."""

    @abstractmethod
    def set(self, transaction_id: str, score: int) -> None: ...
