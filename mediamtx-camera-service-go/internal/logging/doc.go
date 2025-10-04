// Package logging provides structured logging with correlation ID support for the MediaMTX Camera Service.
//
// This package implements a centralized logging system using Logrus with structured
// logging, correlation ID tracking, component identification, and configurable output
// destinations (console, file, both, or disabled).
//
// Architecture Compliance:
//   - Structured Logging: JSON and text formats with consistent field structure
//   - Correlation ID Support: Request tracing across service boundaries
//   - Component Identification: Logger instances tagged with component names
//   - Centralized Configuration: Global logging configuration with factory pattern
//   - Thread Safety: All logger operations are thread-safe
//
// Key Features:
//   - Structured logging with JSON and text formatters
//   - Correlation ID tracking for request tracing
//   - Component-based logger instances
//   - Configurable log levels (debug, info, warn, error, fatal)
//   - File rotation with configurable size limits and backup retention
//   - Console and file output with independent enable/disable
//   - Global logger factory with consistent configuration
//
// Usage Patterns:
//   - Get logger factory: GetLoggerFactory()
//   - Configure globally: ConfigureFactory(config)
//   - Create component logger: factory.CreateLogger("component-name")
//   - Get global logger: GetLogger()
//   - Add correlation ID: WithCorrelationID(ctx)
//
// Logger Creation:
//   - Component loggers: factory.CreateLogger("websocket")
//   - Global logger: GetLogger() for general use
//   - Context-aware: WithCorrelationID(ctx) for request tracing
//
// Field Conventions:
//   - "component": Component name (e.g., "websocket", "mediamtx")
//   - "correlation_id": Request correlation ID for tracing
//   - "client_id": Client identifier for WebSocket connections
//   - "user_id": User identifier for authenticated requests
//   - "method": API method name for JSON-RPC calls
//   - "action": Specific action being performed
//
// Requirements Coverage:
//   - REQ-LOG-001: Structured logging with consistent field format
//   - REQ-LOG-002: Correlation ID support for request tracing
//   - REQ-LOG-003: Component-based logger instances
//   - REQ-LOG-004: Configurable output destinations
//   - REQ-LOG-005: File rotation and retention policies
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/logging.md
package logging
