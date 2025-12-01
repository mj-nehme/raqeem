import io
from unittest.mock import AsyncMock, MagicMock, patch
from uuid import uuid4

import pytest
from app.db.session import get_db
from app.main import app
from app.models.devices import DeviceScreenshot
from httpx import ASGITransport, AsyncClient


@pytest.fixture
def mock_minio():
    """Fixture to mock MinIO service."""
    with patch('app.api.v1.endpoints.screenshots.MinioService') as mock_minio_class:
        mock_instance = MagicMock()
        mock_instance.upload_file.return_value = "test-file-id.png"
        mock_minio_class.return_value = mock_instance
        yield mock_instance


@pytest.fixture
def mock_db_session():
    """Fixture to mock database session for screenshot tests."""
    mock_session = AsyncMock()
    mock_session.add = MagicMock()
    mock_session.commit = AsyncMock()

    # Mock refresh to set screenshotid on the model
    async def mock_refresh(obj):
        if isinstance(obj, DeviceScreenshot):
            obj.screenshotid = uuid4()
    mock_session.refresh = mock_refresh

    return mock_session


@pytest.mark.asyncio
async def test_create_screenshot_file_upload(mock_minio, mock_db_session):
    """Test uploading screenshot file."""
    # Create a fake image file
    fake_image = io.BytesIO(b"fake image content")
    fake_image.name = "test.png"

    # Use a valid UUID for device_id
    device_id = "a1b2c3d4-e5f6-4a5b-8c7d-9e0f1a2b3c4d"

    # Override the database dependency
    async def override_get_db():
        yield mock_db_session

    app.dependency_overrides[get_db] = override_get_db

    try:
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
    finally:
        # Clean up dependency override
        app.dependency_overrides.clear()


@pytest.mark.asyncio
async def test_create_screenshot_file_upload_jpg(mock_minio, mock_db_session):
    """Test uploading JPG screenshot file."""
    fake_image = io.BytesIO(b"fake jpg content")
    fake_image.name = "test.jpg"

    # Use a valid UUID for device_id
    device_id = "b2c3d4e5-f6a7-4b5c-8d7e-9f0a1b2c3d4e"

    # Override the database dependency
    async def override_get_db():
        yield mock_db_session

    app.dependency_overrides[get_db] = override_get_db

    try:
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post(
                "/api/v1/screenshots/",
                data={"device_id": device_id},
                files={"file": ("screenshot.jpg", fake_image, "image/jpeg")},
            )
        assert response.status_code == 201

        # Verify MinIO upload was called
        mock_minio.upload_file.assert_called_once()
    finally:
        # Clean up dependency override
        app.dependency_overrides.clear()
