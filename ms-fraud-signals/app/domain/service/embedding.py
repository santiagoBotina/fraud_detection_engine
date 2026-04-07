"""Transaction embedding builder for similarity search.

Implements Requirements 4.1, 8.3.
"""

from __future__ import annotations

from app.domain.service.fuzzy_logic_scorer import PAYMENT_METHOD_RISK, _ip_risk_score


def build_embedding(amount_in_cents: int, payment_method: str, customer_ip: str) -> list[float]:
    """Build a 3-dimensional embedding vector for similarity search."""
    amount_normalized = min(amount_in_cents / 1_000_000, 1.0)
    payment_risk = PAYMENT_METHOD_RISK.get(payment_method, 5) / 10.0
    ip_risk = _ip_risk_score(customer_ip) / 10.0
    return [amount_normalized, payment_risk, ip_risk]
