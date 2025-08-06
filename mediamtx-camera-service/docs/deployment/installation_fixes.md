# Installation Fixes and Validation

**Version:** 1.0  
**Authors:** Development Team  
**Date:** 2025-01-06  
**Status:** approved  
**Related Epic/Story:** E1 / S3

## Overview

This document summarizes all the fixes applied to prevent installation and configuration issues, ensuring smooth deployment on Ubuntu 22.04+ systems.

## Issues Fixed

### 1. Python Compatibility Issues ✅ FIXED

**Problem**: Scripts referenced `python` instead of `python3` on Ubuntu 22.04+

**Files Fixed**:
- `run_all_tests.py`: Updated `required_tools` to use `python3`
- `run_tests.py`: Updated base pytest command to use `python3`
- `deployment/scripts/install.sh`: Updated systemd service to use `python3`

**Impact**: All Ubuntu 22.04+ installations now work correctly

### 2. Virtual Environment Path Issues ✅ FIXED

**Problem**: Systemd service had incorrect Python path and PYTHONPATH

**Files Fixed**:
- `deployment/scripts/install.sh`: Updated systemd service configuration
  - `ExecStart`: Changed to `$VENV_DIR/bin/python3 -m camera_service.main`
  - `Environment=PYTHONPATH`: Added `$INSTALL_DIR/src`

**Impact**: Service starts correctly with proper Python environment

### 3. Configuration Schema Mismatch ✅ FIXED

**Problem**: Configuration files contained health monitoring parameters but MediaMTXConfig class was missing these fields

**Files Fixed**:
- `src/camera_service/config.py`: Added health monitoring parameters to MediaMTXConfig
  - `health_check_interval: int = 30`
  - `health_failure_threshold: int = 10`
  - `health_circuit_breaker_timeout: int = 60`
  - `health_max_backoff_interval: int = 120`
  - `health_recovery_confirmation_threshold: int = 3`
  - `backoff_base_multiplier: float = 2.0`
  - `backoff_jitter_range: tuple = (0.8, 1.2)`
  - `process_termination_timeout: float = 3.0`
  - `process_kill_timeout: float = 2.0`

**Impact**: Configuration loading no longer fails with "unexpected keyword argument" errors

### 4. API Interface Mismatch ✅ FIXED

**Problem**: Service manager was passing entire config object to MediaMTXController constructor that expects individual parameters

**Files Fixed**:
- `src/camera_service/service_manager.py`: Updated `_start_mediamtx_controller()` method to unpack config object into individual parameters

**Impact**: MediaMTXController instantiation works correctly

### 5. Default Configuration Issues ✅ FIXED

**Problem**: Default configuration files were missing health monitoring parameters

**Files Fixed**:
- `config/default.yaml`: Added all health monitoring parameters
- `deployment/scripts/install.sh`: Updated fallback configuration to include health parameters

**Impact**: Fresh installations get complete configuration by default

## Validation Improvements

### 1. Deployment Validation Script ✅ ADDED

**File**: `scripts/validate_deployment.py`

**Features**:
- Python compatibility testing
- Configuration loading validation
- Component instantiation testing
- MediaMTXController parameter compatibility checking
- Dependency availability verification

### 2. Integration Tests ✅ ADDED

**File**: `tests/integration/test_config_component_integration.py`

**Features**:
- Configuration → component instantiation chain testing
- Parameter type validation
- Configuration serialization testing
- API interface compatibility verification

### 3. Installation Script Validation ✅ ADDED

**File**: `deployment/scripts/install.sh`

**Features**:
- Added `validate_installation()` function
- Runs validation script during installation
- Tests configuration loading manually
- Continues installation with warnings if validation fails

## Prevention Strategy

### 1. Early Detection
- Validation runs during installation
- Integration tests catch mismatches during development
- Configuration schema validation prevents runtime errors

### 2. Comprehensive Testing
- Configuration loading tests
- Component instantiation tests
- API interface compatibility tests
- Python environment validation

### 3. Documentation
- Updated README with validation information
- Added validation script documentation
- Created this fixes summary

## Testing the Fixes

### Manual Validation
```bash
# Run validation script
python3 scripts/validate_deployment.py

# Run integration tests
python3 -m pytest tests/integration/test_config_component_integration.py -v
```

### Fresh Installation Test
```bash
# On a clean Ubuntu 22.04 system
git clone https://github.com/your-org/mediamtx-camera-service
cd mediamtx-camera-service
sudo ./deployment/scripts/install.sh
```

## Future Prevention

### 1. CI/CD Integration
- Add validation script to CI pipeline
- Run integration tests on every commit
- Validate configuration schema changes

### 2. Development Guidelines
- Always update both configuration files and dataclasses together
- Test component instantiation when adding new parameters
- Use integration tests to catch interface mismatches

### 3. Monitoring
- Monitor installation success rates
- Track validation failures
- Alert on configuration schema changes

## Files Modified

1. **Configuration Files**:
   - `src/camera_service/config.py` - Added health monitoring parameters
   - `config/default.yaml` - Added health monitoring parameters

2. **Service Management**:
   - `src/camera_service/service_manager.py` - Fixed parameter unpacking

3. **Installation Scripts**:
   - `deployment/scripts/install.sh` - Added validation and fixed Python paths

4. **Testing**:
   - `run_all_tests.py` - Fixed Python compatibility
   - `run_tests.py` - Fixed Python compatibility
   - `tests/integration/test_config_component_integration.py` - Added integration tests

5. **Validation**:
   - `scripts/validate_deployment.py` - Added deployment validation script

6. **Documentation**:
   - `README.md` - Added validation section
   - `docs/deployment/installation_fixes.md` - This summary document

## Result

A fresh installation on Ubuntu 22.04 should now work smoothly without any of the previously encountered issues:

- ✅ Python compatibility issues resolved
- ✅ Configuration schema mismatches fixed
- ✅ API interface mismatches resolved
- ✅ Default configuration complete
- ✅ Comprehensive validation included
- ✅ Integration tests added

The installation process now includes validation that catches issues early and provides clear error messages if problems are detected.
