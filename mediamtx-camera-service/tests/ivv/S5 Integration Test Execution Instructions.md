# S5 Integration Test Execution Instructions

## Overview

This document provides instructions for running the S5 end-to-end integration smoke tests and interpreting results.

**Created:** 2025-08-04  
**Evidence:** `tests/ivv/acceptance-plan.md`, `tests/ivv/test_integration_smoke.py`  
**Related Story:** E1/S5

## Prerequisites

### System Requirements
- Ubuntu 22.04+ or compatible Linux distribution
- Python 3.10+ with project dependencies installed
- MediaMTX server (can be mocked for core testing)
- Test camera device (USB camera or V4L2 virtual device)

### Environment Setup
```bash
# Navigate to project root
cd mediamtx-camera-service

# Install test dependencies
pip install pytest pytest-asyncio pytest-cov

# Ensure project modules are in Python path
export PYTHONPATH=$PWD/src:$PYTHONPATH

# Create test directories if needed
mkdir -p tests/ivv
mkdir -p tests/integration
```

## Test Execution

### Running All Integration Tests
```bash
# Run complete integration test suite
python3 -m pytest tests/ivv/test_integration_smoke.py -v

# Run with coverage reporting
python3 -m pytest tests/ivv/test_integration_smoke.py -v --cov=src --cov-report=term-missing
```

### Running Specific Test Scenarios
```bash
# Core end-to-end flow test (most critical)
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_end_to_end_camera_flow -v

# Basic connectivity test
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_ping_basic_connectivity -v

# Error recovery test
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_mediamtx_error_recovery -v

# Multiple client test
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_multiple_websocket_clients -v

# Performance validation (longer running)
python3 -m pytest tests/ivv/test_integration_smoke.py::TestPerformanceAndResources::test_resource_usage_limits -v -m slow
```

### Running with Different Test Markers
```bash
# Run only integration tests
python3 -m pytest tests/ivv/ -m integration -v

# Run only slow/performance tests
python3 -m pytest tests/ivv/ -m slow -v

# Skip slow tests
python3 -m pytest tests/ivv/ -m "not slow" -v
```

## Test Configuration

### Mock vs Real Testing
The integration tests are designed to work in two modes:

**Mock Mode (Default):**
- Uses mocked MediaMTX controller and camera monitor
- Suitable for CI/CD and environments without physical cameras
- Tests focus on API contracts and message flows

**Real Mode (Manual Setup):**
- Requires actual MediaMTX server running
- Requires real or virtual camera device
- Tests complete integration including media processing

### Environment Variables
```bash
# Enable real MediaMTX testing (requires MediaMTX running)
export TEST_REAL_MEDIAMTX=true

# Override test camera device
export TEST_CAMERA_DEVICE=/dev/video0

# Override test ports
export TEST_WEBSOCKET_PORT=8002
export TEST_MEDIAMTX_PORT=9997
```

## Expected Results

### Success Criteria
For the test suite to pass S5 acceptance criteria:

1. **Core Flow Test (`test_end_to_end_camera_flow`)** - MUST PASS
   - WebSocket connection established
   - Camera discovery and notification
   - API methods return expected responses
   - Recording and snapshot operations succeed
   - Error handling for invalid requests

2. **Connectivity Test (`test_ping_basic_connectivity`)** - MUST PASS
   - WebSocket server starts successfully
   - JSON-RPC ping/pong exchange works

3. **Error Recovery Test (`test_mediamtx_error_recovery`)** - SHOULD PASS
   - Service handles MediaMTX unavailability gracefully
   - Proper error codes returned

4. **Multiple Clients Test (`test_multiple_websocket_clients`)** - SHOULD PASS
   - Multiple WebSocket connections supported
   - Broadcast notifications delivered to all clients

### Performance Targets
- Test execution time: <5 minutes for full suite
- Memory usage during tests: <200MB
- No memory leaks or resource buildup

## Troubleshooting

### Common Issues

**Import Errors:**
```bash
# Ensure PYTHONPATH includes src directory
export PYTHONPATH=$PWD/src:$PYTHONPATH

# Or install in development mode
pip install -e .
```

**Port Conflicts:**
```bash
# Check if ports are in use
netstat -tuln | grep 8002
netstat -tuln | grep 9997

# Kill conflicting processes or change test ports
export TEST_WEBSOCKET_PORT=8003
```

**Missing Dependencies:**
```bash
# Install all test dependencies
pip install -r requirements-dev.txt

# Or install specific test packages
pip install pytest pytest-asyncio pytest-cov websockets
```

**MediaMTX Connection Issues (Real Mode):**
```bash
# Start MediaMTX server
./mediamtx

# Verify MediaMTX API is accessible
curl http://localhost:9997/v3/paths/list
```

### Log Analysis
Test logs are output to console. For debugging:

```bash
# Run with debug logging
python3 -m pytest tests/ivv/test_integration_smoke.py -v -s --log-cli-level=DEBUG

# Capture logs to file
python3 -m pytest tests/ivv/test_integration_smoke.py -v --log-file=test_logs.txt
```

## Integration with CI/CD

### GitHub Actions Integration
```yaml
# Example CI step for S5 testing
- name: Run S5 Integration Tests
  run: |
    export PYTHONPATH=$PWD/src:$PYTHONPATH
    python3 -m pytest tests/ivv/test_integration_smoke.py -v \
      --cov=src --cov-report=xml \
      -m "not slow"  # Skip performance tests in CI
```

### Test Artifacts
The following artifacts should be collected from test runs:
- Test results (JUnit XML format)
- Coverage reports
- Log files for failed tests
- Performance metrics (if collected)

## Next Steps

After successful S5 test execution:

1. **Document Results:** Record test execution results and any issues found
2. **Address Gaps:** Fix any discovered issues or document as known limitations
3. **CI Integration:** Add tests to automated CI/CD pipeline
4. **Real Environment Testing:** Validate against actual MediaMTX and camera hardware
5. **Performance Baseline:** Establish performance baselines for future regression testing

## Known Limitations and Gaps

See "Discovered Gaps" section below for current limitations and required follow-up work.