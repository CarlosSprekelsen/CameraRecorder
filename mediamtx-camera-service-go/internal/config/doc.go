// Package config provides centralized configuration management for the MediaMTX Camera Service.
//
// This package handles configuration loading, validation, hot reload functionality,
// and provides type-safe access to all service configuration settings.
//
// Architecture Compliance:
//   - Centralized Configuration: Single source of truth for all configuration
//   - Hot Reload: Runtime configuration updates without service restart
//   - Environment Override: Support for environment variable overrides
//   - Validation: Built-in configuration validation and defaults
//   - Type Safety: Strongly typed configuration structures
//
// Key Features:
//   - YAML configuration file loading with Viper
//   - Environment variable override support (CONFIG_* prefix)
//   - Hot reload with file system watching
//   - Configuration validation with meaningful error messages
//   - Default value management and fallback handling
//   - Thread-safe configuration access
//
// Configuration Categories:
//   - Server: WebSocket server settings (host, port, timeouts, limits)
//   - MediaMTX: MediaMTX integration settings (URLs, paths, codecs)
//   - Security: Authentication, authorization, rate limiting, CORS
//   - Storage: File storage paths, usage thresholds, cleanup policies
//   - Camera: Device discovery, capability detection, monitoring
//   - Logging: Log levels, formats, output destinations
//   - Health: Health check intervals, failure thresholds, circuit breaker
//   - Performance: Response time targets, optimization settings
//
// Usage Pattern:
//   - Create ConfigManager with CreateConfigManager()
//   - Load configuration with LoadConfig(path)
//   - Access configuration with GetConfig()
//   - Register for updates with AddUpdateCallback(callback)
//
// Requirements Coverage:
//   - REQ-CFG-001: Centralized configuration management
//   - REQ-CFG-002: Hot reload functionality
//   - REQ-CFG-003: Environment variable override support
//   - REQ-CFG-004: Configuration validation and defaults
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/configuration.md
package config
