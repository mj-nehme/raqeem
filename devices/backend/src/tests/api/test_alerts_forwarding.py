import os
import pytest
from httpx import AsyncClient, ASGITransport
import respx
import httpx
from fastapi import status

# Set minimal required env BEFORE importing app
os.environ.setdefault("SECRET_KEY", "test-secret")
os.environ.setdefault("ACCESS_TOKEN_EXPIRE_MINUTES", "10080")
os.environ.setdefault("MINIO_ENDPOINT", "http://localhost:9000")
os.environ.setdefault("MINIO_ACCESS_KEY", "minio")
os.environ.setdefault("MINIO_SECRET_KEY", "miniosecret")
os.environ.setdefault("MINIO_SECURE", "false")
# DATABASE_URL and MENTOR_API_URL should be set by CI or developer env

from app.main import app  # noqa: E402
from app.db.init_db import init_db  # noqa: E402

pytestmark = pytest.mark.asyncio


async def _ensure_db():
    # Create tables if not exist
    await init_db()


@respx.mock
async def test_post_alerts_is_saved_and_forwarded():
    device_id = "12345678-1234-1234-1234-123456789abc"

    # Mock mentor forwarding endpoint if configured
    mentor_api = os.getenv("MENTOR_API_URL", "http://localhost:8080").rstrip("/")
    route = respx.post(f"{mentor_api}/devices/{device_id}/alerts").mock(
        return_value=httpx.Response(200, json={"ok": True})
    )

    await _ensure_db()

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        payload = [
            {
                "level": "warning",
                "alert_type": "cpu_high",
                "message": "CPU usage high",
                "value": 95,
                "threshold": 80,
            }
        ]
        resp = await client.post(f"/api/v1/devices/{device_id}/alerts", json=payload)
        assert resp.status_code == status.HTTP_200_OK
        body = resp.json()
        assert body.get("inserted") == 1

    # Ensure mentor forwarding was attempted exactly once
    assert route.called
    assert route.call_count == 1
