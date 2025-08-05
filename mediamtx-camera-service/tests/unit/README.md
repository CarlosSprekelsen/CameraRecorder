# Unit Tests - MediaMTX Camera Service

**Version:** 1.0  
**Date:** 2025-08-05  
**Status:** Approved  
**Related Stories:** S3, S4, S5, S14  

## Overview

This directory contains unit tests for the MediaMTX Camera Service project. Unit tests focus on isolated testing of individual modules and components with minimal external dependencies. All tests follow the project's coding standards and architectural principles defined in `docs/development/principles.md` and `docs/development/coding-standards.md`.

**Testing Framework:** pytest with async support (pytest-asyncio)  
**Coverage Target:** 80% minimum for critical modules  
**Professional Standards:** No emojis, structured output, canonical TODO format  

## Directory Structure

```
tests/unit/
├── test_camera_discovery/     # Camera discovery and monitoring module tests
├── test_camera_service/       # Service manager and core service logic tests  
├── test_mediamtx_wrapper/     # MediaMTX integration and stream management tests
├── test_websocket_server/     # WebSocket JSON-RPC API and client management tests
├── test_common/               # Shared utilities and common types tests
└── README.md                  # This documentation file
```

### Directory Purposes

#### `test_camera_discovery/`
**Purpose:** Tests for camera discovery, monitoring, and capability detection  
**Key Modules:**
- `test_hybrid_monitor.py` - Camera discovery and event monitoring
- `test_capability_detection.py` - V4L2 capability probing and validation
- `test_udev_integration.py` - udev event processing and device management
- `test_hybrid_monitor_reconciliation.py` - Component integration validation

**Story Coverage:** S3 (Camera Discovery & Monitoring)

#### `test_camera_service/`
**Purpose:** Tests for service manager, configuration, logging, and startup logic  
**Key Modules:**
- `test_service_manager_lifecycle.py` - Service lifecycle and component orchestration
- `test_main_startup.py` - Application startup and shutdown logic
- `test_config_manager.py` - Configuration loading, validation, and hot reload
- `test_logging_config.py` - Logging setup, formatters, and correlation IDs

**Story Coverage:** S3, S14 (Service Management & Testing Infrastructure)

#### `test_mediamtx_wrapper/`
**Purpose:** Tests for MediaMTX integration, stream management, and media operations  
**Key Modules:**
- `test_controller.py` - MediaMTX API client and stream operations
- `test_health_monitor.py` - Health monitoring and recovery logic
- `test_stream_management.py` - Stream lifecycle and configuration

**Story Coverage:** S4 (MediaMTX Integration)

#### `test_websocket_server/`
**Purpose:** Tests for WebSocket JSON-RPC API and client management  
**Key Modules:**
- `test_server_method_handlers.py` - JSON-RPC method implementations
- `test_client_management.py` - Client connection and subscription handling
- `test_notification_delivery.py` - Real-time notification broadcasting

**Story Coverage:** S3, S4, S5 (API Layer & Integration)

#### `test_common/`
**Purpose:** Tests for shared utilities, types, and cross-module functionality  
**Key Modules:**
- `test_types.py` - Common data types and enums
- `test_retry.py` - Retry logic and backoff utilities
- `test_validation.py` - Input validation and sanitization

**Story Coverage:** S14 (Testing Infrastructure)

## Naming Conventions

### File Naming
- **Test files:** `test_<module_name>.py` or `test_<behavior_area>.py`
- **Use snake_case** throughout (no CamelCase or kebab-case)
- **Descriptive names** that clearly indicate what is being tested

**Examples:**
- `test_hybrid_monitor.py` - Tests hybrid_monitor module
- `test_service_manager_lifecycle.py` - Tests service lifecycle behavior
- `test_capability_detection_varied_formats.py` - Tests specific capability behavior

### Test Function Naming
- **Format:** `test_<behavior>_<scenario>()`
- **Be descriptive** about what behavior is being validated
- **Include success/failure scenario** in the name when relevant

**Examples:**
- `test_camera_discovery_successful_detection()`
- `test_config_loading_with_missing_file_uses_defaults()`
- `test_websocket_notification_delivery_with_client_disconnect()`

