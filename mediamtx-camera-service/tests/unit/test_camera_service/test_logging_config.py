"""
Test suite for logging configuration module.

Tests correlation ID propagation, formatter behavior, rotation functionality,
and environment-based mode detection per Story S14.
"""

import json
import logging
import os
import tempfile
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from camera_service.config import LoggingConfig
from camera_service.logging_config import (
    CorrelationIdFilter,
    JsonFormatter,
    ConsoleFormatter,
    setup_logging,
    get_correlation_filter,
    set_correlation_id,
    get_correlation_id,
    _parse_file_size,
)


class TestCorrelationIdFilter:
    """Test CorrelationIdFilter for thread-local correlation tracking."""

    def test_correlation_filter_basic_functionality(self):
        """Test correlation filter sets and retrieves correlation IDs."""
        # TODO: HIGH: Test correlation filter basic operations [Story:S14]
        correlation_filter = CorrelationIdFilter()

        # Initially no correlation ID
        assert correlation_filter.get_correlation_id() is None

        # Set correlation ID
        correlation_filter.set_correlation_id("test-123")
        assert correlation_filter.get_correlation_id() == "test-123"

        # Filter should add correlation ID to record
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Test message",
            args=(),
            exc_info=None,
        )

        assert correlation_filter.filter(record) is True
        assert hasattr(record, "correlation_id")
        assert record.correlation_id == "test-123"

    def test_correlation_filter_thread_isolation(self):
        """Test correlation IDs are isolated between threads."""
        # TODO: MEDIUM: Test thread isolation for correlation IDs [Story:S14]
        import threading

        correlation_filter = CorrelationIdFilter()
        results = {}

        def thread_func(thread_id, expected_id):
            correlation_filter.set_correlation_id(expected_id)
            # Small delay to ensure threads overlap
            import time

            time.sleep(0.01)
            results[thread_id] = correlation_filter.get_correlation_id()

        # Start multiple threads with different correlation IDs
        threads = []
        for i in range(3):
            thread = threading.Thread(target=thread_func, args=(i, f"thread-{i}"))
            threads.append(thread)
            thread.start()

        # Wait for all threads to complete
        for thread in threads:
            thread.join()

        # Each thread should have its own correlation ID
        assert results[0] == "thread-0"
        assert results[1] == "thread-1"
        assert results[2] == "thread-2"


class TestJsonFormatter:
    """Test JsonFormatter for structured production logging."""

    def test_json_formatter_basic_record(self):
        """Test JSON formatter with basic log record."""
        # TODO: HIGH: Test JSON formatter basic output structure [Story:S14]
        formatter = JsonFormatter()
        record = logging.LogRecord(
            name="test.module",
            level=logging.INFO,
            pathname="test.py",
            lineno=42,
            msg="Test message",
            args=(),
            exc_info=None,
        )
        record.module = "test_module"
        record.funcName = "test_function"

        result = formatter.format(record)
        log_data = json.loads(result)

        assert log_data["level"] == "INFO"
        assert log_data["logger"] == "test.module"
        assert log_data["message"] == "Test message"
        assert log_data["module"] == "test_module"
        assert log_data["function"] == "test_function"
        assert log_data["line"] == 42
        assert "timestamp" in log_data

    def test_json_formatter_with_correlation_id(self):
        """Test JSON formatter includes correlation ID."""
        # TODO: HIGH: Test JSON formatter correlation ID inclusion [Story:S14]
        formatter = JsonFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Test message",
            args=(),
            exc_info=None,
        )
        record.correlation_id = "test-correlation-123"

        result = formatter.format(record)
        log_data = json.loads(result)

        assert log_data["correlation_id"] == "test-correlation-123"

    def test_json_formatter_with_exception(self):
        """Test JSON formatter with exception information."""
        # TODO: HIGH: Test JSON formatter exception handling [Story:S14]
        formatter = JsonFormatter()

        try:
            raise ValueError("Test exception")
        except ValueError:
            exc_info = True

        record = logging.LogRecord(
            name="test",
            level=logging.ERROR,
            pathname="",
            lineno=0,
            msg="Error occurred",
            args=(),
            exc_info=exc_info,
        )

        result = formatter.format(record)
        log_data = json.loads(result)

        assert "exception" in log_data
        assert "Test exception" in log_data["exception"]

    def test_json_formatter_extra_fields(self):
        """Test JSON formatter includes extra fields from log record."""
        # TODO: MEDIUM: Test JSON formatter extra fields handling [Story:S14]
        formatter = JsonFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Test message",
            args=(),
            exc_info=None,
        )
        record.custom_field = "custom_value"
        record.user_id = 12345

        result = formatter.format(record)
        log_data = json.loads(result)

        assert log_data["custom_field"] == "custom_value"
        assert log_data["user_id"] == 12345


