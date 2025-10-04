// Package camera provides camera device discovery, monitoring, and management
// for the MediaMTX Camera Service.
//
// This package implements hybrid camera support (USB, IP, RTSP) through a
// unified interface-based architecture with real-time event-driven monitoring.
//
// Architecture Compliance:
//   - Interface-Based Design: DeviceChecker, V4L2CommandExecutor, DeviceInfoParser
//   - Event-Driven Architecture: Real-time camera connection/disconnection events
//   - Dependency Injection: All implementations injected via constructor
//   - Performance Targets: <100ms discovery, <500ms capability probing
//
// Core Components:
//   - HybridCameraMonitor: Central orchestrator for camera lifecycle management
//   - RealDeviceChecker: File system device existence validation
//   - RealV4L2CommandExecutor: V4L2 command execution via v4l2-ctl
//   - RealDeviceInfoParser: Device capability and format parsing
//
// Architecture References:
//   - Architecture §5: Camera Discovery and Management
//   - CB-TIMING §3: Device discovery performance requirements
//   - OpenAPI §4: Camera status and capability endpoints
//
// Test Strategy:
//   - interfaces_test.go: Interface contract validation with real implementations
//   - real_hardware_test.go: Integration testing with actual camera hardware
//   - No unit tests for real_implementations.go (tested via integration)
//
// Performance Characteristics:
//   - Device Discovery: <100ms for device enumeration
//   - Capability Probing: <500ms per device
//   - Event Delivery: <20ms latency
//   - Memory Footprint: <10MB base
//
// Requirements Coverage:
//   - REQ-CAM-001: USB camera discovery and enumeration
//   - REQ-CAM-002: V4L2 capability detection
//   - REQ-CAM-003: Real-time device status monitoring
//   - REQ-CAM-004: Event notification system
package camera
