# Configuration Tidy-Up Documentation

## Overview

This document describes the configuration tidy-up performed to move hard-coded values to configurable parameters, improving the service's flexibility and maintainability.

## Changes Made

### 1. Security Configuration (Phase 1 Enhancement)

**Added to `config_types.go`:**
```go
type SecurityConfig struct {
    RateLimitRequests int           `mapstructure:"rate_limit_requests"` // Default: 100 requests per window
    RateLimitWindow   time.Duration `mapstructure:"rate_limit_window"`   // Default: 1 minute
    JWTSecretKey      string        `mapstructure:"jwt_secret_key"`
    JWTExpiryHours    int           `mapstructure:"jwt_expiry_hours"` // Default: 24 hours
}
```

**Hard-coded values moved to configuration:**
- `rateLimit: 100` → `security.rate_limit_requests: 100`
- `rateWindow: time.Minute` → `security.rate_limit_window: 1m`

**New constructor added:**
```go
func NewJWTHandlerWithConfig(secretKey string, rateLimit int64, rateWindow time.Duration) (*JWTHandler, error)
```

### 2. Storage Configuration (Phase 2 Enhancement)

**Added to `config_types.go`:**
```go
type StorageConfig struct {
    WarnPercent   int    `mapstructure:"warn_percent"`   // Default: 80% usage warning
    BlockPercent  int    `mapstructure:"block_percent"`  // Default: 90% usage block
    DefaultPath   string `mapstructure:"default_path"`  // Default: "/opt/camera-service/recordings"
    FallbackPath  string `mapstructure:"fallback_path"` // Default: "/tmp/recordings"
}
```

**Hard-coded values moved to configuration:**
- `storageWarnPercent: 80` → `storage.warn_percent: 80`
- `storageBlockPercent: 90` → `storage.block_percent: 90`
- `"/opt/camera-service/recordings"` → `storage.default_path: "/opt/camera-service/recordings"`
- `"/tmp/recordings"` → `storage.fallback_path: "/tmp/recordings"`

**New method added:**
```go
func (rm *RecordingManager) UpdateStorageThresholds(warnPercent, blockPercent int)
```

### 3. Recording Configuration (Phase 2 Enhancement)

**Enhanced `RecordingConfig` in `config_types.go`:**
```go
type RecordingConfig struct {
    // ... existing fields ...
    DefaultRotationSize int64         `mapstructure:"default_rotation_size"` // Default: 100MB
    DefaultMaxDuration  time.Duration `mapstructure:"default_max_duration"`  // Default: 24 hours
    DefaultRetentionDays int          `mapstructure:"default_retention_days"` // Default: 7 days
}
```

**Hard-coded values moved to configuration:**
- `100 * 1024 * 1024` (100MB) → `recording.default_rotation_size: 104857600`
- `24 * time.Hour` → `recording.default_max_duration: 24h`
- `7` (retention days) → `recording.default_retention_days: 7`

### 4. Health Check Configuration (Phase 4 Enhancement)

**Added to `MediaMTXConfig` in `config_types.go`:**
```go
type MediaMTXConfig struct {
    // ... existing fields ...
    HealthCheckTimeout time.Duration `mapstructure:"health_check_timeout"` // Default: 5 seconds
}
```

**Hard-coded values moved to configuration:**
- `5 * time.Second` → `mediamtx.health_check_timeout: 5s`

## Configuration File Structure

The updated `config/default.yaml` now includes all configurable parameters:

```yaml
# Security configuration (Phase 1 enhancement)
security:
  rate_limit_requests: 100  # Requests per window
  rate_limit_window: 1m     # Time window for rate limiting
  jwt_secret_key: "your-secret-key-here"  # Must be set in production
  jwt_expiry_hours: 24      # JWT token expiry in hours

# Storage configuration (Phase 2 enhancement)
storage:
  warn_percent: 80    # Storage usage warning threshold
  block_percent: 90   # Storage usage block threshold
  default_path: "/opt/camera-service/recordings"
  fallback_path: "/tmp/recordings"

# Recording configuration enhancements
recording:
  # ... existing fields ...
  default_rotation_size: 104857600  # 100MB
  default_max_duration: 24h
  default_retention_days: 7

# MediaMTX configuration enhancements
mediamtx:
  # ... existing fields ...
  health_check_timeout: 5s
```

## Benefits

### 1. **Flexibility**
- All hard-coded values are now configurable
- Different environments can have different settings
- Easy to adjust parameters without code changes

### 2. **Maintainability**
- Centralized configuration management
- Clear documentation of all configurable parameters
- Easy to track configuration changes

### 3. **Production Readiness**
- Environment-specific configurations
- Security parameters properly configurable
- Performance tuning capabilities

### 4. **Testing**
- Easy to test different configurations
- Mock configurations for unit tests
- Environment-specific test configurations

## Migration Guide

### For Existing Deployments

1. **Update configuration files** to include new sections:
   ```yaml
   security:
     rate_limit_requests: 100
     rate_limit_window: 1m
     jwt_secret_key: "your-production-secret-key"
     jwt_expiry_hours: 24
   
   storage:
     warn_percent: 80
     block_percent: 90
     default_path: "/opt/camera-service/recordings"
     fallback_path: "/tmp/recordings"
   ```

2. **Update code** to use new constructors:
   ```go
   // Old way
   jwtHandler, err := security.NewJWTHandler(secretKey)
   
   // New way
   jwtHandler, err := security.NewJWTHandlerWithConfig(
       secretKey,
       config.Security.RateLimitRequests,
       config.Security.RateLimitWindow,
   )
   ```

3. **Update RecordingManager** to use configuration:
   ```go
   // After creating RecordingManager
   rm.UpdateStorageThresholds(
       config.Storage.WarnPercent,
       config.Storage.BlockPercent,
   )
   ```

### For New Deployments

1. Copy `config/default.yaml` to your deployment
2. Modify values according to your environment
3. Ensure all required paths exist
4. Set appropriate security keys

## Validation

All configuration changes have been tested to ensure:
- ✅ Type safety (proper Go types)
- ✅ Default values (backward compatibility)
- ✅ Configuration loading (viper compatibility)
- ✅ Runtime behavior (no breaking changes)

## Future Enhancements

Consider these additional configuration improvements:
- Environment variable overrides
- Configuration validation rules
- Dynamic configuration reloading
- Configuration encryption for sensitive values
