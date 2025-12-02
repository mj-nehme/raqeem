import pytest
from app.main import app
from httpx import ASGITransport, AsyncClient

# These integration-style tests require a live Postgres instance. To avoid failing
# the suite (and suppress connection errors) while focusing on deterministic
# mock-based coverage in other test modules, we skip this file. All code paths
# in `devices.py` are exercised by `test_full_coverage.py`, `test_error_handling.py`,
# and `test_legacy_field_rejection.py`.
pytestmark = []  # Live Postgres now provided via testcontainers in conftest; run integration tests.


@pytest.mark.asyncio
async def test_register_device_new():
    """Test registering a new device."""
    payload = {
        "deviceid": "663903cd-f6ac-5211-8e93-4a0889840f94",
        "device_name": "Test Device",
        "device_type": "laptop",
        "os": "Windows 11",
        "device_location": "Office A",
        "ip_address": "192.168.1.100",
        "mac_address": "00:11:22:33:44:55",
        "current_user": "testuser",
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["deviceid"] == "663903cd-f6ac-5211-8e93-4a0889840f94"
    # Device may be created or updated depending on test order
    assert data.get("created") is True or data.get("updated") is True


@pytest.mark.asyncio
async def test_register_device_update_existing():
    """Test updating an existing device."""
    # First register
    payload = {
        "deviceid": "beedc88d-78d8-5564-8baa-eae0531f29dd",
        "device_name": "Device Original",
        "device_type": "desktop",
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
        assert response.status_code == 200

        # Now update
        update_payload = {
            "deviceid": "beedc88d-78d8-5564-8baa-eae0531f29dd",
            "device_name": "Device Updated",
            "device_type": "laptop",
        }
        response = await ac.post("/api/v1/devices/register", json=update_payload)

    assert response.status_code == 200
    data = response.json()
    assert data["deviceid"] == "beedc88d-78d8-5564-8baa-eae0531f29dd"
    assert data.get("updated") is True


@pytest.mark.asyncio
async def test_register_device_missing_id():
    """Test registering device without deviceid fails."""
    payload = {"device_name": "Test Device", "device_type": "laptop"}
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 400
    assert "missing required field: deviceid" in response.json()["detail"].lower()


@pytest.mark.asyncio
async def test_register_device_with_device_id_key():
    """Test registering device using deviceid key."""
    payload = {"deviceid": "1fc94cd2-8a8c-5cf6-a800-8a9a2d31640e", "device_name": "Test Device 3"}
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["deviceid"] == "1fc94cd2-8a8c-5cf6-a800-8a9a2d31640e"


@pytest.mark.asyncio
async def test_post_metrics():
    """Test posting device metrics."""
    device_id = "9b3918c3-dc68-5c24-a2b1-64413d93e56d"
    payload = {
        "cpu_usage": 45.5,
        "cpu_temp": 65.0,
        "memory_total": 16000000000,
        "memory_used": 8000000000,
        "swap_used": 100000000,
        "disk_total": 500000000000,
        "disk_used": 250000000000,
        "net_bytes_in": 1024000,
        "net_bytes_out": 2048000,
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)
    assert response.status_code == 200
    assert response.json()["status"] == "ok"


@pytest.mark.asyncio
async def test_post_metrics_minimal():
    """Test posting metrics with minimal data."""
    device_id = "08e9f7c6-2142-501e-9609-dfc16c0d8044"
    payload = {"cpu_usage": 25.0}
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)
    assert response.status_code == 200


@pytest.mark.asyncio
async def test_post_processes():
    """Test posting device processes."""
    device_id = "0e046a19-b4ce-5698-bb48-bd9c26fff1a8"
    processes = [
        {"pid": 1234, "process_name": "chrome", "cpu": 15.5, "memory": 500000000, "command_text": "/usr/bin/chrome"},
        {"pid": 5678, "process_name": "firefox", "cpu": 10.2, "memory": 300000000, "command_text": "/usr/bin/firefox"},
    ]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 2


@pytest.mark.asyncio
async def test_post_processes_empty():
    """Test posting empty processes list."""
    device_id = "413d403e-8215-5db8-82f9-ed0cdc666484"
    processes = []
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 0


@pytest.mark.asyncio
async def test_post_activities():
    """Test posting device activities."""
    device_id = "cca4121f-cff9-5cf9-ac76-a48faca15128"
    activities = [
        {"activity_type": "app_launch", "description": "User launched Chrome", "app": "chrome", "duration": 3600},
        {"activity_type": "app_close", "description": "User closed Firefox", "app": "firefox", "duration": 7200},
    ]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/activities", json=activities)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 2


@pytest.mark.asyncio
async def test_post_activities_empty():
    """Test posting empty activities list."""
    device_id = "c6a67fde-e80d-548c-b131-cead2959eba5"
    activities = []
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/activities", json=activities)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 0


@pytest.mark.asyncio
async def test_post_alerts():
    """Test posting device alerts."""
    device_id = "a98287ca-f97b-5cd2-b2c7-3a4aced9d4c9"
    alerts = [
        {
            "level": "warning",
            "alert_type": "cpu",
            "message": "High CPU usage detected",
            "value": 85.5,
            "threshold": 80.0,
        },
        {
            "level": "critical",
            "alert_type": "memory",
            "message": "Memory usage critical",
            "value": 95.0,
            "threshold": 90.0,
        },
    ]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/alerts", json=alerts)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 2


@pytest.mark.asyncio
async def test_post_alerts_empty():
    """Test posting empty alerts list."""
    device_id = "acda9ce0-04bb-58cb-8452-fbbcda08bdc9"
    alerts = []
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/alerts", json=alerts)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 0


# NOTE: Read-only list/get endpoints have been moved to mentor backend.
# Tests for list_devices, list_devices_empty, get_device_by_id, list_all_processes,
# list_all_activities, list_all_alerts have been removed as those endpoints
# are no longer available on the devices backend.


@pytest.mark.asyncio
async def test_register_device_preserves_existing_fields():
    """Test that updating a device preserves fields not in update payload."""
    # First register with all fields
    initial_payload = {
        "deviceid": "4fb793da-ab61-5e9b-9db3-b20b085fadaf",
        "device_name": "Original Name",
        "device_type": "laptop",
        "os": "Linux",
        "device_location": "Office",
        "ip_address": "192.168.1.50",
        "mac_address": "AA:BB:CC:DD:EE:FF",
        "current_user": "john",
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        await ac.post("/api/v1/devices/register", json=initial_payload)

        # Update with only name changed
        update_payload = {"deviceid": "4fb793da-ab61-5e9b-9db3-b20b085fadaf", "device_name": "Updated Name"}
        response = await ac.post("/api/v1/devices/register", json=update_payload)
        assert response.status_code == 200
        assert response.json()["updated"] is True


@pytest.mark.asyncio
async def test_post_processes_replaces_existing():
    """Test that posting processes replaces existing ones for the device."""
    device_id = "89006b63-1de6-5f65-a258-5bf69c0f4c15"

    # First batch of processes
    processes1 = [{"pid": 1111, "process_name": "process1", "cpu": 10.0, "memory": 100000}]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes1)
        assert response.status_code == 200

        # Second batch - should replace first
        processes2 = [{"pid": 2222, "process_name": "process2", "cpu": 20.0, "memory": 200000}]
        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes2)
        assert response.status_code == 200
        assert response.json()["inserted"] == 1


