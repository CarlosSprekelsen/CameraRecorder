# Go Testing Guide - MediaMTX Camera Service

**Version:** 1.0  
**Date:** 2025-01-15  
**Status:** Go Implementation Testing Guide  

## 1. Core Principles

### 🚨 **CRITICAL: STOP and Authorization Required**
- **STOP before modifying any code** - Investigate and understand the issue first
- **Ask for explicit authorization** before making any code changes
- **No working in isolation** - Coordinate with team before implementation
- **Present options and recommendations** for team decision
- **Do not create document over population** - Only create requested reports, do not offer free reports outside the chat unless requested

### 🚨 **CRITICAL: API Documentation is Ground Truth**
- **API Documentation**: `docs/api/json_rpc_methods.md` is the ONLY source of truth for API behavior
- **Health Endpoints**: `docs/api/health-endpoints.md` is the ONLY source of truth for health API
- **NEVER use server implementation as reference** - Only use documented API
- **Tests must validate against API documentation** - Not against server implementation
- **If test fails, check API documentation first** - Don't adapt test to broken implementation

### **Ground Truth Enforcement Rules**
1. **API Documentation is FROZEN** - Changes require formal approval process
2. **Server Implementation follows API Documentation** - Not the other way around
3. **Tests validate API compliance** - Not implementation details
4. **Test failures indicate API/implementation mismatch** - Not test bugs
5. **No "accommodation" of broken implementations** - Tests do not fix the implementation - it is ok if a test fails, that's their purpose, to find real bugs not accommodate them

### **Package Declaration Rules - HYBRID PATTERN**
1. **Unit Tests**: Use `package <module>` in internal packages (e.g., `package mediamtx` in `internal/mediamtx/health_monitor_test.go`)
2. **Integration Tests**: Use `package <module>_test` in external test directories (e.g., `package mediamtx_test` in `tests/integration/`)
3. **Unit tests follow Go standard**: `go test ./internal/mediamtx/` works directly
4. **Integration tests require build tags**: `go test -tags="integration" ./tests/integration/`
5. **Coverage measurement**: Unit tests use standard Go coverage, integration tests use `-coverpkg` flag

### **Go Testing Approach Explanation - HYBRID PATTERN**
**Why We Use Hybrid Testing**:
- **Unit Tests (Internal)**: Follow Go standard, easier to maintain, better IDE integration
- **Integration Tests (External)**: Keep test infrastructure separate, better organization by test type
- **Best of Both Worlds**: Standard Go practices for unit tests, organized structure for integration tests
- **Scalability**: Works better for complex projects with multiple components
- **Team Coordination**: Clear boundaries between Developer, IV&V, and PM responsibilities

**Benefits**:
- Unit tests follow Go standard: `go test ./internal/mediamtx/` works directly
- Integration tests organized by type: unit, integration, performance
- Better maintainability and developer experience
- Standard Go tooling support for unit tests

**Correct Usage Examples - HYBRID PATTERN**:
```bash
# Unit Tests (Internal) - Standard Go
go test ./internal/mediamtx/                    # All MediaMTX unit tests
go test ./internal/mediamtx/ -run TestHealth   # Specific test pattern
go test ./internal/mediamtx/ -v                 # Verbose output

# Integration Tests (External) - With build tags
go test -tags="integration" ./tests/integration/    # All integration tests
go test -tags="integration" ./tests/integration/ -run TestMediaMTX  # Specific tests

# Coverage (Unit tests use standard Go coverage)
go test -cover ./internal/mediamtx/                # Unit test coverage
go test -tags="integration" -coverpkg=./internal/mediamtx ./tests/integration/  # Integration coverage
```

**⚠️ CRITICAL: Package Declaration Rules - HYBRID PATTERN**
- **Unit Tests**: MUST use `package <module>` in internal packages (e.g., `package mediamtx`)
- **Integration Tests**: MUST use `package <module>_test` in external test directories
- **NEVER mix**: Don't put unit tests in external packages or integration tests in internal packages
- **Unit test execution**: Use standard Go: `go test ./internal/mediamtx/`
- **Integration test execution**: Use build tags: `go test -tags="integration" ./tests/integration/`

