# Go Testing Guide - MediaMTX Camera Service

**Version:** 2.0 (Simplified)  
**Date:** 2025-01-15  
**Status:** Active Testing Guide  

## 1. Core Principles

### 🚨 **CRITICAL: Authorization & Documentation**
- **STOP before modifying any code** - Investigate and understand first
- **Get explicit authorization** before making changes
- **Document findings** with evidence
- **Maintain requirements traceability** - Every test must trace to requirements

### 🚨 **API Documentation is Ground Truth**
- **API Documentation**: `docs/api/json_rpc_methods.md` - ONLY source for API behavior
- **Health Endpoints**: `docs/api/health-endpoints.md` - ONLY source for health API
- **Tests validate API compliance** - Not implementation details
- **Test failures indicate bugs** - Not test problems

## 2. Test Organization

### Directory Structure
```
internal/                          # Source + unit tests
├── mediamtx/
│   ├── client.go                 # Source code
│   ├── client_test.go            # Unit test (package mediamtx)
│   └── ...

tests/                             # Integration & performance tests  
├── integration/                   # Integration tests (package xxx_test)
├── performance/                   # Performance tests
├── fixtures/                      # Test data files
└── utils/                         # Shared test utilities
```

### Package Declaration Rules
- **Unit tests**: `package mediamtx` (same as source - can test private functions)
- **Integration tests**: `package mediamtx_test` (external - only public API)

### File Naming Convention
- **Unit**: `<component>_test.go` (e.g., `client_test.go`)
- **Integration**: `test_<feature>_integration.go`
- **One file per feature** - no variants (_real, _v2)

## 3. Requirements Traceability (MANDATORY)

### Test File Header Format
```go
/*
Module: MediaMTX Health Monitoring
Purpose: Validates health check functionality

Requirements Coverage:
- REQ-MTX-004: Health monitoring and circuit breaker
- REQ-ERROR-001: Graceful error handling

Test Categories: Unit/Integration/Performance
API Documentation: docs/api/json_rpc_methods.md
*/
package mediamtx

func TestHealthMonitor_CheckHealth(t *testing.T) {
    // REQ-MTX-004: Validate health check with circuit breaker
    // Test implementation that FAILS if requirement violated
}
```

### Requirements Coverage Tracking
- **Document**: `docs/testing/requirements_coverage_analysis.md`
- **Baseline**: 161 frozen requirements
- **Target Coverage**:
  - Critical Requirements: 100% (45 requirements)
  - High Priority: 95% (67 requirements)
  - Overall: 90% (137/161 requirements)

## 4. Running Tests

### Before Running Any Tests
```bash
source .test_env  # Sets JWT secret and test paths
```

### Unit Tests (Standard Go)
```bash
go test ./internal/mediamtx/              # Run all
go test ./internal/mediamtx/ -v           # Verbose
go test -cover ./internal/mediamtx/       # With coverage
go test ./internal/mediamtx/ -run TestHealth  # Specific test
```

### Integration Tests (Require Build Tags)
```bash
go test -tags=integration ./tests/integration/
go test -tags=integration ./tests/integration/ -run TestMediaMTX
```

### Build Tags for Integration Tests
```go
//go:build integration
// +build integration

package mediamtx_test
```

## 5. Testing Approach

### Use Real Components
- ✅ **Real MediaMTX service** (via systemd)
- ✅ **Real filesystem** (use temp directories)  
- ✅ **Real WebSocket connections**
- ✅ **Real JWT tokens** with test secrets

### Only Mock External Dependencies
- External APIs (third-party services)
- Time operations (when testing timeouts)
- Hardware not available in test environment

## 6. API Compliance Testing

### Every API Test Must Validate Against Documentation
```go
func TestGetCameraStatus_APICompliance(t *testing.T) {
    // REQ-API-002: JSON-RPC 2.0 protocol implementation
    
    // Use EXACT format from docs/api/json_rpc_methods.md
    request := JSONRPCRequest{
        JSONRPC: "2.0",
        Method:  "get_camera_status",
        Params:  map[string]interface{}{
            "device": "/dev/video0",  // Parameter name from API doc
        },
        ID: 1,
    }
    
    response := sendRequest(request)
    
    // Validate ALL documented fields are present
    result := response.Result.(map[string]interface{})
    require.Contains(t, result, "status", "Missing 'status' field per API doc")
    require.Contains(t, result, "device", "Missing 'device' field per API doc")
    require.Contains(t, result, "capabilities", "Missing 'capabilities' field per API doc")
}
```

