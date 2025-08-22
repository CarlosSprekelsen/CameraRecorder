# Server Test Guide - MediaMTX Camera Service

**Version:** 4.0  
**Date:** 2025-01-06  
**Status:** Updated with strict structure guidelines and comprehensive markers section  

## 1. Core Principles

### Real System Testing Over Mocking
- **MediaMTX:** Use systemd-managed service, never mock
- **File System:** Use `tempfile`, never mock
- **WebSocket:** Use real connections within system
- **Authentication:** Use real JWT tokens with test secrets
- **API Keys:** Use test-accessible storage location (`/tmp/test_api_keys.json`)

### Strategic Mocking Rules
**MOCK:** External APIs, time operations, expensive hardware simulation  
**NEVER MOCK:** MediaMTX service, filesystem, internal WebSocket, JWT auth, config loading

## 2. Test Organization - STRICT STRUCTURE GUIDELINES

### Mandatory Directory Structure
```
tests/
├── unit/                   # Unit tests (<30 seconds total)
├── integration/            # Integration tests (<5 minutes total)
├── security/              # Security tests
├── performance/           # Performance and load tests
├── health/                # Health monitoring tests
├── fixtures/              # Shared test fixtures and utilities
├── utils/                 # Test utilities and helpers
└── tools/                 # Test runners and orchestration tools
```

### STRICT DIRECTORY RULES

#### **PROHIBITED DIRECTORY CREATION**
- **NO subdirectories** within main test directories (unit/, integration/, etc.)
- **NO feature-specific directories** (e.g., test_camera_discovery/, test_websocket_server/)
- **NO variant directories** (e.g., real/, mock/, v2/)
- **NO temporary directories** (e.g., quarantine/, edge_cases/, e2e/)

#### **MANDATORY FLAT STRUCTURE**
- **All test files** must be directly in their primary directory
- **File naming**: `test_<feature>_<aspect>.py` (e.g., `test_camera_discovery_enumeration.py`)
- **Maximum 1 level** of test directory nesting

