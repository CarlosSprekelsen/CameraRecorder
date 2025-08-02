# Architecture Overview

## Change Log

- **2025-08-02**: Reorganized to separate core architecture from implementation decisions and examples per project standards and IV&V. Moved all implementation rationale, migration plans, version-specific details, and design notes to dedicated "Architecture Decisions & Design Notes" section.

## How to Use This Document

**Ground Truth and Scope**: This document defines the stable, review-approved architecture and interfaces for the MediaMTX Camera Service. All implementation decisions, migration notes, and future enhancements are segregated in the "Architecture Decisions & Design Notes" section below.

- **Core Architecture (above)**: Current system design, stable component responsibilities, and approved interfaces
- **Decisions & Notes (below)**: Implementation rationale, version-specific details, and future planning context

## Status

**Architecture Status**: APPROVED  
All core components and interfaces are finalized and ready for implementation.

---

## System Design

The MediaMTX Camera Service is a lightweight wrapper around MediaMTX, providing:

1. **Real-time USB camera discovery and monitoring**
2. **WebSocket JSON-RPC 2.0 API**
3. **Dynamic MediaMTX configuration management**
4. **Streaming, recording, and snapshot coordination**
5. **Resilient error recovery and health monitoring**
6. **Secure access control and authentication**

## Component Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Client Applications                      │
│            (Web browsers, mobile apps, etc.)               │
└─────────────────────┬──────────────────────────────────────┘
                      │ WebSocket JSON-RPC 2.0
┌─────────────────────▼──────────────────────────────────────┐
│                Camera Service (Python)                     │
├─────────────────────────────────────────────────────────────┤
│            WebSocket JSON-RPC Server                      │
│     • Client connection management                         │
│     • JSON-RPC 2.0 protocol handling                      │
│     • Real-time notifications                             │
│     • Authentication and authorization                     │
├─────────────────────────────────────────────────────────────┤
│             Camera Discovery Monitor                      │
│     • USB camera detection                                │
│     • Camera status tracking                              │
│     • Hot-plug event handling                             │
├─────────────────────────────────────────────────────────────┤
│              MediaMTX Controller                          │
│     • REST API client                                     │
│     • Stream management                                   │
│     • Recording coordination                              │
│     • Health monitoring                                   │
├─────────────────────────────────────────────────────────────┤
│               Health & Monitoring                         │
│     • Service health checks                               │
│     • Resource usage monitoring                           │
│     • Error tracking and recovery                         │
│     • Configuration management                            │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP REST API
┌─────────────────────▼───────────────────────────────────────┐
│                   MediaMTX Server (Go)                      │
├─────────────────────────────────────────────────────────────┤
│                Media Processing                           │
│     • RTSP/WebRTC/HLS streaming                           │
│     • Hardware-accelerated encoding                       │
│     • Multi-protocol support                              │
│     • Recording and snapshot generation                   │
└─────────────────────┬───────────────────────────────────────┘
                      │ FFmpeg + V4L2
┌─────────────────────▼───────────────────────────────────────┐
│                 USB Cameras                                 │
│         /dev/video0, /dev/video1, etc.                     │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### WebSocket JSON-RPC Server
- Client connection management and authentication
- JSON-RPC 2.0 protocol implementation
- Real-time event notifications
- API method routing and response handling

#### Camera Discovery Monitor
- USB camera detection and enumeration
- Device capability probing
- Connection status tracking
- Event generation for connect/disconnect

#### MediaMTX Controller  
- MediaMTX REST API communication
- Stream path creation and deletion
- Recording session management
- Configuration updates

#### Health & Monitoring
- Service component health verification
- Resource usage tracking
- Error detection and recovery coordination
- Configuration validation and hot-reload

## Data Flow

### Camera Discovery Flow
1. Monitor detects USB camera connection
2. Detector probes camera capabilities
3. Controller creates MediaMTX stream configuration
4. Health Monitor verifies configuration acceptance
5. Server broadcasts camera status notification to clients
6. Recovery Handler manages connection failures

### Streaming Flow  
1. Client requests stream via JSON-RPC call
2. Authorization validates client permissions
3. Controller configures MediaMTX path with source
4. Health Monitor verifies stream establishment
5. MediaMTX starts camera capture and encoding
6. Client accesses stream via provided URL

