import os
import tempfile
import pytest
from pathlib import Path

def pytest_sessionstart(session):
    """Set up test environment variables and ensure clean state."""
    # Provide a deterministic secret for JWT tests
    os.environ.setdefault("CAMERA_SERVICE_JWT_SECRET", "test-secret-key")
    os.environ.setdefault("CAMERA_SERVICE_RATE_RPM", "1000")
    
    # Set test-specific environment variables
    os.environ.setdefault("CAMERA_SERVICE_TEST_MODE", "true")
    os.environ.setdefault("CAMERA_SERVICE_DISABLE_HARDWARE", "true")
    
    # Ensure test directories exist
    test_dirs = ["/tmp/test_recordings", "/tmp/test_snapshots", "/tmp/test_logs"]
    for test_dir in test_dirs:
        Path(test_dir).mkdir(parents=True, exist_ok=True)

def pytest_sessionfinish(session, exitstatus):
    """Clean up test environment after session."""
    # Clean up test files if needed
    test_dirs = ["/tmp/test_recordings", "/tmp/test_snapshots", "/tmp/test_logs"]
    for test_dir in test_dirs:
        if Path(test_dir).exists():
            try:
                for file in Path(test_dir).glob("*"):
                    if file.is_file():
                        file.unlink()
            except Exception:
                pass  # Ignore cleanup errors

# No-mock guard for FORBID_MOCKS=1
def pytest_configure(config):
    """Configure pytest with no-mock guard if FORBID_MOCKS=1 is set."""
    if os.environ.get("FORBID_MOCKS") == "1":
        # Monkey patch unittest.mock to raise on import
        import sys
        
        class MockForbiddenError(Exception):
            """Raised when mocks are forbidden but attempted to be used."""
            pass
        
        def forbidden_mock(*args, **kwargs):
            raise MockForbiddenError(
                "Mock usage forbidden when FORBID_MOCKS=1. "
                "Implement real async context manager behavior instead."
            )
        
        # Replace mock classes with forbidden versions
        sys.modules['unittest.mock'] = type('MockModule', (), {
            'Mock': forbidden_mock,
            'MagicMock': forbidden_mock,
            'AsyncMock': forbidden_mock,
            'patch': forbidden_mock,
            'MockForbiddenError': MockForbiddenError,
        })
        
        # Also block pytest-mock if installed
        if 'pytest_mock' in sys.modules:
            raise MockForbiddenError(
                "pytest-mock plugin is loaded but FORBID_MOCKS=1. "
                "Remove pytest-mock from test environment."
            )

@pytest.fixture(scope="session")
def test_environment():
    """Provide a consistent test environment configuration."""
    return {
        "host": "127.0.0.1",  # Use IP instead of localhost
        "api_port": 9997,
        "rtsp_port": 8554,
        "webrtc_port": 8889,
        "hls_port": 8888,
        "websocket_port": 8002,
        "health_port": 8003,
        "test_config_path": "/tmp/test_config.yml",
        "test_recordings_path": "/tmp/test_recordings",
        "test_snapshots_path": "/tmp/test_snapshots",
        "test_logs_path": "/tmp/test_logs",
        "test_device_range": [0, 1, 2],
        "test_jwt_secret": "test-secret-key",
        "test_rate_limit": 1000,
    }

@pytest.fixture(scope="session")
def temp_test_dir():
    """Create a temporary test directory that's cleaned up automatically."""
    with tempfile.TemporaryDirectory() as temp_dir:
        yield temp_dir

@pytest.fixture
def mock_device_paths():
    """Provide mock device paths that work across environments."""
    return {
        "video0": "/dev/video0",
        "video1": "/dev/video1", 
        "video2": "/dev/video2",
        "nonexistent": "/dev/video999",
    }

@pytest.fixture
def mock_stream_urls():
    """Provide mock stream URLs using IP instead of localhost."""
    return {
        "rtsp": "rtsp://127.0.0.1:8554/test_stream",
        "webrtc": "http://127.0.0.1:8889/test_stream",
        "hls": "http://127.0.0.1:8888/test_stream",
    }

