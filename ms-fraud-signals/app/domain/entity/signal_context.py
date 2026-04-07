from __future__ import annotations

from dataclasses import dataclass, field

from app.domain.entity.fraud_signal_request import FraudSignalRequest
from app.domain.entity.signal_result import SignalResult


@dataclass
class SignalContext:
    """Mutable accumulator passed through the signal pipeline."""

    request: FraudSignalRequest
    results: dict[str, SignalResult] = field(default_factory=dict)
    final_score: int | None = None
