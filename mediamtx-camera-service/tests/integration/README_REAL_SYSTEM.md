# Real System Integration Tests

**Version:** 1.0  
**Authors:** Development Team  
**Date:** 2025-01-06  
**Status:** Approved  
**Related Epic/Story:** E1 / S5 - PDR Build Pipeline Remediation

## Overview

The Real System Integration Tests validate actual end-to-end system behavior without excessive mocking. These tests replace mock-based integration tests with real system validation to ensure production-ready functionality.

### Key Objectives

- **Real MediaMTX Server Integration**: Actual MediaMTX server startup, configuration, and API interactions
- **Real Camera Device Simulation**: Test video streams using FFmpeg for realistic camera behavior
- **Real File System Operations**: Actual recording and snapshot file creation and management
- **Real WebSocket Connections**: Authentic WebSocket JSON-RPC communication
- **Real FFmpeg Process Execution**: Actual media processing operations
- **Real Error Scenarios**: Actual service failures and recovery testing

## Test Architecture

### Core Components

1. **RealMediaMTXServer**: Manages actual MediaMTX server process lifecycle
2. **TestVideoStreamSimulator**: Creates real video streams using FFmpeg
3. **WebSocketTestClient**: Real WebSocket client for API testing
4. **TestRealSystemIntegration**: Main test class with comprehensive test scenarios

### Test Scenarios

#### 1. Real MediaMTX Server Startup and Health
- Validates actual MediaMTX server process startup
- Tests real API health endpoint responses
- Verifies configuration file generation
- Confirms directory structure creation

#### 2. Real Camera Discovery and Stream Creation
- Tests actual camera discovery event processing
- Validates real MediaMTX stream creation via API
- Verifies WebSocket notification delivery
- Tests stream URL generation and accessibility

#### 3. Real Recording and Snapshot Operations
- Tests actual recording start/stop via MediaMTX API
- Validates real file system operations for recordings
- Tests snapshot capture and file creation
- Verifies WebSocket notification delivery

#### 4. Real WebSocket Authentication and Control
- Tests actual WebSocket connection establishment
- Validates JSON-RPC method handling
- Tests camera status monitoring
- Verifies error handling and response formatting

#### 5. Real Error Scenarios and Recovery
- Tests MediaMTX server failure and recovery
- Validates WebSocket connection failure handling
- Tests file system error handling
- Verifies process lifecycle management

#### 6. Real Resource Management and Cleanup
- Tests file system cleanup operations
- Validates process termination and cleanup
- Tests memory and resource management
- Verifies temporary file cleanup

#### 7. Real End-to-End Camera Lifecycle
- Complete camera discovery → stream creation → recording → snapshot capture
- WebSocket authentication → camera control → status monitoring
- Real system startup, configuration, and shutdown sequences

## Prerequisites

### System Requirements

- **Operating System**: Linux (Ubuntu 22.04+ recommended)
- **Python**: 3.10+ with asyncio support
- **MediaMTX**: Latest stable version installed and available in PATH
- **FFmpeg**: Latest stable version installed and available in PATH
- **Network**: Available ports for dynamic port allocation

### Dependencies

```bash
# Install MediaMTX
sudo apt update
sudo apt install mediamtx

# Install FFmpeg
sudo apt install ffmpeg

# Install Python dependencies
pip install -r requirements.txt
pip install -r requirements-dev.txt
```

### Environment Setup

```bash
# Set up Python environment
python3 -m venv venv
source venv/bin/activate

# Install project dependencies
pip install -e .
```

## Running Tests

### Using the Test Runner (Recommended)

```bash
# Check dependencies
python3 tests/integration/run_real_integration_tests.py --check-deps

# Run all tests
python3 tests/integration/run_real_integration_tests.py --all

# Run specific test
python3 tests/integration/run_real_integration_tests.py --test test_real_mediamtx_server_startup_and_health
```

### Using pytest

```bash
# Run all real integration tests
python3 -m pytest tests/integration/test_real_system_integration.py -v -s

# Run specific test
python3 -m pytest tests/integration/test_real_system_integration.py::TestRealSystemIntegration::test_real_mediamtx_server_startup_and_health -v -s

# Run with detailed logging
python3 -m pytest tests/integration/test_real_system_integration.py -v -s --log-cli-level=INFO
```

### Using the Main Test Script

```bash
# Run all tests
python3 run_all_tests.py --integration

# Run with real system validation
python3 run_all_tests.py --real-system
```

## Test Configuration

### Dynamic Port Allocation

Tests use dynamic port allocation to avoid conflicts:
- Server port: Automatically assigned
- MediaMTX API port: Automatically assigned
- MediaMTX RTSP port: Automatically assigned
- MediaMTX WebRTC port: Automatically assigned
- MediaMTX HLS port: Automatically assigned

### Temporary Directory Management

Tests create temporary directories for:
- MediaMTX configuration files
- Recording files
- Snapshot files
- Test video streams
- Log files

All temporary files are automatically cleaned up after test completion.

## Test Validation Areas

