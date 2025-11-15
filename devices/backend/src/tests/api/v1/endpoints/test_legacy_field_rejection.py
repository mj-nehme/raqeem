"""Tests for legacy field rejection with clear error messages."""
import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app


@pytest.mark.asyncio
async def test_register_device_rejects_legacy_id():
    """Test that device registration rejects legacy 'id' field."""
    payload = {
        "id": "test-device-id",
        "device_type": "laptop"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 400
    assert "unsupported legacy field: id" in response.json()["detail"]
    assert "use deviceid" in response.json()["detail"]


@pytest.mark.asyncio
async def test_register_device_rejects_legacy_name():
    """Test that device registration rejects legacy 'name' field."""
    payload = {
        "deviceid": "test-device-id",
        "name": "Test Device"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 400
    assert "unsupported legacy field: name" in response.json()["detail"]
    assert "use device_name" in response.json()["detail"]


@pytest.mark.asyncio
async def test_register_device_rejects_legacy_location():
    """Test that device registration rejects legacy 'location' field."""
    payload = {
        "deviceid": "test-device-id",
        "location": "Office"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 400
    assert "unsupported legacy field: location" in response.json()["detail"]
    assert "use device_location" in response.json()["detail"]


@pytest.mark.asyncio
async def test_processes_rejects_legacy_name():
    """Test that process submission rejects legacy 'name' field."""
    processes = [{
        "pid": 1234,
        "name": "chrome",
        "cpu": 10.0
    }]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/test-device/processes", json=processes)
    assert response.status_code == 400
    assert "unsupported legacy field: name" in response.json()["detail"]
    assert "use process_name" in response.json()["detail"]


@pytest.mark.asyncio
async def test_processes_rejects_legacy_command():
    """Test that process submission rejects legacy 'command' field."""
    processes = [{
        "pid": 1234,
        "process_name": "chrome",
        "command": "/usr/bin/chrome"
    }]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/test-device/processes", json=processes)
    assert response.status_code == 400
    assert "unsupported legacy field: command" in response.json()["detail"]
    assert "use command_text" in response.json()["detail"]


@pytest.mark.asyncio
async def test_activities_rejects_legacy_type():
    """Test that activity submission rejects legacy 'type' field."""
    activities = [{
        "type": "app_launch",
        "description": "Launched app"
    }]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/test-device/activities", json=activities)
    assert response.status_code == 400
    assert "unsupported legacy field: type" in response.json()["detail"]
    assert "use activity_type" in response.json()["detail"]


@pytest.mark.asyncio
async def test_alerts_rejects_legacy_type():
    """Test that alert submission rejects legacy 'type' field."""
    alerts = [{
        "level": "warning",
        "type": "cpu_high",
        "message": "CPU usage high"
    }]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/test-device/alerts", json=alerts)
    assert response.status_code == 400
    assert "unsupported legacy field: type" in response.json()["detail"]
    assert "use alert_type" in response.json()["detail"]
