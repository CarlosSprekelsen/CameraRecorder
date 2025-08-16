# Architecture Overview

## Change Log

- **2025-08-13**: Updated architecture to reflect validated MediaMTX FFmpeg integration pattern. Added MediaMTX Path Manager component, updated data flow to show Camera → FFmpeg → MediaMTX → Clients, documented API-based dynamic path creation, and added comprehensive FFmpeg integration guidance for developers.
- **2025-08-02**: Reorganized to separate core architecture from implementation decisions and examples per project standards and IV&V. Moved all implementation rationale, migration plans, version-specific details, and design notes to dedicated "Architecture Decisions & Design Notes" section.

## How to Use This Document

**Ground Truth and Scope**: This document defines the stable, review-approved architecture and interfaces for the MediaMTX Camera Service. All implementation decisions, migration notes, and future enhancements are segregated in the "Architecture Decisions & Design Notes" section below.

- **Core Architecture (above)**: Current system design, stable component responsibilities, and approved interfaces
- **Decisions & Notes (below)**: Implementation rationale, version-specific details, and future planning context

## Status

**Architecture Status**: APPROVED  
All core components and interfaces are finalized and ready for implementation.

**Implementation Readiness Criteria Met**:
- ✅ Component interfaces fully specified with data structures (Lines 258-275)
- ✅ Integration patterns defined with specific protocols (Lines 86-104)  
- ✅ Performance targets quantified with measurable thresholds (Lines 223-227)
- ✅ Technology stack specified with version requirements (Lines 213-219)
- ✅ Deployment architecture documented with operational procedures (Lines 241-254)

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
│            MediaMTX Path Manager                          │
│     • Dynamic path creation via REST API                  │
│     • FFmpeg command generation                           │
│     • Path lifecycle management                           │
│     • Error handling and recovery                         │
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
│     • FFmpeg process management                           │
│     • Multi-protocol support                              │
│     • Recording and snapshot generation                   │
└─────────────────────┬───────────────────────────────────────┘
                      │ FFmpeg Processes
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

#### MediaMTX Path Manager
- MediaMTX REST API communication for dynamic path management
- FFmpeg command generation with optimal encoding parameters
- Path lifecycle management (create/delete/verify)
- Error handling and recovery for path operations
- Automatic path creation on camera detection

#### Health & Monitoring
- Service component health verification
- Resource usage tracking
- Error detection and recovery coordination
- Configuration validation and hot-reload

## Data Flow

### Camera Discovery Flow
1. Monitor detects USB camera connection
2. Detector probes camera capabilities
3. Path Manager creates MediaMTX path via REST API with FFmpeg command
4. Health Monitor verifies path creation and FFmpeg process start
5. Server broadcasts camera status notification to clients
6. Recovery Handler manages connection failures

### Streaming Flow  
1. Client requests stream via JSON-RPC call
2. Authorization validates client permissions
3. Path Manager verifies existing MediaMTX path or creates new one
4. MediaMTX starts FFmpeg process for camera capture and encoding
5. Health Monitor verifies stream establishment
6. Client accesses stream via provided RTSP/WebRTC/HLS URL

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
- FFmpeg process monitoring and restart

### Camera Disconnect Handling  
- Event-based disconnect detection
- Automatic MediaMTX path cleanup via REST API
- FFmpeg process termination and cleanup
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

## MediaMTX FFmpeg Integration Pattern

### Integration Architecture
The system uses MediaMTX as a media server with FFmpeg as the camera capture and encoding bridge. This approach provides:

- **Dynamic Path Creation**: MediaMTX paths created via REST API on camera detection
- **FFmpeg Bridge**: FFmpeg processes handle camera capture and encoding
- **Automatic Management**: Path creation and cleanup handled automatically
- **Multi-Protocol Support**: RTSP, WebRTC, and HLS streaming from single FFmpeg source

