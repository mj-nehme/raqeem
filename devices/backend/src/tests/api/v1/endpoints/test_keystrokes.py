import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app


@pytest.mark.asyncio
async def test_create_keystroke():
    payload = {
        "user_id": "some-valid-uuid",
        "keylog": "a",
        "timestamp": "2025-06-25T10:00:00Z"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/keystrokes/", json=payload)
    assert response.status_code == 201
    data = response.json()
    assert data["keylog"] == payload["keylog"]
    assert "id" in data


@pytest.mark.asyncio
async def test_get_keystrokes_list():
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/keystrokes/")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