### Recording Flow
1. Client requests recording start via JSON-RPC
2. Authorization validates recording permissions
3. Controller enables recording in MediaMTX
4. Health Monitor tracks recording status
5. MediaMTX captures video to file
6. Server notifies client of completion or failure

## Error Recovery and Resilience

### MediaMTX Failure Handling
- Continuous REST API health monitoring
- Automatic service restart with backoff
- Circuit breaker for repeated failures
- State preservation and restoration
- Client notification of status changes
- Graceful degradation when unavailable

### Camera Disconnect Handling  
- Event-based disconnect detection
- Automatic MediaMTX path cleanup
- Real-time client notifications
- Automatic reconnection handling
- Configurable timeout management

### Network and System Failures
- WebSocket connection resilience
- Configuration backup and rollback
- Resource usage monitoring
- Rate limiting and abuse protection

## Security Model

### API Authentication
- Token-based authentication
- Role-based access control
- Session management
- Certificate-based authentication support

### Stream Access Control
- Time-limited access tokens
- IP address restrictions
- Per-camera permissions
- Session tracking

### System Security
- Process isolation with minimal privileges
- File system access restrictions
- Configurable network binding
- Comprehensive audit logging

## Configuration Management

### Configuration Sources
- Primary YAML configuration files
- Environment variable overrides
- Runtime configuration updates
- Schema validation

### Configuration Features
- Hot reload without service restart
- Change validation and rollback
- Version tracking and backup
- Real-time update notifications

## Health Monitoring and Observability

### Health Endpoints
- Service liveness checking
- Readiness verification
- Backend connectivity status
- Camera summary reporting

### Metrics and Monitoring
- Performance and error metrics
- Resource usage tracking
- Camera connection metrics
- Integration response metrics

### Logging and Diagnostics
- Structured logging with correlation IDs
- Configurable log levels and rotation
- Diagnostic and troubleshooting endpoints

## Technology Stack

- **Camera Service**: Python 3.10+, asyncio, websockets, aiohttp
- **Media Server**: MediaMTX (Go binary)
- **Camera Interface**: V4L2, FFmpeg 6.0+
- **Protocols**: WebSocket, JSON-RPC 2.0, REST, RTSP, WebRTC, HLS
- **Security**: JWT, TLS 1.3, optional mTLS
- **Deployment**: Systemd services, Linux (Ubuntu 22.04+)
- **Monitoring**: Prometheus metrics, structured JSON logging

## Non-Functional Requirements

### Performance Targets
- Camera Detection: Sub-200ms USB connect/disconnect detection
- API Response: <50ms for status queries, <100ms for control operations
- Memory Usage: <30MB base service footprint, <100MB with 10 cameras
- CPU Usage: <5% idle, <20% with active streaming and recording

### Scalability Limits
- Concurrent Cameras: Up to 16 USB cameras per service instance
- WebSocket Connections: Up to 100 concurrent client connections  
- Recording Capacity: Limited by available disk space and I/O bandwidth
- Stream Bandwidth: Limited by network capacity and MediaMTX performance

### Reliability Requirements
- Uptime Target: 99.9% availability excluding planned maintenance
- Recovery Time: <30 seconds for service restart, <10 seconds for camera reconnect
- Data Integrity: Zero-loss recording with atomic file operations
- Error Rate: <0.1% API call failure rate under normal conditions

## Deployment Architecture

### Production Deployment
- Service isolation with dedicated system user
- Systemd process management with automatic restart
- Configurable network ports with firewall integration
- Structured file system layout under `/opt/camera-service/`
- Integration with system logging and log rotation

### Development Environment
- Native Linux development and testing
- Mock services for camera and MediaMTX simulation
- Enhanced logging and diagnostic capabilities
- Hot reload for configuration and code changes

## Supported Data Structures

### Camera Status Response Fields
**Standard Fields** (always included):
- `device`: Camera device path (string)
- `status`: Connection status (string: "CONNECTED", "DISCONNECTED", "ERROR")
- `name`: Camera display name (string)
- `resolution`: Current resolution setting (string, e.g., "1920x1080")
- `fps`: Current frame rate (number)
- `streams`: Available stream URLs (object with rtsp, webrtc, hls fields)

