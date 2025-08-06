# Validation Scripts

**Version:** 1.0  
**Authors:** Development Team  
**Date:** 2025-01-06  
**Status:** approved  
**Related Epic/Story:** E1 / S3

## Overview

Validation scripts ensure deployment compatibility and catch configuration issues early in the development and deployment process.

## Scripts

### `validate_deployment.py`

Comprehensive deployment validation script that tests configuration and component compatibility.

**Features:**
- Python compatibility testing
- Configuration loading validation
- Component instantiation testing
- MediaMTXController parameter compatibility checking
- Dependency availability verification

**Usage:**
```bash
# Run validation script
python3 scripts/validate_deployment.py

# Run during installation (automatic)
sudo ./deployment/scripts/install.sh
```

**Test Coverage:**
- Python3 availability and version
- Required Python dependencies
- Configuration loading without errors
- ServiceManager instantiation
- MediaMTXController parameter compatibility
- Health monitoring parameter validation

**Exit Codes:**
- `0`: All validation tests passed
- `1`: One or more validation tests failed

## Integration

### Installation Script Integration

The validation script is automatically run during installation via `deployment/scripts/install.sh`:

```bash
# Validation runs automatically during installation
sudo ./deployment/scripts/install.sh
```

### Test Automation Integration

Validation tests are included in the test automation script (`run_all_tests.py`):

```bash
# Run all tests including validation
python3 run_all_tests.py

# Run validation tests only
python3 run_all_tests.py --only-validation
```

### CI/CD Integration

Validation script can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Validate Deployment
  run: python3 scripts/validate_deployment.py
```

## Error Handling

If validation fails:

1. **Python Compatibility Issues:**
   - Ensure `python3` is available
   - Check Python version (3.10+ required)

2. **Configuration Issues:**
   - Verify all required parameters are present
   - Check parameter types and values
   - Ensure configuration files are valid

3. **Component Issues:**
   - Verify component constructor signatures
   - Check parameter compatibility
   - Ensure all dependencies are installed

4. **Dependency Issues:**
   - Install missing Python packages
   - Check virtual environment setup
   - Verify import paths

## Related Documentation

- [Installation Fixes](docs/deployment/installation_fixes.md)
- [Integration Tests](tests/integration/README.md)
- [Test Automation](run_all_tests.py)
