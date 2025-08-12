# Test Environment Setup and Configuration

**Document Type:** Development Guide  
**Version:** 1.0  
**Last Updated:** 2025-01-27  
**Status:** Active

## Overview

This document describes the test environment setup, configuration, and best practices for the MediaMTX Camera Service. The test environment is designed to be consistent, isolated, and reproducible across different development environments.

## Test Environment Architecture

### Directory Structure

```
tests/
├── conftest.py                    # Global test configuration
├── unit/                          # Unit tests
│   ├── conftest.py               # Unit test fixtures
│   ├── test_camera_service/      # Camera service tests
│   ├── test_mediamtx_wrapper/    # MediaMTX wrapper tests
│   ├── test_websocket_server/    # WebSocket server tests
│   └── test_camera_discovery/    # Camera discovery tests
├── integration/                   # Integration tests
├── ivv/                          # IV&V tests
├── security/                     # Security tests
├── installation/                 # Installation tests
├── production/                   # Production validation tests
└── performance/                  # Performance tests
```

### Test Categories

| Category | Purpose | Scope | Dependencies |
|----------|---------|-------|--------------|
| **Unit Tests** | Test individual components | Isolated | Minimal |
| **Integration Tests** | Test component interactions | Service-level | Mocked external services |
| **IV&V Tests** | Independent validation | System-level | Full environment |
| **Security Tests** | Security validation | Security-focused | Authentication/authorization |
| **Installation Tests** | Installation validation | Deployment | System packages |
| **Production Tests** | Production readiness | Production-like | Full stack |

## Environment Configuration

### Environment Variables

The test environment uses the following environment variables:

| Variable | Purpose | Default | Notes |
|----------|---------|---------|-------|
| `CAMERA_SERVICE_JWT_SECRET` | JWT signing secret | `test-secret-key` | Deterministic for tests |
| `CAMERA_SERVICE_RATE_RPM` | Rate limiting | `1000` | High limit for tests |
| `CAMERA_SERVICE_TEST_MODE` | Test mode flag | `true` | Enables test optimizations |
| `CAMERA_SERVICE_DISABLE_HARDWARE` | Disable hardware access | `true` | Prevents hardware dependencies |

### Test Directories

| Directory | Purpose | Cleanup |
|-----------|---------|---------|
| `/tmp/test_recordings` | Test recording files | Automatic |
| `/tmp/test_snapshots` | Test snapshot files | Automatic |
| `/tmp/test_logs` | Test log files | Automatic |
| `/tmp/test_config.yml` | Test configuration | Automatic |

## Fixtures and Configuration

### Global Fixtures

**`test_environment`** - Provides consistent test environment configuration:
```python
@pytest.fixture(scope="session")
def test_environment():
    return {
        "host": "127.0.0.1",  # Use IP instead of localhost
        "api_port": 9997,
        "rtsp_port": 8554,
        "webrtc_port": 8889,
        "hls_port": 8888,
        "websocket_port": 8002,
        "health_port": 8003,
        # ... other configuration
    }
```

**`temp_test_dir`** - Creates temporary test directories:
```python
@pytest.fixture(scope="session")
def temp_test_dir():
    with tempfile.TemporaryDirectory() as temp_dir:
        yield temp_dir
```

**`mock_device_paths`** - Provides mock device paths:
```python
@pytest.fixture
def mock_device_paths():
    return {
        "video0": "/dev/video0",
        "video1": "/dev/video1", 
        "video2": "/dev/video2",
        "nonexistent": "/dev/video999",
    }
```

### Unit Test Fixtures

**`test_controller_config`** - MediaMTX controller configuration:
```python
@pytest.fixture
def test_controller_config():
    return {
        "host": "127.0.0.1",
        "api_port": 9997,
        # ... other configuration
    }
```

**`temp_test_files`** - Temporary test files:
```python
@pytest.fixture
def temp_test_files():
    with tempfile.TemporaryDirectory() as temp_dir:
        # Create test files
        yield {
            "temp_dir": temp_dir,
            "config_path": str(config_file),
            "recordings_path": str(recordings_dir),
            "snapshots_path": str(snapshots_dir),
        }
```

## Mocking Strategy

### Hardware Dependencies

**Camera Devices** - Mocked using `mock_device_paths` fixture:
```python
def test_camera_operations(mock_device_paths):
    device_path = mock_device_paths["video0"]
    # Test with mocked device path
```

