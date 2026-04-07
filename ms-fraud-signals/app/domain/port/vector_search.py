from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass


@dataclass(frozen=True)
class SimilarityMatch:
    transaction_id: str
    score: float          # cosine similarity 0–1
    status: str           # "approved" or "declined"


class VectorSearch(ABC):
    """Port for querying vector similarity against historical transactions."""

    @abstractmethod
    def find_similar(self, embedding: list[float], top_k: int) -> list[SimilarityMatch]: ...