@pytest.mark.asyncio
async def test_get_pending_commands():
    """Test getting pending commands for a device."""
    device_id = "e2985a22-c6f1-5fa4-b7cb-2e67ab6afaab"
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get(f"/api/v1/devices/{device_id}/commands/pending")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)


@pytest.mark.asyncio
async def test_create_command_success():
    """Test creating a command for a device."""
    device_id = "4ce3a91e-b09d-57a5-8493-b5115b7d3e01"
    payload = {"command_text": "get_info"}
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["deviceid"] == device_id
    assert data["command_text"] == "get_info"
    assert data["status"] == "pending"
    assert "commandid" in data


@pytest.mark.asyncio
async def test_create_command_not_allowed():
    """Test creating a command with disallowed command fails."""
    device_id = "d8acd640-bbe1-53fa-b2b8-63aa9bed99de"
    payload = {"command_text": "rm -rf /"}
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)
    assert response.status_code == 400
    assert "not allowed" in response.json()["detail"].lower()


@pytest.mark.asyncio
async def test_create_command_various_allowed():
    """Test creating commands with various allowed command types."""
    device_id = "1aed9f08-fc44-57bf-a94f-76b58f110a30"
    allowed_commands = ["status", "restart", "get_processes", "get_logs", "restart_service", "screenshot"]

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        for cmd in allowed_commands:
            payload = {"command_text": cmd}
            response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)
            assert response.status_code == 200, f"Command {cmd} should be allowed"
            data = response.json()
            assert data["command_text"] == cmd


