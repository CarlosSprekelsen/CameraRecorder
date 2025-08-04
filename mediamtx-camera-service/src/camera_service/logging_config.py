"""
Enhanced logging configuration with correlation IDs, structured formats,
and configurable rotation per architecture overview AD-8.

Provides JSON structured logging for production with correlation ID tracking,
human-readable console format for development, and configurable log rotation.
"""

import json
import logging
import logging.handlers
import os
import sys
import threading
import re
from pathlib import Path
from typing import Optional

from camera_service.config import LoggingConfig


class CorrelationIdFilter(logging.Filter):
    """
    Thread-local correlation ID filter for request tracking.

    Automatically adds correlation IDs to log records for distributed
    tracing and debugging support across request boundaries.
    """

    def __init__(self):
        super().__init__()
        self._local = threading.local()

    def filter(self, record: logging.LogRecord) -> bool:
        """Add correlation ID to log record if available."""
        correlation_id = self.get_correlation_id()
        if correlation_id:
            record.correlation_id = correlation_id
        return True

    def set_correlation_id(self, correlation_id: str) -> None:
        """Set correlation ID for the current thread."""
        self._local.correlation_id = correlation_id

    def get_correlation_id(self) -> Optional[str]:
        """Get correlation ID for the current thread."""
        return getattr(self._local, "correlation_id", None)


class JsonFormatter(logging.Formatter):
    """
    JSON formatter for structured logging in production environments.

    Outputs log records as JSON objects suitable for log aggregation
    systems and automated processing per architecture AD-8 specification.
    """

    def format(self, record: logging.LogRecord) -> str:
        """Format log record as JSON."""
        log_entry = {
            "timestamp": self.formatTime(record, self.datefmt),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }

        # Add correlation ID if present
        if hasattr(record, "correlation_id"):
            log_entry["correlation_id"] = record.correlation_id

        # Add exception info if present
        if record.exc_info:
            log_entry["exception"] = self.formatException(record.exc_info)

        # Add any extra fields
        for key, value in record.__dict__.items():
            if key not in [
                "name",
                "msg",
                "args",
                "levelname",
                "levelno",
                "pathname",
                "filename",
                "module",
                "lineno",
                "funcName",
                "created",
                "msecs",
                "relativeCreated",
                "thread",
                "threadName",
                "processName",
                "process",
                "exc_info",
                "exc_text",
                "stack_info",
                "correlation_id",
            ]:
                log_entry[key] = value

        return json.dumps(log_entry, default=str)


class ConsoleFormatter(logging.Formatter):
    """
    Human-readable formatter for development environments.

    Provides clear, readable log output suitable for console viewing
    during development and debugging per architecture AD-8 specification.
    """

    def format(self, record: logging.LogRecord) -> str:
        """Format log record for console display."""
        # Add correlation ID to the format if present
        correlation_part = ""
        if hasattr(record, "correlation_id"):
            correlation_part = f"[{record.correlation_id}] "

        # Format: timestamp - logger - level - [correlation_id] message
        formatted = super().format(record)
        if correlation_part:
            # Insert correlation ID after the level
            parts = formatted.split(" - ", 3)
            if len(parts) >= 4:
                formatted = f"{parts[0]} - {parts[1]} - {parts[2]} - {correlation_part}{parts[3]}"

        return formatted


def _parse_file_size(size_str: str) -> int:
    """
    Parse file size string to bytes.

    Args:
        size_str: Size string like '10MB', '500KB', '1GB'

    Returns:
        Size in bytes

    Raises:
        ValueError: If size string format is invalid
    """
    # TODO: MEDIUM: Add comprehensive size parsing validation [Story:S14]
    size_str = size_str.upper().strip()

    # Extract number and unit
    match = re.match(r"^(\d+(?:\.\d+)?)\s*([KMGT]?B?)$", size_str)
    if not match:
        raise ValueError(f"Invalid file size format: {size_str}")

    number, unit = match.groups()
    number = float(number)

    # Convert to bytes
    multipliers = {
        "B": 1,
        "KB": 1024,
        "MB": 1024**2,
        "GB": 1024**3,
        "TB": 1024**4,
        "": 1,  # No unit means bytes
    }

    return int(number * multipliers.get(unit, 1))


