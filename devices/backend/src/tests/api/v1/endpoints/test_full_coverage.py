"""
Comprehensive unit tests for 100% coverage of devices/backend/src/app/api/v1/endpoints.

These tests mock all external dependencies (database, MinIO, mentor API) to ensure
complete code path coverage without requiring external services.
"""

import io
from contextlib import contextmanager
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch
from uuid import UUID, uuid4

import pytest
from app.db.session import get_db
from app.main import app
from httpx import ASGITransport, AsyncClient


# ==============================================================================
# Helper fixtures and context managers
# ==============================================================================


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


def create_mock_db_session():
    """Create a mock database session with common async methods."""
    mock_session = AsyncMock()
    mock_session.add = MagicMock()
    mock_session.commit = AsyncMock()
    mock_session.refresh = AsyncMock()
    mock_session.execute = AsyncMock()
    return mock_session


# ==============================================================================
# Tests for devices.py - register_device
# ==============================================================================


class TestRegisterDeviceFullCoverage:
    """Tests for register_device endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_register_device_creates_new_device(self):
        """Test creating a new device when it doesn't exist."""
        mock_session = create_mock_db_session()

        # Mock: device doesn't exist (res.scalars().first() returns None)
        mock_result = MagicMock()
        mock_result.scalars.return_value.first.return_value = None
        mock_session.execute.return_value = mock_result

        device_id = str(uuid4())
        payload = {
            "deviceid": device_id,
            "device_name": "Test Device",
            "device_type": "laptop",
            "os": "Linux",
            "device_location": "Office",
            "ip_address": "192.168.1.1",
            "mac_address": "AA:BB:CC:DD:EE:FF",
            "current_user": "testuser",
        }

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = None  # No forwarding
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.post("/api/v1/devices/register", json=payload)

        assert response.status_code == 200
        data = response.json()
        assert data["deviceid"] == device_id
        assert data["created"] is True
        mock_session.add.assert_called()
        mock_session.commit.assert_called()

    @pytest.mark.asyncio
    async def test_register_device_updates_existing_device(self):
        """Test updating an existing device."""
        mock_session = create_mock_db_session()

        # Mock: device exists
        existing_device = MagicMock()
        existing_device.device_name = "Old Name"
        existing_device.device_type = "desktop"
        existing_device.os = "Windows"
        existing_device.device_location = "Lab"
        existing_device.ip_address = "10.0.0.1"
        existing_device.mac_address = "11:22:33:44:55:66"
        existing_device.current_user = "olduser"

        mock_result = MagicMock()
        mock_result.scalars.return_value.first.return_value = existing_device
        mock_session.execute.return_value = mock_result

        device_id = str(uuid4())
        payload = {
            "deviceid": device_id,
            "device_name": "Updated Name",
            "device_type": "laptop",
        }

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = None  # No forwarding
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.post("/api/v1/devices/register", json=payload)

        assert response.status_code == 200
        data = response.json()
        assert data["deviceid"] == device_id
        assert data["updated"] is True
        assert existing_device.device_name == "Updated Name"
        assert existing_device.device_type == "laptop"
        mock_session.commit.assert_called()

    @pytest.mark.asyncio
    async def test_register_device_with_mentor_forwarding(self):
        """Test device registration with mentor API forwarding."""
        mock_session = create_mock_db_session()

        # Mock: device doesn't exist
        mock_result = MagicMock()
        mock_result.scalars.return_value.first.return_value = None
        mock_session.execute.return_value = mock_result

        device_id = str(uuid4())
        payload = {
            "deviceid": device_id,
            "device_name": "Test Device",
        }

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = "http://mentor:8080"
                with patch("app.api.v1.endpoints.devices.post_with_retry") as mock_post:
                    mock_post.return_value = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post("/api/v1/devices/register", json=payload)

        assert response.status_code == 200
        mock_post.assert_called_once()
        # Verify forwarding URL
        call_args = mock_post.call_args
        assert "http://mentor:8080/devices/register" in call_args[0]


# ==============================================================================
# Tests for devices.py - post_metrics
# ==============================================================================


