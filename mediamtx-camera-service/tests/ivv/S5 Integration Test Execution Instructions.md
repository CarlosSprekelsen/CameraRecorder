# S5 Integration Test Execution Instructions

**Version:** 1.0  
**Date:** 2025-08-05  
**Related Story:** E1/S5 - Core Integration IV&V  
**Evidence Files:** `tests/ivv/acceptance-plan.md`, `tests/ivv/test_integration_smoke.py`

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
pip install pytest pytest-asyncio pytest-cov websockets

# Ensure project modules are in Python path
export PYTHONPATH=$PWD/src:$PYTHONPATH

# Create test directories if needed
mkdir -p tests/ivv
mkdir -p tests/integration
```

## Test Execution

### Running All S5 Integration Tests
```bash
# Run complete S5 integration test suite
python3 -m pytest tests/ivv/test_integration_smoke.py -v

# Run with coverage reporting
python3 -m pytest tests/ivv/test_integration_smoke.py -v --cov=src --cov-report=term-missing

# Run integration tests only (skip slow performance tests)
python3 -m pytest tests/ivv/test_integration_smoke.py -v -m "integration and not slow"
```

### Running Specific Test Scenarios
```bash
# Core end-to-end flow test (MUST PASS)
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_end_to_end_camera_flow -v

# Basic connectivity test (MUST PASS)
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_ping_basic_connectivity -v

# Error recovery test (SHOULD PASS)
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_mediamtx_error_recovery -v

# Multiple clients test (SHOULD PASS)
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_multiple_websocket_clients -v

# Notification flow test (MUST PASS)
python3 -m pytest tests/ivv/test_integration_smoke.py::TestEndToEndIntegration::test_notification_delivery_flow -v

# Performance validation (longer running)
python3 -m pytest tests/ivv/test_integration_smoke.py::TestPerformanceAndResources::test_resource_usage_limits -v -m slow
```

### Using Test Framework Integration
```bash
# Run through project test framework (if available)
python3 run_all_tests.py --only-integration

# Or run integration tests via pytest discovery
python3 -m pytest tests/ivv/ -m integration -v
```

## Test Configuration

### Mock vs Real Testing Modes

**Mock Mode (Default):**
- Uses mocked MediaMTX controller and camera monitor
- Suitable for CI/CD and environments without physical cameras
- Tests focus on API contracts and message flows
- No external dependencies required

**Real Mode (Advanced Setup):**
- Requires actual MediaMTX server running
- Requires real or virtual camera device
- Tests complete integration including media processing

### Environment Variables
```bash
# Override test ports (if defaults conflict)
export TEST_WEBSOCKET_PORT=8003
export TEST_MEDIAMTX_PORT=9998

# Enable debug logging during tests
export TEST_LOG_LEVEL=DEBUG

# Specify test camera device for real mode
export TEST_CAMERA_DEVICE=/dev/video0

# Enable real MediaMTX testing (requires MediaMTX running)
export TEST_REAL_MEDIAMTX=true
```

## Expected Results

### Critical Success Criteria
For S5 acceptance, these tests **MUST PASS**:

1. **Core Flow Test (`test_end_to_end_camera_flow`)**
   - WebSocket connection established
   - Camera discovery and status retrieval
   - API methods return expected responses
   - Recording and snapshot operations succeed
   - Error handling for invalid requests

2. **Connectivity Test (`test_ping_basic_connectivity`)**
   - WebSocket server starts successfully
   - JSON-RPC ping/pong exchange works

3. **Notification Flow Test (`test_notification_delivery_flow`)**
   - Camera status notifications delivered correctly
   - Recording status notifications delivered correctly
   - Notification schema compliance verified

4. **Invalid Requests Test (`test_invalid_api_requests`)**
   - Proper error codes returned for invalid requests
   - Service stability maintained during errors

### Optional Success Criteria
These tests **SHOULD PASS** but failures are acceptable:

1. **Error Recovery Test (`test_mediamtx_error_recovery`)**
   - Service handles MediaMTX unavailability gracefully
   - Proper error codes returned (-1003: MediaMTX error)

2. **Multiple Clients Test (`test_multiple_websocket_clients`)**
   - Multiple WebSocket connections supported
   - Broadcast notifications delivered to all clients

3. **Resource Usage Test (`test_resource_usage_limits`)**
   - Memory usage <150MB during testing
   - No significant memory leaks detected

### Performance Targets
- Test execution time: <5 minutes for full suite
- Memory usage during tests: <200MB peak
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
pip install pytest pytest-asyncio pytest-cov websockets psutil
```

**MediaMTX Connection Issues (Real Mode Only):**
```bash
# Start MediaMTX server
./mediamtx

# Verify MediaMTX API is accessible
curl http://localhost:9997/v3/paths/list
```

### Log Analysis
Test logs are output to console by default. For debugging:

```bash
# Run with debug logging enabled
python3 -m pytest tests/ivv/test_integration_smoke.py -v -s --log-cli-level=DEBUG

# Capture logs to file
python3 -m pytest tests/ivv/test_integration_smoke.py -v --log-file=test_logs.txt

# Run with maximum verbosity
python3 -m pytest tests/ivv/test_integration_smoke.py -vvv -s
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
      -m "integration and not slow"  # Skip performance tests in CI
```

### Test Artifacts Collection
The following artifacts should be collected from test runs:
- Test results (JUnit XML format): `--junit-xml=test-results.xml`
- Coverage reports: `--cov-report=xml:coverage.xml`
- Log files for failed tests: `--log-file=integration-test.log`
- Performance metrics (if psutil available)

## S5 Acceptance Gate

### IV&V Control Point Requirements
To pass the S5 IV&V control point, the following must be demonstrated:

1. **All critical tests pass** (marked as MUST PASS above)
2. **Test execution completes** within performance targets
3. **No unhandled exceptions** during test execution
4. **Resource usage** remains within acceptable limits
5. **Error scenarios** are handled gracefully with proper error codes

### Documentation Evidence
- Test plan: `tests/ivv/acceptance-plan.md`
- Test implementation: `tests/ivv/test_integration_smoke.py`
- Execution results: Generated test report with pass/fail status
- Architecture compliance: References to approved components and interfaces

## Next Steps

After successful S5 test execution:

1. **Document Results:** Record test execution results and any issues found
2. **Address Gaps:** Fix any discovered issues or document as known limitations  
3. **CI Integration:** Add tests to automated CI/CD pipeline
4. **Real Environment Testing:** Validate against actual MediaMTX and camera hardware
5. **Performance Baseline:** Establish performance baselines for future regression testing

## Known Limitations and Gaps

**Current Testing Gaps:**
- Authentication testing deferred to S6 (security implementation)
- Load testing deferred to S7 (performance validation)
- Multi-camera concurrency requires additional test infrastructure
- Physical camera disconnect testing requires manual hardware manipulation

**Mitigation:**
- Mock mode provides adequate coverage for CI/CD and automated testing
- Real mode available for manual validation with actual hardware
- Performance tests use resource monitoring for memory leak detection
- Error injection scenarios cover most failure modes without hardware manipulation

---

**S5 Completion Criteria:**
All critical tests must pass to advance to E2 (Security and Production Hardening). Optional tests provide additional confidence but are not blocking for S5 completion.