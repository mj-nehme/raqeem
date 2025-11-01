import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app


@pytest.mark.asyncio
async def test_create_location():
    payload = {
        "user_id": "some-valid-uuid",  # ideally, get or create a user in fixture or test setup
        "latitude": 51.5074,
        "longitude": -0.1278
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/locations/", json=payload)
    # Adjust status based on your logic, probably 201 if created successfully
    assert response.status_code == 201
    data = response.json()
    assert data["latitude"] == payload["latitude"]
    assert data["longitude"] == payload["longitude"]
    assert "id" in data


@pytest.mark.asyncio
async def test_get_locations_list():
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/locations/")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
