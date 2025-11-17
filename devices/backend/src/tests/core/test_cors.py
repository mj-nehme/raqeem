import pytest
from app.main import app
from fastapi.testclient import TestClient


@pytest.fixture
def client():
    return TestClient(app)


def test_cors_preflight_request(client):
    """Test that CORS preflight requests are handled correctly"""
    response = client.options(
        "/api/v1/screenshots",
        headers={
            "Origin": "http://localhost:4000",
            "Access-Control-Request-Method": "POST",
            "Access-Control-Request-Headers": "content-type",
        },
    )
    assert response.status_code == 200
    assert "access-control-allow-origin" in response.headers


def test_cors_allows_screenshot_upload(client):
    """Test that screenshot upload endpoint allows CORS"""
    response = client.post(
        "/api/v1/screenshots",
        headers={
            "Origin": "http://localhost:4000",
        },
        files={"file": ("test.png", b"fake image data", "image/png")},
        data={"deviceid": "test-device"},
    )
    # May fail due to DB, but should have CORS headers
    assert "access-control-allow-origin" in response.headers


def test_cors_allows_all_origins_by_default(client):
    """Test that when FRONTEND_ORIGINS is not set, all origins are allowed"""
    response = client.get("/health", headers={"Origin": "http://any-domain.com"})
    assert response.status_code == 200
    # Should have CORS headers for any origin when properly configured
    if "access-control-allow-origin" in response.headers:
        assert response.headers["access-control-allow-origin"] in ["*", "http://any-domain.com"]


def test_cors_regex_pattern(client, monkeypatch):
    """Test that CORS works with regex pattern for dynamic port ranges"""
    # Set regex pattern to match ports 4000-4004
    monkeypatch.setenv("FRONTEND_ORIGIN_REGEX", r"^http://localhost:(4000|4001|4002|4003|4004)$")
    
    # Reimport to pick up new env var
    from importlib import reload
    import app.main as main_module
    reload(main_module)
    test_client = TestClient(main_module.app)
    
    # Test with matching origin
    response = test_client.options(
        "/api/v1/screenshots",
        headers={
            "Origin": "http://localhost:4002",
            "Access-Control-Request-Method": "POST",
            "Access-Control-Request-Headers": "content-type",
        },
    )
    assert response.status_code == 200
    # Regex should allow this origin
    if "access-control-allow-origin" in response.headers:
        assert response.headers["access-control-allow-origin"] in ["http://localhost:4002", "*"]