**Udev Events** - Mocked using `mock_udev_device` fixture:
```python
def test_udev_processing(mock_udev_device):
    device = mock_udev_device(device_node="/dev/video0", action="add")
    # Test with mocked udev device
```

**V4L2 Commands** - Mocked using `mock_v4l2_outputs` fixture:
```python
def test_capability_detection(mock_v4l2_outputs):
    outputs = mock_v4l2_outputs
    # Test with mocked v4l2-ctl outputs
```

### Network Dependencies

**Localhost References** - Replaced with IP addresses:
```python
# Before
host = "localhost"

# After
host = "127.0.0.1"
```

**Port Configuration** - Consistent port assignments:
```python
test_ports = {
    "api": 9997,
    "rtsp": 8554,
    "webrtc": 8889,
    "hls": 8888,
    "websocket": 8002,
    "health": 8003,
}
```

## Best Practices

### Test Isolation

1. **Use Temporary Directories** - Always use `tempfile.TemporaryDirectory()` for file operations
2. **Mock External Dependencies** - Mock hardware, network, and external services
3. **Clean State** - Reset state between tests using fixtures
4. **Deterministic Results** - Use fixed seeds and deterministic values

### Environment Independence

1. **Avoid Host-Specific Paths** - Use relative paths or temporary directories
2. **Use IP Addresses** - Prefer `127.0.0.1` over `localhost`
3. **Mock Hardware Access** - Don't rely on actual camera devices
4. **Environment Variables** - Use environment variables for configuration

### Performance Optimization

1. **Session-Scoped Fixtures** - Use session scope for expensive setup
2. **Parallel Execution** - Tests should be able to run in parallel
3. **Minimal Dependencies** - Keep test dependencies minimal
4. **Fast Execution** - Tests should complete quickly

## Running Tests

### Basic Test Execution

```bash
# Run all tests
python3 -m pytest tests/ -v

# Run specific test category
python3 -m pytest tests/unit/ -v
python3 -m pytest tests/integration/ -v

# Run with coverage
python3 -m pytest tests/ --cov=src --cov-report=html
```

### Test Configuration

**pytest.ini** - Global pytest configuration:
```ini
[pytest]
minversion = 7.0
testpaths = tests/unit tests/integration
python_files = test_*.py
addopts = -ra -q
pythonpath = src
markers =
    unit: unit-level tests
    integration: integration-level tests
```

### Environment Setup

**Prerequisites** - No additional setup required:
- Python 3.10+
- pytest
- pytest-asyncio
- pytest-cov

**Automatic Setup** - Test environment is automatically configured:
- Environment variables set
- Test directories created
- Fixtures available

## Troubleshooting

### Common Issues

**Import Errors** - Ensure `pythonpath = src` in pytest.ini
**Permission Errors** - Tests use `/tmp` directory (should be writable)
**Network Errors** - Tests use `127.0.0.1` (should be available)
**Hardware Errors** - All hardware access is mocked

### Debug Mode

Enable debug mode for detailed test information:
```bash
python3 -m pytest tests/ -v -s --tb=long
```

### Test Isolation

If tests interfere with each other:
1. Check for shared state in fixtures
2. Ensure proper cleanup in `pytest_sessionfinish`
3. Use `scope="function"` for test-specific fixtures

## Continuous Integration

### CI Environment

The test environment is designed to work in CI environments:
- No hardware dependencies
- Minimal system requirements
- Fast execution
- Deterministic results

### CI Configuration

```yaml
# Example GitHub Actions configuration
- name: Run Tests
  run: |
    python3 -m pytest tests/ -v --cov=src --cov-report=xml
```

## Maintenance

### Regular Tasks

1. **Update Fixtures** - Keep fixtures up to date with code changes
2. **Review Dependencies** - Remove unnecessary test dependencies
3. **Performance Monitoring** - Monitor test execution time
4. **Coverage Analysis** - Maintain high test coverage

### Version Compatibility

- **Python Version** - Tests support Python 3.10+
- **pytest Version** - Tests require pytest 7.0+
- **Dependencies** - Keep test dependencies minimal

---

**Test Environment Guide:** ✅ **COMPLETE**  
**Next Steps:** Use this guide for consistent test development and maintenance
