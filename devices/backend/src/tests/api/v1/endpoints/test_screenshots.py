import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app
import io


@pytest.mark.asyncio
async def test_create_screenshot_json():
    """Test creating screenshot via JSON endpoint."""
    payload = {
        "user_id": "some-valid-uuid",
        "image_url": "https://example.com/screenshot.png",
        "timestamp": "2025-06-25T10:00:00Z"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/screenshots/json", json=payload)
    assert response.status_code == 201
    data = response.json()
    assert data["image_url"] == payload["image_url"]
    assert "id" in data


@pytest.mark.asyncio
async def test_create_screenshot_json_minimal():
    """Test creating screenshot with minimal fields."""
    payload = {
        "user_id": "user-123",
        "image_url": "screenshot.png"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/screenshots/json", json=payload)
    assert response.status_code == 201
    data = response.json()
    assert data["user_id"] == "user-123"
    assert "id" in data


@pytest.mark.asyncio
async def test_create_screenshot_file_upload():
    """Test uploading screenshot file."""
    # Create a fake image file
    fake_image = io.BytesIO(b"fake image content")
    fake_image.name = "test.png"
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(
            "/api/v1/screenshots/",
            data={"device_id": "device-upload-001"},
            files={"file": ("screenshot.png", fake_image, "image/png")}
        )
    assert response.status_code == 201
    data = response.json()
    assert data["status"] == "success"
    assert "id" in data
    assert "image_url" in data


@pytest.mark.asyncio
async def test_create_screenshot_file_upload_jpg():
    """Test uploading JPG screenshot file."""
    fake_image = io.BytesIO(b"fake jpg content")
    fake_image.name = "test.jpg"
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(
            "/api/v1/screenshots/",
            data={"device_id": "device-upload-002"},
            files={"file": ("screenshot.jpg", fake_image, "image/jpeg")}
        )
    assert response.status_code == 201


@pytest.mark.asyncio
async def test_get_screenshots_list():
    """Test getting list of screenshots."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/screenshots/")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)


@pytest.mark.asyncio
async def test_get_screenshots_list_with_data():
    """Test getting screenshots list after creating one."""
    # First create a screenshot
    payload = {
        "user_id": "user-list-test",
        "image_url": "test-screenshot.png"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        await ac.post("/api/v1/screenshots/json", json=payload)
        
        # Now get the list
        response = await ac.get("/api/v1/screenshots/")
    
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    # Check that our screenshot is in the list
    user_ids = [s["user_id"] for s in data]
    assert "user-list-test" in user_ids
