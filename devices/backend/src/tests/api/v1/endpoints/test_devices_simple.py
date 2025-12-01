import pytest
from app.main import app
from httpx import ASGITransport, AsyncClient

# NOTE: Read-only list endpoints (list_processes, list_activities, list_alerts)
# have been moved to the mentor backend. These tests now verify that the
# endpoints return 404 on the devices backend, as expected.


@pytest.mark.asyncio
async def test_list_processes():
    """Test GET /api/v1/devices/processes endpoint returns 404 (moved to mentor)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/processes")
        # Endpoint moved to mentor backend - expect 404
        assert response.status_code == 404


@pytest.mark.asyncio
async def test_list_activities():
    """Test GET /api/v1/devices/activities endpoint returns 404 (moved to mentor)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/activities")
        # Endpoint moved to mentor backend - expect 404
        assert response.status_code == 404


@pytest.mark.asyncio
async def test_list_alerts():
    """Test GET /api/v1/devices/alerts endpoint returns 404 (moved to mentor)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/alerts")
        # Endpoint moved to mentor backend - expect 404
        assert response.status_code == 404
