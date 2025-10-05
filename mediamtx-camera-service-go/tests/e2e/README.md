# E2E Test Suite - Prerequisites and Setup

This directory contains end-to-end tests that validate complete user workflows across all system boundaries.

## Prerequisites

### Required Services

**MediaMTX Server:**
- Must be running on `localhost:9997`
- API endpoint accessible at `http://localhost:9997/v3/`
- RTSP server on port 8554
- WebRTC server on port 8889
- HLS server on port 8888

**WebSocket Server:**
- Camera service WebSocket server must be running on `localhost:8002`
- WebSocket endpoint: `ws://localhost:8002/ws`
- JSON-RPC 2.0 protocol support

### File System Requirements

**Write Access Required:**
- `/tmp/e2e-test-recordings/` - For recording file creation and verification
- `/tmp/e2e-test-snapshots/` - For snapshot file creation and verification
- `/tmp/` - General temporary file operations

**Disk Space:**
- Minimum 500MB available for test recordings
- Additional space for snapshot files and temporary data

### Environment Variables

**Required:**
```bash
export CAMERA_SERVICE_JWT_SECRET="e2e_test_secret_key_for_testing_only"
```

**Optional (for debugging):**
```bash
export CAMERA_SERVICE_LOG_LEVEL="debug"
export CAMERA_SERVICE_LOG_FORMAT="json"
```

### Camera Devices

**Option 1: Real Camera Devices**
- USB cameras connected and accessible via `/dev/video*` devices
- V4L2 compatible cameras for direct capture testing
- Network cameras accessible via RTSP URLs

**Option 2: Mock/Test Devices**
- MediaMTX configured with test video sources
- Virtual cameras or test patterns for consistent testing
- FFmpeg test sources for recording validation

### Go Environment

**Required:**
- Go 1.21+ installed and in PATH
- GOPATH configured correctly
- Module mode enabled

**Dependencies:**
- All project dependencies installed via `go mod download`
- Test dependencies available in vendor or module cache

## Running E2E Tests

### Basic Test Execution
```bash
# Run all E2E tests
go test ./tests/e2e/... -v -timeout=30m

# Run specific workflow category
go test ./tests/e2e/camera_workflows_test.go -v

# Run with coverage analysis
go test ./tests/e2e/... -coverpkg=./internal/... -coverprofile=coverage/e2e/e2e.out -v -timeout=30m
```

### Test Isolation Verification
```bash
# Run tests in random order
go test ./tests/e2e/... -shuffle=on -v

# Run tests in parallel
go test ./tests/e2e/... -parallel=2 -v
```

### Coverage Analysis
```bash
# Generate coverage report
go tool cover -func=coverage/e2e/e2e.out

# Generate HTML coverage report
go tool cover -html=coverage/e2e/e2e.out -o coverage/e2e/e2e.html
```

## Test Configuration

Tests use the E2E-specific configuration fixture:
- `tests/fixtures/config_e2e_test.yaml`
- Optimized timeouts for faster execution
- Test-specific file paths and JWT secrets
- All features enabled for comprehensive testing

## Troubleshooting

### Common Issues

**Service Not Available:**
- Verify MediaMTX server is running on port 9997
- Check camera service WebSocket server on port 8002
- Ensure no port conflicts with other services

**Permission Denied:**
- Verify write access to `/tmp/e2e-test-*` directories
- Check file permissions on test directories
- Ensure test user has necessary privileges

**Authentication Failures:**
- Verify `CAMERA_SERVICE_JWT_SECRET` environment variable is set
- Check JWT token generation in test helpers
- Ensure token expiry times are appropriate for test duration

**Camera Device Issues:**
- List available devices: `ls /dev/video*`
- Test device access: `ffmpeg -f v4l2 -list_formats all -i /dev/video0`
- Verify V4L2 compatibility for direct capture tests

### Debug Mode

Enable debug logging for troubleshooting:
```bash
export CAMERA_SERVICE_LOG_LEVEL="debug"
go test ./tests/e2e/... -v -timeout=30m
```

### Test Isolation Issues

If tests interfere with each other:
- Check for shared file paths or temporary directories
- Verify cleanup procedures are working correctly
- Ensure no global state is being modified between tests

## Test Categories

- **Camera Workflows** - Device discovery, status queries, capabilities
- **Recording Workflows** - Start/stop recording, file verification, multiple cameras
- **Snapshot Workflows** - Image capture, format validation, file verification
- **Health Workflows** - System health checks, metrics collection, monitoring
- **Security Workflows** - Authentication, authorization, session management

## Success Criteria

- All 19 E2E tests passing consistently
- Zero mocks - all real components and services
- Actual file creation verified on disk with content validation
- Real state changes verified through system queries
- Business outcomes validated (files playable, images viewable, etc.)
- Test isolation verified with shuffle and parallel execution
- 75% E2E coverage achieved with milestone validation