### Data Flow: Camera → FFmpeg → MediaMTX → Clients
1. **Camera Detection**: USB camera detected via V4L2
2. **Path Creation**: MediaMTX path created via REST API with FFmpeg command
3. **FFmpeg Process**: MediaMTX starts FFmpeg process for camera capture
4. **Stream Publishing**: FFmpeg publishes encoded stream to MediaMTX
5. **Client Access**: Clients access streams via MediaMTX protocols (RTSP/WebRTC/HLS)

### FFmpeg Command Template
```bash
ffmpeg -f v4l2 -i {device_path} -c:v libx264 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:{rtsp_port}/{path_name}
```

**Parameters**:
- `-f v4l2`: Video4Linux2 input format
- `-i {device_path}`: Camera device path (e.g., /dev/video0)
- `-c:v libx264`: H.264 video codec
- `-pix_fmt yuv420p`: Widely compatible pixel format
- `-preset ultrafast`: Minimal encoding latency
- `-b:v 600k`: Balanced quality/bandwidth
- `-f rtsp`: RTSP output format

### Dynamic vs. Static Configuration
- **Dynamic Configuration**: Paths created via MediaMTX REST API on camera detection
- **No Static Files**: No manual MediaMTX configuration files required
- **Automatic Management**: Path creation and deletion handled by Path Manager
- **Real-time Updates**: Configuration changes applied without service restart

### Path Lifecycle Management
1. **Creation**: Path created when camera detected
2. **Configuration**: Path configured with `runOnDemand` FFmpeg command
3. **On-Demand Activation**: FFmpeg process starts only when stream accessed
4. **Monitoring**: FFmpeg process health monitored during active periods
5. **Cleanup**: Path deleted when camera disconnected
6. **Recovery**: Automatic restart on FFmpeg process failure

### On-Demand Stream Activation
The system implements power-efficient on-demand stream activation:

- **Initial State**: Paths created but FFmpeg processes not running (`source: null`, `ready: false`)
- **Activation Trigger**: First access to stream (recording, snapshot, or streaming request)
- **Process Start**: MediaMTX starts FFmpeg process via `runOnDemand` configuration
- **Stream Ready**: FFmpeg publishes stream to MediaMTX (`source: {...}`, `ready: true`)
- **Power Efficiency**: No unnecessary processes running when not needed

**Configuration Impact**:
- `auto_start_streams: true` creates MediaMTX paths on camera detection
- FFmpeg processes start on-demand when operations are requested
- Provides optimal balance of responsiveness and power efficiency

## Non-Functional Requirements

### Performance Targets
- Camera Detection: Sub-200ms USB connect/disconnect detection
- Path Creation: <100ms (API call + verification)
- API Response: <50ms for status queries, <100ms for control operations
- FFmpeg Process Start: <200ms from path creation to stream availability
- Memory Usage: <30MB base service footprint, <100MB with 10 cameras
- CPU Usage: <5% idle, <20% with active streaming and recording

### Python/Go Integration Performance
- MediaMTX REST API Calls: <10ms per request (local HTTP)
- Cross-Language Data Serialization: <1ms for typical payloads (<10KB)
- Process Communication Overhead: <5% CPU impact under normal load
- Error Propagation Latency: <20ms for service-to-client error reporting
- FFmpeg Process Management: <50ms process start/stop operations

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

### Decision AD-11: MediaMTX FFmpeg Integration Pattern
**Date**: 2025-08-13  
**Components**: MediaMTX Path Manager, MediaMTX Server  
**Decision**: Use MediaMTX as media server with FFmpeg as camera capture and encoding bridge via API-driven dynamic path creation.  
**Rationale**: Provides optimal balance of functionality, reliability, and maintainability while preserving the "plug-and-play" concept.  
**Implementation Specifications**:
- Dynamic path creation via MediaMTX REST API on camera detection
- FFmpeg command template with optimized encoding parameters
- Automatic path lifecycle management (create/verify/monitor/cleanup)
- FFmpeg process monitoring and automatic restart on failure
- No manual MediaMTX configuration files required
**Validation**: IV&V confirmed 100% success rate with 4 cameras, all streams accessible  
**IV&V Reference**: MediaMTX FFmpeg Validation Report, 2025-08-13

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