### Test Class Naming
- **Format:** `TestModuleName` or `TestBehaviorArea`
- **Use CamelCase** for test class names
- **Group related test scenarios** under logical test classes

**Examples:**
- `TestServiceManagerLifecycle`
- `TestCapabilityDetection`
- `TestConfigurationLoading`

## Story-to-Test Mapping

### S3: Camera Discovery & Monitoring
**Test Coverage Areas:**
- Camera device discovery and udev event processing
- Capability detection and metadata validation  
- Camera status tracking and state management
- Error handling and device reconnection scenarios

**Key Test Files:**
- `test_camera_discovery/test_hybrid_monitor.py`
- `test_camera_discovery/test_capability_detection.py`
- `test_camera_service/test_service_manager_lifecycle.py` (camera events)

**Completion Checklist:**
- [ ] Camera discovery event flow validated
- [ ] Capability detection accuracy confirmed
- [ ] Device metadata reconciliation tested
- [ ] Error recovery scenarios covered

### S4: MediaMTX Integration
**Test Coverage Areas:**
- MediaMTX API client operations (stream creation, health checks)
- Recording and snapshot file operations
- Stream URL generation and validation
- Health monitoring and recovery logic

**Key Test Files:**
- `test_mediamtx_wrapper/test_controller.py`
- `test_mediamtx_wrapper/test_health_monitor.py`
- `test_websocket_server/test_server_method_handlers.py` (media operations)

**Completion Checklist:**
- [ ] Stream creation and teardown validated
- [ ] Recording duration accuracy confirmed
- [ ] Snapshot capture completeness verified
- [ ] Health monitor recovery tested

### S5: Core Integration IV&V
**Test Coverage Areas:**
- End-to-end flow validation (camera → MediaMTX → WebSocket)
- Component coordination and orchestration
- Error propagation and recovery across boundaries
- Real-time notification delivery and schema compliance

**Key Test Files:**
- `tests/ivv/test_integration_smoke.py` (integration tests, not unit tests)
- Unit tests that validate component interfaces for integration

**Completion Checklist:**
- [ ] End-to-end happy path validated
- [ ] Error recovery across components tested
- [ ] Notification schema compliance verified
- [ ] Resource cleanup on shutdown confirmed

### S14: Automated Testing & CI
**Test Coverage Areas:**
- Test execution infrastructure and quality gates
- Coverage measurement and reporting
- Linting, formatting, and type checking integration
- Test organization and maintainability

**Key Test Files:**
- All test files contribute to S14 completion
- `test_camera_service/test_main_startup.py` (service lifecycle)
- `test_camera_service/test_config_manager.py` (configuration testing)
- `test_camera_service/test_logging_config.py` (logging infrastructure)

**Completion Checklist:**
- [ ] Unit test suite execution passes
- [ ] Coverage thresholds met (80% minimum)
- [ ] Quality gates (lint, format, type check) pass
- [ ] Test documentation complete

## Adding New Tests

### Creating New Test Files

1. **Determine correct subdirectory** based on module being tested
2. **Create test file** following naming conventions
3. **Include required imports** and fixtures
4. **Add descriptive docstring** explaining test purpose
5. **Use canonical TODO format** for placeholders

**Template Structure:**
```python
"""
Unit tests for [module_name] [brief_description].

Tests [specific_behavior] functionality including [key_aspects]
as specified in the architecture.
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch

from src.module_path.target_module import TargetClass


class TestTargetClass:
    """Test [TargetClass] [behavior_area]."""

    @pytest.fixture
    def target_instance(self):
        """Create target instance for testing."""
        return TargetClass()

    def test_behavior_success_scenario(self, target_instance):
        """Test [behavior] under normal conditions."""
        # TODO: HIGH: Test [specific_behavior] [Story:SXX]
        pass

    @pytest.mark.asyncio
    async def test_async_behavior_error_scenario(self, target_instance):
        """Test [async_behavior] error handling."""
        # TODO: HIGH: Test [error_handling] [Story:SXX]
        pass
```

### Canonical TODO Format

For test placeholders and deferred implementation:
```python
# TODO: PRIORITY: description [Story:Reference]
```

**Priority Levels:** HIGH, MEDIUM, LOW  
**Story References:** Story:S3, Story:S4, Story:S5, Story:S14, IV&V:S5, etc.

