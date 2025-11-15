import io
import os

import httpx
import pytest
import respx
from fastapi import status
from httpx import ASGITransport, AsyncClient

# Set minimal required env BEFORE importing app
os.environ.setdefault("SECRET_KEY", "test-secret")
os.environ.setdefault("ACCESS_TOKEN_EXPIRE_MINUTES", "10080")
os.environ.setdefault("MINIO_ENDPOINT", "http://localhost:9000")
os.environ.setdefault("MINIO_ACCESS_KEY", "minio")
os.environ.setdefault("MINIO_SECRET_KEY", "miniosecret")
os.environ.setdefault("MINIO_SECURE", "false")
# DATABASE_URL and MENTOR_API_URL should be set by CI or developer env

from app.db.init_db import init_db
from app.main import app

pytestmark = pytest.mark.asyncio


async def _ensure_db():
    # Create tables if not exist
    await init_db()


@respx.mock
async def test_post_screenshot_is_saved_and_forwarded():
    """Test that screenshots are saved and forwarded to mentor backend."""
    device_id = "a9d2f637-8409-57be-9246-ae620165a769"

    # Mock mentor forwarding endpoint if configured
    mentor_api = os.getenv("MENTOR_API_URL", "http://localhost:8080").rstrip("/")
    route = respx.post(f"{mentor_api}/devices/screenshots").mock(return_value=httpx.Response(200, json={"ok": True}))

    await _ensure_db()

    # Create a fake image file
    fake_image = io.BytesIO(b"fake image content for testing")
    fake_image.name = "test.png"

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.post(
            "/api/v1/screenshots/",
            data={"deviceid": device_id},
            files={"file": ("screenshot.png", fake_image, "image/png")},
        )
        assert response.status_code == status.HTTP_201_CREATED
        body = response.json()
        assert body.get("status") == "success"
        assert "id" in body

    # Ensure mentor forwarding was attempted
    assert route.called
    assert route.call_count == 1

    # Verify the forwarded payload contains correct fields
    forwarded_request = route.calls[0].request
    forwarded_data = forwarded_request.json()
    assert forwarded_data["deviceid"] == device_id
    assert "path" in forwarded_data
    assert "size" in forwarded_data
    assert "resolution" in forwarded_data