## 7. Common Test Patterns

### Unit Test Example
```go
// internal/mediamtx/client_test.go
package mediamtx

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestClient_parseHealthResponse(t *testing.T) {
    // REQ-MTX-004: Health monitoring
    
    // Can access private methods in unit tests
    client := &Client{baseURL: "http://localhost:9997"}
    health := client.parseHealthResponse(data) // private method
    require.Equal(t, "healthy", health.Status)
}
```

### Integration Test Example
```go
// tests/integration/test_mediamtx_integration.go
//go:build integration
// +build integration

package mediamtx_test

import (
    "testing"
    "github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
    "github.com/stretchr/testify/require"
)

func TestMediaMTX_RealSystemIntegration(t *testing.T) {
    // REQ-MTX-001: MediaMTX service integration
    
    // Only public API accessible in integration tests
    client := mediamtx.NewClient("http://localhost:9997")
    err := client.HealthCheck()
    require.NoError(t, err, "Real MediaMTX service should be healthy")
}
```

### Using Test Utilities
```go
import "github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"

func TestWithEnvironment(t *testing.T) {
    // REQ-TEST-001: Test environment setup
    
    env := testtestutils.SetupTestEnvironment(t)
    defer testtestutils.TeardownTestEnvironment(t, env)
    
    // env.ConfigManager, env.Logger ready to use
    // env.TempDir for test files
}
```

## 8. Port Management

### Test Servers (Use Dynamic Ports)
```go
// For test WebSocket servers - use GetFreePort
port := testtestutils.GetFreePort()
config := &WebSocketConfig{Port: port}
server.Start()
```

### Real Services (Keep Hardcoded)
```go
// MediaMTX ports - DO NOT CHANGE
mediamtx.NewClient("http://localhost:9997")  // API port
rtspURL := "rtsp://localhost:8554/stream"     // RTSP port
webrtcURL := "http://localhost:8889/"         // WebRTC port
hlsURL := "http://localhost:8888/"            // HLS port
```

## 9. Performance Targets

- **Unit tests**: <30 seconds total
- **Integration tests**: <5 minutes total
- **Full suite**: <10 minutes total
- **Flaky rate**: <1% (must be deterministic)

## 10. Quality Checklist

### Before Submitting Tests
- [ ] **Requirements documented** in test file header
- [ ] **REQ-XXX references** in test functions
- [ ] **API compliance** validated against documentation
- [ ] **Test passes** with `source .test_env`
- [ ] **No hardcoded ports** for test servers
- [ ] **Cleanup in defer** statements

### Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| Port conflicts | Use `testtestutils.GetFreePort()` for test servers |
| Auth failures | Run `source .test_env` before tests |
| MediaMTX not found | Start service: `systemctl start mediamtx` |
| Tests not running | Check build tags for integration tests |
| Flaky tests | Add retries or increase timeouts appropriately |

## 11. Test Categories

### Primary Classification (Required)
- **Unit**: Component isolation tests
- **Integration**: Component interaction tests
- **Performance**: Load and stress tests
- **Security**: Authentication and authorization tests

### Build Tag Usage
```go
// Unit tests - NO tags needed
package mediamtx

// Integration tests - tags REQUIRED
//go:build integration
// +build integration
package mediamtx_test

// Performance tests
//go:build performance
// +build performance
package performance_test
```

## 12. Compliance Requirements

### API Compliance Audit
Every test calling server APIs must:
1. Use exact request format from API documentation
2. Validate all documented response fields
3. Check documented error codes and messages
4. Follow documented authentication flow
5. Never test undocumented behavior

### Test Quality Gates
- **Critical Requirements**: 100% coverage required
- **High Priority Requirements**: 95% coverage required
- **Overall Coverage**: 90% coverage required
- **All tests must pass** before merge
- **No unauthorized changes** without explicit approval

---

**Quick Reference Card:**

```bash
# Setup
source .test_env

# Run unit tests
go test ./internal/mediamtx/

# Run integration tests  
go test -tags=integration ./tests/integration/

# Run with coverage
go test -cover ./internal/mediamtx/

# Run specific test
go test ./internal/mediamtx/ -run TestHealthMonitor
```

**Remember**: 
- Requirements traceability is MANDATORY
- API documentation is the source of truth
- Get authorization before changes
- Use real components, mock only external deps