**Examples:**
```python
# TODO: HIGH: Test configuration loading with malformed YAML [Story:S14]
# TODO: MEDIUM: Test hot reload graceful degradation [Story:S14]
# TODO: LOW: Test edge case with unusual device paths [Story:S3]
```

### Marking Tests Complete

When implementing a placeholder test:
1. **Remove the TODO comment**
2. **Implement the test logic**
3. **Ensure test passes** and provides meaningful validation
4. **Update story checklist** if applicable
5. **Verify coverage** includes the new test

## Running Tests

### Quick Unit Test Execution
```bash
# Run all unit tests
python3 -m pytest tests/unit/ -v

# Run specific module tests
python3 -m pytest tests/unit/test_camera_discovery/ -v

# Run with coverage
python3 -m pytest tests/unit/ -v --cov=src --cov-report=term-missing
```

### Using Project Test Runner
```bash
# Run complete quality pipeline (includes unit tests)
python3 run_all_tests.py

# Run only unit tests
python3 run_all_tests.py --only-unit

# Run with custom coverage threshold
python3 run_all_tests.py --threshold=85
```

### Test Debugging
```bash
# Run with detailed output
python3 -m pytest tests/unit/test_module/test_file.py::test_specific_function -vvv -s

# Run with debug logging
python3 -m pytest tests/unit/ -v --log-cli-level=DEBUG

# Run and stop on first failure
python3 -m pytest tests/unit/ -x
```

## Interpreting Test Results

### Successful Unit Test Run
```
============================= test session starts ==============================
collected 42 items

tests/unit/test_camera_discovery/test_hybrid_monitor.py::test_discovery PASSED
tests/unit/test_camera_service/test_service_manager.py::test_lifecycle PASSED
...

==================== 42 passed in 2.34s ====================

Name                                    Stmts   Miss  Cover   Missing
---------------------------------------------------------------------
src/camera_discovery/hybrid_monitor.py   450     45    90%    120-125
src/camera_service/service_manager.py    380     30    92%    245-250
---------------------------------------------------------------------
TOTAL                                     830     75    91%
```

### Test Failure Analysis
- **Import errors:** Check PYTHONPATH and module structure
- **Assertion failures:** Review test logic and expected behavior
- **Async errors:** Ensure proper `@pytest.mark.asyncio` decoration
- **Mock errors:** Verify mock setup matches actual interfaces

### Coverage Interpretation
- **Green (>80%):** Good coverage, meets project standards
- **Yellow (60-80%):** Acceptable but should be improved
- **Red (<60%):** Insufficient coverage, needs attention

**Missing lines** are shown in the coverage report - focus on testing those code paths.

## Contribution Guidelines

### Before Adding Tests
1. **Check existing tests** to avoid duplication
2. **Review architecture docs** to understand module responsibilities
3. **Follow coding standards** including type annotations and docstrings
4. **Use appropriate mocking** - mock external dependencies, test real logic

### Test Quality Standards
- **Test real behavior** not just mock interactions
- **Use descriptive assertions** with clear failure messages
- **Include edge cases** and error conditions
- **Avoid over-mocking** - mock only external dependencies
- **Maintain test independence** - tests should not depend on each other

### Integration with CI/CD
- All tests must pass before code can be merged
- Coverage thresholds must be maintained
- Linting and formatting checks must pass
- Type checking must succeed

## Related Documentation

- **Architecture Overview:** `docs/architecture/overview.md`
- **Coding Standards:** `docs/development/coding-standards.md`
- **Development Principles:** `docs/development/principles.md`
- **Testing Guide:** `docs/development/tests/testing-guide.md`
- **Integration Tests:** `tests/ivv/S5 Integration Test Execution Instructions.md`
- **Project Roadmap:** `docs/roadmap.md`

## Support

For questions about testing:
1. **Check existing test examples** in the relevant subdirectory
2. **Review architectural documentation** for module responsibilities
3. **Follow established patterns** from similar test files
4. **Ensure compliance** with project coding standards

**Test Execution Issues:**
- Verify Python path includes `src/` directory: `export PYTHONPATH=$PWD/src:$PYTHONPATH`
- Install test dependencies: `pip install pytest pytest-asyncio pytest-cov`
- Use project test runner: `python3 run_all_tests.py --help`