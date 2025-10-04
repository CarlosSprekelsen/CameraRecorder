// Package constants provides shared constants for the MediaMTX Camera Service.
//
// This package contains centralized constants for API error codes, timeouts,
// response values, and configuration defaults to ensure consistency across
// the entire project and eliminate magic numbers.
//
// Architecture Compliance:
//   - Single Source of Truth: All constants defined in one location
//   - Shared Implementation/Test: Constants used by both production and test code
//   - API Documentation Alignment: Constants match documented API specifications
//   - Version Control: Constants organized with clear versioning and deprecation
//
// Constant Categories:
//   - JSON-RPC Error Codes: Standard and service-specific error codes
//   - WebSocket Server: Connection timeouts, buffer sizes, limits
//   - API Response Values: Exact string values for API responses
//   - Configuration Defaults: Default values for service configuration
//   - Test Constants: Shared values for testing scenarios
//   - Legacy Support: Deprecated constants with backward compatibility
//
// Design Principles:
//   - Ground Truth: Constants based on API documentation specifications
//   - Consistency: Eliminate magic number duplication across codebase
//   - Maintainability: Clear organization and naming conventions
//   - Backward Compatibility: Legacy constants maintained for compatibility
//
// Usage Pattern:
//   - Import constants: import "github.com/camerarecorder/mediamtx-camera-service-go/internal/constants"
//   - Use error codes: constants.JSONRPC_INVALID_REQUEST
//   - Use timeouts: constants.WEBSOCKET_READ_TIMEOUT
//   - Use response values: constants.RESPONSE_STATUS_SUCCESS
//
// Requirements Coverage:
//   - REQ-API-001: JSON-RPC 2.0 protocol constants
//   - REQ-API-002: Standardized error codes
//   - REQ-TEST-001: Shared test constants
//   - REQ-CFG-001: Configuration default values
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/api/json_rpc_methods.md
package constants
