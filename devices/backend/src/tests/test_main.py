import pytest
from app.main import app
from httpx import ASGITransport, AsyncClient


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
    assert hasattr(app, "router")


@pytest.mark.asyncio
async def test_api_router_included():
    """Test that API router is included with correct prefix."""
    # Check that routes are registered
    routes = [getattr(route, "path", None) for route in app.routes if hasattr(route, "path")]
    assert any("/api/v1" in route for route in routes if route)


@pytest.mark.asyncio
async def test_lifespan_context():
    """Test lifespan context manager."""
    from app.main import lifespan

    # Test that lifespan works as context manager
    async with lifespan(app):
        # In context
        assert app is not None
    # After context - should complete successfully


@pytest.mark.asyncio
async def test_root_redirect():
    """Test root endpoint redirects to docs."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test", follow_redirects=False) as ac:
        response = await ac.get("/")
    # RedirectResponse uses 307 by default
    assert response.status_code == 307
    assert response.headers.get("location") == "/docs"


@pytest.mark.asyncio
async def test_devices_patch_returns_405():
    """Test PATCH /api/v1/devices/ returns 405 Method Not Allowed."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.patch("/api/v1/devices/")
    assert response.status_code == 405


@pytest.mark.asyncio
async def test_devices_patch_without_trailing_slash_returns_405():
    """Test PATCH /api/v1/devices returns 405 Method Not Allowed."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.patch("/api/v1/devices")
    assert response.status_code == 405


@pytest.mark.asyncio
async def test_devices_get_returns_404():
    """Test GET /api/v1/devices/ returns 404 (endpoint moved to mentor)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/")
    assert response.status_code == 404


@pytest.mark.asyncio
async def test_health_check_has_timestamp():
    """Test health check includes timestamp field."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert "timestamp" in data