def setup_logging(config: LoggingConfig, development_mode: bool = None) -> None:
    """
    Initialize logging configuration for the camera service.

    Args:
        config: Logging configuration object containing all settings
        development_mode: If True, use console-friendly logging.
                         If False, use structured JSON logging.
                         If None, determine from environment or config.

    This function configures the root logger and sets up appropriate
    formatters, handlers, and filters based on the environment per AD-8.
    """
    # Determine logging mode
    if development_mode is None:
        development_mode = (
            os.getenv("CAMERA_SERVICE_ENV", "production").lower() == "development"
            or os.getenv("LOG_FORMAT", "json").lower() == "console"
        )

    # Clear any existing handlers
    root_logger = logging.getLogger()
    for handler in root_logger.handlers[:]:
        root_logger.removeHandler(handler)

    # Set logging level with graceful fallback
    try:
        level = getattr(logging, config.level.upper(), logging.INFO)
    except AttributeError:
        level = logging.INFO

    root_logger.setLevel(level)

    # Create correlation ID filter
    correlation_filter = CorrelationIdFilter()

    # Setup console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.addFilter(correlation_filter)

    if development_mode:
        # Development: Human-readable console format
        console_formatter = ConsoleFormatter(
            fmt=config.format, datefmt="%Y-%m-%d %H:%M:%S"
        )
    else:
        # Production: JSON format for structured logging
        console_formatter = JsonFormatter(datefmt="%Y-%m-%dT%H:%M:%S")

    console_handler.setFormatter(console_formatter)
    root_logger.addHandler(console_handler)

    # Setup file handler with rotation if enabled
    if config.file_enabled and config.file_path:
        # Ensure log directory exists
        log_path = Path(config.file_path)
        log_path.parent.mkdir(parents=True, exist_ok=True)

        # Create rotating file handler with configuration-based rotation
        try:
            max_bytes = _parse_file_size(config.max_file_size)
            file_handler = logging.handlers.RotatingFileHandler(
                config.file_path, maxBytes=max_bytes, backupCount=config.backup_count
            )
        except (ValueError, AttributeError) as e:
            # Fallback to basic file handler if rotation config is invalid
            # TODO: LOW: Log rotation configuration validation warning [Story:S14]
            file_handler = logging.FileHandler(config.file_path)

        file_handler.addFilter(correlation_filter)

        # Use JSON format for file logging in production, console for development
        if development_mode:
            file_formatter = ConsoleFormatter(
                fmt=config.format, datefmt="%Y-%m-%d %H:%M:%S"
            )
        else:
            file_formatter = JsonFormatter(datefmt="%Y-%m-%dT%H:%M:%S")

        file_handler.setFormatter(file_formatter)
        file_handler.setLevel(level)
        root_logger.addHandler(file_handler)


def get_correlation_filter() -> Optional[CorrelationIdFilter]:
    """
    Get the correlation ID filter from the root logger.

    Returns:
        The CorrelationIdFilter instance if found, None otherwise.

    This function allows other parts of the application to access
    the correlation filter for setting request-specific correlation IDs.
    """
    root_logger = logging.getLogger()
    for handler in root_logger.handlers:
        for filter_obj in handler.filters:
            if isinstance(filter_obj, CorrelationIdFilter):
                return filter_obj
    return None


def set_correlation_id(correlation_id: str) -> bool:
    """
    Set correlation ID for the current thread across all handlers.

    Args:
        correlation_id: Correlation ID to set for current thread

    Returns:
        True if correlation filter was found and ID was set, False otherwise
    """
    correlation_filter = get_correlation_filter()
    if correlation_filter:
        correlation_filter.set_correlation_id(correlation_id)
        return True
    return False


def get_correlation_id() -> Optional[str]:
    """
    Get the current thread's correlation ID.

    Returns:
        The correlation ID for the current thread, or None if not set
    """
    correlation_filter = get_correlation_filter()
    if correlation_filter:
        return correlation_filter.get_correlation_id()
    return None
