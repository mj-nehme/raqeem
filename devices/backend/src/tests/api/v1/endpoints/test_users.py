# tests/api/v1/endpoints/test_users.py

import pytest
import pytest_asyncio
from httpx import AsyncClient, ASGITransport
from app.main import app

@pytest_asyncio.fixture
async def async_client():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://testserver") as client:
        yield client

@pytest.mark.asyncio
async def test_create_user(async_client):
    payload = {
        "deviceid": "test-device-001",
        "name": "Test User"
    }
    response = await async_client.post("/api/v1/users/", json=payload)
    assert response.status_code == 201
    data = response.json()
    assert "id" in data
    assert data["deviceid"] == payload["deviceid"]
    assert data["name"] == payload["name"]

@pytest.mark.asyncio
async def test_get_user_list(async_client):
    # Create user first so list is not empty
    payload = {
        "deviceid": "test-device-002",
        "name": "Test User 2"
    }
    create_resp = await async_client.post("/api/v1/users/", json=payload)
    assert create_resp.status_code == 201

    # Now get the user list
    response = await async_client.get("/api/v1/users/")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    # Check the user we created is in the list by device_id
    assert any(user["deviceid"] == payload["deviceid"] for user in data)