class TestConsoleFormatter:
    """Test ConsoleFormatter for development logging."""

    def test_console_formatter_basic_record(self):
        """Test console formatter with basic log record."""
        # TODO: HIGH: Test console formatter readable output [Story:S14]
        formatter = ConsoleFormatter("%(levelname)s - %(name)s - %(message)s")
        record = logging.LogRecord(
            name="test.module",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Test message",
            args=(),
            exc_info=None,
        )

        result = formatter.format(record)

        assert "INFO" in result
        assert "test.module" in result
        assert "Test message" in result

    def test_console_formatter_with_correlation_id(self):
        """Test console formatter includes correlation ID."""
        # TODO: HIGH: Test console formatter correlation ID display [Story:S14]
        formatter = ConsoleFormatter("%(levelname)s - %(name)s - %(message)s")
        record = logging.LogRecord(
            name="test.module",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Test message",
            args=(),
            exc_info=None,
        )
        record.correlation_id = "test-456"

        result = formatter.format(record)

        assert "[test-456]" in result
        assert "Test message" in result


class TestFileSizeParsing:
    """Test file size parsing utility function."""

    def test_parse_file_size_basic_units(self):
        """Test parsing of standard file size units."""
        # TODO: MEDIUM: Test file size parsing accuracy [Story:S14]
        assert _parse_file_size("1024B") == 1024
        assert _parse_file_size("1KB") == 1024
        assert _parse_file_size("1MB") == 1024**2
        assert _parse_file_size("1GB") == 1024**3
        assert _parse_file_size("10MB") == 10 * 1024**2

    def test_parse_file_size_invalid_formats(self):
        """Test file size parsing with invalid formats."""
        # TODO: MEDIUM: Test file size parsing error handling [Story:S14]
        with pytest.raises(ValueError):
            _parse_file_size("invalid")

        with pytest.raises(ValueError):
            _parse_file_size("10XB")


class TestSetupLogging:
    """Test setup_logging function configuration."""

    @pytest.fixture
    def logging_config(self):
        """Create a LoggingConfig for testing."""
        return LoggingConfig(
            level="INFO",
            format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
            file_enabled=False,
            file_path="/tmp/test.log",
            max_file_size="1MB",
            backup_count=3,
        )

    def test_setup_logging_development_mode(self, logging_config):
        """Test logging setup in development mode."""
        # TODO: HIGH: Test development mode console formatter usage [Story:S14]
        # Clear any existing handlers
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)

        setup_logging(logging_config, development_mode=True)

        # Should have console handler
        assert len(root_logger.handlers) >= 1

        # Should have correlation filter
        handler = root_logger.handlers[0]
        correlation_filters = [
            f for f in handler.filters if isinstance(f, CorrelationIdFilter)
        ]
        assert len(correlation_filters) == 1

        # Should use console formatter in development mode
        assert isinstance(handler.formatter, ConsoleFormatter)

    def test_setup_logging_production_mode(self, logging_config):
        """Test logging setup in production mode."""
        # TODO: HIGH: Test production mode JSON formatter usage [Story:S14]
        # Clear any existing handlers
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)

        setup_logging(logging_config, development_mode=False)

        # Should use JSON formatter in production mode
        handler = root_logger.handlers[0]
        assert isinstance(handler.formatter, JsonFormatter)

    def test_setup_logging_auto_mode_detection(self, logging_config):
        """Test automatic development/production mode detection."""
        # TODO: MEDIUM: Test environment-based mode detection [Story:S14]
        with patch.dict(os.environ, {"CAMERA_SERVICE_ENV": "development"}):
            # Clear any existing handlers
            root_logger = logging.getLogger()
            for handler in root_logger.handlers[:]:
                root_logger.removeHandler(handler)

            setup_logging(logging_config, development_mode=None)

            # Should detect development mode and use console formatter
            handler = root_logger.handlers[0]
            assert isinstance(handler.formatter, ConsoleFormatter)

    def test_setup_logging_with_rotation(self):
        """Test logging setup with file rotation enabled."""
        # TODO: HIGH: Test log rotation handler creation [Story:S14]
        config = LoggingConfig(
            level="DEBUG",
            file_enabled=True,
            file_path="/tmp/test_rotation.log",
            max_file_size="1MB",
            backup_count=3,
        )

        with tempfile.TemporaryDirectory() as temp_dir:
            config.file_path = str(Path(temp_dir) / "test.log")

            # Clear any existing handlers
            root_logger = logging.getLogger()
            for handler in root_logger.handlers[:]:
                root_logger.removeHandler(handler)

            setup_logging(config, development_mode=True)

            # Should have both console and file handlers
            assert len(root_logger.handlers) >= 2

            # File handler should be RotatingFileHandler
            file_handlers = [
                h
                for h in root_logger.handlers
                if isinstance(h, logging.handlers.RotatingFileHandler)
            ]
            assert len(file_handlers) == 1

            # File should be created
            assert Path(config.file_path).exists()

    def test_setup_logging_level_configuration(self, logging_config):
        """Test logging level is set correctly."""
        # TODO: MEDIUM: Test logging level configuration [Story:S14]
        logging_config.level = "DEBUG"

        # Clear any existing handlers
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)

        setup_logging(logging_config)

        assert root_logger.level == logging.DEBUG

    def test_setup_logging_creates_log_directory(self):
        """Test that logging setup creates log directory if needed."""
        # TODO: MEDIUM: Test log directory creation [Story:S14]
        with tempfile.TemporaryDirectory() as temp_dir:
            log_path = Path(temp_dir) / "logs" / "service.log"
            config = LoggingConfig(file_enabled=True, file_path=str(log_path))

            # Clear any existing handlers
            root_logger = logging.getLogger()
            for handler in root_logger.handlers[:]:
                root_logger.removeHandler(handler)

            setup_logging(config)

            # Directory should be created
            assert log_path.parent.exists()

    def test_setup_logging_rotation_fallback(self):
        """Test fallback to basic FileHandler when rotation config is invalid."""
        # TODO: MEDIUM: Test rotation configuration fallback [Story:S14]
        config = LoggingConfig(
            file_enabled=True,
            file_path="/tmp/test_fallback.log",
            max_file_size="invalid_size",  # Invalid size should trigger fallback
            backup_count=3,
        )

        with tempfile.TemporaryDirectory() as temp_dir:
            config.file_path = str(Path(temp_dir) / "test.log")

            # Clear any existing handlers
            root_logger = logging.getLogger()
            for handler in root_logger.handlers[:]:
                root_logger.removeHandler(handler)

            setup_logging(config, development_mode=True)

            # Should still have file handler (basic FileHandler as fallback)
            file_handlers = [
                h for h in root_logger.handlers if isinstance(h, logging.FileHandler)
            ]
            assert len(file_handlers) == 1


