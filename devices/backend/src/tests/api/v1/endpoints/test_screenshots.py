import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app
import io


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
