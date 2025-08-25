"""
Real integration tests for logging configuration without mocks.

Tests the logging configuration with real file system operations,
real environment variables, and real component integration.

Requirements:
- REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
- REQ-HEALTH-002: System shall support structured logging for production environments
- REQ-HEALTH-003: System shall enable correlation ID tracking across components
- REQ-ERROR-006: System shall handle logging configuration failures gracefully

Story Coverage: S14 - Logging Configuration
IV&V Control Point: Real system integration validation
"""

import json
import logging
import os
import sys
import tempfile
from pathlib import Path

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


class TestLoggingConfigRealIntegration:
    """Real integration tests for logging configuration without mocks."""

    @pytest.fixture
    def temp_log_dir(self):
        """Create a temporary directory for log files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    @pytest.mark.unit
    def test_real_correlation_id_filter_thread_isolation(self):
        """Test correlation IDs are isolated between threads with real threading."""
        import threading
        import time

        correlation_filter = CorrelationIdFilter()
        results = {}

        def thread_func(thread_id, expected_id):
            correlation_filter.set_correlation_id(expected_id)
            # Small delay to ensure threads overlap
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

    @pytest.mark.unit
    def test_real_json_formatter_with_exception(self):
        """Test JSON formatter with real exception information."""
        formatter = JsonFormatter()

        try:
            raise ValueError("Test exception")
        except ValueError:
            exc_info = sys.exc_info()

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

    @pytest.mark.unit
    def test_real_setup_logging_development_mode(self, temp_log_dir):
        """Test logging setup in development mode with real file system."""
        log_file = temp_log_dir / "test.log"
        config = LoggingConfig(
            level="INFO",
            file_enabled=True,
            file_path=str(log_file),
            max_file_size="1MB",
            backup_count=3,
        )

        # Clear any existing handlers and filters
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)

        setup_logging(config, development_mode=True)

        # Should have console handler
        assert len(root_logger.handlers) >= 1

        # Should have correlation filter on handler
        handler = root_logger.handlers[0]
        correlation_filters = [
            f for f in handler.filters if isinstance(f, CorrelationIdFilter)
        ]
        assert len(correlation_filters) == 1

        # Should also have correlation filter on root logger
        root_correlation_filters = [
            f for f in root_logger.filters if isinstance(f, CorrelationIdFilter)
        ]
        assert len(root_correlation_filters) == 1

        # Test actual logging
        logger = logging.getLogger("test.module")
        logger.info("Test message")

        # Verify log file was created
        assert log_file.exists()

    @pytest.mark.unit
    def test_real_setup_logging_production_mode(self, temp_log_dir):
        """Test logging setup in production mode with real file system."""
        log_file = temp_log_dir / "test_prod.log"
        config = LoggingConfig(
            level="INFO",
            file_enabled=True,
            file_path=str(log_file),
            max_file_size="1MB",
            backup_count=3,
        )

        # Clear any existing handlers and filters
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)

        setup_logging(config, development_mode=False)

        # Should use JSON formatter in production mode
        handler = root_logger.handlers[0]
        assert isinstance(handler.formatter, JsonFormatter)
        
        # Should have correlation filter on root logger
        root_correlation_filters = [
            f for f in root_logger.filters if isinstance(f, CorrelationIdFilter)
        ]
        assert len(root_correlation_filters) == 1

        # Test actual logging
        logger = logging.getLogger("test.module")
        logger.info("Test production message")

        # Verify log file was created
        assert log_file.exists()

    @pytest.mark.unit
    def test_real_setup_logging_auto_mode_detection(self, temp_log_dir):
        """Test automatic development/production mode detection with real environment."""
        log_file = temp_log_dir / "test_auto.log"
        config = LoggingConfig(
            level="INFO",
            file_enabled=True,
            file_path=str(log_file),
            max_file_size="1MB",
            backup_count=3,
        )
        
        # Test development mode detection
        original_env = os.environ.copy()
        try:
            os.environ["CAMERA_SERVICE_ENV"] = "development"
            
            # Clear any existing handlers and filters
            root_logger = logging.getLogger()
            for handler in root_logger.handlers[:]:
                root_logger.removeHandler(handler)
            for filter_obj in root_logger.filters[:]:
                root_logger.removeFilter(filter_obj)

            setup_logging(config, development_mode=None)

            # Should detect development mode and use console formatter
            handler = root_logger.handlers[0]
            assert isinstance(handler.formatter, ConsoleFormatter)
            
            # Should have correlation filter on root logger
            root_correlation_filters = [
                f for f in root_logger.filters if isinstance(f, CorrelationIdFilter)
            ]
            assert len(root_correlation_filters) == 1
        finally:
            # Restore original environment
            os.environ.clear()
            os.environ.update(original_env)
        
        # Test production mode detection (default)
        original_env = os.environ.copy()
        try:
            os.environ["CAMERA_SERVICE_ENV"] = "production"
            
            # Clear any existing handlers and filters
            root_logger = logging.getLogger()
            for handler in root_logger.handlers[:]:
                root_logger.removeHandler(handler)
            for filter_obj in root_logger.filters[:]:
                root_logger.removeFilter(filter_obj)

            setup_logging(config, development_mode=None)

            # Should detect production mode and use JSON formatter
            handler = root_logger.handlers[0]
            assert isinstance(handler.formatter, JsonFormatter)
        finally:
            # Restore original environment
            os.environ.clear()
            os.environ.update(original_env)

    def test_real_setup_logging_with_rotation(self, temp_log_dir):
        """Test logging setup with file rotation enabled using real file system."""
        config = LoggingConfig(
            level="DEBUG",
            file_enabled=True,
            file_path=str(temp_log_dir / "test_rotation.log"),
            max_file_size="1MB",
            backup_count=3,
        )

        # Clear any existing handlers and filters
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)

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
        
        # Should have correlation filter on root logger
        root_correlation_filters = [
            f for f in root_logger.filters if isinstance(f, CorrelationIdFilter)
        ]
        assert len(root_correlation_filters) == 1

        # Test actual logging
        logger = logging.getLogger("test.module")
        logger.info("Test rotation message")

    def test_real_setup_logging_creates_log_directory(self, temp_log_dir):
        """Test that logging setup creates log directory if needed using real file system."""
        log_path = temp_log_dir / "logs" / "service.log"
        config = LoggingConfig(file_enabled=True, file_path=str(log_path))

        # Clear any existing handlers and filters
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)

        setup_logging(config)

        # Directory should be created
        assert log_path.parent.exists()

        # Test actual logging
        logger = logging.getLogger("test.module")
        logger.info("Test directory creation message")

        # File should be created
        assert log_path.exists()

    def test_real_setup_logging_rotation_fallback(self, temp_log_dir):
        """Test fallback to basic FileHandler when rotation config is invalid using real file system."""
        config = LoggingConfig(
            file_enabled=True,
            file_path=str(temp_log_dir / "test_fallback.log"),
            max_file_size="invalid_size",  # Invalid size should trigger fallback
            backup_count=3,
        )

        # Clear any existing handlers and filters
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)

        setup_logging(config, development_mode=True)

        # Should still have file handler (basic FileHandler as fallback)
        file_handlers = [
            h for h in root_logger.handlers if isinstance(h, logging.FileHandler)
        ]
        assert len(file_handlers) == 1

        # Test actual logging
        logger = logging.getLogger("test.module")
        logger.info("Test fallback message")

        # File should be created
        assert Path(config.file_path).exists()


class TestGlobalHelperFunctionsReal:
    """Real integration tests for global helper functions."""

    def test_real_get_correlation_filter_function(self):
        """Test get_correlation_filter function with real logging system."""
        # Clear existing logging configuration
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)
        
        # Setup real logging with correlation filter
        config = LoggingConfig(level="INFO", file_enabled=False)
        setup_logging(config, development_mode=True)
        
        # Test that we can retrieve the correlation filter
        result = get_correlation_filter()
        assert result is not None
        assert isinstance(result, CorrelationIdFilter)

    def test_real_get_correlation_filter_not_found(self):
        """Test get_correlation_filter when no filter exists."""
        # Clear existing logging configuration
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)
        
        # Test that no correlation filter is found when logging is not set up
        result = get_correlation_filter()
        assert result is None

    def test_real_set_correlation_id_function(self):
        """Test global set_correlation_id function with real system."""
        # Clear existing logging configuration
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)
        
        # Setup real logging with correlation filter
        config = LoggingConfig(level="INFO", file_enabled=False)
        setup_logging(config, development_mode=True)
        
        # Test setting correlation ID
        result = set_correlation_id("global-test-id")
        assert result is True
        
        # Verify correlation ID was set
        retrieved_id = get_correlation_id()
        assert retrieved_id == "global-test-id"

    def test_real_get_correlation_id_function(self):
        """Test global get_correlation_id function with real system."""
        # Clear existing logging configuration
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)
        
        # Setup real logging with correlation filter
        config = LoggingConfig(level="INFO", file_enabled=False)
        setup_logging(config, development_mode=True)
        
        # Test getting correlation ID when none is set
        result = get_correlation_id()
        assert result is None
        
        # Set correlation ID and test retrieval
        set_correlation_id("test-correlation-id")
        result = get_correlation_id()
        assert result == "test-correlation-id"


class TestLoggingIntegrationReal:
    """Real integration tests for logging functionality."""

    @pytest.fixture
    def temp_log_dir(self):
        """Create a temporary directory for log files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    def test_real_end_to_end_logging_flow(self, temp_log_dir):
        """Test complete logging flow with correlation IDs using real file system."""
        # Clear existing logging configuration
        root_logger = logging.getLogger()
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)
        
        # Setup logging in development mode with real file handler
        log_file = temp_log_dir / "test_flow.log"
        config = LoggingConfig(level="INFO", file_enabled=True, file_path=str(log_file))
        setup_logging(config, development_mode=True)
        
        # Set correlation ID
        set_correlation_id("test-flow-123")
        
        # Log a message
        logger = logging.getLogger("test.module")
        logger.info("Test message with correlation ID")
        
        # Verify log file was created and contains the message
        assert log_file.exists()
        
        with open(log_file, 'r') as f:
            log_content = f.read()
            assert "test-flow-123" in log_content
            assert "Test message with correlation ID" in log_content
        
        # Test production mode JSON output
        # Clear handlers and setup JSON logging
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        for filter_obj in root_logger.filters[:]:
            root_logger.removeFilter(filter_obj)
        
        # Setup production mode logging with real file handler
        json_log_file = temp_log_dir / "test_json.log"
        config.file_path = str(json_log_file)
        setup_logging(config, development_mode=False)
        
        set_correlation_id("test-json-456")
        logger.info("Test JSON message")
        
        # Verify JSON log file was created
        assert json_log_file.exists()
        
        with open(json_log_file, 'r') as f:
            json_log_content = f.read().strip()
            log_data = json.loads(json_log_content)
            assert log_data["correlation_id"] == "test-json-456"
            assert log_data["message"] == "Test JSON message"
            assert log_data["level"] == "INFO"
