// Package common provides common interfaces and utilities for the MediaMTX Camera Service.
//
// This package contains shared interfaces and helper functions used across
// multiple components to ensure consistent behavior and graceful shutdown patterns.
//
// Architecture Compliance:
//   - Context-Aware Cancellation: Use context.Context for operation cancellation
//   - Graceful Shutdown: Consistent shutdown patterns across all services
//   - Interface-Based Design: Common interfaces for shared behaviors
//   - Concurrency Safety: Thread-safe shutdown coordination
//
// Key Components:
//   - Stoppable: Interface for services requiring graceful shutdown
//   - StopWithTimeout: Helper function for timeout-based shutdown
//
// Usage Pattern:
//   - Implement Stoppable interface for services requiring shutdown
//   - Use StopWithTimeout() for consistent timeout-based shutdown
//   - Pass context for cancellation and timeout enforcement
//
// Requirements Coverage:
//   - REQ-COM-001: Common shutdown interface for all services
//   - REQ-COM-002: Context-aware cancellation support
//   - REQ-COM-003: Timeout-based shutdown coordination
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/common-interfaces.md
package common
