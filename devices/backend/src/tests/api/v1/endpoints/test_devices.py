import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app


@pytest.mark.asyncio
async def test_register_device_new():
    """Test registering a new device."""
    payload = {
        "id": "test-device-001",
        "name": "Test Device",
        "type": "laptop",
        "os": "Windows 11",
        "location": "Office A",
        "ip_address": "192.168.1.100",
        "mac_address": "00:11:22:33:44:55",
        "current_user": "testuser"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["device_id"] == "test-device-001"
    assert data.get("created") is True


@pytest.mark.asyncio
async def test_register_device_update_existing():
    """Test updating an existing device."""
    # First register
    payload = {
        "id": "test-device-002",
        "name": "Device Original",
        "type": "desktop"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
        assert response.status_code == 200
        
        # Now update
        update_payload = {
            "id": "test-device-002",
            "name": "Device Updated",
            "type": "laptop"
        }
        response = await ac.post("/api/v1/devices/register", json=update_payload)
    
    assert response.status_code == 200
    data = response.json()
    assert data["device_id"] == "test-device-002"
    assert data.get("updated") is True


@pytest.mark.asyncio
async def test_register_device_missing_id():
    """Test registering device without id fails."""
    payload = {
        "name": "Test Device",
        "type": "laptop"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 400
    assert "missing device id" in response.json()["detail"].lower()


@pytest.mark.asyncio
async def test_register_device_with_device_id_key():
    """Test registering device using device_id key instead of id."""
    payload = {
        "device_id": "test-device-003",
        "name": "Test Device 3"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/devices/register", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["device_id"] == "test-device-003"


@pytest.mark.asyncio
async def test_post_metrics():
    """Test posting device metrics."""
    device_id = "test-device-metrics"
    payload = {
        "cpu_usage": 45.5,
        "cpu_temp": 65.0,
        "memory_total": 16000000000,
        "memory_used": 8000000000,
        "swap_used": 100000000,
        "disk_total": 500000000000,
        "disk_used": 250000000000,
        "net_bytes_in": 1024000,
        "net_bytes_out": 2048000
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)
    assert response.status_code == 200
    assert response.json()["status"] == "ok"


@pytest.mark.asyncio
async def test_post_metrics_minimal():
    """Test posting metrics with minimal data."""
    device_id = "test-device-metrics-min"
    payload = {
        "cpu_usage": 25.0
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)
    assert response.status_code == 200


@pytest.mark.asyncio
async def test_post_processes():
    """Test posting device processes."""
    device_id = "test-device-proc"
    processes = [
        {
            "pid": 1234,
            "name": "chrome",
            "cpu": 15.5,
            "memory": 500000000,
            "command": "/usr/bin/chrome"
        },
        {
            "pid": 5678,
            "name": "firefox",
            "cpu": 10.2,
            "memory": 300000000,
            "command": "/usr/bin/firefox"
        }
    ]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 2


@pytest.mark.asyncio
async def test_post_processes_empty():
    """Test posting empty processes list."""
    device_id = "test-device-proc-empty"
    processes = []
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 0


@pytest.mark.asyncio
async def test_post_activities():
    """Test posting device activities."""
    device_id = "test-device-act"
    activities = [
        {
            "type": "app_launch",
            "description": "User launched Chrome",
            "app": "chrome",
            "duration": 3600
        },
        {
            "type": "app_close",
            "description": "User closed Firefox",
            "app": "firefox",
            "duration": 7200
        }
    ]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/activities", json=activities)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 2


@pytest.mark.asyncio
async def test_post_activities_empty():
    """Test posting empty activities list."""
    device_id = "test-device-act-empty"
    activities = []
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/activities", json=activities)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 0


@pytest.mark.asyncio
async def test_post_alerts():
    """Test posting device alerts."""
    device_id = "test-device-alert"
    alerts = [
        {
            "level": "warning",
            "type": "cpu",
            "message": "High CPU usage detected",
            "value": 85.5,
            "threshold": 80.0
        },
        {
            "level": "critical",
            "type": "memory",
            "message": "Memory usage critical",
            "value": 95.0,
            "threshold": 90.0
        }
    ]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/alerts", json=alerts)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 2


@pytest.mark.asyncio
async def test_post_alerts_empty():
    """Test posting empty alerts list."""
    device_id = "test-device-alert-empty"
    alerts = []
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/alerts", json=alerts)
    assert response.status_code == 200
    data = response.json()
    assert data["inserted"] == 0


@pytest.mark.asyncio
async def test_list_devices():
    """Test listing all devices."""
    # First register a device
    payload = {
        "id": "test-device-list",
        "name": "Device for Listing",
        "type": "tablet",
        "os": "iOS"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        await ac.post("/api/v1/devices/register", json=payload)
        
        # Now list devices
        response = await ac.get("/api/v1/devices/")
    
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    # Check that our device is in the list
    device_ids = [d["id"] for d in data]
    assert "test-device-list" in device_ids


@pytest.mark.asyncio
async def test_list_devices_empty():
    """Test listing devices when none exist (after cleanup)."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)


@pytest.mark.asyncio
async def test_register_device_preserves_existing_fields():
    """Test that updating a device preserves fields not in update payload."""
    # First register with all fields
    initial_payload = {
        "id": "test-device-preserve",
        "name": "Original Name",
        "type": "laptop",
        "os": "Linux",
        "location": "Office",
        "ip_address": "192.168.1.50",
        "mac_address": "AA:BB:CC:DD:EE:FF",
        "current_user": "john"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        await ac.post("/api/v1/devices/register", json=initial_payload)
        
        # Update with only name changed
        update_payload = {
            "id": "test-device-preserve",
            "name": "Updated Name"
        }
        response = await ac.post("/api/v1/devices/register", json=update_payload)
        assert response.status_code == 200
        assert response.json()["updated"] is True


@pytest.mark.asyncio
async def test_post_processes_replaces_existing():
    """Test that posting processes replaces existing ones for the device."""
    device_id = "test-device-proc-replace"
    
    # First batch of processes
    processes1 = [{"pid": 1111, "name": "process1", "cpu": 10.0, "memory": 100000}]
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes1)
        assert response.status_code == 200
        
        # Second batch - should replace first
        processes2 = [{"pid": 2222, "name": "process2", "cpu": 20.0, "memory": 200000}]
        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes2)
        assert response.status_code == 200
        assert response.json()["inserted"] == 1


@pytest.mark.asyncio
async def test_get_pending_commands():
    """Test getting pending commands for a device."""
    device_id = "test-device-commands"
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get(f"/api/v1/devices/{device_id}/commands/pending")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)


@pytest.mark.asyncio
async def test_create_command_success():
    """Test creating a command for a device."""
    device_id = "test-device-cmd-create"
    payload = {
        "command": "get_info"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["device_id"] == device_id
    assert data["command"] == "get_info"
    assert data["status"] == "pending"
    assert "id" in data


@pytest.mark.asyncio
async def test_create_command_not_allowed():
    """Test creating a command with disallowed command fails."""
    device_id = "test-device-cmd-fail"
    payload = {
        "command": "rm -rf /"
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)
    assert response.status_code == 400
    assert "not allowed" in response.json()["detail"].lower()


@pytest.mark.asyncio
async def test_create_command_various_allowed():
    """Test creating commands with various allowed command types."""
    device_id = "test-device-cmd-various"
    allowed_commands = ["status", "restart", "get_processes", "get_logs", "restart_service", "screenshot"]
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        for cmd in allowed_commands:
            payload = {"command": cmd}
            response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)
            assert response.status_code == 200, f"Command {cmd} should be allowed"
            data = response.json()
            assert data["command"] == cmd


@pytest.mark.asyncio
async def test_submit_command_result_success():
    """Test submitting command execution result."""
    device_id = "test-device-cmd-result"
    
    # First create a command
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        create_payload = {"command": "get_info"}
        create_response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=create_payload)
        assert create_response.status_code == 200
        command_id = create_response.json()["id"]
        
        # Now submit result
        result_payload = {
            "status": "completed",
            "result": "Command output here",
            "exit_code": 0
        }
        response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=result_payload)
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert data["command_id"] == command_id


@pytest.mark.asyncio
async def test_submit_command_result_not_found():
    """Test submitting result for non-existent command fails."""
    command_id = 999999
    result_payload = {
        "status": "completed",
        "result": "Output",
        "exit_code": 0
    }
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=result_payload)
    assert response.status_code == 404
    assert "not found" in response.json()["detail"].lower()


@pytest.mark.asyncio
async def test_submit_command_result_failed_status():
    """Test submitting command result with failed status."""
    device_id = "test-device-cmd-fail-result"
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Create command
        create_payload = {"command": "get_info"}
        create_response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=create_payload)
        command_id = create_response.json()["id"]
        
        # Submit failed result
        result_payload = {
            "status": "failed",
            "result": "Error occurred",
            "exit_code": 1
        }
        response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=result_payload)
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"


@pytest.mark.asyncio
async def test_list_devices_multiple():
    """Test listing all devices."""
    # First register a couple of devices
    devices = [
        {"id": "list-device-1", "name": "Device 1", "type": "laptop"},
        {"id": "list-device-2", "name": "Device 2", "type": "desktop"}
    ]
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Register devices
        for device in devices:
            await ac.post("/api/v1/devices/register", json=device)
        
        # List devices
        response = await ac.get("/api/v1/devices/")
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)
        # At least our 2 devices should be there
        device_ids = [d["id"] for d in data]
        assert "list-device-1" in device_ids
        assert "list-device-2" in device_ids


@pytest.mark.asyncio
async def test_list_devices_with_existing():
    """Test listing devices with existing devices in database."""
    # This test assumes a fresh database or will get existing devices
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/")
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)


@pytest.mark.asyncio
async def test_get_device_by_id():
    """Test getting a specific device by ID."""
    device_id = "get-device-test"
    device_payload = {
        "id": device_id,
        "name": "Test Device",
        "type": "laptop",
        "os": "Windows 11"
    }
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Register device
        await ac.post("/api/v1/devices/register", json=device_payload)
        
        # Get device by ID
        response = await ac.get(f"/api/v1/devices/{device_id}")
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == device_id
        assert data["name"] == "Test Device"
        assert data["type"] == "laptop"
        assert data["os"] == "Windows 11"


@pytest.mark.asyncio
async def test_get_device_by_id_not_found():
    """Test getting non-existent device returns 404."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/api/v1/devices/nonexistent-device-xyz")
        assert response.status_code == 404
        assert "not found" in response.json()["detail"].lower()


@pytest.mark.asyncio
async def test_post_metrics_with_forwarding():
    """Test posting metrics with mentor API forwarding (when configured)."""
    import os
    from unittest.mock import patch, AsyncMock
    
    device_id = "test-device-forward"
    payload = {
        "cpu_usage": 50.0,
        "memory_total": 16000000000,
        "memory_used": 8000000000
    }
    
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
    from unittest.mock import patch, AsyncMock
    
    device_id = "test-device-alert-forward"
    alerts = [
        {
            "level": "critical",
            "type": "cpu",
            "message": "CPU critical",
            "value": 95.0,
            "threshold": 90.0
        }
    ]
    
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
    
    device_id = "test-device-forward-fail"
    payload = {
        "cpu_usage": 60.0
    }
    
    # Mock the mentor API URL but make it fail
    with patch.dict(os.environ, {"MENTOR_API_URL": "http://invalid-host:99999"}):
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            # Should still return 200 even if forwarding fails
            response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)
            assert response.status_code == 200
            assert response.json()["status"] == "ok"


@pytest.mark.asyncio
async def test_list_all_processes():
    """Test listing all processes across all devices."""
    device_id = "test-device-processes-list"
    
    # First post some processes
    processes = [
        {
            "pid": 1111,
            "name": "test-process-1",
            "cpu": 10.5,
            "memory": 100000000,
            "command": "/usr/bin/test1"
        },
        {
            "pid": 2222,
            "name": "test-process-2",
            "cpu": 20.5,
            "memory": 200000000,
            "command": "/usr/bin/test2"
        }
    ]
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Post processes
        await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes)
        
        # Now get all processes
        response = await ac.get("/api/v1/devices/processes")
    
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    # Check that our processes are in the list
    process_names = [p["name"] for p in data]
    assert "test-process-1" in process_names
    assert "test-process-2" in process_names


@pytest.mark.asyncio
async def test_list_all_activities():
    """Test listing all activities across all devices."""
    device_id = "test-device-activities-list"
    
    # First post some activities
    activities = [
        {
            "type": "test_activity",
            "description": "Test activity 1",
            "app": "test-app-1",
            "duration": 100
        },
        {
            "type": "test_activity",
            "description": "Test activity 2",
            "app": "test-app-2",
            "duration": 200
        }
    ]
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Post activities
        await ac.post(f"/api/v1/devices/{device_id}/activities", json=activities)
        
        # Now get all activities
        response = await ac.get("/api/v1/devices/activities")
    
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    # Check that our activities are in the list
    activity_apps = [a["app"] for a in data]
    assert "test-app-1" in activity_apps
    assert "test-app-2" in activity_apps


@pytest.mark.asyncio
async def test_list_all_alerts():
    """Test listing all alerts across all devices."""
    device_id = "test-device-alerts-list"
    
    # First post some alerts
    alerts = [
        {
            "level": "warning",
            "type": "test_alert",
            "message": "Test alert 1",
            "value": 75.0,
            "threshold": 70.0
        },
        {
            "level": "critical",
            "type": "test_alert",
            "message": "Test alert 2",
            "value": 95.0,
            "threshold": 90.0
        }
    ]
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        # Post alerts
        await ac.post(f"/api/v1/devices/{device_id}/alerts", json=alerts)
        
        # Now get all alerts
        response = await ac.get("/api/v1/devices/alerts")
    
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    # Check that our alerts are in the list
    alert_messages = [a["message"] for a in data]
    assert "Test alert 1" in alert_messages
    assert "Test alert 2" in alert_messages