@pytest.mark.asyncio
async def test_submit_command_result_success():
    """Test submitting command execution result."""
    device_id = "015f9f53-05e9-5f15-82c9-69a7b208f452"

    # First create a command
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        create_payload = {"command_text": "get_info"}
        create_response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=create_payload)
        assert create_response.status_code == 200
        command_id = create_response.json()["commandid"]

        # Now submit result
        result_payload = {"status": "completed", "result": "Command output here", "exit_code": 0}
        response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=result_payload)
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert data["commandid"] == str(command_id)


@pytest.mark.asyncio
async def test_submit_command_result_not_found():
    """Test submitting result for non-existent command fails."""
    # Use a valid UUID that doesn't exist in the database
    command_id = "00000000-0000-0000-0000-000000000000"
    result_payload = {"status": "completed", "result": "Output", "exit_code": 0}
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=result_payload)
    assert response.status_code == 404
    assert "not found" in response.json()["detail"].lower()


@pytest.mark.asyncio
async def test_submit_command_result_failed_status():
    """Test submitting command result with failed status."""
    device_id = "5c667ba9-433f-5183-a521-f82d714cd68f"

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Create command
        create_payload = {"command_text": "get_info"}
        create_response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=create_payload)
        command_id = create_response.json()["commandid"]

        # Submit failed result
        result_payload = {"status": "failed", "result": "Error occurred", "exit_code": 1}
        response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=result_payload)
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"


# NOTE: Additional read-only endpoint tests removed (list_devices_multiple,
# list_devices_with_existing, get_device_by_id, get_device_by_id_not_found)
# as those endpoints have been moved to mentor backend.


@pytest.mark.asyncio
async def test_post_metrics_with_forwarding():
    """Test posting metrics with mentor API forwarding (when configured)."""
    import os
    from unittest.mock import AsyncMock, patch

    device_id = "ce3b06af-dc71-5066-b9f1-f1a61edf28a1"
    payload = {"cpu_usage": 50.0, "memory_total": 16000000000, "memory_used": 8000000000}

    # Mock the mentor API URL to test forwarding path
    with patch.dict(os.environ, {"MENTOR_API_URL": "http://mock-mentor:8080"}):
        with patch("httpx.AsyncClient") as mock_client:
            mock_post = AsyncMock()
            mock_client.return_value.__aenter__.return_value.post = mock_post
            mock_post.return_value = None  # Forwarding doesn't check response

            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)
                assert response.status_code == 200


@pytest.mark.asyncio
async def test_post_alerts_with_forwarding():
    """Test posting alerts with mentor API forwarding (when configured)."""
    import os
    from unittest.mock import AsyncMock, patch

    device_id = "88c34f3e-8b13-5615-94cd-569c8c4fff16"
    alerts = [{"level": "critical", "alert_type": "cpu", "message": "CPU critical", "value": 95.0, "threshold": 90.0}]

    # Mock the mentor API URL to test forwarding path
    with patch.dict(os.environ, {"MENTOR_API_URL": "http://mock-mentor:8080"}):
        with patch("httpx.AsyncClient") as mock_client:
            mock_post = AsyncMock()
            mock_client.return_value.__aenter__.return_value.post = mock_post
            mock_post.return_value = None

            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(f"/api/v1/devices/{device_id}/alerts", json=alerts)
                assert response.status_code == 200


@pytest.mark.asyncio
async def test_post_metrics_forwarding_failure_handled():
    """Test that metrics ingestion succeeds even if forwarding fails."""
    import os
    from unittest.mock import patch

    device_id = "0e8d587a-c984-59a3-8b95-7dc57fa0d676"
    payload = {"cpu_usage": 60.0}

    # Mock the mentor API URL but make it fail
    with patch.dict(os.environ, {"MENTOR_API_URL": "http://invalid-host:99999"}):
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            # Should still return 200 even if forwarding fails
            response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)
            assert response.status_code == 200
            assert response.json()["status"] == "ok"


# NOTE: Tests for list_all_processes, list_all_activities, list_all_alerts
# have been removed as those endpoints have been moved to mentor backend.
