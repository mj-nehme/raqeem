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
    # Status can be "ok" or "degraded" depending on database availability
    assert data["status"] in ["ok", "degraded"]
    assert data["service"] == "devices-backend"
    # Database status should be present
    assert "database" in data


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


@pytest.mark.asyncio
async def test_lifespan_init_db_exception():
    """Test that lifespan handles init_db exception gracefully."""
    from unittest.mock import patch, AsyncMock
    from app.main import lifespan
    
    # Mock init_db to raise an exception
    with patch('app.main.init_db', new=AsyncMock(side_effect=Exception("DB init failed"))):
        async with lifespan(app):
            # Should not raise - exception is caught
            pass


@pytest.mark.asyncio  
async def test_lifespan_context():
    """Test lifespan context manager."""
    from app.main import lifespan
    
    # Test that lifespan works as context manager
    async with lifespan(app):
        # In context
        assert app is not None
    # After context - should complete successfully