class TestPostMetricsFullCoverage:
    """Tests for post_metrics endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_post_metrics_success(self):
        """Test posting metrics successfully."""
        mock_session = create_mock_db_session()

        device_id = str(uuid4())
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

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = None  # No forwarding
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)

        assert response.status_code == 200
        assert response.json()["status"] == "ok"
        mock_session.add.assert_called()
        mock_session.commit.assert_called()

    @pytest.mark.asyncio
    async def test_post_metrics_with_mentor_forwarding(self):
        """Test posting metrics with mentor API forwarding."""
        mock_session = create_mock_db_session()

        device_id = str(uuid4())
        payload = {"cpu_usage": 50.0}

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = "http://mentor:8080"
                with patch("app.api.v1.endpoints.devices.post_with_retry") as mock_post:
                    mock_post.return_value = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)

        assert response.status_code == 200
        mock_post.assert_called_once()
        call_args = mock_post.call_args
        assert "http://mentor:8080/devices/metrics" in call_args[0]


# ==============================================================================
# Tests for devices.py - post_processes
# ==============================================================================


class TestPostProcessesFullCoverage:
    """Tests for post_processes endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_post_processes_success(self):
        """Test posting processes successfully."""
        mock_session = create_mock_db_session()

        # Mock execute for both delete and insert operations
        mock_session.execute = AsyncMock()

        device_id = str(uuid4())
        processes = [
            {"pid": 1234, "process_name": "chrome", "cpu": 15.5, "memory": 500000000, "command_text": "/usr/bin/chrome"},
            {"pid": 5678, "process_name": "firefox", "cpu": 10.2, "memory": 300000000, "command_text": "/usr/bin/firefox"},
        ]

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = None  # No forwarding
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes)

        assert response.status_code == 200
        data = response.json()
        assert data["inserted"] == 2
        mock_session.commit.assert_called()

    @pytest.mark.asyncio
    async def test_post_processes_with_mentor_forwarding(self):
        """Test posting processes with mentor API forwarding."""
        mock_session = create_mock_db_session()
        mock_session.execute = AsyncMock()

        device_id = str(uuid4())
        processes = [{"pid": 1234, "process_name": "chrome", "cpu": 15.5, "memory": 500000000}]

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = "http://mentor:8080"
                with patch("app.api.v1.endpoints.devices.post_with_retry") as mock_post:
                    mock_post.return_value = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post(f"/api/v1/devices/{device_id}/processes", json=processes)

        assert response.status_code == 200
        mock_post.assert_called_once()
        call_args = mock_post.call_args
        assert "http://mentor:8080/devices/processes" in call_args[0]


# ==============================================================================
# Tests for devices.py - post_activity
# ==============================================================================


class TestPostActivityFullCoverage:
    """Tests for post_activity endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_post_activity_success(self):
        """Test posting activities successfully."""
        mock_session = create_mock_db_session()
        mock_session.execute = AsyncMock()

        device_id = str(uuid4())
        activities = [
            {"activity_type": "app_launch", "description": "User launched Chrome", "app": "chrome", "duration": 3600},
            {"activity_type": "app_close", "description": "User closed Firefox", "app": "firefox", "duration": 7200},
        ]

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = None  # No forwarding
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.post(f"/api/v1/devices/{device_id}/activities", json=activities)

        assert response.status_code == 200
        data = response.json()
        assert data["inserted"] == 2
        mock_session.commit.assert_called()

    @pytest.mark.asyncio
    async def test_post_activity_with_empty_legacy_type_returns_422(self):
        """Test that empty legacy 'type' field returns 422."""
        mock_session = create_mock_db_session()

        device_id = str(uuid4())
        # Empty 'type' without 'activity_type' should trigger 422
        activities = [{"type": "", "description": "Test activity"}]

        with override_db_dependency(mock_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(f"/api/v1/devices/{device_id}/activities", json=activities)

        assert response.status_code == 422
        assert "invalid field: use activity_type instead of type" in response.json()["detail"]

    @pytest.mark.asyncio
    async def test_post_activity_with_mentor_forwarding(self):
        """Test posting activities with mentor API forwarding."""
        mock_session = create_mock_db_session()
        mock_session.execute = AsyncMock()

        device_id = str(uuid4())
        activities = [{"activity_type": "app_launch", "description": "Test", "app": "test", "duration": 100}]

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = "http://mentor:8080"
                with patch("app.api.v1.endpoints.devices.post_with_retry") as mock_post:
                    mock_post.return_value = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post(f"/api/v1/devices/{device_id}/activities", json=activities)

        assert response.status_code == 200
        mock_post.assert_called_once()
        call_args = mock_post.call_args
        assert "http://mentor:8080/devices/activity" in call_args[0]


# ==============================================================================
# Tests for devices.py - post_alerts
# ==============================================================================


class TestPostAlertsFullCoverage:
    """Tests for post_alerts endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_post_alerts_success(self):
        """Test posting alerts successfully."""
        mock_session = create_mock_db_session()
        mock_session.execute = AsyncMock()

        device_id = str(uuid4())
        alerts = [
            {"level": "warning", "alert_type": "cpu", "message": "High CPU usage", "value": 85.5, "threshold": 80.0},
            {"level": "critical", "alert_type": "memory", "message": "Memory critical", "value": 95.0, "threshold": 90.0},
        ]

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = None  # No forwarding
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.post(f"/api/v1/devices/{device_id}/alerts", json=alerts)

        assert response.status_code == 200
        data = response.json()
        assert data["inserted"] == 2
        mock_session.commit.assert_called()

    @pytest.mark.asyncio
    async def test_post_alerts_with_mentor_forwarding(self):
        """Test posting alerts with mentor API forwarding."""
        mock_session = create_mock_db_session()
        mock_session.execute = AsyncMock()

        device_id = str(uuid4())
        alerts = [{"level": "warning", "alert_type": "cpu", "message": "High CPU", "value": 85.0, "threshold": 80.0}]

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = "http://mentor:8080"
                with patch("app.api.v1.endpoints.devices.post_with_retry") as mock_post:
                    mock_post.return_value = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post(f"/api/v1/devices/{device_id}/alerts", json=alerts)

        assert response.status_code == 200
        mock_post.assert_called_once()
        call_args = mock_post.call_args
        assert f"http://mentor:8080/devices/{device_id}/alerts" in call_args[0]