### Real System Testing Over Mocking
- **MediaMTX:** Use systemd-managed service, never mock
- **File System:** Use `tempfile`, never mock
- **WebSocket:** Use real connections within system
- **Authentication:** Use real JWT tokens with test secrets
- **API Keys:** Use test-accessible storage location (`/tmp/test_api_keys.json`)

### Strategic Mocking Rules
**MOCK:** External APIs, time operations, expensive hardware simulation  
**NEVER MOCK:** MediaMTX service, filesystem, internal WebSocket, JWT auth, config loading

### **Authorization Process**
1. **Investigate First**: Understand the issue, root cause, and impact
2. **Document Findings**: Write clear investigation report with evidence
3. **Present Options**: Provide multiple solutions with pros/cons
4. **Request Authorization**: Ask for explicit approval before implementation
5. **Wait for Approval**: Do not proceed without team authorization
6. **Implement Approved Solution**: Follow approved approach exactly
7. **Document Changes**: Update documentation and create issues as needed

## 2. Test Organization - HYBRID PATTERN GUIDELINES

### **🚨 CRITICAL: New Hybrid Testing Pattern**
**IMPLEMENTED: January 2025 - All developers must follow this pattern**

**Unit Tests (Internal)**:
- **Location**: `internal/<module>/<module>_test.go`
- **Package**: `package <module>` (same as source)
- **Execution**: `go test ./internal/mediamtx/` (standard Go)
- **Coverage**: `go test -cover ./internal/mediamtx/` (standard Go)
- **Purpose**: Test individual components in isolation

**Integration Tests (External)**:
- **Location**: `tests/integration/test_<module>_integration.go`
- **Package**: `package <module>_test` (external testing)
- **Execution**: `go test -tags="integration" ./tests/integration/`
- **Coverage**: `go test -tags="integration" -coverpkg=./internal/mediamtx ./tests/integration/`
- **Purpose**: Test component interactions and real system integration

**Why This Pattern**:
- ✅ **Unit tests follow Go standard** - Better IDE support, easier debugging
- ✅ **Integration tests organized** - Clear separation by test type
- ✅ **Maintainability** - Unit tests co-located with source code
- ✅ **Developer Experience** - Standard Go tooling for unit tests
- ✅ **Team Coordination** - Clear boundaries between test types

## 3. Test Organization - STRICT STRUCTURE GUIDELINES

### Mandatory Directory Structure - HYBRID PATTERN
```
internal/                           # Source code with unit tests
├── mediamtx/
│   ├── health_monitor.go          # Source code
│   ├── health_monitor_test.go     # Unit tests (package mediamtx)
│   ├── client.go                  # Source code
│   ├── client_test.go             # Unit tests (package mediamtx)
│   └── ...                        # Other source + unit test files

tests/                              # External integration and performance tests
├── integration/                    # End-to-end API tests
|   ├── test_mediamtx_integration.go    # MediaMTX integration (package mediamtx_test)
|   ├── test_websocket_integration.go   # WebSocket integration (package websocket_test)
|   └── test_e2e_integration.go        # End-to-end workflows (package e2e_test)
├── performance/                    # Load/stress tests for reliability
├── fixtures/                       # Test data files (.yaml, .json, mock data)
├── utils/                          # Helper functions (setupTest, generateToken, etc.)
└── tools/                          # Scripts that run tests
```

### STRICT DIRECTORY RULES

#### **PROHIBITED DIRECTORY CREATION**
- **NO subdirectories** within main test directories (unit/, integration/, etc.)
- **NO feature-specific directories** (e.g., test_camera_discovery/, test_websocket_server/)
- **NO variant directories** (e.g., real/, mock/, v2/)
- **NO temporary directories** (e.g., quarantine/, edge_cases/, e2e/)

#### **MANDATORY FLAT STRUCTURE**
- **All test files** must be directly in their primary directory
- **File naming**: `test_<feature>_<aspect>_test.go` (e.g., `test_camera_discovery_enumeration_test.go`)
- **Maximum 1 level** of test directory nesting

