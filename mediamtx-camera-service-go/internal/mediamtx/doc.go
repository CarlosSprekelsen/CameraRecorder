// Package mediamtx implements the MediaMTX controller and integration layer.
//
// This package serves as the central orchestration component (Layer 5) coordinating
// all video operations and business logic. It provides API abstraction between
// external identifiers (camera0, camera1) and internal device paths (/dev/videoN).
//
// Architecture Compliance:
//   - Single Source of Truth: All business logic resides in the controller
//   - Interface-Based Design: All major components defined as interfaces
//   - Dependency Inversion: High-level interfaces for low-level implementations
//   - Event-Driven Architecture: Event notification interfaces for real-time updates
//   - Optional Component Pattern: Interfaces support nil implementations
//
// Key Responsibilities:
//   - Camera operations coordination and abstraction layer management
//   - Stream lifecycle management with path reuse optimization
//   - Recording orchestration using stateless MediaMTX API queries
//   - Snapshot capture with multi-tier fallback (V4L2 → FFmpeg → RTSP)
//   - Health monitoring and system readiness coordination
//   - Event notification for real-time client updates
//
// Interface Categories:
//   - MediaMTX Integration: Client, PathManager, StreamManager interfaces
//   - Event Notification: MediaMTXEventNotifier, SystemEventNotifier for real-time updates
//   - Business Logic: RecordingManager, SnapshotManager for high-level operations
//   - Health Monitoring: HealthMonitor interface with circuit breaker support
//   - Configuration: ConfigIntegration for centralized configuration access
//   - API Abstraction: DeviceToCameraIDMapper for camera0 ↔ /dev/video0 mapping
//
// Requirements Coverage:
//   - REQ-MTX-001: MediaMTX service integration with REST API
//   - REQ-MTX-002: Stream management with on-demand FFmpeg processes
//   - REQ-MTX-003: Path creation, deletion, and lifecycle management
//   - REQ-MTX-004: Health monitoring with circuit breaker pattern
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/api/json_rpc_methods.md
package mediamtx