class TestGlobalHelperFunctions:
    """Test global helper functions for correlation ID management."""

    def test_get_correlation_filter_function(self):
        """Test get_correlation_filter function."""
        # TODO: MEDIUM: Test correlation filter retrieval [Story:S14]
        with patch("logging.getLogger") as mock_get_logger:
            mock_logger = Mock()
            mock_handler = Mock()
            mock_filter = CorrelationIdFilter()
            mock_handler.filters = [mock_filter]
            mock_logger.handlers = [mock_handler]
            mock_get_logger.return_value = mock_logger

            result = get_correlation_filter()

            assert result is mock_filter

    def test_get_correlation_filter_not_found(self):
        """Test get_correlation_filter when no filter exists."""
        # TODO: MEDIUM: Test correlation filter not found case [Story:S14]
        with patch("logging.getLogger") as mock_get_logger:
            mock_logger = Mock()
            mock_logger.handlers = []
            mock_get_logger.return_value = mock_logger

            result = get_correlation_filter()

            assert result is None

    def test_set_correlation_id_function(self):
        """Test global set_correlation_id function."""
        # TODO: MEDIUM: Test global correlation ID setting [Story:S14]
        with patch(
            "camera_service.logging_config.get_correlation_filter"
        ) as mock_get_filter:
            mock_filter = Mock()
            mock_get_filter.return_value = mock_filter

            result = set_correlation_id("global-test-id")

            mock_filter.set_correlation_id.assert_called_once_with("global-test-id")
            assert result is True

    def test_get_correlation_id_function(self):
        """Test global get_correlation_id function."""
        # TODO: MEDIUM: Test global correlation ID retrieval [Story:S14]
        with patch(
            "camera_service.logging_config.get_correlation_filter"
        ) as mock_get_filter:
            mock_filter = Mock()
            mock_filter.get_correlation_id.return_value = "current-id"
            mock_get_filter.return_value = mock_filter

            result = get_correlation_id()

            assert result == "current-id"


class TestLoggingIntegration:
    """Integration tests for logging functionality."""

    def test_end_to_end_logging_flow(self):
        """Test complete logging flow with correlation IDs."""
        # TODO: LOW: Test end-to-end logging integration [Story:S14]
        # This would test the complete flow:
        # 1. Setup logging
        # 2. Set correlation ID
        # 3. Log messages
        # 4. Verify output format includes correlation ID
        # 5. Verify structured/console format as appropriate

        # Implementation deferred until integration test phase
        pass
