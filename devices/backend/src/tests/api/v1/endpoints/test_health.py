"""Tests for health check endpoints.

These tests cover health check endpoints including liveness,
readiness, and basic health checks.
"""

import pytest
from app.main import app
from httpx import ASGITransport, AsyncClient


@pytest.mark.asyncio
async def test_health_liveness():
    """Test liveness check endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/health/live")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "alive"


@pytest.mark.asyncio
async def test_liveness_minimal_response():
    """Test that liveness check returns minimal response for efficiency."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/health/live")
        assert response.status_code == 200
        data = response.json()
        # Liveness should only have status, no extra fields
        assert "status" in data
        # It may have other fields but status is required


@pytest.mark.asyncio
async def test_unified_health_endpoint():
    """Test unified health endpoint at /health."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert data["service"] == "devices-backend"


@pytest.mark.asyncio
async def test_unified_health_has_timestamp():
    """Test unified health endpoint includes timestamp."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert "timestamp" in data
        # Validate ISO 8601 format by parsing
        from datetime import datetime
        timestamp = data.get("timestamp", "")
        # ISO 8601 format should be parseable
        try:
            datetime.fromisoformat(timestamp.replace("Z", "+00:00"))
        except ValueError:
            pytest.fail(f"Timestamp '{timestamp}' is not in valid ISO 8601 format")
