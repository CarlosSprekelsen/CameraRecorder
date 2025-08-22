# Deployment Automation

## Overview

The deployment process has been automated to eliminate repetitive manual steps and reduce human error during development and testing cycles.

## Automated Deployment Script

### Location
`deployment/scripts/deploy.sh`

### Features

1. **Automated Uninstall/Install Cycle**
   - Handles the complete uninstall → install cycle
   - Bypasses user confirmation with `--force-uninstall`
   - Skips uninstall step with `--skip-uninstall` (useful for first-time install)

2. **Automatic JWT Secret Synchronization**
   - Extracts JWT secret from deployed service (`/opt/camera-service/.env`)
   - Updates test environment (`.test_env`) automatically
   - Creates backup of previous `.test_env` before updating
   - Reports when JWT secret changes

3. **Deployment Validation**
   - Verifies both services are running (camera-service, mediamtx)
   - Checks health endpoint responsiveness
   - Reports deployment success/failure

4. **Error Handling**
   - Graceful handling of missing files/services
   - Clear error messages and status reporting
   - Automatic cleanup of temporary files

## Usage

### Basic Usage
```bash
# Full deployment with user confirmation
sudo ./deployment/scripts/deploy.sh

# Automated deployment without confirmation
sudo ./deployment/scripts/deploy.sh --force-uninstall

# Install only (skip uninstall)
sudo ./deployment/scripts/deploy.sh --skip-uninstall

# Skip validation (useful for debugging)
sudo ./deployment/scripts/deploy.sh --force-uninstall --skip-validation
```

### Options
- `--force-uninstall`: Bypass user confirmation for uninstall
- `--skip-uninstall`: Skip uninstall step (useful for first-time install)
- `--skip-validation`: Skip deployment validation
- `--help`: Show usage information

### Uninstall Script Options

The uninstall script (`uninstall.sh`) supports:
- `--force`: Skip confirmation prompt
- `--remove-user`: Remove service users (use with caution)
- `--help`: Show usage information

## Before vs After

### Before (Manual Process)
```bash
# 1. Manual uninstall with confirmation
sudo ./deployment/scripts/uninstall.sh
# [User must type 'y' to confirm]

# 2. Manual install
sudo ./deployment/scripts/install.sh

# 3. Manual JWT secret extraction
sudo cat /opt/camera-service/.env

# 4. Manual .test_env update
# Edit .test_env file manually with new JWT secret

# 5. Manual validation
systemctl status camera-service
systemctl status mediamtx
curl http://localhost:8003/health/ready
```

### Alternative: Automated Uninstall
```bash
# Automated uninstall without confirmation
sudo ./deployment/scripts/uninstall.sh --force
```

### After (Automated Process)
```bash
# Single command handles everything
sudo ./deployment/scripts/deploy.sh --force-uninstall

# Or use individual scripts with automation
sudo ./deployment/scripts/uninstall.sh --force
sudo ./deployment/scripts/install.sh
```

## Benefits

1. **Time Savings**: Reduces deployment time from ~5 minutes to ~1 minute
2. **Error Reduction**: Eliminates manual JWT secret copy/paste errors
3. **Consistency**: Ensures same process every time
4. **Developer Experience**: One command instead of multiple manual steps
5. **CI/CD Ready**: Can be easily integrated into automated pipelines

## Implementation Details

### JWT Secret Sync
- Reads from `/opt/camera-service/.env`
- Updates `.test_env` with `sed` command
- Creates timestamped backup: `.test_env.backup.YYYYMMDD_HHMMSS`
- Reports secret changes in deployment summary

### Force Uninstall
- Creates temporary modified uninstall script
- Replaces `read -p` with `REPLY=y` for auto-confirmation
- Cleans up temporary files after execution

### Validation
- Checks systemd service status
- Tests health endpoint with curl
- Reports detailed success/failure status

## Troubleshooting

### Common Issues

1. **Permission Denied**
   - Ensure script is run with `sudo`
   - Check file permissions: `chmod +x deployment/scripts/deploy.sh`

2. **Service Not Starting**
   - Check logs: `journalctl -u camera-service -f`
   - Verify dependencies are installed

3. **JWT Secret Not Synced**
   - Check if `/opt/camera-service/.env` exists
   - Verify `.test_env` file permissions
   - Check backup files for previous secrets

### Manual Override
If automation fails, you can still use the original scripts:
```bash
sudo ./deployment/scripts/uninstall.sh
sudo ./deployment/scripts/install.sh
# Then manually update .test_env
```

## Future Enhancements

1. **Configuration Management**
   - Support for different deployment environments
   - Environment-specific configuration files

2. **Rollback Capability**
   - Automatic rollback on deployment failure
   - Backup of previous deployment

3. **Health Checks**
   - More comprehensive service validation
   - Performance metrics collection

4. **Logging**
   - Structured deployment logs
   - Integration with monitoring systems
