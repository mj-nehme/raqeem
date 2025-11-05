import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app


@pytest.mark.asyncio
async def test_cors_preflight_request():
    """Test that CORS preflight (OPTIONS) requests are handled correctly."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Simulate a browser preflight request
        response = await ac.options(
            "/api/v1/screenshots/",
            headers={
                "Origin": "http://localhost:4002",
                "Access-Control-Request-Method": "POST",
                "Access-Control-Request-Headers": "content-type",
            }
        )
    
    # Should return 200 OK for preflight
    assert response.status_code == 200
    
    # Check required CORS headers are present
    assert "access-control-allow-origin" in response.headers
    assert "access-control-allow-methods" in response.headers
    assert "access-control-allow-headers" in response.headers
    
    # Verify the origin is allowed
    assert response.headers["access-control-allow-origin"] == "http://localhost:4002"


@pytest.mark.asyncio
async def test_cors_actual_get_request():
    """Test that actual GET requests include CORS headers."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Simulate an actual cross-origin GET request (doesn't need DB)
        response = await ac.get(
            "/health",
            headers={
                "Origin": "http://localhost:4002",
            }
        )
    
    # Check CORS headers are present in the response
    assert "access-control-allow-origin" in response.headers
    assert response.headers["access-control-allow-origin"] == "http://localhost:4002"


@pytest.mark.asyncio
async def test_cors_allows_multiple_origins():
    """Test that CORS configuration allows multiple configured origins."""
    test_origins = [
        "http://localhost:4000",
        "http://localhost:4002",
        "http://localhost:5002",
    ]
    
    for origin in test_origins:
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.options(
                "/api/v1/screenshots/",
                headers={
                    "Origin": origin,
                    "Access-Control-Request-Method": "POST",
                }
            )
        
        assert response.status_code == 200
        assert "access-control-allow-origin" in response.headers
        assert response.headers["access-control-allow-origin"] == origin
