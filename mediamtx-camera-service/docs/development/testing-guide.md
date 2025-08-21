# Server Test Guide - MediaMTX Camera Service

**Version:** 3.0  
**Date:** 2025-01-06  
**Status:** Updated with baseline rebuild and tools reorganization  

## 1. Core Principles

### Real System Testing Over Mocking
- **MediaMTX:** Use systemd-managed service, never mock
- **File System:** Use `tempfile`, never mock
- **WebSocket:** Use real connections within system
- **Authentication:** Use real JWT tokens with test secrets

### Strategic Mocking Rules
**MOCK:** External APIs, time operations, expensive hardware simulation  
**NEVER MOCK:** MediaMTX service, filesystem, internal WebSocket, JWT auth, config loading

## 2. Test Organization

### Directory Structure
```
tests/
├── unit/                    # <30 seconds total
├── integration/             # <5 minutes total  
├── fixtures/                # Shared utilities
├── performance/             # Load tests
├── tools/                   # Test runners and utilities
│   ├── run_all_tests.py     # Comprehensive test automation
│   ├── run_tests.py         # Basic test runner
│   ├── run_individual_tests.py # Individual test execution
│   ├── run_critical_error_tests.py # Critical error testing
│   ├── run_integration_tests.py # Integration test runner
│   ├── setup_test_environment.py # Environment setup
│   ├── validate_test_environment.py # Environment validation
│   └── README.md            # Tools documentation
└── requirements/            # Requirements coverage mapping
```

### File Rules
- **One file per feature** - no variants (_real, _v2)
- **REQ-* references required** in every test file docstring
- **Shared utilities over duplication**
- **Test tools in tests/tools/** - separate from actual test files

## 3. Requirements Traceability

### Mandatory Format for Test Files
```python
"""
Module description.

Requirements Coverage:
- REQ-XXX-001: Requirement description
- REQ-XXX-002: Additional requirement

Test Categories: Unit/Integration
"""

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

## 4. Test Tools and Runners

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

## 5. Performance Targets

- **Unit tests:** <30 seconds total
- **Integration tests:** <5 minutes total  
- **Full suite:** <10 minutes total
- **Flaky rate:** <1%

### Test Markers
```python
@pytest.mark.unit          # Fast isolated tests
@pytest.mark.integration   # Real component integration
@pytest.mark.real_mediamtx # Requires systemd MediaMTX
@pytest.mark.performance   # Load/performance tests
```

## 6. Standard Patterns

### MediaMTX Integration
```python
@pytest.mark.real_mediamtx
async def test_stream_creation():
    controller = MediaMTXController("http://localhost:9997")
    stream_id = await controller.create_stream("test", "/dev/video0")
    assert stream_id is not None
```

### Authentication Testing
```python
async def test_valid_auth():
    token = generate_valid_test_token("test_user", "operator")
    # Test with real JWT token
```

**Note**: Source `.test_env` before running tests to provide JWT token environment variables.

## 7. Quality Assurance

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

## 8. Documentation Standards

### Test File Documentation
- **Requirements Coverage**: Mandatory in every test file docstring
- **Test Categories**: Unit/Integration/Performance
- **Real Component Usage**: Document when real components are used

### Tool Documentation
- **Purpose**: What the tool does, not requirements coverage
- **Usage**: Command-line examples and options
- **Location**: `tests/tools/README.md`

### Coverage Analysis
- **Location**: `docs/test/requirements_coverage_analysis.md`
- **Updates**: After major test changes
- **Focus**: Critical and high-priority requirements gaps

## 9. Compliance and Validation

### Testing Guide Compliance
- **Test Files**: Must follow requirements traceability format
- **Test Tools**: Must follow script conventions (no requirements coverage)
- **Coverage Analysis**: Must be updated after major changes

### Quality Gates
- **Critical Requirements**: 100% coverage required
- **High Priority Requirements**: 95% coverage required
- **Overall Coverage**: 90% coverage required
- **Performance Testing**: Must be implemented for critical requirements

---

**Status**: **UPDATED** - Reflects baseline rebuild, tools reorganization, and comprehensive coverage analysis approach.