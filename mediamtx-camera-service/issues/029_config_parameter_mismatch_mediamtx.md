# Bug Report: Configuration parameter mismatch between MediaMTXConfig and MediaMTXController

**Issue ID:** 029  
**Severity:** MEDIUM  
**Status:** OPEN  
**Date:** 2025-01-06  
**Reporter:** IV&V Team  

## Summary
There is a parameter mismatch between `MediaMTXConfig` and `MediaMTXController` classes, where the controller has a `ffmpeg_config` parameter that is not present in the config class, causing configuration validation tests to fail.

## Description
When running configuration validation tests, the following failures occur:

```
FAILED tests/unit/test_configuration_validation.py::TestConfigurationSchemaValidation::test_mediamtx_config_controller_compatibility
Failed: MediaMTXController parameters missing from MediaMTXConfig: {'ffmpeg_config'}

FAILED tests/unit/test_configuration_validation.py::TestConfigurationFileValidation::test_config_parameter_consistency
AssertionError: MediaMTXController has parameters not in MediaMTXConfig: {'ffmpeg_config'}
assert 1 == 0
 +  where 1 = len({'ffmpeg_config'})
```

## Root Cause
The `MediaMTXController` class has a `ffmpeg_config` parameter that is not defined in the `MediaMTXConfig` class, creating an inconsistency between the configuration schema and the controller implementation.

## Impact
- Configuration validation tests fail
- Potential runtime issues when using configuration parameters
- Inconsistency between configuration schema and implementation
- Violates testing guide requirements for configuration validation

## Steps to Reproduce
1. Run configuration validation tests: `python3 -m pytest tests/unit/test_configuration_validation.py -v`
2. Observe parameter mismatch failures

## Expected Behavior
All MediaMTXController parameters should be present in MediaMTXConfig, and vice versa.

## Actual Behavior
Configuration validation tests fail due to parameter mismatch.

## Affected Files
- `tests/unit/test_configuration_validation.py`
- MediaMTXConfig class definition
- MediaMTXController class definition

## Requirements Impact
- Configuration management requirements
- Parameter validation and consistency
- System configuration integrity

## Fix Required
1. Add the missing `ffmpeg_config` parameter to `MediaMTXConfig` class, or
2. Remove the `ffmpeg_config` parameter from `MediaMTXController` class, or
3. Update the configuration validation logic to handle optional parameters

## Testing Guide Compliance
- **Violation**: Configuration schema and implementation are inconsistent
- **Impact**: Prevents proper configuration validation
- **Priority**: MEDIUM - Configuration consistency is important for system reliability
