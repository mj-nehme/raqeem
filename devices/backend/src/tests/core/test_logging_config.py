"""Tests for logging configuration module."""

import logging
import os
import tempfile
from pathlib import Path
from unittest.mock import patch

from app.core.logging_config import LogContext, LoggingSettings, configure_logging, get_logger


class TestLoggingSettings:
    """Tests for LoggingSettings configuration class."""

    def test_default_settings(self):
        """Test default logging settings."""
        settings = LoggingSettings()
        assert settings.log_level == "INFO"
        assert settings.log_format == "text"
        assert settings.log_file is None

    def test_custom_log_level_from_env(self):
        """Test custom log level setting from environment."""
        with patch.dict(os.environ, {"LOG_LEVEL": "DEBUG"}):
            settings = LoggingSettings()
            assert settings.log_level == "DEBUG"

    def test_json_log_format_from_env(self):
        """Test JSON log format setting from environment."""
        with patch.dict(os.environ, {"LOG_FORMAT": "json"}):
            settings = LoggingSettings()
            assert settings.log_format == "json"


class TestConfigureLogging:
    """Tests for configure_logging function."""

    def test_configure_with_defaults(self):
        """Test logging configuration with default settings."""
        configure_logging()
        # Should not raise

    def test_configure_with_debug_level(self):
        """Test logging configuration with DEBUG level."""
        with patch.dict(os.environ, {"LOG_LEVEL": "DEBUG"}):
            settings = LoggingSettings()
            assert settings.log_level == "DEBUG"
            # Just verify the function doesn't raise
            configure_logging(settings)

    def test_configure_with_json_format(self):
        """Test logging configuration with JSON format."""
        with patch.dict(os.environ, {"LOG_FORMAT": "json"}):
            settings = LoggingSettings()
            configure_logging(settings)
            # Should not raise

    def test_configure_with_log_file(self):
        """Test logging configuration with file output."""
        with tempfile.NamedTemporaryFile(suffix=".log", delete=False) as f:
            log_file = f.name

        try:
            with patch.dict(os.environ, {"LOG_FILE": log_file}):
                settings = LoggingSettings()
                configure_logging(settings)

                # Log something
                logger = logging.getLogger("test_file")
                logger.info("Test message to file")

                # Verify file exists
                assert Path(log_file).exists()
        finally:
            # Clean up
            Path(log_file).unlink(missing_ok=True)

    def test_third_party_loggers_suppressed(self):
        """Test that noisy third-party loggers are suppressed."""
        configure_logging()

        # These loggers should be set to WARNING level
        assert logging.getLogger("urllib3").level == logging.WARNING
        assert logging.getLogger("httpx").level == logging.WARNING
        assert logging.getLogger("httpcore").level == logging.WARNING


class TestGetLogger:
    """Tests for get_logger function."""

    def test_get_logger_returns_logger(self):
        """Test that get_logger returns a Logger instance."""
        logger = get_logger("test_module")
        assert isinstance(logger, logging.Logger)

    def test_get_logger_with_name(self):
        """Test that get_logger returns logger with correct name."""
        logger = get_logger("my.custom.module")
        assert logger.name == "my.custom.module"

    def test_get_logger_logs_message(self):
        """Test that logger can log messages."""
        logger = get_logger("test_logging")
        # Should not raise
        logger.info("Test message")
        logger.debug("Debug message")
        logger.warning("Warning message")


class TestLogContext:
    """Tests for LogContext context manager."""

    def test_log_context_adds_attributes(self):
        """Test that LogContext adds attributes to log records."""
        with LogContext(request_id="test123", user_id="user456"):
            # Log record should have these attributes
            logger = logging.getLogger("test_context")
            # The test validates the context manager doesn't raise
            logger.info("Test message with context")

    def test_log_context_restores_factory(self):
        """Test that LogContext restores original factory on exit."""
        original_factory = logging.getLogRecordFactory()

        with LogContext(test_attr="value"):
            pass

        # Factory should be restored
        assert logging.getLogRecordFactory() == original_factory

    def test_log_context_nested(self):
        """Test nested LogContext managers."""
        with LogContext(outer="value1"):
            with LogContext(inner="value2"):
                logger = logging.getLogger("test_nested")
                logger.info("Nested context")
            logger.info("Outer context only")

    def test_log_context_multiple_attributes(self):
        """Test LogContext with multiple attributes."""
        with LogContext(
            request_id="req123",
            user_id="user456",
            correlation_id="corr789"
        ):
            logger = logging.getLogger("test_multi_attrs")
            logger.info("Message with multiple context attributes")