# ==============================================================================
# Tests for devices.py - get_pending_commands
# ==============================================================================


class TestGetPendingCommandsFullCoverage:
    """Tests for get_pending_commands endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_get_pending_commands_returns_list(self):
        """Test getting pending commands for a device."""
        mock_session = create_mock_db_session()

        # Mock: return a list of pending commands
        mock_cmd = MagicMock()
        mock_cmd.commandid = uuid4()
        mock_cmd.deviceid = str(uuid4())
        mock_cmd.command_text = "get_info"
        mock_cmd.status = "pending"
        mock_cmd.created_at = datetime.now(timezone.utc)
        mock_cmd.completed_at = None
        mock_cmd.result = None
        mock_cmd.exit_code = None

        mock_result = MagicMock()
        mock_result.scalars.return_value.all.return_value = [mock_cmd]
        mock_session.execute.return_value = mock_result

        device_id = str(uuid4())

        with override_db_dependency(mock_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.get(f"/api/v1/devices/{device_id}/commands/pending")

        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)

    @pytest.mark.asyncio
    async def test_get_pending_commands_empty_list(self):
        """Test getting pending commands when none exist."""
        mock_session = create_mock_db_session()

        mock_result = MagicMock()
        mock_result.scalars.return_value.all.return_value = []
        mock_session.execute.return_value = mock_result

        device_id = str(uuid4())

        with override_db_dependency(mock_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.get(f"/api/v1/devices/{device_id}/commands/pending")

        assert response.status_code == 200
        data = response.json()
        assert data == []


# ==============================================================================
# Tests for devices.py - submit_command_result
# ==============================================================================


class TestSubmitCommandResultFullCoverage:
    """Tests for submit_command_result endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_submit_command_result_success(self):
        """Test submitting command result successfully."""
        mock_session = create_mock_db_session()

        # Mock: command exists
        command_id = uuid4()
        mock_cmd = MagicMock()
        mock_cmd.commandid = command_id
        mock_cmd.status = "pending"
        mock_cmd.result = ""
        mock_cmd.exit_code = 0

        mock_result = MagicMock()
        mock_result.scalars.return_value.first.return_value = mock_cmd
        mock_session.execute.return_value = mock_result

        payload = {"status": "completed", "result": "Command output here", "exit_code": 0}

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = None  # No forwarding
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=payload)

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert data["commandid"] == str(command_id)
        assert mock_cmd.status == "completed"
        assert mock_cmd.result == "Command output here"

    @pytest.mark.asyncio
    async def test_submit_command_result_with_none_values(self):
        """Test submitting command result with None result and exit_code."""
        mock_session = create_mock_db_session()

        command_id = uuid4()
        mock_cmd = MagicMock()
        mock_cmd.commandid = command_id
        mock_cmd.status = "pending"
        mock_cmd.result = ""
        mock_cmd.exit_code = 0

        mock_result = MagicMock()
        mock_result.scalars.return_value.first.return_value = mock_cmd
        mock_session.execute.return_value = mock_result

        # Payload with None values to test default handling
        payload = {"status": "completed"}

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = None
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=payload)

        assert response.status_code == 200
        # Default values should be applied
        assert mock_cmd.result == ""
        assert mock_cmd.exit_code == 0

    @pytest.mark.asyncio
    async def test_submit_command_result_with_mentor_forwarding(self):
        """Test submitting command result with mentor API forwarding."""
        mock_session = create_mock_db_session()

        command_id = uuid4()
        mock_cmd = MagicMock()
        mock_cmd.commandid = command_id
        mock_cmd.status = "pending"
        mock_cmd.result = ""
        mock_cmd.exit_code = 0

        mock_result = MagicMock()
        mock_result.scalars.return_value.first.return_value = mock_cmd
        mock_session.execute.return_value = mock_result

        payload = {"status": "completed", "result": "Output", "exit_code": 0}

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.devices.settings") as mock_settings:
                mock_settings.mentor_api_url = "http://mentor:8080"
                with patch("app.api.v1.endpoints.devices.post_with_retry") as mock_post:
                    mock_post.return_value = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=payload)

        assert response.status_code == 200
        mock_post.assert_called_once()
        call_args = mock_post.call_args
        assert "http://mentor:8080/commands/status" in call_args[0]


