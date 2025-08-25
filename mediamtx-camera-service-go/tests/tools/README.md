# Test Tools - MediaMTX Camera Service Go

**Location**: `tests/tools/`  
**Purpose**: Test orchestration and automation tools  
**Status**: Test execution utilities  

## Available Tools

### Core Test Runners

#### `run_all_tests.sh`
**Purpose**: Comprehensive test automation with quality gates  
**Usage**: 
```bash
./tests/tools/run_all_tests.sh
```
**Features**:
- Runs all test categories (unit, integration, security, performance, health)
- Enforces quality gates and performance targets
- Generates coverage reports
- Validates API compliance

#### `run_tests.sh`
**Purpose**: Basic test runner with Go test integration  
**Usage**:
```bash
./tests/tools/run_tests.sh [category]
```
**Examples**:
```bash
./tests/tools/run_tests.sh unit
./tests/tools/run_tests.sh integration
./tests/tools/run_tests.sh security
```

#### `run_individual_tests.sh`
**Purpose**: Individual test execution with failure categorization  
**Usage**:
```bash
./tests/tools/run_individual_tests.sh [test_file]
```
**Features**:
- Runs specific test files
- Categorizes failures by type
- Provides detailed error reporting

### Specialized Test Runners

#### `run_critical_error_tests.sh`
**Purpose**: Critical error handling test runner  
**Usage**:
```bash
./tests/tools/run_critical_error_tests.sh
```
**Features**:
- Tests error handling scenarios
- Validates error codes and messages
- Ensures graceful failure handling

#### `run_integration_tests.sh`
**Purpose**: Real system integration test runner  
**Usage**:
```bash
./tests/tools/run_integration_tests.sh
```
**Features**:
- Tests with real MediaMTX service
- Validates system integration
- Ensures real component compatibility

### Environment Management

#### `setup_test_environment.sh`
**Purpose**: Test environment setup  
**Usage**:
```bash
./tests/tools/setup_test_environment.sh
```
**Features**:
- Creates test environment configuration
- Sets up test API keys and JWT secrets
- Configures test directories

#### `validate_test_environment.sh`
**Purpose**: Environment validation  
**Usage**:
```bash
./tests/tools/validate_test_environment.sh
```
**Features**:
- Validates test environment configuration
- Checks required services and permissions
- Ensures test readiness

## Usage Guidelines

### Standard Testing Workflow
1. **Setup Environment**: `./tests/tools/setup_test_environment.sh`
2. **Validate Environment**: `./tests/tools/validate_test_environment.sh`
3. **Run Tests**: Use appropriate test runner based on needs
4. **Review Results**: Check coverage and quality gate results

### For Development
```bash
# Quick unit tests during development
go test -tags=unit ./...

# Integration tests with real system
./tests/tools/run_integration_tests.sh

# Full test suite before commit
./tests/tools/run_all_tests.sh
```

### For CI/CD
```bash
# Automated testing in CI
./tests/tools/setup_test_environment.sh
./tests/tools/run_all_tests.sh
```

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

## Notes
- **Not Test Files**: These tools orchestrate test execution, they don't validate requirements directly
- **No Requirements Coverage**: Tools focus on test execution, not requirements validation
- **Documentation**: Each tool is documented with purpose and usage examples