**Optional Fields** (implementation-dependent):
- `metrics`: Performance metrics (object with bytes_sent, readers, uptime fields)
  - *Note*: Inclusion pending clarification per roadmap blocked item
- `capabilities`: Device capabilities (object, populated when capability detection enabled)

### Notification Payload Fields
**Camera Status Update**: Uses same fields as Camera Status Response  
**Recording Status Update**: device, status, filename, duration fields

---

## Architecture Decisions & Design Notes

### Decision AD-1: MediaMTX Version Compatibility Strategy
**Date**: 2025-01-15  
**Components**: MediaMTX Controller, Technology Stack  
**Decision**: Target latest stable MediaMTX version, pin minimum version after compatibility confirmation, document upgrade and testing procedures.  
**Rationale**: Balance modern features with stability. Minimum version pinning prevents compatibility issues.  
**IV&V Reference**: Architecture Decisions v6, item 1

### Decision AD-2: Camera Discovery Implementation Method
**Date**: 2025-01-15  
**Components**: Camera Discovery Monitor  
**Decision**: Hybrid udev + polling approach with configurable switching based on environment.  
**Rationale**: Udev provides real-time events but may be unreliable in some environments. Polling provides guaranteed coverage with configurable intervals (0.1s default).  
**Implementation Notes**: Environment variable `CAMERA_DISCOVERY_METHOD` allows override to "udev", "polling", or "hybrid".  
**IV&V Reference**: Architecture Decisions v6, item 2

### Decision AD-3: Configuration Management Strategy
**Date**: 2025-01-15  
**Components**: Configuration Management, Health & Monitoring  
**Decision**: YAML primary configuration with environment variable overrides and comprehensive schema validation.  
**Rationale**: Provides flexibility for different deployment scenarios while maintaining configuration integrity.  
**Implementation Details**: 
- Configuration hierarchy: defaults < YAML file < environment variables
- Schema validation using JSON Schema before applying changes
- Hot reload triggers validation and rollback on failure
- Example environment override: `CAMERA_SERVICE_PORT=8003` overrides `server.port`  
**IV&V Reference**: Architecture Decisions v6, item 3

### Decision AD-4: Error Recovery Strategy Implementation
**Date**: 2025-01-15  
**Components**: Health & Monitoring, MediaMTX Controller  
**Decision**: Multi-layered approach with health monitoring, exponential backoff, circuit breaker pattern, and structured health event logging.  
**Rationale**: Ensures service resilience and automatic recovery from transient failures.  
**Implementation Specifications**:
- Health checks every 5 seconds with 10-second timeout
- Exponential backoff: 1s, 2s, 4s, 8s, max 60s intervals
- Circuit breaker: Open after 5 consecutive failures, half-open after 30s
- Health events logged with correlation IDs for traceability  
**IV&V Reference**: Architecture Decisions v6, item 4

### Decision AD-5: API Versioning Strategy
**Date**: 2025-01-15  
**Components**: WebSocket JSON-RPC Server  
**Decision**: Method-level JSON-RPC versioning with structured deprecation and migration support.  
**Rationale**: Allows independent evolution of API methods while maintaining backward compatibility.  
**Implementation Framework**:
- Version header in method calls: `{"method": "get_camera_list", "version": "1.1"}`
- Deprecation warnings in responses for old versions
- Migration guides for version transitions
- Version support matrix documentation  
**Future Implementation**: Version negotiation during WebSocket handshake  
**IV&V Reference**: Architecture Decisions v6, item 5

### Decision AD-6: API Protocol Selection
**Date**: 2025-01-15  
**Components**: WebSocket JSON-RPC Server  
**Decision**: WebSocket-only JSON-RPC with minimal REST endpoints for health checks only.  
**Rationale**: WebSocket provides real-time notifications essential for camera events. REST limited to `/health/*` endpoints for monitoring systems.  
**Implementation Note**: No REST API for camera control to maintain protocol consistency.  
**IV&V Reference**: Architecture Decisions v6, item 6