#### **UTILITY DIRECTORY RULES**
- **fixtures/**: Shared test fixtures, testdata files, common setup
- **utils/**: Test utilities, helpers, mock factories
- **tools/**: Test runners, orchestration scripts, automation tools

#### **ENFORCEMENT**
- **Violation**: Any new directory creation requires IV&V approval
- **Migration**: Existing subdirectories must be flattened
- **Documentation**: All structure changes must be documented

### File Organization Rules
- **One file per feature** - no variants (_real, _v2)
- **REQ-* references required** in every test file docstring
- **Shared utilities over duplication**
- **Test tools in tests/tools/** - separate from actual test files

## 3. Test Markers - COMPREHENSIVE CLASSIFICATION

### Primary Classification (Test Level) - HYBRID PATTERN
```go
// Unit Tests (Internal) - NO build tags needed
// internal/mediamtx/health_monitor_test.go
package mediamtx

func TestHealthMonitor(t *testing.T) {
    // Standard Go unit test
}

// Integration Tests (External) - Build tags required
// tests/integration/test_mediamtx_integration.go
//go:build integration
// +build integration

package mediamtx_test

func TestMediaMTXIntegration(t *testing.T) {
    // Integration test
}
```

### Secondary Classification (Test Characteristics)
```go
//go:build real_mediamtx
// +build real_mediamtx

//go:build real_websocket
// +build real_websocket

//go:build real_system
// +build real_system

//go:build sudo_required
// +build sudo_required
```

### Tertiary Classification (Test Scope)
```go
//go:build edge_case
// +build edge_case

//go:build sanity
// +build sanity

//go:build hardware
// +build hardware

//go:build network
// +build network
```

### Marker Usage Rules

#### **MANDATORY MARKERS**
- **Every test function** must have at least one primary marker
- **Real system tests** must include appropriate `real_*` marker
- **Build tags** must be defined in test configuration

#### **MARKER COMBINATIONS**
```go
//go:build unit
// +build unit

func TestFeatureBehavior(t *testing.T) {
    // Standard unit test
}

//go:build integration
//go:build real_mediamtx
// +build integration,real_mediamtx

func TestRealSystemIntegration(t *testing.T) {
    // Integration test with real system
}

//go:build performance
//go:build timeout
// +build performance,timeout

func TestLoadPerformance(t *testing.T) {
    // Performance test with timeout
}
```

#### **MARKER DEFINITION REQUIREMENTS**
- **All markers** must be defined in test configuration
- **No undefined markers** allowed in test files
- **Clear descriptions** required for each marker
- **Regular validation** of marker usage vs definition

### Go Test Configuration Alignment
```go
// test_config.go
const (
    // Primary Classification
    TestUnit       = "unit"
    TestIntegration = "integration"
    TestSecurity   = "security"
    TestPerformance = "performance"
    TestHealth     = "health"
    
    // Secondary Classification
    TestRealMediaMTX = "real_mediamtx"
    TestRealWebSocket = "real_websocket"
    TestRealSystem   = "real_system"
    TestSudoRequired = "sudo_required"
    
    // Tertiary Classification
    TestEdgeCase    = "edge_case"
    TestSanity      = "sanity"
    TestHardware    = "hardware"
    TestNetwork     = "network"
)
```

## 4. API Compliance Testing - MANDATORY

### **🚨 CRITICAL: API Documentation Compliance**
Every test that calls server APIs MUST validate against API documentation, not implementation.

### **Mandatory API Compliance Rules**
1. **Test against documented API format** - Use exact request/response formats from `json_rpc_methods.md`
2. **Validate documented error codes** - Use error codes and messages from API documentation
3. **Test documented authentication flow** - Follow authentication flow exactly as documented
4. **Verify documented response fields** - Check all required fields are present and correct
5. **No implementation-specific testing** - Don't test server internals, only documented behavior

### **API Compliance Test Template**
```go
/*
API Compliance Test for [Method Name]

API Documentation Reference: docs/api/json_rpc_methods.md
Method: [method_name]
Expected Request Format: [documented format]
Expected Response Format: [documented format]
Expected Error Codes: [documented codes]
*/

