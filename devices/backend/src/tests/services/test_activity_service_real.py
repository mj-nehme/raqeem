"""Real tests for ActivityService - no mocking, direct function calls."""

import pytest
from app.services.activity_service import ActivityService


class TestActivityServiceReal:
    """Test ActivityService with real function calls."""

    @pytest.fixture
    def service(self):
        """Create a real ActivityService instance."""
        return ActivityService()

    # Activity Type Validation Tests
    def test_validate_activity_type_app_launch(self, service):
        """Test app_launch is a valid activity type."""
        assert service.validate_activity_type("app_launch") is True

    def test_validate_activity_type_app_close(self, service):
        """Test app_close is a valid activity type."""
        assert service.validate_activity_type("app_close") is True

    def test_validate_activity_type_file_access(self, service):
        """Test file_access is a valid activity type."""
        assert service.validate_activity_type("file_access") is True

    def test_validate_activity_type_web_visit(self, service):
        """Test web_visit is a valid activity type."""
        assert service.validate_activity_type("web_visit") is True

    def test_validate_activity_type_idle(self, service):
        """Test idle is a valid activity type."""
        assert service.validate_activity_type("idle") is True

    def test_validate_activity_type_invalid(self, service):
        """Test invalid activity type."""
        assert service.validate_activity_type("invalid_type") is False

    def test_validate_activity_type_empty(self, service):
        """Test empty activity type."""
        assert service.validate_activity_type("") is False

    def test_validate_activity_type_uppercase(self, service):
        """Test uppercase activity type is invalid (case-sensitive)."""
        assert service.validate_activity_type("APP_LAUNCH") is False

    def test_validate_activity_type_mixed_case(self, service):
        """Test mixed case activity type is invalid."""
        assert service.validate_activity_type("App_Launch") is False

    # Duration Validation Tests
    def test_validate_duration_zero(self, service):
        """Test zero duration is valid."""
        assert service.validate_duration(0) is True

    def test_validate_duration_positive(self, service):
        """Test positive duration is valid."""
        assert service.validate_duration(30) is True

    def test_validate_duration_large(self, service):
        """Test large duration is valid."""
        assert service.validate_duration(86400) is True  # 1 day

    def test_validate_duration_negative(self, service):
        """Test negative duration is invalid."""
        assert service.validate_duration(-1) is False

    def test_validate_duration_large_negative(self, service):
        """Test large negative duration is invalid."""
        assert service.validate_duration(-3600) is False

    def test_validate_duration_float_positive(self, service):
        """Test positive float duration is valid."""
        assert service.validate_duration(30.5) is True

    def test_validate_duration_float_negative(self, service):
        """Test negative float duration is invalid."""
        assert service.validate_duration(-0.5) is False

    # Activity Pattern Analysis Tests
    def test_analyze_patterns_empty(self, service):
        """Test pattern analysis with empty list."""
        result = service.analyze_activity_patterns([])
        assert result["most_used_app"] is None
        assert result["total_activities"] == 0
        assert result["unique_apps"] == 0

    def test_analyze_patterns_single_activity(self, service):
        """Test pattern analysis with single activity."""
        activities = [{"type": "app_launch", "app": "chrome"}]
        result = service.analyze_activity_patterns(activities)
        assert result["most_used_app"] == "chrome"
        assert result["total_activities"] == 1
        assert result["unique_apps"] == 1

    def test_analyze_patterns_multiple_same_app(self, service):
        """Test pattern analysis with multiple activities of same app."""
        activities = [
            {"type": "app_launch", "app": "chrome"},
            {"type": "app_launch", "app": "chrome"},
            {"type": "app_launch", "app": "chrome"},
        ]
        result = service.analyze_activity_patterns(activities)
        assert result["most_used_app"] == "chrome"
        assert result["total_activities"] == 3
        assert result["unique_apps"] == 1

    def test_analyze_patterns_multiple_different_apps(self, service):
        """Test pattern analysis with multiple different apps."""
        activities = [
            {"type": "app_launch", "app": "chrome"},
            {"type": "app_launch", "app": "slack"},
            {"type": "app_launch", "app": "vscode"},
        ]
        result = service.analyze_activity_patterns(activities)
        assert result["total_activities"] == 3
        assert result["unique_apps"] == 3

    def test_analyze_patterns_most_used_app(self, service):
        """Test pattern analysis correctly identifies most used app."""
        activities = [
            {"type": "app_launch", "app": "chrome"},
            {"type": "app_launch", "app": "slack"},
            {"type": "app_launch", "app": "chrome"},
            {"type": "app_launch", "app": "vscode"},
            {"type": "app_launch", "app": "chrome"},
        ]
        result = service.analyze_activity_patterns(activities)
        assert result["most_used_app"] == "chrome"
        assert result["total_activities"] == 5
        assert result["unique_apps"] == 3

    def test_analyze_patterns_missing_app_field(self, service):
        """Test pattern analysis with missing app field."""
        activities = [
            {"type": "app_launch", "app": "chrome"},
            {"type": "file_access"},  # No app field
            {"type": "app_launch", "app": "slack"},
        ]
        result = service.analyze_activity_patterns(activities)
        assert result["total_activities"] == 3
        assert result["unique_apps"] == 2

    def test_analyze_patterns_empty_app_field(self, service):
        """Test pattern analysis with empty app field."""
        activities = [
            {"type": "app_launch", "app": "chrome"},
            {"type": "app_launch", "app": ""},
            {"type": "app_launch", "app": "slack"},
        ]
        result = service.analyze_activity_patterns(activities)
        # Empty string is filtered out
        assert result["unique_apps"] == 2

    def test_analyze_patterns_none_app_field(self, service):
        """Test pattern analysis with None app field."""
        activities = [
            {"type": "app_launch", "app": "chrome"},
            {"type": "app_launch", "app": None},
            {"type": "app_launch", "app": "slack"},
        ]
        result = service.analyze_activity_patterns(activities)
        # None is filtered out
        assert result["unique_apps"] == 2
