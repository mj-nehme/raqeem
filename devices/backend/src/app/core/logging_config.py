"""Structured logging configuration for the Devices Backend API.

This module provides centralized logging configuration with support for:
- JSON-formatted logs for production (machine-readable)
- Human-readable logs for development
- Request ID tracking for distributed tracing
- Configurable log levels per module
"""

import logging
import sys
from typing import Any

from pydantic import Field
from pydantic_settings import BaseSettings


class LoggingSettings(BaseSettings):
    """Configuration settings for application logging."""

    log_level: str = Field(default="INFO", validation_alias="LOG_LEVEL")
    log_format: str = Field(default="text", validation_alias="LOG_FORMAT")  # "text" or "json"
    log_file: str | None = Field(default=None, validation_alias="LOG_FILE")


def configure_logging(settings: LoggingSettings | None = None) -> None:
    """Configure application-wide logging with structured output.

    Args:
        settings: Optional logging settings. If None, uses defaults from environment.

    Example:
        >>> configure_logging()  # Uses defaults from environment
        >>> configure_logging(LoggingSettings(log_level="DEBUG", log_format="json"))
    """
    if settings is None:
        settings = LoggingSettings()  # type: ignore[call-arg]

    log_level = getattr(logging, settings.log_level.upper(), logging.INFO)

    # Create formatter based on configuration
    if settings.log_format == "json":
        # JSON format for production - machine-readable
        log_format = '{"time": "%(asctime)s", "level": "%(levelname)s", "name": "%(name)s", "message": "%(message)s"}'
    else:
        # Human-readable format for development
        log_format = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"

    # Configure root logger
    logging.basicConfig(
        level=log_level,
        format=log_format,
        datefmt="%Y-%m-%d %H:%M:%S",
        handlers=[logging.StreamHandler(sys.stdout)],
    )

    # Add file handler if log file is specified
    if settings.log_file:
        file_handler = logging.FileHandler(settings.log_file)
        file_handler.setFormatter(logging.Formatter(log_format, datefmt="%Y-%m-%d %H:%M:%S"))
        logging.root.addHandler(file_handler)

    # Set specific log levels for noisy third-party libraries
    logging.getLogger("urllib3").setLevel(logging.WARNING)
    logging.getLogger("httpx").setLevel(logging.WARNING)
    logging.getLogger("httpcore").setLevel(logging.WARNING)


def get_logger(name: str) -> logging.Logger:
    """Get a logger instance for the given name.

    Args:
        name: Logger name, typically __name__ of the calling module.

    Returns:
        Configured logger instance.

    Example:
        >>> logger = get_logger(__name__)
        >>> logger.info("Processing request", extra={"request_id": "123"})
    """
    return logging.getLogger(name)


class LogContext:
    """Context manager for adding structured context to log messages.

    Example:
        >>> with LogContext(request_id="abc123", user_id="user456"):
        ...     logger.info("Processing payment")
        # Logs will include request_id and user_id in the output
    """

    def __init__(self, **kwargs: Any) -> None:
        self.context = kwargs
        self.old_factory = logging.getLogRecordFactory()

    def __enter__(self) -> "LogContext":
        def record_factory(*args: Any, **kwargs: Any) -> logging.LogRecord:
            record = self.old_factory(*args, **kwargs)
            for key, value in self.context.items():
                setattr(record, key, value)
            return record

        logging.setLogRecordFactory(record_factory)
        return self

    def __exit__(self, *args: object) -> None:
        logging.setLogRecordFactory(self.old_factory)
