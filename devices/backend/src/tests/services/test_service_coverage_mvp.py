"""Additional service layer tests to improve coverage for MVP."""

import pytest
from app.services.alert_service import AlertService
from app.services.device_service import DeviceService
from app.services.file_service import FileService
from app.services.metrics_service import MetricsService
from app.services.security_service import SecurityService


class TestAlertServiceCoverage:
    """Tests for AlertService to improve coverage."""

    def test_should_create_alert_above_threshold(self):
        """Test alert creation when value exceeds threshold."""
        service = AlertService()
        assert service.should_create_alert(85.0, 80.0) is True
        assert service.should_create_alert(100.0, 90.0) is True

    def test_should_create_alert_below_threshold(self):
        """Test alert not created when value below threshold."""
        service = AlertService()
        assert service.should_create_alert(70.0, 80.0) is False
        assert service.should_create_alert(50.0, 90.0) is False

    def test_should_create_alert_equal_threshold(self):
        """Test alert created when value equals threshold."""
        service = AlertService()
        assert service.should_create_alert(80.0, 80.0) is True

    def test_determine_alert_level_info(self):
        """Test alert level determination for info level."""
        service = AlertService()
        assert service.determine_alert_level(75.0, 80.0) == "info"

    def test_determine_alert_level_warning(self):
        """Test alert level determination for warning level."""
        service = AlertService()
        # Delta < WARNING_DELTA (10)
        assert service.determine_alert_level(85.0, 80.0) == "warning"
        assert service.determine_alert_level(88.0, 80.0) == "warning"

    def test_determine_alert_level_error(self):
        """Test alert level determination for error level."""
        service = AlertService()
        # WARNING_DELTA <= Delta < ERROR_DELTA (10 <= delta < 20)
        assert service.determine_alert_level(95.0, 80.0) == "error"
        assert service.determine_alert_level(98.0, 80.0) == "error"

    def test_determine_alert_level_critical(self):
        """Test alert level determination for critical level."""
        service = AlertService()
        # Delta >= ERROR_DELTA (20)
        assert service.determine_alert_level(100.0, 80.0) == "critical"
        assert service.determine_alert_level(150.0, 80.0) == "critical"


class TestDeviceServiceCoverage:
    """Tests for DeviceService to improve coverage."""

    def test_device_service_initialization(self):
        """Test DeviceService can be instantiated."""
        service = DeviceService()
        assert service is not None


class TestFileServiceCoverage:
    """Tests for FileService to improve coverage."""

    def test_file_service_initialization(self):
        """Test FileService can be instantiated."""
        service = FileService()
        assert service is not None


class TestMetricsServiceCoverage:
    """Tests for MetricsService to improve coverage."""

    def test_metrics_service_initialization(self):
        """Test MetricsService can be instantiated."""
        service = MetricsService()
        assert service is not None


class TestSecurityServiceCoverage:
    """Tests for SecurityService to improve coverage."""

    def test_security_service_initialization(self):
        """Test SecurityService can be instantiated."""
        service = SecurityService()
        assert service is not None
