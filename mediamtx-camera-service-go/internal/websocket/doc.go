// Package websocket implements the Layer 6 (API) WebSocket JSON-RPC 2.0 server.
//
// This package provides the protocol layer implementation with NO business logic,
// following the architectural constraint that all operations are delegated to
// the MediaMTX controller (single source of truth).
//
// Architecture Compliance:
//   - Protocol Layer Only: No business logic, pure JSON-RPC 2.0 implementation
//   - Delegation Pattern: All operations forwarded to MediaMTX controller
//   - High Concurrency: Supports 1000+ simultaneous WebSocket connections
//   - Security Integration: JWT authentication with role-based access control
//   - Event System: Real-time client notifications via event manager
//
// Key Responsibilities:
//   - WebSocket connection management and lifecycle
//   - JSON-RPC 2.0 protocol implementation and message handling
//   - Authentication enforcement and session management
//   - Input validation and security protection (rate limiting)
//   - Real-time event broadcasting to connected clients
//   - Performance metrics collection and monitoring
//
// Thread Safety: All components are designed for concurrent access with
// appropriate synchronization primitives protecting shared state.
//
// Performance Targets:
//   - 1000+ simultaneous WebSocket connections
//   - <2s response time for 95% of V4L2 hardware operations
//   - <20ms event notification delivery latency
//
// Requirements Coverage:
//   - REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint on port 8002
//   - REQ-API-002: Complete JSON-RPC 2.0 protocol implementation
//   - REQ-API-003: Request/response message handling with proper error codes
//   - REQ-API-011: API methods respond within specified time limits (<2s for V4L2 hardware)
//
// Method Categories:
//   - Core Methods: ping, authenticate, system status
//   - Camera Methods: get_camera_list, get_camera_status, camera operations
//   - Recording Methods: start_recording, stop_recording, recording management
//   - Snapshot Methods: take_snapshot with multi-tier fallback
//   - File Methods: list_recordings, list_snapshots, file operations
//   - Health Methods: get_system_health, component status
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/api/json_rpc_methods.md
package websocket