# ==============================================================================
# Tests for devices.py - create_command
# ==============================================================================


class TestCreateCommandFullCoverage:
    """Tests for create_command endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_create_command_success(self):
        """Test creating a command successfully."""
        mock_session = create_mock_db_session()

        # Mock: create command
        command_id = uuid4()

        async def mock_refresh(obj):
            obj.commandid = command_id
            obj.status = "pending"
            obj.created_at = datetime.now(timezone.utc)

        mock_session.refresh = mock_refresh

        device_id = str(uuid4())
        payload = {"command_text": "get_info"}

        with override_db_dependency(mock_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)

        assert response.status_code == 200
        data = response.json()
        assert data["deviceid"] == device_id
        assert data["command_text"] == "get_info"
        assert data["status"] == "pending"
        mock_session.add.assert_called()
        mock_session.commit.assert_called()

    @pytest.mark.asyncio
    async def test_create_command_empty_command_text(self):
        """Test creating a command with empty command_text."""
        mock_session = create_mock_db_session()

        device_id = str(uuid4())
        payload = {"command_text": ""}

        with override_db_dependency(mock_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)

        # Empty command_text fails pydantic validation (min_length=1)
        assert response.status_code == 422

    @pytest.mark.asyncio
    async def test_create_command_not_allowed(self):
        """Test creating a command with a disallowed command."""
        mock_session = create_mock_db_session()

        device_id = str(uuid4())
        payload = {"command_text": "rm -rf /"}

        with override_db_dependency(mock_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(f"/api/v1/devices/{device_id}/commands", json=payload)

        assert response.status_code == 400
        assert "not allowed" in response.json()["detail"].lower()
        assert "Allowed commands" in response.json()["detail"]


# ==============================================================================
# Tests for devices.py - submit_command_result (404 case)
# ==============================================================================


class TestSubmitCommandResultNotFound:
    """Additional tests for submit_command_result 404 handling."""

    @pytest.mark.asyncio
    async def test_submit_command_result_not_found(self):
        """Test submitting result for non-existent command."""
        mock_session = create_mock_db_session()

        # Mock: command doesn't exist
        mock_result = MagicMock()
        mock_result.scalars.return_value.first.return_value = None
        mock_session.execute.return_value = mock_result

        command_id = uuid4()
        payload = {"status": "completed", "result": "Output", "exit_code": 0}

        with override_db_dependency(mock_session):
            async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                response = await ac.post(f"/api/v1/devices/commands/{command_id}/result", json=payload)

        assert response.status_code == 404
        assert "Command not found" in response.json()["detail"]


# ==============================================================================
# Tests for health.py - readiness_check
# ==============================================================================


class TestReadinessCheckFullCoverage:
    """Tests for readiness_check endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_readiness_check_success(self):
        """Test readiness check when everything is ok."""
        mock_session = create_mock_db_session()

        # Mock database query success
        mock_result = MagicMock()
        mock_result.scalar.return_value = 1
        mock_session.execute.return_value = mock_result

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.health.settings") as mock_settings:
                mock_settings.database_url = "postgresql://test"
                mock_settings.secret_key = "test-secret-key"
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.get("/health/ready")

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ready"
        assert data["checks"]["database"] == "ok"
        assert data["checks"]["config"] == "ok"

    @pytest.mark.asyncio
    async def test_readiness_check_database_failure(self):
        """Test readiness check when database check fails."""
        mock_session = create_mock_db_session()

        # Mock database query failure
        mock_session.execute.side_effect = Exception("Database connection failed")

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.health.settings") as mock_settings:
                mock_settings.database_url = "postgresql://test"
                mock_settings.secret_key = "test-secret-key"
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.get("/health/ready")

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "not_ready"
        assert "error" in data["checks"]["database"]

    @pytest.mark.asyncio
    async def test_readiness_check_missing_database_url(self):
        """Test readiness check when database_url is not configured."""
        mock_session = create_mock_db_session()

        # Mock database query success
        mock_result = MagicMock()
        mock_result.scalar.return_value = 1
        mock_session.execute.return_value = mock_result

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.health.settings") as mock_settings:
                mock_settings.database_url = None  # Not configured
                mock_settings.secret_key = "test-secret-key"
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.get("/health/ready")

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "not_ready"
        assert "database_url not configured" in data["checks"]["config"]

    @pytest.mark.asyncio
    async def test_readiness_check_missing_secret_key(self):
        """Test readiness check when secret_key is not configured."""
        mock_session = create_mock_db_session()

        # Mock database query success
        mock_result = MagicMock()
        mock_result.scalar.return_value = 1
        mock_session.execute.return_value = mock_result

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.health.settings") as mock_settings:
                mock_settings.database_url = "postgresql://test"
                mock_settings.secret_key = None  # Not configured
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.get("/health/ready")

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "not_ready"
        assert "secret_key not configured" in data["checks"]["config"]

    @pytest.mark.asyncio
    async def test_readiness_check_config_exception(self):
        """Test readiness check when settings access throws exception."""
        mock_session = create_mock_db_session()

        # Mock database query success
        mock_result = MagicMock()
        mock_result.scalar.return_value = 1
        mock_session.execute.return_value = mock_result

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.health.settings") as mock_settings:
                # First call succeeds (database_url check), second raises exception
                type(mock_settings).database_url = property(lambda self: (_ for _ in ()).throw(Exception("Config error")))
                async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                    response = await ac.get("/health/ready")

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "not_ready"


