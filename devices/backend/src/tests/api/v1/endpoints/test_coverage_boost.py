"""
Simple tests to boost coverage for endpoints that just need GET requests tested.
These tests call endpoints without complex database setup.
"""
import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app


# Test devices list endpoints
@pytest.mark.asyncio
async def test_get_devices_list():
    """Test GET /api/v1/devices/ endpoint returns list."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/")
        assert response.status_code == 200
        assert isinstance(response.json(), list)


@pytest.mark.asyncio
async def test_get_devices_processes():
    """Test GET /api/v1/devices/processes endpoint returns list."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/processes")
        assert response.status_code == 200
        assert isinstance(response.json(), list)


@pytest.mark.asyncio
async def test_get_devices_activities():
    """Test GET /api/v1/devices/activities endpoint returns list."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/activities")
        assert response.status_code == 200
        assert isinstance(response.json(), list)


@pytest.mark.asyncio
async def test_get_devices_alerts():
    """Test GET /api/v1/devices/alerts endpoint returns list."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/alerts")
        assert response.status_code == 200
        assert isinstance(response.json(), list)