//go:build integration
// +build integration

func TestMethodNameAPICompliance(t *testing.T) {
    // 1. Use documented request format
    request := JSONRPCRequest{
        JSONRPC: "2.0",
        Method:  "[method_name]",
        Params: map[string]interface{}{
            // Use exact parameter names from API documentation
        },
        ID: 1,
    }
    
    // 2. Validate documented response format
    response, err := sendRequest(request)
    require.NoError(t, err)
    
    // 3. Check all documented fields are present
    require.Contains(t, response, "result", "Response must contain 'result' field per API documentation")
    result := response["result"]
    
    // 4. Validate documented response structure
    requiredFields := []string{"field1", "field2"} // From API documentation
    for _, field := range requiredFields {
        require.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
    }
    
    // 5. Validate documented error handling
    // Test error cases exactly as documented
}
```

## 5. Requirements Traceability

### Mandatory Format for Test Files
```go
/*
Module description.

Requirements Coverage:
- REQ-XXX-001: Requirement description
- REQ-XXX-002: Additional requirement

Test Categories: Unit/Integration/Security/Performance/Health
API Documentation Reference: docs/api/json_rpc_methods.md
*/

//go:build unit
// +build unit

func TestFeatureBehaviorReqXXX001(t *testing.T) {
    // REQ-XXX-001: Specific requirement validation
    // Test that would FAIL if requirement violated
}
```

### Requirements Coverage Analysis
- **Location**: `docs/testing/requirements_coverage_analysis.md`
- **Purpose**: Track coverage against frozen baseline (161 requirements)
- **Focus**: Critical and high-priority requirements
- **Updates**: After major test changes or baseline updates

### Coverage Categories
- **Critical Requirements**: 45 requirements (93% covered)
- **High Priority Requirements**: 67 requirements (85% covered)
- **Overall Coverage**: 85% (137/161 requirements)

## 6. Test Tools and Runners

### Test Tools Location
All test runners and utilities are located in `tests/tools/`:
- **Not test files** - they orchestrate test execution
- **No requirements coverage** - they don't validate requirements directly
- **Script conventions** - follow tool documentation standards
- **Documentation**: `tests/tools/README.md`

### Available Tools
- `run_all_tests.sh`: Comprehensive test automation with quality gates
- `run_tests.sh`: Basic test runner with Go test integration
- `run_individual_tests.sh`: Individual test execution with failure categorization
- `run_critical_error_tests.sh`: Critical error handling test runner
- `run_integration_tests.sh`: Real system integration test runner
- `setup_test_environment.sh`: Test environment setup
- `validate_test_environment.sh`: Environment validation

### **🚨 MANDATORY: Build Tags by Test Type - HYBRID PATTERN**

**Unit Tests (Internal) - NO build tags needed:**
```bash
# CORRECT - Standard Go unit tests
go test ./internal/mediamtx/                    # All MediaMTX unit tests
go test ./internal/mediamtx/ -run TestHealth   # Specific test pattern
go test -cover ./internal/mediamtx/             # With coverage

# WRONG - Don't use build tags for unit tests
go test -tags="unit" ./internal/mediamtx/      # ❌ Unnecessary
```

**Integration Tests (External) - Build tags REQUIRED:**
```bash
# REQUIRED - External testing needs build tags
go test -tags="integration" ./tests/integration/                    # All integration tests
go test -tags="integration" ./tests/integration/ -run TestMediaMTX  # Specific tests
go test -tags="integration" -coverpkg=./internal/mediamtx ./tests/integration/  # With coverage

# WRONG - This will exclude all integration test files
go test ./tests/integration/                    # ❌ No tests will run
```

**Why:** External testing uses build constraints that exclude files without matching tags.

### Usage Guidelines
```bash
# For most testing needs, use go test directly
go test -tags="unit" ./tests/unit/
go test -tags="integration,real_system" ./tests/integration/
go test -tags="performance" ./tests/performance/

