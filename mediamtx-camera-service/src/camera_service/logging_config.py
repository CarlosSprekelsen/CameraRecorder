"""
Logging configuration for the MediaMTX Camera Service.

Provides structured JSON logging for production and human-readable console
logging for development environments. Supports correlation IDs and configurable
log levels as required by the architecture specification.
"""

import json
import logging
import logging.handlers
import os
import sys
import threading
import uuid
from pathlib import Path
from typing import Optional, Dict, Any

from .config import LoggingConfig


class CorrelationIdFilter(logging.Filter):
    """
    Logging filter that adds correlation IDs to log records.
    
    Correlation IDs help trace related log entries across the service
    for debugging and monitoring purposes.
    """
    
    def __init__(self):
        super().__init__()
        self._local = threading.local()
    
    def filter(self, record: logging.LogRecord) -> bool:
        """Add correlation ID to the log record."""
        # TODO: Implement full correlation ID management with request context
        # For now, generate a simple correlation ID per thread
        if not hasattr(self._local, 'correlation_id'):
            self._local.correlation_id = str(uuid.uuid4())[:8]
        
        record.correlation_id = self._local.correlation_id
        return True
    
    def set_correlation_id(self, correlation_id: str) -> None:
        """Set correlation ID for the current thread."""
        self._local.correlation_id = correlation_id
    
    def get_correlation_id(self) -> Optional[str]:
        """Get correlation ID for the current thread."""
        return getattr(self._local, 'correlation_id', None)


class JsonFormatter(logging.Formatter):
    """
    JSON formatter for structured logging in production environments.
    
    Outputs log records as JSON objects suitable for log aggregation
    systems and automated processing.
    """
    
    def format(self, record: logging.LogRecord) -> str:
        """Format log record as JSON."""
        log_entry = {
            'timestamp': self.formatTime(record, self.datefmt),
            'level': record.levelname,
            'logger': record.name,
            'message': record.getMessage(),
            'module': record.module,
            'function': record.funcName,
            'line': record.lineno
        }
        
        # Add correlation ID if present
        if hasattr(record, 'correlation_id'):
            log_entry['correlation_id'] = record.correlation_id
        
        # Add exception info if present
        if record.exc_info:
            log_entry['exception'] = self.formatException(record.exc_info)
        
        # Add any extra fields
        for key, value in record.__dict__.items():
            if key not in ['name', 'msg', 'args', 'levelname', 'levelno', 
                          'pathname', 'filename', 'module', 'lineno', 
                          'funcName', 'created', 'msecs', 'relativeCreated',
                          'thread', 'threadName', 'processName', 'process',
                          'exc_info', 'exc_text', 'stack_info', 'correlation_id']:
                log_entry[key] = value
        
        return json.dumps(log_entry, default=str)


class ConsoleFormatter(logging.Formatter):
    """
    Human-readable formatter for development environments.
    
    Provides clear, readable log output suitable for console viewing
    during development and debugging.
    """
    
    def format(self, record: logging.LogRecord) -> str:
        """Format log record for console display."""
        # Add correlation ID to the format if present
        correlation_part = ""
        if hasattr(record, 'correlation_id'):
            correlation_part = f"[{record.correlation_id}] "
        
        # Format: timestamp - logger - level - [correlation_id] message
        formatted = super().format(record)
        if correlation_part:
            # Insert correlation ID after the level
            parts = formatted.split(' - ', 3)
            if len(parts) >= 4:
                formatted = f"{parts[0]} - {parts[1]} - {parts[2]} - {correlation_part}{parts[3]}"
        
        return formatted


def setup_logging(config: LoggingConfig, development_mode: bool = None) -> None:
    """
    Initialize logging configuration for the camera service.
    
    Args:
        config: Logging configuration object containing all settings
        development_mode: If True, use console-friendly logging.
                         If False, use structured JSON logging.
                         If None, determine from environment or config.
    
    This function configures the root logger and sets up appropriate
    formatters, handlers, and filters based on the environment.
    """
    # Determine logging mode
    if development_mode is None:
        development_mode = (
            os.getenv('CAMERA_SERVICE_ENV', 'production').lower() == 'development'
            or os.getenv('LOG_FORMAT', 'json').lower() == 'console'
        )
    
    # Clear any existing handlers
    root_logger = logging.getLogger()
    for handler in root_logger.handlers[:]:
        root_logger.removeHandler(handler)
    
    # Set logging level
    level = getattr(logging, config.level.upper(), logging.INFO)
    root_logger.setLevel(level)
    
    # Create correlation ID filter
    correlation_filter = CorrelationIdFilter()
    
    # Setup console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.addFilter(correlation_filter)
    
    if development_mode:
        # Development: Human-readable console format
        console_formatter = ConsoleFormatter(
            fmt=config.format,
            datefmt='%Y-%m-%d %H:%M:%S'
        )
    else:
        # Production: JSON format for structured logging
        console_formatter = JsonFormatter(datefmt='%Y-%m-%dT%H:%M:%S')
    
    console_handler.setFormatter(console_formatter)
    root_logger.addHandler(console_handler)
    
    # Setup file handler if enabled
    if config.file_enabled and config.file_path:
        # Ensure log directory exists
        log_path = Path(config.file_path)
        log_path.parent.mkdir(parents=True, exist_ok=True)
        
        # Create rotating file handler
        # TODO: Implement log rotation based on config.max_file_size and config.backup_count
        file_handler = logging.FileHandler(config.file_path)
        file_handler.addFilter(correlation_filter)
        
        # Always use JSON format for file logging in production
        if development_mode:
            file_formatter = ConsoleFormatter(
                fmt=config.format,
                datefmt='%Y-%m-%d %H:%M:%S'
            )
        else:
            file_formatter = JsonFormatter(datefmt='%Y-%m-%dT%H:%M:%S')
        
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


def set_correlation_id(correlation_id: str) -> None:
    """
    Set correlation ID for the current thread/request.
    
    Args:
        correlation_id: Unique identifier for correlating related log entries
        
    This function should be called at the beginning of request processing
    to ensure all log entries for that request share the same correlation ID.
    """
    correlation_filter = get_correlation_filter()
    if correlation_filter:
        correlation_filter.set_correlation_id(correlation_id)


def get_correlation_id() -> Optional[str]:
    """
    Get the current correlation ID for the thread/request.
    
    Returns:
        Current correlation ID if set, None otherwise.
    """
    correlation_filter = get_correlation_filter()
    if correlation_filter:
        return correlation_filter.get_correlation_id()
    return None