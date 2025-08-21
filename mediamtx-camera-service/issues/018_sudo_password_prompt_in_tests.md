# Issue 018: Sudo Password Prompt in Tests

**Date:** 2025-01-15  
**Priority:** HIGH  
**Status:** INVESTIGATION COMPLETE - ROOT CAUSE CONFIRMED  
**Type:** Test Infrastructure Bug  

## Description

Tests are unexpectedly prompting for sudo passwords during execution, which is causing test failures and interrupting the automated test flow.

## Evidence

### Test Output Showing Sudo Prompts

```
sudo: a password is required
========================================= test session starts =========================================
platform linux -- Python 3.10.12, pytest-8.4.1, pluggy-1.6.0
```

### Affected Tests

Multiple tests are showing this behavior:
- `test_ffmpeg_integration.py::test_ffmpeg_integration`
- `test_critical_interfaces.py::test_ping_method`
- `test_mediamtx_real_integration.py::TestMediaMTXRealIntegration::test_real_error_handling_scenarios`
- `test_api_contracts.py::TestAPIContracts::test_json_rpc_contract_compliance`

## Root Cause Analysis

### Root Cause Identified

The sudo password prompt is caused by a **pytest configuration check** in `tests/conftest.py` that runs during test collection:

```python
# Check if sudo is available for sudo_required tests
import subprocess
try:
    subprocess.run(["sudo", "-n", "true"], check=True, timeout=5)
    config.sudo_available = True
except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
    config.sudo_available = False
```

### Why This Happens

1. **Pytest Collection Phase**: This check runs during pytest's collection phase, before any tests execute
2. **Sudo Configuration**: The `sudo -n true` command should not prompt for password, but it's failing
3. **Environment Issue**: The system's sudo configuration may not allow passwordless sudo for the test user
4. **Timing**: This happens even for tests that don't require sudo privileges

### Investigation Findings

**Confirmed Root Cause:** The sudo password prompt occurs during pytest collection when the `conftest.py` file runs the sudo availability check.

**Environment Behavior:**
- `sudo -n true` works correctly when run directly in terminal
- `sudo -n true` fails with "sudo: a password is required" when run via subprocess from the mediamtx-camera-service directory
- This suggests a directory-specific environment issue or sudo configuration

**Additional Sudo Usage Found:**
- `tests/integration/test_installation_validation.py` contains multiple sudo calls for systemctl operations
- `tests/performance/test_performance_validation.py` contains sudo calls for MediaMTX restart
- These tests are marked with `@pytest.mark.sudo_required` and should be skipped when sudo is not available

### Additional Causes (Secondary)

1. **FFmpeg System Calls**: Some tests call FFmpeg commands that may require elevated privileges
2. **File System Operations**: Tests writing to protected directories (`/opt/camera-service/`, `/etc/mediamtx/`)
3. **Port Binding**: Tests trying to bind to privileged ports (< 1024)
4. **Service Management**: Tests trying to start/stop system services
5. **Device Access**: Tests accessing camera devices that require root access

## Impact

### Test Execution Issues
- Tests fail when sudo password is not provided
- Automated CI/CD pipelines are interrupted
- Test results are unreliable
- Development workflow is disrupted

### Security Concerns
- Tests should not require elevated privileges
- Sudo prompts indicate potential security misconfiguration
- Tests should run in isolated, unprivileged environment

## Investigation Plan

### Phase 1: Identify Sudo Usage ✅ COMPLETED
```bash
# Search for sudo usage in test files
grep -r "sudo" tests/
grep -r "subprocess.*sudo" tests/
grep -r "os.system.*sudo" tests/
```

### Phase 2: Check File Permissions ✅ COMPLETED
```bash
# Check permissions of test directories
ls -la tests/
ls -la /opt/camera-service/
ls -la /etc/mediamtx/
```

### Phase 3: Review Port Usage ✅ COMPLETED
```bash
# Check if tests use privileged ports
grep -r "8002\|8003\|8554\|8888\|8889\|9997" tests/
```

### Phase 4: Analyze FFmpeg Integration ✅ COMPLETED
```bash
# Check FFmpeg command execution
grep -r "ffmpeg" tests/
grep -r "subprocess.*ffmpeg" tests/
```

## Expected Behavior

### Tests Should:
- Run without requiring elevated privileges
- Use temporary directories for file operations
- Use non-privileged ports (> 1024)
- Mock system services when possible
- Run in isolated environment

### Tests Should NOT:
- Require sudo passwords
- Write to system directories
- Bind to privileged ports
- Start/stop system services
- Access protected devices

## Proposed Solutions

### 1. Fix Pytest Configuration (IMMEDIATE)

**File:** `tests/conftest.py`

**Problem:** The sudo check runs during pytest collection and prompts for password
**Solution:** Make the sudo check non-blocking and silent

```python
# Current problematic code:
try:
    subprocess.run(["sudo", "-n", "true"], check=True, timeout=5)
    config.sudo_available = True
except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
    config.sudo_available = False

# Fixed code:
try:
    # Use capture_output=True to suppress any output
    result = subprocess.run(
        ["sudo", "-n", "true"], 
        check=False,  # Don't raise exception on failure
        timeout=2,    # Shorter timeout
        capture_output=True,  # Suppress output
        text=True
    )
    config.sudo_available = result.returncode == 0
except (subprocess.TimeoutExpired, FileNotFoundError):
    config.sudo_available = False
```

### 2. Alternative: Remove Sudo Check Entirely

If sudo is not needed for most tests, remove the check entirely:

```python
# Remove the sudo check and default to False
config.sudo_available = False
```

### 3. Use Temporary Directories
```python
# Instead of system directories
import tempfile
temp_dir = tempfile.mkdtemp(prefix="test_")
```

### 4. Use Non-Privileged Ports
```python
# Use ports > 1024 for testing
test_port = 9000 + random.randint(1, 999)
```

### 5. Mock System Services
```python
# Mock MediaMTX service instead of starting real one
@patch('mediamtx_wrapper.controller.MediaMTXController')
def test_with_mocked_service(mock_controller):
    # Test with mocked service
```

### 6. Use Test-Specific Configuration
```python
# Use test configuration that doesn't require privileges
test_config = Config(
    mediamtx=MediaMTXConfig(
        host="127.0.0.1",
        api_port=9997,
        recordings_path="/tmp/test_recordings",
        snapshots_path="/tmp/test_snapshots"
    )
)
```

## Files to Investigate

### Test Files
- `tests/integration/test_ffmpeg_integration.py`
- `tests/integration/test_critical_interfaces.py`
- `tests/integration/test_mediamtx_real_integration.py`
- `tests/integration/test_api_contracts.py`

### Configuration Files
- `pytest.ini`
- `conftest.py`
- `tests/fixtures/`

### Service Files
- `src/camera_service/service_manager.py`
- `src/mediamtx_wrapper/controller.py`
- `src/health_server.py`

## Acceptance Criteria

- [ ] No sudo password prompts during test execution
- [ ] All tests run without elevated privileges
- [ ] Tests use temporary directories for file operations
- [ ] Tests use non-privileged ports
- [ ] Tests are properly isolated from system services
- [ ] CI/CD pipeline runs without interruption

## Next Steps

1. **Immediate**: Fix pytest configuration sudo check
2. **Short-term**: Fix file permission and port binding issues
3. **Medium-term**: Implement proper test isolation
4. **Long-term**: Review and improve test infrastructure

---

**Bug ID:** 018  
**Severity:** HIGH  
**Assigned:** TBD  
**Estimated Effort:** 1-2 days
