import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app


@pytest.mark.asyncio
async def test_health_check():
    """Test the health check endpoint."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"
    assert data["service"] == "devices-backend"


@pytest.mark.asyncio
async def test_app_startup():
    """Test that the app starts up correctly."""
    # The app should have lifespan context manager
    assert app is not None
    assert hasattr(app, 'router')


@pytest.mark.asyncio
async def test_api_router_included():
    """Test that API router is included with correct prefix."""
    # Check that routes are registered
    routes = [route.path for route in app.routes]
    assert any('/api/v1' in route for route in routes)