#### **UTILITY DIRECTORY RULES**
- **fixtures/**: Shared test fixtures, conftest.py files, common setup
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

### Primary Classification (Test Level)
```python
@pytest.mark.unit          # Unit-level tests (<30s)
@pytest.mark.integration   # Integration tests (<5min)
@pytest.mark.security      # Security validation tests
@pytest.mark.performance   # Performance and load tests
@pytest.mark.health        # Health monitoring tests
```

### Secondary Classification (Test Characteristics)
```python
@pytest.mark.asyncio       # Async test functions
@pytest.mark.timeout       # Tests with specific timeouts
@pytest.mark.slow          # Long-running tests
@pytest.mark.real_mediamtx # Requires real MediaMTX service
@pytest.mark.real_websocket # Real WebSocket connections
@pytest.mark.real_system   # Real system integration
@pytest.mark.sudo_required # Requires elevated privileges
```

### Tertiary Classification (Test Scope)
```python
@pytest.mark.edge_case     # Edge case testing
@pytest.mark.sanity        # Basic functionality validation
@pytest.mark.hardware      # Hardware-dependent tests (mocked)
@pytest.mark.network       # Network-dependent tests (mocked)
```

### Marker Usage Rules

#### **MANDATORY MARKERS**
- **Every test function** must have at least one primary marker
- **Async tests** must include `@pytest.mark.asyncio`
- **Real system tests** must include appropriate `real_*` marker

#### **MARKER COMBINATIONS**
```python
# Standard unit test
@pytest.mark.unit
def test_feature_behavior():
    pass

# Async integration test with real system
@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_mediamtx
async def test_real_system_integration():
    pass

# Performance test with timeout
@pytest.mark.performance
@pytest.mark.timeout(300)
def test_load_performance():
    pass
```

#### **MARKER DEFINITION REQUIREMENTS**
- **All markers** must be defined in `pytest.ini`
- **No undefined markers** allowed in test files
- **Clear descriptions** required for each marker
- **Regular validation** of marker usage vs definition

### Pytest Configuration Alignment
```ini
# pytest.ini markers section
markers =
    # Primary Classification
    unit: unit-level tests
    integration: integration-level tests
    security: security-focused tests
    performance: performance and load tests
    health: health monitoring tests
    
    # Secondary Classification
    asyncio: async test functions
    timeout: tests with specific timeouts
    slow: long-running tests
    real_mediamtx: requires real MediaMTX service
    real_websocket: real WebSocket connections
    real_system: real system integration
    sudo_required: requires elevated privileges
    
    # Tertiary Classification
    edge_case: edge case testing
    sanity: basic functionality validation
    hardware: hardware-dependent tests (mocked)
    network: network-dependent tests (mocked)
```

## 4. Requirements Traceability

### Mandatory Format for Test Files
```python
"""
Module description.

Requirements Coverage:
- REQ-XXX-001: Requirement description
- REQ-XXX-002: Additional requirement

Test Categories: Unit/Integration/Security/Performance/Health
"""

@pytest.mark.unit
def test_feature_behavior_req_xxx_001(self):
    """REQ-XXX-001: Specific requirement validation."""
    # Test that would FAIL if requirement violated
```

### Requirements Coverage Analysis
- **Location**: `docs/test/requirements_coverage_analysis.md`
- **Purpose**: Track coverage against frozen baseline (161 requirements)
- **Focus**: Critical and high-priority requirements
- **Updates**: After major test changes or baseline updates

### Coverage Categories
- **Critical Requirements**: 45 requirements (93% covered)
- **High Priority Requirements**: 67 requirements (85% covered)
- **Overall Coverage**: 85% (137/161 requirements)

## 5. Test Tools and Runners

### Test Tools Location
All test runners and utilities are located in `tests/tools/`:
- **Not test files** - they orchestrate test execution
- **No requirements coverage** - they don't validate requirements directly
- **Script conventions** - follow tool documentation standards
- **Documentation**: `tests/tools/README.md`

### Available Tools
- `run_all_tests.py`: Comprehensive test automation with quality gates
- `run_tests.py`: Basic test runner with pytest integration
- `run_individual_tests.py`: Individual test execution with failure categorization
- `run_critical_error_tests.py`: Critical error handling test runner
- `run_integration_tests.py`: Real system integration test runner
- `setup_test_environment.py`: Test environment setup
- `validate_test_environment.py`: Environment validation

### Usage Guidelines
```bash
# For most testing needs, use pytest directly
pytest
pytest -m unit
pytest -m integration

# Use tools only for specialized orchestration
python3 tests/tools/run_all_tests.py
python3 tests/tools/run_critical_error_tests.py
```

## 6. Performance Targets

- **Unit tests:** <30 seconds total
- **Integration tests:** <5 minutes total  
- **Full suite:** <10 minutes total
- **Flaky rate:** <1%

## 7. Standard Patterns

### MediaMTX Integration
```python
@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_mediamtx
async def test_stream_creation():
    controller = MediaMTXController("http://localhost:9997")
    stream_id = await controller.create_stream("test", "/dev/video0")
    assert stream_id is not None
```

### Authentication Testing
```python
@pytest.mark.security
@pytest.mark.asyncio
async def test_valid_auth():
    token = generate_valid_test_token("test_user", "operator")
    # Test with real JWT token
```

**Note**: Source `.test_env` before running tests to provide JWT token and API key storage environment variables.

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

**Deployment Script Protection:**
- ✅ **`deploy.sh`**: Modified to preserve `CAMERA_SERVICE_API_KEYS_PATH` when updating JWT secrets
- ✅ **`setup_test_environment.py`**: Modified to include `CAMERA_SERVICE_API_KEYS_PATH` in generated files
- ✅ **Test Environment**: Automatically maintained across deployments

## 8. Quality Assurance

### Requirements Coverage Monitoring
- **Baseline**: 161 requirements (frozen ground truth)
- **Target**: 100% coverage for critical requirements
- **Current**: 93% critical requirements coverage
- **Gaps**: Performance testing (67% coverage) - **IMMEDIATE PRIORITY**

### Critical Gaps Identified
1. **Performance Testing**: 3 critical requirements missing
2. **API Method Coverage**: 9 high-priority methods missing
3. **Notification Testing**: 4 high-priority requirements missing

### Improvement Priorities
1. **Phase 1**: Implement performance test suite (CRITICAL)
2. **Phase 2**: Complete API method coverage (HIGH)
3. **Phase 3**: Add notification testing (HIGH)

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
- **Location**: `docs/test/requirements_coverage_analysis.md`
- **Updates**: After major test changes
- **Focus**: Critical and high-priority requirements gaps

## 10. Compliance and Validation

### Testing Guide Compliance
- **Test Files**: Must follow requirements traceability format
- **Test Tools**: Must follow script conventions (no requirements coverage)
- **Coverage Analysis**: Must be updated after major changes
- **Directory Structure**: Must follow strict structure guidelines
- **Markers**: Must be properly defined and used

### Quality Gates
- **Critical Requirements**: 100% coverage required
- **High Priority Requirements**: 95% coverage required
- **Overall Coverage**: 90% coverage required
- **Performance Testing**: Must be implemented for critical requirements
- **Structure Compliance**: No unauthorized directory creation
- **Marker Compliance**: All markers defined and properly used

---

**Status**: **UPDATED** - Reflects strict structure guidelines, comprehensive markers section, and enhanced compliance requirements.