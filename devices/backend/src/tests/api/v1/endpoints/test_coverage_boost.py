"""
Simple tests to boost coverage for endpoints that just need GET requests tested.
These tests call endpoints without complex database setup.

NOTE: Read-only list endpoints have been moved to the mentor backend.
Tests verify that these endpoints return 404 on the devices backend, as expected.
"""

import pytest
from app.main import app
from httpx import ASGITransport, AsyncClient


# Test devices list endpoints - now return 404 since moved to mentor backend
@pytest.mark.asyncio
async def test_get_devices_list():
    """Test GET /api/v1/devices/ endpoint - returns 404 (moved to mentor)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/")
        # Endpoint moved to mentor backend - expect 404
        assert response.status_code == 404


@pytest.mark.asyncio
async def test_get_devices_processes():
    """Test GET /api/v1/devices/processes endpoint - returns 404 (moved to mentor)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/processes")
        # Endpoint moved to mentor backend - expect 404
        assert response.status_code == 404


@pytest.mark.asyncio
async def test_get_devices_activities():
    """Test GET /api/v1/devices/activities endpoint - returns 404 (moved to mentor)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/activities")
        # Endpoint moved to mentor backend - expect 404
        assert response.status_code == 404


@pytest.mark.asyncio
async def test_get_devices_alerts():
    """Test GET /api/v1/devices/alerts endpoint - returns 404 (moved to mentor)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/alerts")
        # Endpoint moved to mentor backend - expect 404
        assert response.status_code == 404