# ==============================================================================
# Tests for screenshots.py - create_screenshot
# ==============================================================================


class TestCreateScreenshotFullCoverage:
    """Tests for create_screenshot endpoint to achieve full coverage."""

    @pytest.mark.asyncio
    async def test_create_screenshot_missing_device_id(self):
        """Test screenshot creation with missing device_id."""
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            fake_image = io.BytesIO(b"fake image content")
            response = await ac.post(
                "/api/v1/screenshots/",
                data={},  # No device_id
                files={"file": ("screenshot.png", fake_image, "image/png")},
            )

        assert response.status_code == 422
        assert "device_id is required" in response.json()["detail"]

    @pytest.mark.asyncio
    async def test_create_screenshot_using_deviceid_field(self):
        """Test screenshot creation using 'deviceid' field instead of 'device_id'."""
        mock_session = create_mock_db_session()

        # Mock refresh to set screenshotid
        async def mock_refresh(obj):
            obj.screenshotid = uuid4()

        mock_session.refresh = mock_refresh

        device_id = str(uuid4())
        fake_image = io.BytesIO(b"fake image content")

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.screenshots.MinioService") as mock_minio_class:
                mock_minio = MagicMock()
                mock_minio.upload_file.return_value = None
                mock_minio_class.return_value = mock_minio
                with patch("app.api.v1.endpoints.screenshots.settings") as mock_settings:
                    mock_settings.mentor_api_url = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post(
                            "/api/v1/screenshots/",
                            data={"deviceid": device_id},  # Using deviceid
                            files={"file": ("screenshot.png", fake_image, "image/png")},
                        )

        assert response.status_code == 201
        data = response.json()
        assert data["status"] == "success"

    @pytest.mark.asyncio
    async def test_create_screenshot_minio_failure(self):
        """Test screenshot creation when MinIO upload fails (should continue)."""
        mock_session = create_mock_db_session()

        async def mock_refresh(obj):
            obj.screenshotid = uuid4()

        mock_session.refresh = mock_refresh

        device_id = str(uuid4())
        fake_image = io.BytesIO(b"fake image content")

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.screenshots.MinioService") as mock_minio_class:
                mock_minio = MagicMock()
                mock_minio.upload_file.side_effect = Exception("MinIO connection failed")
                mock_minio_class.return_value = mock_minio
                with patch("app.api.v1.endpoints.screenshots.settings") as mock_settings:
                    mock_settings.mentor_api_url = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post(
                            "/api/v1/screenshots/",
                            data={"device_id": device_id},
                            files={"file": ("screenshot.png", fake_image, "image/png")},
                        )

        # Should still succeed despite MinIO failure
        assert response.status_code == 201
        data = response.json()
        assert data["status"] == "success"

    @pytest.mark.asyncio
    async def test_create_screenshot_with_mentor_forwarding(self):
        """Test screenshot creation with mentor API forwarding."""
        mock_session = create_mock_db_session()

        async def mock_refresh(obj):
            obj.screenshotid = uuid4()

        mock_session.refresh = mock_refresh

        device_id = str(uuid4())
        fake_image = io.BytesIO(b"fake image content")

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.screenshots.MinioService") as mock_minio_class:
                mock_minio = MagicMock()
                mock_minio.upload_file.return_value = None
                mock_minio_class.return_value = mock_minio
                with patch("app.api.v1.endpoints.screenshots.settings") as mock_settings:
                    mock_settings.mentor_api_url = "http://mentor:8080"
                    with patch("app.api.v1.endpoints.screenshots.post_with_retry") as mock_post:
                        mock_post.return_value = None
                        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                            response = await ac.post(
                                "/api/v1/screenshots/",
                                data={"device_id": device_id},
                                files={"file": ("screenshot.png", fake_image, "image/png")},
                            )

        assert response.status_code == 201
        mock_post.assert_called_once()
        call_args = mock_post.call_args
        assert "http://mentor:8080/devices/screenshots" in call_args[0]

    @pytest.mark.asyncio
    async def test_create_screenshot_general_exception(self):
        """Test screenshot creation when an unexpected exception occurs."""
        mock_session = create_mock_db_session()

        # Mock commit to raise an exception
        mock_session.commit.side_effect = Exception("Database error")

        device_id = str(uuid4())
        fake_image = io.BytesIO(b"fake image content")

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.screenshots.MinioService") as mock_minio_class:
                mock_minio = MagicMock()
                mock_minio.upload_file.return_value = None
                mock_minio_class.return_value = mock_minio
                with patch("app.api.v1.endpoints.screenshots.settings") as mock_settings:
                    mock_settings.mentor_api_url = None
                    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                        response = await ac.post(
                            "/api/v1/screenshots/",
                            data={"device_id": device_id},
                            files={"file": ("screenshot.png", fake_image, "image/png")},
                        )

        assert response.status_code == 500
        assert "Screenshot upload failed" in response.json()["detail"]

    @pytest.mark.asyncio
    async def test_create_screenshot_cleanup_error(self):
        """Test screenshot creation with temp file cleanup error."""
        mock_session = create_mock_db_session()

        async def mock_refresh(obj):
            obj.screenshotid = uuid4()

        mock_session.refresh = mock_refresh

        device_id = str(uuid4())
        fake_image = io.BytesIO(b"fake image content")

        with override_db_dependency(mock_session):
            with patch("app.api.v1.endpoints.screenshots.MinioService") as mock_minio_class:
                mock_minio = MagicMock()
                mock_minio.upload_file.return_value = None
                mock_minio_class.return_value = mock_minio
                with patch("app.api.v1.endpoints.screenshots.settings") as mock_settings:
                    mock_settings.mentor_api_url = None
                    # Patch Path to simulate cleanup error
                    with patch("app.api.v1.endpoints.screenshots.Path") as mock_path:
                        mock_path_instance = MagicMock()
                        mock_path_instance.exists.return_value = True
                        mock_path_instance.unlink.side_effect = Exception("Permission denied")
                        mock_path.return_value = mock_path_instance
                        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
                            response = await ac.post(
                                "/api/v1/screenshots/",
                                data={"device_id": device_id},
                                files={"file": ("screenshot.png", fake_image, "image/png")},
                            )

        # Should still succeed despite cleanup error
        assert response.status_code == 201
