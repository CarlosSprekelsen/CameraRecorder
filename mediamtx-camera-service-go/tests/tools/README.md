# Test Tools - MediaMTX Camera Service Go

**Location**: `tests/tools/`  
**Purpose**: Test orchestration and automation tools  
**Status**: Test execution utilities following testing guidelines  

## Available Tools

### Core Test Runners

#### `run_tests.sh` (Main Runner)
**Purpose**: Main test runner that delegates to specialized runners  
**Usage**: 
```bash
./tests/tools/run_tests.sh [category]
```
**Examples**:
```bash
./tests/tools/run_tests.sh unit        # Uses run_unit_tests.sh
./tests/tools/run_tests.sh integration # Uses run_integration_tests.sh
./tests/tools/run_tests.sh all         # Runs all categories in sequence
```

**Features**:
- Delegates to specialized runners for unit and integration tests
- Runs security, performance, and health tests directly
- Coordinates test execution sequence
- Provides unified interface for all test categories

#### `run_unit_tests.sh` (Specialized Unit Runner)
**Purpose**: Run unit tests with proper coverage measurement following testing guidelines  
**Usage**: 
```bash
./tests/tools/run_unit_tests.sh
```

**Following Testing Guidelines**:
- ✅ Uses `-coverpkg` flag for cross-package coverage measurement
- ✅ Tests each package individually to avoid package conflicts
- ✅ Generates coverage profiles per package
- ✅ Enforces 90% coverage threshold per guidelines
- ✅ Follows external testing pattern (`package *_test`)
- ✅ **NEW: Enables parallel test execution for faster CI/CD**

**Coverage Measurement**:
```bash
# Tests each package individually with proper flags and parallel execution
go test -tags="unit" -coverpkg="./internal/websocket" -parallel 4 ./tests/unit/test_websocket_*.go
go test -tags="unit" -coverpkg="./internal/mediamtx" -parallel 4 ./tests/unit/test_mediamtx_*.go
go test -tags="unit" -coverpkg="./internal/config" -parallel 4 ./tests/unit/test_config_*.go
# ... and more packages
```

**Output**:
- Individual coverage files per package
- Combined coverage report
- HTML coverage visualization
- Coverage threshold validation (90% required)

**Parallel Execution**:
- **Unit Tests**: `-parallel 4` for maximum concurrency (safe due to isolated test environments)
- **Integration Tests**: `-parallel 2` for conservative concurrency (real services may have port conflicts)
- **Performance Impact**: 2-4x faster test execution on multi-core systems
- **Coverage Maintained**: All coverage measurement flags preserved

#### `run_integration_tests.sh` (Specialized Integration Runner)
**Purpose**: Run integration tests with real system testing following testing guidelines  
**Usage**: 
```bash
./tests/tools/run_integration_tests.sh
```

**Following Testing Guidelines**:
- ✅ Uses real system testing over mocking
- ✅ Tests end-to-end workflows
- ✅ Validates API compliance against documentation
- ✅ Tests real MediaMTX service, filesystem, WebSocket connections
- ✅ No "invented fixes" like timeouts

**Real System Testing**:
- **MediaMTX**: Uses systemd-managed service, never mock
- **File System**: Uses real filesystem, never mock
- **WebSocket**: Uses real connections within system
- **Authentication**: Uses real JWT tokens with test secrets

**Parallel Execution Strategy**:
- **Conservative Approach**: Uses `-parallel 2` to avoid port conflicts with real services
- **Service Isolation**: Each test gets unique temporary directories and resources
- **Conflict Prevention**: Tests are designed to avoid shared resource contention

**Test Categories**:
- Standard integration tests
- Quarantined tests (if any)
- End-to-end workflow tests
- API compliance validation

**Service Validation**:
- Checks MediaMTX service status
- Validates WebSocket port availability
- Confirms camera device presence
- Ensures test environment readiness

### Legacy Test Runners

#### `run_all_tests.sh`
**Purpose**: Comprehensive test automation with quality gates  
**Status**: Legacy - use `run_tests.sh all` instead

#### `run_individual_tests.sh`
**Purpose**: Individual test execution with failure categorization  
**Status**: Legacy - use specialized runners instead

#### `run_critical_error_tests.sh`
**Purpose**: Critical error handling test runner  
**Status**: Legacy - use specialized runners instead

#### `run_integration_tests.sh` (Old)
**Purpose**: Real system integration test runner  
**Status**: Replaced by new `run_integration_tests.sh`

## Usage Guidelines

### Standard Testing Workflow
1. **Setup Environment**: `./tests/tools/setup_test_environment.sh`
2. **Run Unit Tests**: `./tests/tools/run_tests.sh unit`
3. **Run Integration Tests**: `./tests/tools/run_tests.sh integration`
4. **Run All Tests**: `./tests/tools/run_tests.sh all`

### For Development
```bash
# Quick unit tests with coverage measurement
./tests/tools/run_tests.sh unit

# Integration tests with real system
./tests/tools/run_tests.sh integration

# Full test suite before commit
./tests/tools/run_tests.sh all
```

### For CI/CD
```bash
# Automated testing in CI
./tests/tools/setup_test_environment.sh
./tests/tools/run_tests.sh all
```

## Coverage Measurement

### Unit Test Coverage
- **Per Package**: Individual coverage files for each internal package
- **Combined Report**: Overall coverage across all packages
- **Threshold**: 90% coverage required per guidelines
- **Output**: Coverage files in `coverage/unit/` directory

### Integration Test Coverage
- **Real System**: Coverage of real component interactions
- **API Compliance**: Validation against API documentation
- **End-to-End**: Complete workflow coverage
- **Output**: Coverage files in `coverage/integration/` directory

## Tool Conventions

### Script Standards
- **Shebang**: All scripts use `#!/bin/bash`
- **Error Handling**: Scripts exit on first error (`set -e`)
- **Logging**: Consistent logging format across all tools
- **Documentation**: Each script includes usage and purpose

### Environment Requirements
- **Test Environment**: Must source `.test_env` before running
- **Permissions**: Some tools require specific user permissions
- **Services**: MediaMTX service must be running for integration tests

### Output Standards
- **Success**: Exit code 0 with clear success message
- **Failure**: Non-zero exit code with detailed error information
- **Reports**: Generate coverage and quality reports in standard formats

## Testing Guidelines Compliance

### ✅ What These Tools Follow
- **External Testing**: Uses `package *_test` pattern
- **Coverage Measurement**: Proper `-coverpkg` flag usage
- **Real System Testing**: No mocking of core components
- **API Documentation**: Validates against ground truth
- **Coverage Thresholds**: Enforces 90% requirement

### ❌ What These Tools Avoid
- **Package Conflicts**: Tests packages individually
- **Mocking**: Uses real components where guidelines specify
- **Coverage Hiding**: No artificial test passing
- **Guideline Violations**: Follows testing guide exactly

## Notes
- **Not Test Files**: These tools orchestrate test execution, they don't validate requirements directly
- **No Requirements Coverage**: Tools focus on test execution, not requirements validation
- **Documentation**: Each tool is documented with purpose and usage examples
- **Guidelines Compliance**: All tools follow testing guidelines exactly as specified
