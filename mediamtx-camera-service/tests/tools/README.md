# Test Tools

This directory contains tools and scripts specifically for test execution, environment setup, and test validation.

## Test Runners

### `run_all_tests.py`
**Comprehensive test automation script** that runs all quality gates in priority order:
- Critical tests (type checking, unit tests, integration tests)
- Code quality checks (formatting, linting)

**Usage:**
```bash
python3 tests/tools/run_all_tests.py                    # Run all stages
python3 tests/tools/run_all_tests.py --no-lint          # Skip linting
python3 tests/tools/run_all_tests.py --only-unit        # Unit tests only
python3 tests/tools/run_all_tests.py --threshold=85     # Custom coverage threshold
```

### `run_tests.py`
**Basic test runner** that sets up the test environment and executes pytest with various options.

**Usage:**
```bash
python3 tests/tools/run_tests.py                    # Run all tests
python3 tests/tools/run_tests.py --unit             # Run only unit tests
python3 tests/tools/run_tests.py --integration      # Run only integration tests
python3 tests/tools/run_tests.py --coverage         # Run with coverage report
```

### `run_individual_tests.py`
**Individual test execution** with failure categorization and timeout protection.

**Usage:**
```bash
python3 tests/tools/run_individual_tests.py              # Run all tests individually
python3 tests/tools/run_individual_tests.py --timeout=60 # Custom timeout
python3 tests/tools/run_individual_tests.py --output=json # JSON output format
```

### `run_critical_error_tests.py`
**Critical error handling test runner** that focuses on failure scenarios that could break the system.

**Usage:**
```bash
python3 tests/tools/run_critical_error_tests.py          # Run all critical error tests
python3 tests/tools/run_critical_error_tests.py --timeout=180 # Custom timeout
python3 tests/tools/run_critical_error_tests.py --retries=3   # Custom retry count
```

### `run_integration_tests.py`
**Real system integration test runner** with proper setup and teardown.

**Usage:**
```bash
python3 tests/tools/run_integration_tests.py             # Run all integration tests
python3 tests/tools/run_integration_tests.py --check-deps # Check dependencies only
python3 tests/tools/run_integration_tests.py --setup-only # Setup environment only
```

### `run_health_tests.py`
**Health test runner** that executes comprehensive health monitoring tests against real system components.

**Usage:**
```bash
python3 tests/tools/run_health_tests.py                    # Run all health tests
python3 tests/tools/run_health_tests.py --components      # Detailed component info only
python3 tests/tools/run_health_tests.py --kubernetes      # Kubernetes probes only
python3 tests/tools/run_health_tests.py --json            # JSON response format only
python3 tests/tools/run_health_tests.py --ok-response     # 200 OK response only
python3 tests/tools/run_health_tests.py --error-response  # 500 error response only
```

### `run_performance_tests.py`
**Performance test runner** that executes comprehensive performance tests against real system components.

**Usage:**
```bash
python3 tests/tools/run_performance_tests.py              # Run all performance tests
python3 tests/tools/run_performance_tests.py --status     # Status methods only
python3 tests/tools/run_performance_tests.py --control    # Control methods only
python3 tests/tools/run_performance_tests.py --file-ops   # File operations only
python3 tests/tools/run_performance_tests.py --concurrent # Concurrent connections only
python3 tests/tools/run_performance_tests.py --iterations=200 # Custom iterations
python3 tests/tools/run_performance_tests.py --connections=50 # Custom connection count
```

### `run_individual_tests_no_mocks.py`
**No-mock test execution** script for testing with real components only.

## Test Environment Tools

### `setup_test_environment.py`
**Test environment setup script** that prepares the testing environment.

### `validate_test_environment.py`
**Test environment validation script** that verifies the testing environment is properly configured.

## Test Utilities

### `ws_auth_perf_smoke.py`
**WebSocket authentication performance smoke test** for quick validation.

### `e2e_ws_system_test.py`
**End-to-end WebSocket system test** for comprehensive system validation.

## Important Notes

- These are **test tools**, not test files themselves
- They do **NOT** contain requirements coverage - they orchestrate test execution
- They follow **tool conventions**, not testing guide rules
- They are located in `tests/tools/` to keep them close to the test files
- Use `pytest` directly for most test execution needs

## Running Tests

For most testing needs, use pytest directly:

```bash
# Run all tests
pytest

# Run specific test categories
pytest -m unit
pytest -m integration

# Run with coverage
pytest --cov=src --cov-report=html
```

Use these tools only for specialized test orchestration needs.
