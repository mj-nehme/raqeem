"""Tests for DeviceService."""

from datetime import datetime, timedelta, timezone

from app.services.device_service import (
    DEVICE_ONLINE_THRESHOLD_MINUTES,
    VALID_DEVICE_TYPES,
    DeviceService,
)


class TestDeviceServiceConstants:
    """Tests for DeviceService constants."""

    def test_device_online_threshold(self):
        """Test device online threshold constant."""
        assert DEVICE_ONLINE_THRESHOLD_MINUTES == 5

    def test_valid_device_types(self):
        """Test valid device types constant."""
        expected_types = {"laptop", "desktop", "server", "mobile", "tablet"}
        assert expected_types == VALID_DEVICE_TYPES


class TestDeviceServiceValidateDeviceData:
    """Tests for DeviceService.validate_device_data method."""

    def setup_method(self):
        """Set up test fixtures."""
        self.service = DeviceService()

    def test_validate_device_data_valid(self):
        """Test validation of valid device data."""
        data = {"deviceid": "test-device-123"}
        assert self.service.validate_device_data(data) is True

    def test_validate_device_data_missing_deviceid(self):
        """Test validation fails when deviceid is missing."""
        data = {"device_name": "Test Device"}
        assert self.service.validate_device_data(data) is False

    def test_validate_device_data_with_valid_device_type(self):
        """Test validation with valid device type."""
        data = {"deviceid": "test-device-123", "device_type": "laptop"}
        assert self.service.validate_device_data(data) is True

    def test_validate_device_data_with_invalid_device_type(self):
        """Test validation fails with invalid device type."""
        data = {"deviceid": "test-device-123", "device_type": "unknown"}
        assert self.service.validate_device_data(data) is False

    def test_validate_device_data_with_none_device_type(self):
        """Test validation passes when device_type is None."""
        data = {"deviceid": "test-device-123", "device_type": None}
        assert self.service.validate_device_data(data) is True


class TestDeviceServiceIsDeviceOnline:
    """Tests for DeviceService.is_device_online method."""

    def setup_method(self):
        """Set up test fixtures."""
        self.service = DeviceService()

    def test_device_online_recent(self):
        """Test device is online when recently seen."""
        now = datetime.now(timezone.utc)
        last_seen = now - timedelta(minutes=1)  # 1 minute ago
        assert self.service.is_device_online(last_seen) is True

    def test_device_online_at_threshold(self):
        """Test device is online at threshold boundary."""
        now = datetime.now(timezone.utc)
        # Just inside threshold (4 minutes ago)
        last_seen = now - timedelta(minutes=DEVICE_ONLINE_THRESHOLD_MINUTES - 1)
        assert self.service.is_device_online(last_seen) is True

    def test_device_offline_past_threshold(self):
        """Test device is offline past threshold."""
        now = datetime.now(timezone.utc)
        last_seen = now - timedelta(minutes=DEVICE_ONLINE_THRESHOLD_MINUTES + 1)
        assert self.service.is_device_online(last_seen) is False

    def test_device_online_naive_datetime(self):
        """Test device online check with naive datetime."""
        now = datetime.now()
        last_seen = now - timedelta(minutes=1)  # Naive datetime
        # Should handle naive datetime by assuming UTC
        result = self.service.is_device_online(last_seen)
        assert isinstance(result, bool)

    def test_device_offline_way_past_threshold(self):
        """Test device is offline when far past threshold."""
        now = datetime.now(timezone.utc)
        last_seen = now - timedelta(hours=1)  # 1 hour ago
        assert self.service.is_device_online(last_seen) is False


class TestDeviceServiceValidateDeviceType:
    """Tests for DeviceService.validate_device_type method."""

    def setup_method(self):
        """Set up test fixtures."""
        self.service = DeviceService()

    def test_validate_device_type_laptop(self):
        """Test validation of laptop device type."""
        assert self.service.validate_device_type("laptop") is True

    def test_validate_device_type_desktop(self):
        """Test validation of desktop device type."""
        assert self.service.validate_device_type("desktop") is True

    def test_validate_device_type_server(self):
        """Test validation of server device type."""
        assert self.service.validate_device_type("server") is True

    def test_validate_device_type_mobile(self):
        """Test validation of mobile device type."""
        assert self.service.validate_device_type("mobile") is True

    def test_validate_device_type_tablet(self):
        """Test validation of tablet device type."""
        assert self.service.validate_device_type("tablet") is True

    def test_validate_device_type_invalid(self):
        """Test validation of invalid device type."""
        assert self.service.validate_device_type("unknown") is False

    def test_validate_device_type_empty_string(self):
        """Test validation of empty string device type."""
        assert self.service.validate_device_type("") is False

    def test_validate_device_type_case_sensitive(self):
        """Test that device type validation is case sensitive."""
        assert self.service.validate_device_type("Laptop") is False
        assert self.service.validate_device_type("LAPTOP") is False
