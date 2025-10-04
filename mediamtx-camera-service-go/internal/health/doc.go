// Package health provides health monitoring and HTTP health endpoints for the MediaMTX Camera Service.
//
// This package implements comprehensive health monitoring with HTTP endpoints for
// liveness and readiness probes, component status tracking, and system metrics
// collection following Kubernetes health check patterns.
//
// Architecture Compliance:
//   - Thin Delegation Pattern: HTTP server delegates all operations to HealthAPI
//   - Interface-Based Design: HealthAPI interface for component integration
//   - Event-Based Progressive Readiness: Component status tracking with timestamps
//   - Configuration Integration: Centralized configuration management
//   - Structured Logging: Consistent logging with component identification
//
// Key Components:
//   - HealthAPI: Interface for health monitoring components
//   - HealthMonitor: Core health monitoring logic implementation
//   - HTTPHealthServer: HTTP endpoint server with thin delegation
//   - ComponentStatus: Individual component health tracking
//   - Health Responses: Structured health response types
//
// Health Endpoints:
//   - /health: Basic health status (healthy/unhealthy/degraded)
//   - /health/detailed: Comprehensive health with components and metrics
//   - /ready: Readiness probe for Kubernetes
//   - /alive: Liveness probe for Kubernetes
//
// Health Status Semantics:
//   - healthy: All components operational, system ready for requests
//   - degraded: Some components failing but core functionality available
//   - unhealthy: Critical components failing, system not ready
//
// Component Integration:
//   - Register components with RegisterComponent()
//   - Update component status with UpdateComponentStatus()
//   - Automatic timestamp tracking for all status updates
//   - Configurable health check intervals and thresholds
//
// Requirements Coverage:
//   - REQ-HEALTH-001: Health monitoring with component status tracking
//   - REQ-HEALTH-002: HTTP health endpoints for liveness and readiness probes
//   - REQ-HEALTH-003: System metrics collection and reporting
//   - REQ-HEALTH-004: Event-based progressive readiness architecture
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/api/health-endpoints.md
package health
