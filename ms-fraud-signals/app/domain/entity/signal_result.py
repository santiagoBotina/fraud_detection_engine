from __future__ import annotations

from dataclasses import dataclass, field


@dataclass(frozen=True)
class SignalResult:
    """Represents the output of a single Signal execution."""

    signal_id: str
    executed: bool
    value: float | None
    metadata: dict = field(default_factory=dict)
    skip_reason: str | None = None
    error: str | None = None
