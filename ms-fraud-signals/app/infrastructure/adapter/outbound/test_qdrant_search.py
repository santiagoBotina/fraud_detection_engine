from __future__ import annotations

from unittest.mock import MagicMock, patch

import pytest

from app.domain.port.vector_search import SimilarityMatch
from app.infrastructure.adapter.outbound.qdrant_search import (
    COLLECTION_NAME,
    QdrantVectorSearch,
)


def _make_scored_point(transaction_id: str, score: float, status: str) -> MagicMock:
    point = MagicMock()
    point.payload = {"transaction_id": transaction_id, "status": status}
    point.score = score
    return point


@patch("app.infrastructure.adapter.outbound.qdrant_search.QdrantClient")
class TestQdrantVectorSearch:
    def test_find_similar_returns_similarity_matches(self, mock_client_cls):
        mock_client = mock_client_cls.return_value
        mock_response = MagicMock()
        mock_response.points = [
            _make_scored_point("txn-1", 0.95, "approved"),
            _make_scored_point("txn-2", 0.80, "declined"),
        ]
        mock_client.query_points.return_value = mock_response

        adapter = QdrantVectorSearch(host="localhost", port=6333)
        results = adapter.find_similar([0.5, 0.3, 0.7], top_k=5)

        assert len(results) == 2
        assert results[0] == SimilarityMatch(transaction_id="txn-1", score=0.95, status="approved")
        assert results[1] == SimilarityMatch(transaction_id="txn-2", score=0.80, status="declined")

    def test_find_similar_passes_correct_params_to_qdrant(self, mock_client_cls):
        mock_client = mock_client_cls.return_value
        mock_response = MagicMock()
        mock_response.points = []
        mock_client.query_points.return_value = mock_response

        adapter = QdrantVectorSearch(host="myhost", port=1234)
        adapter.find_similar([0.1, 0.2, 0.3], top_k=3)

        mock_client_cls.assert_called_once_with(host="myhost", port=1234)
        mock_client.query_points.assert_called_once_with(
            collection_name=COLLECTION_NAME,
            query=[0.1, 0.2, 0.3],
            limit=3,
        )

    def test_find_similar_empty_results(self, mock_client_cls):
        mock_client = mock_client_cls.return_value
        mock_response = MagicMock()
        mock_response.points = []
        mock_client.query_points.return_value = mock_response

        adapter = QdrantVectorSearch(host="localhost", port=6333)
        results = adapter.find_similar([0.5, 0.3, 0.7], top_k=5)

        assert results == []

    def test_find_similar_propagates_qdrant_errors(self, mock_client_cls):
        mock_client = mock_client_cls.return_value
        mock_client.query_points.side_effect = ConnectionError("Qdrant unreachable")

        adapter = QdrantVectorSearch(host="localhost", port=6333)

        with pytest.raises(ConnectionError, match="Qdrant unreachable"):
            adapter.find_similar([0.5, 0.3, 0.7], top_k=5)