# Use tools only for specialized orchestration
./tests/tools/run_all_tests.sh
./tests/tools/run_critical_error_tests.sh

# Clear cache if tests behave unexpectedly
go clean -cache
```

## 7. Performance Targets

- **Unit tests:** <30 seconds total
- **Integration tests:** <5 minutes total  
- **Full suite:** <10 minutes total
- **Flaky rate:** <1%

## 8. Standard Patterns

### MediaMTX Integration
```go
//go:build integration && real_mediamtx
// +build integration,real_mediamtx

func TestStreamCreation(t *testing.T) {
    controller := NewMediaMTXController("http://localhost:9997")
    streamID, err := controller.CreateStream("test", "/dev/video0")
    require.NoError(t, err)
    assert.NotEmpty(t, streamID)
}
```

### Authentication Testing
```go
//go:build security
// +build security

func TestValidAuth(t *testing.T) {
    token := generateValidTestToken("test_user", "operator")
    // Test with real JWT token
}
```

### WebSocket Testing
```go
//go:build integration
//go:build real_websocket
// +build integration,real_websocket

func TestWebSocketConnection(t *testing.T) {
    conn, err := websocket.Dial("ws://localhost:8002/ws", "", "http://localhost")
    require.NoError(t, err)
    defer conn.Close()
    
    // Test WebSocket communication
}
```

### Test Environment Configuration
**CRITICAL**: Always source the test environment before running tests:
```bash
source .test_env
```

**Required Environment Variables:**
- `CAMERA_SERVICE_JWT_SECRET`: Test JWT secret for authentication
- `CAMERA_SERVICE_API_KEYS_PATH`: Test API key storage location (`/tmp/test_api_keys.json`)

**Why This Matters:**
- Tests run as regular user, not `camera-service` user
- Production API key storage (`/opt/camera-service/keys/`) requires elevated permissions
- Test environment redirects to user-accessible location (`/tmp/`)
- Without this configuration, 90% of tests will fail with authentication errors


## 9. Documentation Standards

### Test File Documentation
- **Requirements Coverage**: Mandatory in every test file docstring
- **Test Categories**: Unit/Integration/Security/Performance/Health
- **Real Component Usage**: Document when real components are used

### Tool Documentation
- **Purpose**: What the tool does, not requirements coverage
- **Usage**: Command-line examples and options
- **Location**: `tests/tools/README.md`

### Coverage Analysis
- **Location**: `docs/testing/requirements_coverage_analysis.md`
- **Updates**: After major test changes
- **Focus**: Critical and high-priority requirements gaps

## 11. Compliance and Validation

### **🚨 MANDATORY: API Compliance**
Every test that calls server APIs MUST be audited against API documentation.

### **Audit Requirements**
1. **Pre-commit Audit**: All API tests must be validated against `json_rpc_methods.md`
2. **Response Format Validation**: Verify all response fields match API documentation
3. **Error Code Validation**: Verify error codes and messages match API documentation
4. **Authentication Flow Validation**: Verify authentication follows documented flow
5. **Parameter Validation**: Verify parameter names and types match API documentation

### Testing Guide Compliance
- **Test Files**: Must follow requirements traceability format
- **Test Tools**: Must follow script conventions (no requirements coverage)
- **Coverage Analysis**: Must be updated after major changes
- **Directory Structure**: Must follow strict structure guidelines
- **Markers**: Must be properly defined and used
- **API Compliance**: Must validate against API documentation

### Quality Gates
- **Authorization Required**: All code changes must be explicitly authorized
- **Critical Requirements**: 100% coverage required
- **High Priority Requirements**: 95% coverage required
- **Overall Coverage**: 90% coverage required
- **Performance Testing**: Must be implemented for critical requirements
- **Structure Compliance**: No unauthorized directory creation
- **Marker Compliance**: All markers defined and properly used
- **API Compliance**: All tests must validate against API documentation

## 12. Migration to Hybrid Pattern - COMPLETED

### **✅ IMPLEMENTED: January 2025 - Hybrid Testing Pattern**
**All developers must now follow this pattern - old patterns are deprecated**

#### **What Was Changed:**
1. **Unit Tests Moved**: From `tests/unit/` to `internal/<module>/<module>_test.go`
2. **Package Declarations**: Unit tests now use `package <module>` (standard Go)
3. **Build Tags Removed**: Unit tests no longer need build tags
4. **Test Execution**: Unit tests use standard Go: `go test ./internal/mediamtx/`
5. **Old Test Files**: All old scattered test files have been removed and consolidated

#### **Current Structure:**
```
internal/mediamtx/                    # Source code + unit tests
├── health_monitor.go                # Source
├── health_monitor_test.go           # Unit tests (package mediamtx)
├── client.go                        # Source  
├── client_test.go                   # Unit tests (package mediamtx)
├── errors.go                        # Source
├── errors_test.go                   # Unit tests (package mediamtx)
├── manager_test.go                  # Unit tests (package mediamtx)
└── ...                              # Other source files

