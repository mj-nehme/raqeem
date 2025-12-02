from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from app.main import app
from httpx import ASGITransport, AsyncClient
from app.db.session import get_db


def _make_mock_session(device_exists: bool = False):
    """Create a lightweight AsyncSession mock used to avoid real DB connections.

    device_exists=False -> scalars().first() returns None (new device path)
    device_exists=True  -> returns a MagicMock representing an existing device
    """
    session = AsyncMock()
    session.add = MagicMock()
    session.commit = AsyncMock()
    session.refresh = AsyncMock()

    mock_result = MagicMock()
    if device_exists:
        existing = MagicMock()
        existing.device_name = "Existing"
        mock_result.scalars.return_value.first.return_value = existing
    else:
        mock_result.scalars.return_value.first.return_value = None
    # For queries that use .all() (pending commands etc.) let it return empty list
    mock_result.scalars.return_value.all.return_value = []
    session.execute = AsyncMock(return_value=mock_result)
    return session


@pytest.fixture
def override_db_new_device():
    """Override get_db to always yield a mock session (device does not yet exist)."""
    mock_session = _make_mock_session(device_exists=False)

    async def _override():
        yield mock_session

    app.dependency_overrides[get_db] = _override
    yield mock_session
    app.dependency_overrides.clear()


@pytest.fixture
def override_db_existing_device():
    """Override get_db to simulate existing device record (update path)."""
    mock_session = _make_mock_session(device_exists=True)

    async def _override():
        yield mock_session

    app.dependency_overrides[get_db] = _override
    yield mock_session
    app.dependency_overrides.clear()


@pytest.mark.asyncio
async def test_register_device_forwards_to_mentor(override_db_new_device):
    """Test that device registration is forwarded to mentor backend when configured."""
    payload = {
        "deviceid": "a843a399-701f-5011-aff3-4b69d8f21b11",
        "device_name": "Test Device for Forwarding",
        "device_type": "laptop",
        "os": "Linux",
        "device_location": "Test Lab",
        "ip_address": "192.168.1.200",
        "mac_address": "11:22:33:44:55:66",
        "current_user": "testuser",
    }

    # Patch settings so forwarding branch triggers and mock httpx client call
    with patch("app.api.v1.endpoints.devices.settings") as mock_settings, patch("httpx.AsyncClient") as mock_client:
        mock_settings.mentor_api_url = "http://mentor:8080"
        mock_response = AsyncMock()
        mock_response.status_code = 200
        mock_client.return_value.__aenter__.return_value.post = AsyncMock(return_value=mock_response)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post("/api/v1/devices/register", json=payload)

        assert response.status_code == 200
        data = response.json()
        assert data["deviceid"] == "a843a399-701f-5011-aff3-4b69d8f21b11"
        # Verify forwarding invoked
        mock_client.return_value.__aenter__.return_value.post.assert_called()


@pytest.mark.asyncio
async def test_register_device_updates_and_forwards(override_db_existing_device):
    """Exercise the 'existing device' update path plus forwarding."""
    payload = {
        "deviceid": "6f2a9c2c-0c5d-5e25-9acf-2e6b9e7f3321",
        "device_name": "Updated Device",
        "device_type": "desktop",
    }
    with patch("app.api.v1.endpoints.devices.settings") as mock_settings, patch("httpx.AsyncClient") as mock_client:
        mock_settings.mentor_api_url = "http://mentor:8080"
        mock_client.return_value.__aenter__.return_value.post = AsyncMock(return_value=AsyncMock(status_code=200))
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["updated"] is True
    mock_client.return_value.__aenter__.return_value.post.assert_called()


@pytest.mark.asyncio
async def test_register_device_survives_mentor_forwarding_failure(override_db_new_device):
    """Test that device registration succeeds even if mentor forwarding fails."""
    payload = {
        "deviceid": "e35e27a7-5808-5ea8-9ac5-acc284f75552",
        "device_name": "Test Device",
        "device_type": "laptop",
    }

    # Mock forwarding failure while mentor URL configured
    with patch("app.api.v1.endpoints.devices.settings") as mock_settings, patch("httpx.AsyncClient") as mock_client:
        mock_settings.mentor_api_url = "http://mentor:8080"
        # Simulate network error during forwarding
        mock_client.return_value.__aenter__.return_value.post = AsyncMock(side_effect=Exception("Network error"))

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post("/api/v1/devices/register", json=payload)

        # Registration should still succeed despite forwarding failure
        assert response.status_code == 200
        data = response.json()
        assert data["deviceid"] == "e35e27a7-5808-5ea8-9ac5-acc284f75552"


@pytest.mark.asyncio
async def test_metrics_forwarding_to_mentor(override_db_new_device):
    """Test that metrics are forwarded to mentor backend when configured."""
    device_id = "33f9ce74-d0ce-515e-bb95-2464e9faa707"
    payload = {
        "cpu_usage": 55.5,
        "cpu_temp": 70.0,
        "memory_total": 16000000000,
        "memory_used": 10000000000,
        "swap_used": 200000000,
        "disk_total": 500000000000,
        "disk_used": 300000000000,
        "net_bytes_in": 2048000,
        "net_bytes_out": 4096000,
    }

    with patch("app.api.v1.endpoints.devices.settings") as mock_settings, patch("httpx.AsyncClient") as mock_client:
        mock_settings.mentor_api_url = "http://mentor:8080"
        mock_response = AsyncMock()
        mock_response.status_code = 200
        mock_client.return_value.__aenter__.return_value.post = AsyncMock(return_value=mock_response)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)

        assert response.status_code == 200
        assert response.json()["status"] == "ok"
        mock_client.return_value.__aenter__.return_value.post.assert_called()
