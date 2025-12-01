"""Real tests for MetricsService - no mocking, direct function calls."""

import pytest
from app.services.metrics_service import MetricsService


class TestMetricsServiceReal:
    """Test MetricsService with real function calls."""

    @pytest.fixture
    def service(self):
        """Create a real MetricsService instance."""
        return MetricsService()

    # CPU Usage Validation Tests
    def test_validate_cpu_usage_zero(self, service):
        """Test CPU usage at 0% is valid."""
        assert service.validate_cpu_usage(0) is True

    def test_validate_cpu_usage_hundred(self, service):
        """Test CPU usage at 100% is valid."""
        assert service.validate_cpu_usage(100) is True

    def test_validate_cpu_usage_fifty(self, service):
        """Test CPU usage at 50% is valid."""
        assert service.validate_cpu_usage(50.0) is True

    def test_validate_cpu_usage_decimal(self, service):
        """Test CPU usage with decimal is valid."""
        assert service.validate_cpu_usage(75.5) is True

    def test_validate_cpu_usage_negative(self, service):
        """Test negative CPU usage is invalid."""
        assert service.validate_cpu_usage(-1) is False

    def test_validate_cpu_usage_over_hundred(self, service):
        """Test CPU usage over 100% is invalid."""
        assert service.validate_cpu_usage(101) is False

    def test_validate_cpu_usage_large_negative(self, service):
        """Test large negative CPU usage is invalid."""
        assert service.validate_cpu_usage(-100) is False

    def test_validate_cpu_usage_large_positive(self, service):
        """Test large positive CPU usage is invalid."""
        assert service.validate_cpu_usage(200) is False

    # Memory Usage Validation Tests
    def test_validate_memory_usage_valid(self, service):
        """Test valid memory usage."""
        assert service.validate_memory_usage(4 * 1024**3, 8 * 1024**3) is True

    def test_validate_memory_usage_zero(self, service):
        """Test zero memory used is valid."""
        assert service.validate_memory_usage(0, 8 * 1024**3) is True

    def test_validate_memory_usage_full(self, service):
        """Test full memory usage is valid."""
        assert service.validate_memory_usage(8 * 1024**3, 8 * 1024**3) is True

    def test_validate_memory_usage_exceeds_total(self, service):
        """Test memory used exceeding total is invalid."""
        assert service.validate_memory_usage(16 * 1024**3, 8 * 1024**3) is False

    def test_validate_memory_usage_negative_used(self, service):
        """Test negative memory used is invalid."""
        assert service.validate_memory_usage(-1, 8 * 1024**3) is False

    def test_validate_memory_usage_negative_total(self, service):
        """Test negative total memory is invalid."""
        assert service.validate_memory_usage(4 * 1024**3, -1) is False

    def test_validate_memory_usage_none_total(self, service):
        """Test None total memory is invalid."""
        assert service.validate_memory_usage(4 * 1024**3, None) is False

    def test_validate_memory_usage_zero_total(self, service):
        """Test zero total memory with zero used is valid."""
        assert service.validate_memory_usage(0, 0) is True

    # Disk Usage Validation Tests
    def test_validate_disk_usage_valid(self, service):
        """Test valid disk usage."""
        assert service.validate_disk_usage(500 * 1024**3, 1024**4) is True

    def test_validate_disk_usage_zero(self, service):
        """Test zero disk used is valid."""
        assert service.validate_disk_usage(0, 1024**4) is True

    def test_validate_disk_usage_full(self, service):
        """Test full disk usage is valid."""
        assert service.validate_disk_usage(1024**4, 1024**4) is True

    def test_validate_disk_usage_exceeds_total(self, service):
        """Test disk used exceeding total is invalid."""
        assert service.validate_disk_usage(2 * 1024**4, 1024**4) is False

    def test_validate_disk_usage_negative_used(self, service):
        """Test negative disk used is invalid."""
        assert service.validate_disk_usage(-1, 1024**4) is False

    def test_validate_disk_usage_negative_total(self, service):
        """Test negative total disk is invalid."""
        assert service.validate_disk_usage(100, -1) is False

    # Network Bytes Validation Tests
    def test_validate_network_bytes_valid(self, service):
        """Test valid network bytes."""
        assert service.validate_network_bytes(1000000) is True

    def test_validate_network_bytes_zero(self, service):
        """Test zero network bytes is valid."""
        assert service.validate_network_bytes(0) is True

    def test_validate_network_bytes_negative(self, service):
        """Test negative network bytes is invalid."""
        assert service.validate_network_bytes(-1) is False

    def test_validate_network_bytes_large(self, service):
        """Test large network bytes is valid."""
        assert service.validate_network_bytes(10 * 1024**4) is True

    # Average Metrics Calculation Tests
    def test_calculate_average_metrics_empty(self, service):
        """Test average calculation with empty data."""
        result = service.calculate_average_metrics([])
        assert result == {"avg_cpu": 0.0, "avg_memory": 0.0}

    def test_calculate_average_metrics_single(self, service):
        """Test average calculation with single data point."""
        data = [{"cpu_usage": 50.0, "memory_used": 4 * 1024**3}]
        result = service.calculate_average_metrics(data)
        assert result["avg_cpu"] == 50.0
        assert result["avg_memory"] == 4 * 1024**3

    def test_calculate_average_metrics_multiple(self, service):
        """Test average calculation with multiple data points."""
        data = [
            {"cpu_usage": 40.0, "memory_used": 4 * 1024**3},
            {"cpu_usage": 60.0, "memory_used": 6 * 1024**3},
            {"cpu_usage": 50.0, "memory_used": 5 * 1024**3},
        ]
        result = service.calculate_average_metrics(data)
        assert result["avg_cpu"] == 50.0
        assert result["avg_memory"] == 5 * 1024**3

    def test_calculate_average_metrics_missing_fields(self, service):
        """Test average calculation with missing fields defaults to 0."""
        data = [
            {"cpu_usage": 50.0},
            {"memory_used": 5 * 1024**3},
            {},
        ]
        result = service.calculate_average_metrics(data)
        assert result["avg_cpu"] == pytest.approx(50.0 / 3)
        assert result["avg_memory"] == pytest.approx(5 * 1024**3 / 3)

    def test_calculate_average_metrics_all_zeros(self, service):
        """Test average calculation with all zero values."""
        data = [
            {"cpu_usage": 0, "memory_used": 0},
            {"cpu_usage": 0, "memory_used": 0},
        ]
        result = service.calculate_average_metrics(data)
        assert result["avg_cpu"] == 0.0
        assert result["avg_memory"] == 0.0
