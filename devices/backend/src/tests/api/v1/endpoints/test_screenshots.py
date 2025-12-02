import io
from contextlib import contextmanager
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


@contextmanager
def override_db_dependency(mock_session):
    """Context manager to override database dependency and ensure cleanup."""
    async def override_get_db():
        yield mock_session

    app.dependency_overrides[get_db] = override_get_db
    try:
        yield
    finally:
        app.dependency_overrides.clear()


@pytest.mark.asyncio
async def test_create_screenshot_file_upload(mock_minio, mock_db_session):
    """Test uploading screenshot file."""
    # Create a fake image file
    fake_image = io.BytesIO(b"fake image content")
    fake_image.name = "test.png"

    # Use a valid UUID for device_id
    device_id = "a1b2c3d4-e5f6-4a5b-8c7d-9e0f1a2b3c4d"

    with override_db_dependency(mock_db_session):
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
async def test_create_screenshot_file_upload_jpg(mock_minio, mock_db_session):
    """Test uploading JPG screenshot file."""
    fake_image = io.BytesIO(b"fake jpg content")
    fake_image.name = "test.jpg"

    # Use a valid UUID for device_id
    device_id = "b2c3d4e5-f6a7-4b5c-8d7e-9f0a1b2c3d4e"

    with override_db_dependency(mock_db_session):
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post(
                "/api/v1/screenshots/",
                data={"device_id": device_id},
                files={"file": ("screenshot.jpg", fake_image, "image/jpeg")},
            )
    assert response.status_code == 201

    # Verify MinIO upload was called
    mock_minio.upload_file.assert_called_once()


@pytest.mark.asyncio
async def test_create_screenshot_missing_device_id(mock_minio, mock_db_session):
    """Test error when device identifier missing (device_id/deviceid)."""
    fake_image = io.BytesIO(b"img")
    fake_image.name = "missing.png"
    with override_db_dependency(mock_db_session):
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post(
                "/api/v1/screenshots/",
                data={},
                files={"file": ("screenshot.png", fake_image, "image/png")},
            )
    assert response.status_code == 422
    assert "device_id is required" in response.text


@pytest.mark.asyncio
async def test_create_screenshot_minio_failure(mock_db_session):
    """Test that MinIO failure is logged but endpoint still succeeds."""
    fake_image = io.BytesIO(b"fake content")
    fake_image.name = "test.png"
    device_id = str(uuid4())

    # Patch MinioService to raise exception on upload
    with patch('app.api.v1.endpoints.screenshots.MinioService') as mock_minio_class:
        mock_instance = MagicMock()
        def _raise(*args, **kwargs):
            raise RuntimeError("minio down")
        mock_instance.upload_file.side_effect = _raise
        mock_minio_class.return_value = mock_instance

        with override_db_dependency(mock_db_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(
                    "/api/v1/screenshots/",
                    data={"device_id": device_id},
                    files={"file": ("screenshot.png", fake_image, "image/png")},
                )
    assert response.status_code == 201
    data = response.json()
    assert data["status"] == "success"


@pytest.mark.asyncio
async def test_create_screenshot_forwarding_called(mock_db_session):
    """Test that forwarding logic calls post_with_retry when mentor_api_url set."""
    fake_image = io.BytesIO(b"forward content")
    fake_image.name = "forward.png"
    device_id = str(uuid4())

    # Patch settings and post_with_retry
    with patch('app.api.v1.endpoints.screenshots.settings') as mock_settings, patch('app.api.v1.endpoints.screenshots.post_with_retry') as mock_post:
        mock_settings.mentor_api_url = "http://mentor-backend"  # trigger forwarding
        mock_post.return_value = AsyncMock()
        with override_db_dependency(mock_db_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(
                    "/api/v1/screenshots/",
                    data={"device_id": device_id},
                    files={"file": ("screenshot.png", fake_image, "image/png")},
                )
    assert response.status_code == 201
    mock_post.assert_awaited()


@pytest.mark.asyncio
async def test_create_screenshot_tempfile_cleanup_error(mock_db_session):
    """Test that a cleanup unlink failure is caught and does not affect success response."""
    fake_image = io.BytesIO(b"cleanup content")
    fake_image.name = "cleanup.png"
    device_id = str(uuid4())

    # Simulate Path.unlink raising exception
    with patch('app.api.v1.endpoints.screenshots.Path.unlink', side_effect=RuntimeError("unlink fail")):
        with override_db_dependency(mock_db_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(
                    "/api/v1/screenshots/",
                    data={"device_id": device_id},
                    files={"file": ("screenshot.png", fake_image, "image/png")},
                )
    assert response.status_code == 201
    assert response.json()["status"] == "success"


@pytest.mark.asyncio
async def test_create_screenshot_with_deviceid_field(mock_db_session):
    """Test using legacy 'deviceid' form field name instead of device_id."""
    fake_image = io.BytesIO(b"legacy deviceid content")
    fake_image.name = "legacy.png"
    device_id = str(uuid4())
    with override_db_dependency(mock_db_session):
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post(
                "/api/v1/screenshots/",
                data={"deviceid": device_id},
                files={"file": ("screenshot.png", fake_image, "image/png")},
            )
    assert response.status_code == 201
    assert response.json()["status"] == "success"


@pytest.mark.asyncio
async def test_create_screenshot_database_failure(mock_db_session):
    """Test that a database failure triggers 500 error path."""
    fake_image = io.BytesIO(b"db fail content")
    fake_image.name = "dbfail.png"
    device_id = str(uuid4())
    # Force add() to raise to enter exception block
    mock_db_session.add.side_effect = RuntimeError("db add failed")
    with patch('app.api.v1.endpoints.screenshots.MinioService') as mock_minio_class:
        mock_minio_class.return_value = MagicMock()  # MinIO succeeds
        with override_db_dependency(mock_db_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(
                    "/api/v1/screenshots/",
                    data={"device_id": device_id},
                    files={"file": ("screenshot.png", fake_image, "image/png")},
                )
    assert response.status_code == 500
    assert "Screenshot upload failed" in response.text
