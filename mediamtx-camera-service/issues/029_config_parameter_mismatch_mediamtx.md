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

**API Documentation Analysis:**
- **MediaMTXConfig**: Has `ffmpeg_config` parameter documented in API (`docs/api/configuration-reference.md`)
- **MediaMTXController**: Has `ffmpeg_config` parameter in implementation but **NOT DOCUMENTED** in API documentation
- **Ground Truth**: The `ffmpeg_config` parameter exists in the implementation but is **NOT PART OF THE OFFICIAL API DOCUMENTATION**

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
**Option 1 (Recommended)**: Document the `ffmpeg_config` parameter in the MediaMTXController API documentation
- Add API documentation for MediaMTXController in `docs/api/`
- Define the `ffmpeg_config` parameter interface and usage
- Ensure consistency between documented API and implementation

**Option 2**: Remove the `ffmpeg_config` parameter from `MediaMTXController` class
- Remove the parameter from the constructor
- Update implementation to not use ffmpeg_config
- Ensure this doesn't break existing functionality

**Option 3**: Update the configuration validation logic to handle optional parameters
- Modify tests to allow for optional parameters not in config
- This is less ideal as it masks the real API inconsistency

**Recommendation**: Choose Option 1 to properly document the API and maintain consistency.

## Testing Guide Compliance
- **Violation**: Configuration schema and implementation are inconsistent
- **Impact**: Prevents proper configuration validation
- **Priority**: MEDIUM - Configuration consistency is important for system reliability

## Research Findings
**Ground Truth Analysis**: After examining the API documentation in `docs/api/`:
- **MediaMTXConfig API**: Fully documented in `docs/api/configuration-reference.md` with STANAG 4406 H.264 codec parameters
- **MediaMTXController API**: **NOT DOCUMENTED** in `docs/api/` - no official API documentation exists
- **Test Design**: The configuration validation tests are correctly identifying API inconsistencies
- **Risk Assessment**: Adapting tests to non-compliant code would violate API contract and risk stakeholder acceptance

**Conclusion**: This is a **SOFTWARE BUG** where the implementation has parameters not defined in the API documentation. The test design is correct and should not be modified.
