# Integration Tests - Prerequisites and Setup

This document outlines the required environment and verification steps for running integration tests in the MediaMTX Camera Service.

## Required Environment

### 1. MediaMTX Server
- **Purpose:** MediaMTX integration tests require a running MediaMTX server
- **URL:** `localhost:9997`
- **API Endpoint:** `http://localhost:9997/v3/`

### 2. Filesystem Access
- **Path:** `/tmp/mediamtx_test_data/`
- **Purpose:** Test file operations and temporary data storage
- **Permissions:** Read/write access required

### 3. Environment Variables
- **CAMERA_SERVICE_JWT_SECRET:** Required for security integration tests
- **Purpose:** JWT token validation and authentication testing

### 4. System Dependencies
- **Go Version:** 1.21+ (required for integration tests)
- **v4l2-ctl:** Required for camera hardware integration tests

## Verification Commands

Run these commands to verify your environment is properly configured:

### 1. Verify MediaMTX Server
```bash
# Test MediaMTX server connectivity
curl http://localhost:9997/v3/config/global/get

# Expected: JSON response or HTTP 200
```

### 2. Verify Filesystem Access
```bash
# Create test directory and verify write access
mkdir -p /tmp/mediamtx_test_data
touch /tmp/mediamtx_test_data/test.txt
ls -la /tmp/mediamtx_test_data/test.txt

# Expected: File created successfully
```

### 3. Verify Environment Variables
```bash
# Check JWT secret is set
echo $CAMERA_SERVICE_JWT_SECRET

# Expected: Non-empty string (the actual secret value)
```

### 4. Verify Go Version
```bash
# Check Go version
go version

# Expected: go version go1.21.x or higher
```

### 5. Verify v4l2-ctl (for camera tests)
```bash
# Check v4l2-ctl availability
which v4l2-ctl
v4l2-ctl --version

# Expected: Command found and version displayed
```

## Running Integration Tests

### Basic Test Execution
```bash
# Run all integration tests
go test ./tests/integration/... -v

# Run with coverage
go test ./tests/integration/... -coverpkg=./internal/... -coverprofile=coverage/integration/integration.out -v
```

### Test Isolation Verification
```bash
# Verify tests can run in random order
go test ./tests/integration/... -shuffle=on -v

# Verify tests can run in parallel
go test ./tests/integration/... -parallel=4 -v
```

### Coverage Analysis
```bash
# View function-level coverage
go tool cover -func=coverage/integration/integration.out

# Generate HTML coverage report
go tool cover -html=coverage/integration/integration.out
```

## Troubleshooting

### MediaMTX Server Issues
- Ensure MediaMTX is installed and running
- Check if port 9997 is available and not blocked
- Verify MediaMTX configuration allows API access

### Filesystem Permission Issues
- Ensure user has write access to `/tmp/`
- Check if SELinux or other security policies block access
- Verify disk space is available

### Environment Variable Issues
- Ensure `CAMERA_SERVICE_JWT_SECRET` is set in your shell
- Check if the variable is exported (use `export CAMERA_SERVICE_JWT_SECRET=your_secret`)
- Verify the secret is not empty

### Go Version Issues
- Upgrade Go to version 1.21 or higher
- Ensure `go` command is in your PATH
- Check if multiple Go versions are installed

### v4l2-ctl Issues
- Install v4l-utils package: `sudo apt-get install v4l-utils` (Ubuntu/Debian)
- For other systems, check your package manager for v4l2 utilities
- Ensure the command is in your PATH

## Integration Test Files

This directory contains the following integration test files:

- `mediamtx_client_test.go` - MediaMTX HTTP client integration
- `config_loading_test.go` - Configuration loading and validation
- `security_auth_test.go` - Security and authentication integration
- `camera_monitor_test.go` - Camera hardware integration

Each file follows the integration testing guidelines and includes proper requirements traceability.
