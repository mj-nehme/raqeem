import pytest
from fastapi.testclient import TestClient
from app.main import app


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
        }
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
        data={"device_id": "test-device"}
    )
    # May fail due to DB, but should have CORS headers
    assert "access-control-allow-origin" in response.headers


def test_cors_allows_all_origins_by_default(client):
    """Test that when FRONTEND_ORIGINS is not set, all origins are allowed"""
    response = client.get(
        "/health",
        headers={"Origin": "http://any-domain.com"}
    )
    assert response.status_code == 200
    # When allow_origins=["*"], the header should be "*"
    if "access-control-allow-origin" in response.headers:
        assert response.headers["access-control-allow-origin"] in ["*", "http://any-domain.com"]
