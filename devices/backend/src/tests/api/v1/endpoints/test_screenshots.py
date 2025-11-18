import io
from unittest.mock import MagicMock, patch

import pytest
from app.main import app
from httpx import ASGITransport, AsyncClient


@pytest.fixture
def mock_minio():
    """Fixture to mock MinIO service."""
    with patch('app.api.v1.endpoints.screenshots.MinioService') as mock_minio_class:
        mock_instance = MagicMock()
        mock_instance.upload_file.return_value = "test-file-id.png"
        mock_minio_class.return_value = mock_instance
        yield mock_instance


@pytest.mark.asyncio
async def test_create_screenshot_file_upload(mock_minio):
    """Test uploading screenshot file."""
    # Create a fake image file
    fake_image = io.BytesIO(b"fake image content")
    fake_image.name = "test.png"

    # Use a valid UUID for device_id
    device_id = "a1b2c3d4-e5f6-4a5b-8c7d-9e0f1a2b3c4d"

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(
            "/api/v1/screenshots/",
            data={"device_id": device_id},
            files={"file": ("screenshot.png", fake_image, "image/png")},
        )
    assert response.status_code == 201
    data = response.json()
    assert data["status"] == "success"
    assert "id" in data
    assert "image_url" in data

    # Verify MinIO upload was called
    mock_minio.upload_file.assert_called_once()


@pytest.mark.asyncio
async def test_create_screenshot_file_upload_jpg(mock_minio):
    """Test uploading JPG screenshot file."""
    fake_image = io.BytesIO(b"fake jpg content")
    fake_image.name = "test.jpg"

    # Use a valid UUID for device_id
    device_id = "b2c3d4e5-f6a7-4b5c-8d7e-9f0a1b2c3d4e"

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(
            "/api/v1/screenshots/",
            data={"device_id": device_id},
            files={"file": ("screenshot.jpg", fake_image, "image/jpeg")},
        )
    assert response.status_code == 201

    # Verify MinIO upload was called
    mock_minio.upload_file.assert_called_once()
