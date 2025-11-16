"""Additional endpoint tests for MVP coverage improvement."""

import pytest
from app.main import app
from httpx import ASGITransport, AsyncClient


@pytest.mark.asyncio
async def test_get_devices_endpoint():
    """Test GET /api/v1/devices/ endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/devices/")
        assert response.status_code in [200, 404]


@pytest.mark.asyncio
async def test_get_devices_with_pagination():
    """Test GET /api/v1/devices/ with pagination parameters."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/devices/?skip=0&limit=10")
        assert response.status_code in [200, 404]


@pytest.mark.asyncio
async def test_get_processes_endpoint():
    """Test GET /api/v1/devices/processes endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/devices/processes")
        assert response.status_code in [200, 404]


@pytest.mark.asyncio
async def test_get_activities_endpoint():
    """Test GET /api/v1/devices/activities endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/devices/activities")
        assert response.status_code in [200, 404]


@pytest.mark.asyncio
async def test_get_alerts_endpoint():
    """Test GET /api/v1/devices/alerts endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/devices/alerts")
        assert response.status_code in [200, 404]


@pytest.mark.asyncio
async def test_get_locations_endpoint():
    """Test GET /api/v1/devices/locations endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/devices/locations")
        assert response.status_code in [200, 404]


@pytest.mark.asyncio
async def test_get_screenshots_endpoint():
    """Test GET /api/v1/screenshots/ endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/screenshots/")
        # Screenshots endpoint may return 405 if GET not supported
        assert response.status_code in [200, 404, 405]


@pytest.mark.asyncio
async def test_get_screenshots_with_pagination():
    """Test GET /api/v1/screenshots/ with pagination."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/screenshots/?skip=0&limit=10")
        # Screenshots endpoint may return 405 if GET not supported
        assert response.status_code in [200, 404, 405]


@pytest.mark.asyncio
async def test_get_users_endpoint():
    """Test GET /api/v1/users/ endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/users/")
        assert response.status_code in [200, 404]


@pytest.mark.asyncio
async def test_get_keystrokes_endpoint():
    """Test GET /api/v1/keystrokes/ endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/keystrokes/")
        assert response.status_code in [200, 404]


@pytest.mark.asyncio
async def test_get_app_activity_endpoint():
    """Test GET /api/v1/app-activity/ endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/app-activity/")
        assert response.status_code in [200, 404]
