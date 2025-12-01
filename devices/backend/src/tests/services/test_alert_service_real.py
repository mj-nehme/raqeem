"""Real tests for AlertService - no mocking, direct function calls."""

import pytest
from app.services.alert_service import AlertService


class TestAlertServiceReal:
    """Test AlertService with real function calls."""

    @pytest.fixture
    def service(self):
        """Create a real AlertService instance."""
        return AlertService()

    # should_create_alert Tests
    def test_should_create_alert_value_exceeds_threshold(self, service):
        """Test alert is created when value exceeds threshold."""
        assert service.should_create_alert(85.0, 80.0) is True

    def test_should_create_alert_value_equals_threshold(self, service):
        """Test alert is created when value equals threshold."""
        assert service.should_create_alert(80.0, 80.0) is True

    def test_should_create_alert_value_below_threshold(self, service):
        """Test alert is not created when value is below threshold."""
        assert service.should_create_alert(75.0, 80.0) is False

    def test_should_create_alert_zero_values(self, service):
        """Test with zero values."""
        assert service.should_create_alert(0.0, 0.0) is True

    def test_should_create_alert_large_delta(self, service):
        """Test alert with large delta."""
        assert service.should_create_alert(100.0, 50.0) is True

    def test_should_create_alert_small_delta_below(self, service):
        """Test no alert with small value below threshold."""
        assert service.should_create_alert(79.9, 80.0) is False

    # determine_alert_level Tests
    def test_determine_alert_level_info(self, service):
        """Test info level when below threshold."""
        assert service.determine_alert_level(75.0, 80.0) == "info"

    def test_determine_alert_level_warning(self, service):
        """Test warning level when just above threshold."""
        # Delta is 5, which is < 10 (WARNING_DELTA)
        assert service.determine_alert_level(85.0, 80.0) == "warning"

    def test_determine_alert_level_error(self, service):
        """Test error level when moderately above threshold."""
        # Delta is 15, which is >= 10 but < 20 (ERROR_DELTA)
        assert service.determine_alert_level(95.0, 80.0) == "error"

    def test_determine_alert_level_critical(self, service):
        """Test critical level when well above threshold."""
        # Delta is 25, which is >= 20 (ERROR_DELTA)
        assert service.determine_alert_level(105.0, 80.0) == "critical"

    def test_determine_alert_level_at_threshold(self, service):
        """Test warning level when exactly at threshold."""
        # Delta is 0, which is >= 0 and < 10
        assert service.determine_alert_level(80.0, 80.0) == "warning"

    def test_determine_alert_level_boundary_warning_error(self, service):
        """Test boundary between warning and error."""
        # Delta is 10, which is >= 10 (WARNING_DELTA)
        assert service.determine_alert_level(90.0, 80.0) == "error"

    def test_determine_alert_level_boundary_error_critical(self, service):
        """Test boundary between error and critical."""
        # Delta is 20, which is >= 20 (ERROR_DELTA)
        assert service.determine_alert_level(100.0, 80.0) == "critical"

    def test_determine_alert_level_zero_threshold(self, service):
        """Test with zero threshold."""
        # Delta is 50
        assert service.determine_alert_level(50.0, 0.0) == "critical"

    def test_determine_alert_level_negative_delta(self, service):
        """Test with negative delta (well below threshold)."""
        assert service.determine_alert_level(50.0, 80.0) == "info"

    # calculate_threshold_percentage Tests
    def test_calculate_threshold_percentage_above(self, service):
        """Test percentage calculation when above threshold."""
        # (90-80)/80 * 100 = 12.5%
        assert service.calculate_threshold_percentage(90.0, 80.0) == pytest.approx(12.5)

    def test_calculate_threshold_percentage_below(self, service):
        """Test percentage is 0 when below threshold."""
        assert service.calculate_threshold_percentage(75.0, 80.0) == 0.0

    def test_calculate_threshold_percentage_at_threshold(self, service):
        """Test percentage is 0 when at threshold."""
        assert service.calculate_threshold_percentage(80.0, 80.0) == 0.0

    def test_calculate_threshold_percentage_zero_threshold(self, service):
        """Test with zero threshold returns 0."""
        assert service.calculate_threshold_percentage(50.0, 0.0) == 0.0

    def test_calculate_threshold_percentage_double(self, service):
        """Test 100% over threshold."""
        # (160-80)/80 * 100 = 100%
        assert service.calculate_threshold_percentage(160.0, 80.0) == pytest.approx(100.0)

    def test_calculate_threshold_percentage_small_excess(self, service):
        """Test small excess over threshold."""
        # (81-80)/80 * 100 = 1.25%
        assert service.calculate_threshold_percentage(81.0, 80.0) == pytest.approx(1.25)

    def test_calculate_threshold_percentage_large_excess(self, service):
        """Test large excess over threshold."""
        # (180-80)/80 * 100 = 125%
        assert service.calculate_threshold_percentage(180.0, 80.0) == pytest.approx(125.0)
