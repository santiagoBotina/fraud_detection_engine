from __future__ import annotations

from qdrant_client import QdrantClient
from qdrant_client.models import ScoredPoint

from app.domain.port.vector_search import SimilarityMatch, VectorSearch

COLLECTION_NAME = "transaction-embeddings"


class QdrantVectorSearch(VectorSearch):
    """Qdrant-backed implementation of the VectorSearch port."""

    def __init__(self, host: str, port: int) -> None:
        self._client = QdrantClient(host=host, port=port)

    def find_similar(self, embedding: list[float], top_k: int) -> list[SimilarityMatch]:
        results: list[ScoredPoint] = self._client.query_points(
            collection_name=COLLECTION_NAME,
            query=embedding,
            limit=top_k,
        ).points

        return [
            SimilarityMatch(
                transaction_id=point.payload["transaction_id"],
                score=point.score,
                status=point.payload["status"],
            )
            for point in results
        ]