### Critical Validation Areas

1. **End-to-End Camera Discovery and Stream Management**
   - Camera device detection and event processing
   - MediaMTX stream creation and configuration
   - Stream URL generation and accessibility
   - Camera status tracking and updates

2. **Real Authentication and Authorization Flows**
   - WebSocket connection establishment
   - JSON-RPC method authentication
   - Error handling and response validation
   - Connection recovery and reconnection

3. **Actual File System Operations and Cleanup**
   - Recording file creation and management
   - Snapshot file generation and storage
   - Directory structure creation and cleanup
   - File permission and access validation

4. **Real Process Lifecycle Management**
   - MediaMTX server startup and shutdown
   - FFmpeg process creation and termination
   - Service manager component coordination
   - Resource cleanup and memory management

5. **System Behavior Under Actual Failure Conditions**
   - MediaMTX server failure and recovery
   - WebSocket connection failure handling
   - File system error scenarios
   - Process termination and restart

## Success Criteria

### Integration Test Validation

- ✅ **Real MediaMTX Server Integration**: Actual server startup, configuration, and API interactions
- ✅ **Real Camera Device Simulation**: Test video streams with realistic camera behavior
- ✅ **Real File System Operations**: Actual recording and snapshot file operations
- ✅ **Real WebSocket Connections**: Authentic JSON-RPC communication
- ✅ **Real FFmpeg Process Execution**: Actual media processing operations
- ✅ **Real Error Scenarios**: Actual service failures and recovery testing

### Performance Requirements

- **Test Execution Time**: Complete test suite completes within 10 minutes
- **Resource Usage**: Tests use less than 2GB RAM and 1GB disk space
- **Network Usage**: Tests use less than 100MB network bandwidth
- **Process Management**: All test processes terminate cleanly

### Quality Gates

- **Test Coverage**: 100% of critical system paths validated
- **Error Handling**: All error scenarios properly tested and handled
- **Resource Cleanup**: No resource leaks or orphaned processes
- **Logging**: Comprehensive logging for debugging and monitoring

## Troubleshooting

### Common Issues

#### MediaMTX Not Found
```bash
# Check MediaMTX installation
which mediamtx
mediamtx --version

# Install MediaMTX if missing
sudo apt install mediamtx
```

#### FFmpeg Not Found
```bash
# Check FFmpeg installation
which ffmpeg
ffmpeg -version

# Install FFmpeg if missing
sudo apt install ffmpeg
```

#### Port Conflicts
```bash
# Check for port conflicts
sudo netstat -tlnp | grep -E ':(8554|8889|8888|9997)'

# Kill conflicting processes
sudo pkill -f mediamtx
```

#### Permission Issues
```bash
# Check file permissions
ls -la /dev/video*
sudo chmod 666 /dev/video*

# Check directory permissions
ls -la /tmp/
chmod 755 /tmp/
```

### Debug Mode

```bash
# Run with debug logging
python3 -m pytest tests/integration/test_real_system_integration.py -v -s --log-cli-level=DEBUG

# Run single test with debug
python3 tests/integration/run_real_integration_tests.py --test test_real_mediamtx_server_startup_and_health
```

### Log Analysis

```bash
# View test logs
tail -f /tmp/real_integration_runner_*/test.log

# Analyze MediaMTX logs
journalctl -u mediamtx -f

# Check system resources
htop
df -h
```

## Integration with CI/CD

### GitHub Actions

```yaml
# .github/workflows/real-integration-tests.yml
name: Real System Integration Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  real-integration-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.10'
    
    - name: Install dependencies
      run: |
        sudo apt update
        sudo apt install -y mediamtx ffmpeg
        pip install -r requirements.txt
        pip install -r requirements-dev.txt
    
    - name: Run real integration tests
      run: |
        python3 tests/integration/run_real_integration_tests.py --all
```

### Local Development

```bash
# Pre-commit hook for real integration tests
# .git/hooks/pre-commit
#!/bin/bash
python3 tests/integration/run_real_integration_tests.py --check-deps
if [ $? -ne 0 ]; then
    echo "Real integration test dependencies not met"
    exit 1
fi
```

## Related Documentation

- [Integration Test Overview](README.md)
- [Service Manager Documentation](../src/camera_service/service_manager.py)
- [MediaMTX Integration](../src/mediamtx_wrapper/)
- [WebSocket Server](../src/websocket_server/)
- [Camera Discovery](../camera_discovery/)

## Contributing

When adding new real system integration tests:

1. **Follow Naming Convention**: `test_real_<feature>_<scenario>()`
2. **Use Real Components**: Avoid excessive mocking
3. **Include Error Scenarios**: Test failure and recovery paths
4. **Add Documentation**: Update this README with new test details
5. **Update Test Runner**: Add new tests to the test runner script

## Support

For issues with real system integration tests:

1. Check the troubleshooting section above
2. Review test logs for detailed error information
3. Verify all dependencies are properly installed
4. Ensure sufficient system resources are available
5. Contact the development team for assistance