### Decision AD-7: Authentication Strategy
**Date**: 2025-01-15  
**Components**: Security Model, WebSocket JSON-RPC Server  
**Decision**: JWT tokens for user authentication, API keys for service authentication, optional mTLS for high-security deployments.  
**Rationale**: Provides flexible authentication suitable for different deployment scenarios and security requirements.  
**Implementation Details**:
- JWT with configurable expiry (default 24 hours)
- API keys stored in secure configuration with bcrypt hashing
- mTLS certificate validation for enterprise environments
- Role-based access: viewer, operator, admin permission levels  
**IV&V Reference**: Architecture Decisions v6, item 7

### Decision AD-8: Logging Format Strategy
**Date**: 2025-01-15  
**Components**: Health & Monitoring, All Components  
**Decision**: Structured JSON logging with correlation IDs, human-readable format for development environments.  
**Rationale**: Enables effective log aggregation and troubleshooting while maintaining developer productivity.  
**Implementation Specifications**:
- Production: JSON format with correlation IDs for log aggregation
- Development: Human-readable console format with correlation IDs
- Correlation ID propagation through all request/response cycles
- Log levels: ERROR, WARN, INFO, DEBUG with configurable filtering  
**Example JSON Log**: `{"timestamp": "2025-01-15T14:30:00Z", "level": "INFO", "correlation_id": "req-abc123", "message": "Camera connected", "device": "/dev/video0"}`  
**IV&V Reference**: Architecture Decisions v6, item 8

### Decision AD-9: Performance Target Specifications
**Date**: 2025-01-15  
**Components**: All Components  
**Decision**: Specific performance targets with monitoring and alerting thresholds.  
**Rationale**: Ensures consistent performance characteristics and enables proactive monitoring.  
**Target Details**:
- Camera detection: <200ms (hybrid udev + polling approach)
- API response: <50ms status queries, <100ms control operations
- Concurrent support: 16 cameras, 100 WebSocket clients maximum
- Resource limits: <100MB RAM total, <20% CPU usage with full load  
**Monitoring Integration**: Prometheus metrics with alerting on threshold breaches  
**IV&V Reference**: Architecture Decisions v6, item 9

### Decision AD-10: Resource Management Strategy
**Date**: 2025-01-15  
**Components**: Health & Monitoring, Configuration Management  
**Decision**: Comprehensive resource limits with automatic cleanup and monitoring.  
**Rationale**: Prevents resource exhaustion and maintains system stability over extended operation.  
**Implementation Specifications**:
- Memory limit: <100MB total service footprint
- CPU monitoring with <20% sustained usage target
- Log rotation: 500MB maximum with 5 backup files
- Disk usage warnings at 80% capacity for recordings directory
- Automatic cleanup of recordings older than 30 days (configurable)  
**Monitoring**: Resource usage tracked and exposed via health endpoints  
**IV&V Reference**: Architecture Decisions v6, item 10

### Future Extensibility Framework (Not Current Implementation)
**Note**: The following items are planned extensions NOT included in the current architecture implementation. These are design considerations for post-1.0 releases.

**Components**: All Components (Future)  
**Future Enhancement Categories**:
- **Camera Source Extensions**: Network cameras (RTSP/HTTP), virtual cameras, mobile device integration, plugin architecture for custom drivers
- **Protocol Extensions**: SRT, NDI, custom protocols, cloud integration (AWS Kinesis, Azure Media Services), message queue integration (MQTT, RabbitMQ)
- **Advanced Features**: AI integration (object detection, analytics), multi-node distributed clusters, cloud storage automation, synchronized multi-camera recording

**Implementation Note**: Extensibility hooks planned but not implemented in initial version. Plugin architecture design deferred to post-1.0 release.

### Version Compatibility Matrix
**MediaMTX Integration**:
- Minimum supported: v0.23.x (to be confirmed during integration testing)
- Target version: Latest stable at release time
- Compatibility testing required for major MediaMTX version updates
- Breaking change handling: Version-specific adapter pattern

**Deprecation Process**:
1. Announce deprecation with 2 minor version notice period
2. Include migration documentation and automated tooling where possible
3. Maintain backward compatibility during deprecation period
4. Remove deprecated features only in major version releases