tests/                                # External integration tests only
├── integration/                     # Integration tests (package *_test)
│   ├── test_mediamtx_integration.go # MediaMTX integration
│   └── ...                          # Other integration tests
├── performance/                     # Performance tests
└── ...                              # Other test types
```

#### **Test Execution Commands:**
```bash
# Unit Tests (Internal) - Standard Go
go test ./internal/mediamtx/                    # All MediaMTX unit tests
go test ./internal/mediamtx/ -run TestHealth   # Specific test pattern
go test -cover ./internal/mediamtx/             # With coverage

# Integration Tests (External) - With build tags
go test -tags="integration" ./tests/integration/    # All integration tests
go test -tags="integration" ./tests/integration/ -run TestMediaMTX  # Specific tests
```

#### **Benefits Achieved:**
- ✅ **Standard Go Compliance**: Unit tests follow Go conventions
- ✅ **Better IDE Support**: Standard Go tooling works directly
- ✅ **Easier Debugging**: Unit tests co-located with source
- ✅ **Cleaner Structure**: No more scattered test files
- ✅ **Maintainability**: Single test file per component
- ✅ **Developer Experience**: Standard `go test` commands work

## 13. Mandatory IV&V Validation Checklist

### **🚨 ARCHITECTURE COMPLIANCE VALIDATION**
**Before any IV&V approval, MUST validate:**

□ **Single Responsibility Principle**: Each component has one clear purpose
□ **No Duplicate Implementations**: Reuses existing logger, config, utilities
□ **Proper Dependency Injection**: No public methods created for testing shortcuts
□ **Architecture Integration**: Follows documented patterns and interfaces
□ **Component Boundaries**: Respects existing component responsibilities

### **🚨 TEST QUALITY VALIDATION**  
**Before any IV&V approval, MUST validate:**

□ **Requirements-Based Testing**: Tests validate requirements, not implementation details
□ **Error Detection Design**: Tests designed to catch errors, not just pass
□ **Real Functionality Testing**: Tests exercise real behavior, minimal mocking
□ **Failure Conditions**: Tests validate error handling and edge cases
□ **Integration Testing**: Tests verify component interactions work properly

### **🚨 TECHNICAL DEBT ASSESSMENT**
**Before any IV&V approval, MUST assess:**

□ **Architecture Violations**: Document any departures from established patterns
□ **Code Quality Issues**: Identify maintainability and readability problems
□ **Integration Risks**: Assess impact on existing system components
□ **Technical Debt Quantification**: Measure and report debt accumulation
□ **Remediation Requirements**: Identify what must be fixed before progression

### **IV&V FAILURE CONDITIONS**
**MUST fail validation if:**
❌ Architecture violations found without justified waivers
❌ Duplicate implementations created instead of reusing existing components  
❌ Tests designed to pass rather than validate requirements
❌ Technical debt above established thresholds
❌ Missing requirements traceability documentation


**Status**: **CREATED** - Go implementation testing guide with strict structure guidelines, comprehensive markers section, enhanced compliance requirements, and critical authorization rules.
