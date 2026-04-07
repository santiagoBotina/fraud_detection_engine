from __future__ import annotations

from abc import ABC, abstractmethod

from app.domain.entity.signal_context import SignalContext


class Signal(ABC):
    @property
    @abstractmethod
    def signal_id(self) -> str: ...

    @abstractmethod
    def execute(self, context: SignalContext) -> SignalContext: ...

    @abstractmethod
    def should_skip(self, context: SignalContext) -> tuple[bool, str | None]: ...
