"""FastAPI health check router.

Provides a /health endpoint returning service status.
Implements Requirement 6.1.
"""

from fastapi import APIRouter

health_router = APIRouter()


@health_router.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "ok"}
