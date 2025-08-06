# Integration Tests

**Version:** 1.0  
**Authors:** Development Team  
**Date:** 2025-01-06  
**Status:** approved  
**Related Epic/Story:** E1 / S3

## Overview

Integration tests validate the complete configuration → component instantiation chain to prevent deployment issues.

## Test Files

### `test_config_component_integration.py`

Validates that configuration can be loaded and used to instantiate all components without parameter mismatches.

**Test Coverage:**
- MediaMTXConfig → MediaMTXController instantiation
- ServiceManager instantiation with configuration
- Configuration schema completeness validation
- Parameter type validation
- Configuration serialization testing
- API interface compatibility verification

**Key Validations:**
- All health monitoring parameters are present and correctly typed
- Configuration can be serialized and deserialized without data loss
- MediaMTXController constructor accepts all config parameters
- ServiceManager can be instantiated with configuration object

## Running Tests

```bash
# Run all integration tests
python3 -m pytest tests/integration/ -v

# Run specific test file
python3 -m pytest tests/integration/test_config_component_integration.py -v

# Run with test automation script
python3 run_all_tests.py
```

## Test Execution

Integration tests are automatically included in the test automation script (`run_all_tests.py`) and run after unit tests but before code quality checks.

**Execution Order:**
1. Type checking (critical)
2. Unit tests (critical)
3. Integration tests (if not unit-only mode)
4. Validation tests (if not unit-only mode)
5. Code formatting (cosmetic)
6. Linting (cosmetic)

## Validation Integration

Integration tests work together with the deployment validation script (`scripts/validate_deployment.py`) to ensure:

- Configuration schema consistency
- Component instantiation compatibility
- API interface alignment
- Python environment compatibility

## Failure Handling

If integration tests fail:
1. Check configuration schema mismatches
2. Verify component constructor signatures
3. Ensure all required parameters are present
4. Validate Python environment and dependencies

## Related Documentation

- [Installation Fixes](docs/deployment/installation_fixes.md)
- [Configuration Management](src/camera_service/config.py)
- [Service Manager](src/camera_service/service_manager.py)
