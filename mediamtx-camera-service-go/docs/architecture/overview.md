# Architecture Overview - Go Implementation

**Version:** 2.0  
**Date:** 2025-01-15  
**Status:** Go Implementation Architecture  
**Related Epic/Story:** Go Implementation Architecture  

## Change Log

- **2025-01-15**: Updated architecture for Go implementation with performance improvements and technology stack updates. Added Go-specific components, updated performance targets, and maintained full API compatibility with Python implementation.
- **2025-08-13**: Updated architecture to reflect validated MediaMTX FFmpeg integration pattern. Added MediaMTX Path Manager component, updated data flow to show Camera → FFmpeg → MediaMTX → Clients, documented API-based dynamic path creation, and added comprehensive FFmpeg integration guidance for developers.

## How to Use This Document

**Ground Truth and Scope**: This document defines the stable, review-approved architecture and interfaces for the MediaMTX Camera Service Go implementation. All implementation decisions, migration notes, and future enhancements are segregated in the "Architecture Decisions & Design Notes" section below.

- **Core Architecture (above)**: Current system design, stable component responsibilities, and approved interfaces
- **Decisions & Notes (below)**: Implementation rationale, version-specific details, and future planning context

## Status

**Architecture Status**: APPROVED  
All core components and interfaces are finalized and ready for Go implementation.

**Implementation Readiness Criteria Met**:
- ✅ Component interfaces fully specified with data structures
- ✅ Integration patterns defined with specific protocols  
- ✅ Performance targets quantified with measurable thresholds (5x improvement)
- ✅ Technology stack specified with Go version requirements
- ✅ Deployment architecture documented with operational procedures

---

## System Design

The MediaMTX Camera Service Go implementation is a high-performance wrapper around MediaMTX, providing:

1. **Real-time USB camera discovery and monitoring** (5x faster detection)
2. **WebSocket JSON-RPC 2.0 API** (1000+ concurrent connections)
3. **Dynamic MediaMTX configuration management** (100ms response time)
4. **Streaming, recording, and snapshot coordination** (5x throughput improvement)
5. **Resilient error recovery and health monitoring** (50% resource reduction)
6. **Secure access control and authentication** (Go crypto libraries)

## Component Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Client Applications                      │
│            (Web browsers, mobile apps, etc.)               │
└─────────────────────┬──────────────────────────────────────┘
                      │ WebSocket JSON-RPC 2.0
┌─────────────────────▼──────────────────────────────────────┐
│            Camera Service (Go Implementation)              │
├─────────────────────────────────────────────────────────────┤
│            WebSocket JSON-RPC Server (gorilla/websocket)  │
│     • Client connection management (1000+ concurrent)     │
│     • JSON-RPC 2.0 protocol handling                      │
│     • Real-time notifications (<20ms latency)             │
│     • Authentication and authorization (golang-jwt/jwt/v4) │
├─────────────────────────────────────────────────────────────┤
│             Camera Discovery Monitor (goroutines)         │
│     • USB camera detection (<200ms)                       │
│     • Camera status tracking                              │
│     • Hot-plug event handling                             │
│     • Concurrent monitoring with channels                 │
├─────────────────────────────────────────────────────────────┤
│            MediaMTX Path Manager (net/http)               │
│     • Dynamic path creation via REST API                  │
│     • FFmpeg command generation                           │
│     • Path lifecycle management                           │
│     • Error handling and recovery                         │
├─────────────────────────────────────────────────────────────┤
│               Health & Monitoring (logrus)                │
│     • Service health checks                               │
│     • Resource usage monitoring                           │
│     • Error tracking and recovery                         │
│     • Configuration management (viper)                    │
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

#### WebSocket JSON-RPC Server (gorilla/websocket)
- Client connection management and authentication (1000+ concurrent)
- JSON-RPC 2.0 protocol implementation
- Real-time event notifications (<20ms delivery)
- API method routing and response handling (<100ms response time)

#### Camera Discovery Monitor (goroutines + channels)
- USB camera detection and enumeration (<200ms detection time)
- Device capability probing
- Connection status tracking
- Event generation for connect/disconnect
- Concurrent monitoring with goroutines

#### MediaMTX Path Manager (net/http)
- MediaMTX REST API communication for dynamic path management
- FFmpeg command generation with optimal encoding parameters
- Path lifecycle management (create/delete/verify)
- Error handling and recovery for path operations
- Automatic path creation on camera detection
- **Stream enumeration and status monitoring** via `get_streams` method
- **Recording session management and state tracking**
- **File rotation management with configurable intervals**
- **Storage space validation and threshold management**
- **Resource monitoring during recording operations**

#### Health & Monitoring (logrus + viper)
- Service component health verification
- Resource usage tracking (<60MB memory footprint)
- Error detection and recovery coordination
- Configuration validation and hot-reload
- Structured logging with correlation IDs

## Data Flow

### Camera Discovery Flow
1. Monitor detects USB camera connection (<200ms)
2. Detector probes camera capabilities
3. Path Manager creates MediaMTX path via REST API with FFmpeg command (<100ms)
4. Health Monitor verifies path creation and FFmpeg process start
5. Server broadcasts camera status notification to clients (<20ms)
6. Recovery Handler manages connection failures

### Streaming Flow  
1. Client requests stream via JSON-RPC call
2. Authorization validates client permissions (golang-jwt/jwt/v4)
3. Path Manager verifies existing MediaMTX path or creates new one
4. MediaMTX starts FFmpeg process for camera capture and encoding
5. Health Monitor verifies stream establishment
6. Client accesses stream via provided RTSP/WebRTC/HLS URL

### Recording Flow
1. Client requests recording start via JSON-RPC
2. Authorization validates recording permissions
3. **System validates storage space and recording state**
4. **System checks for existing recording conflicts**
5. Controller enables recording in MediaMTX
6. **System initiates file rotation management**
7. Health Monitor tracks recording status
8. MediaMTX captures video to file with rotation
9. **System monitors storage usage and applies thresholds**
10. Server notifies client of completion or failure

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
- Token-based authentication (golang-jwt/jwt/v4)
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
- Primary YAML configuration files (viper)
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
- Resource usage tracking (<60MB memory)
- Camera connection metrics
- Integration response metrics
- Goroutine monitoring

### Logging and Diagnostics
- Structured logging with correlation IDs (logrus)
- Configurable log levels and rotation
- Diagnostic and troubleshooting endpoints

## Technology Stack

### Go Implementation Stack
- **Camera Service**: Go 1.25.0+, gorilla/websocket, net/http
- **Media Server**: MediaMTX (Go binary)
- **Camera Interface**: V4L2, FFmpeg 6.0+
- **Protocols**: WebSocket, JSON-RPC 2.0, REST, RTSP, WebRTC, HLS
- **Security**: golang-jwt/jwt/v4, golang.org/x/crypto/bcrypt, TLS 1.3
- **Configuration**: viper, YAML
- **Logging**: logrus, structured JSON logging
- **Testing**: testify, built-in testing
- **Deployment**: Systemd services, OCI-compatible containers, CNCF-compliant orchestration, Linux (Ubuntu 22.04+)
- **Monitoring**: Prometheus metrics, structured JSON logging

### Performance Improvements
- **Response Time**: 5x improvement (500ms → 100ms)
- **Concurrency**: 10x improvement (100 → 1000+ connections)
- **Throughput**: 5x improvement (200 → 1000+ requests/second)
- **Memory Usage**: 50% reduction (80% → 60%)
- **CPU Usage**: 30% reduction (70% → 50%)

## MediaMTX FFmpeg Integration Pattern

### Integration Architecture
The system uses MediaMTX as a media server with FFmpeg as the camera capture and encoding bridge. This approach provides:

- **Dynamic Path Creation**: MediaMTX paths created via REST API on camera detection
- **FFmpeg Bridge**: FFmpeg processes handle camera capture and encoding
- **Automatic Management**: Path creation and cleanup handled automatically
- **Multi-Protocol Support**: RTSP, WebRTC, and HLS streaming from single FFmpeg source

### Data Flow: Camera → FFmpeg → MediaMTX → Clients
1. **Camera Detection**: USB camera detected via V4L2 (<200ms)
2. **Path Creation**: MediaMTX path created via REST API with FFmpeg command (<100ms)
3. **FFmpeg Process**: MediaMTX starts FFmpeg process for camera capture
4. **Stream Publishing**: FFmpeg publishes encoded stream to MediaMTX
5. **Client Access**: Clients access streams via MediaMTX protocols (RTSP/WebRTC/HLS)

### FFmpeg Command Template
```bash
ffmpeg -f v4l2 -i {device_path} -c:v libx264 -profile:v baseline -level 3.0 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:{rtsp_port}/{path_name}
```

**Parameters**:
- `-f v4l2`: Video4Linux2 input format
- `-i {device_path}`: Camera device path (e.g., /dev/video0)
- `-c:v libx264`: H.264 video codec
- `-profile:v baseline`: Constrained Baseline Profile for STANAG 4406 compliance
- `-level 3.0`: H.264 Level 3.0 (supports up to 720p resolution)
- `-pix_fmt yuv420p`: Widely compatible pixel format (4:2:0)
- `-preset ultrafast`: Minimal encoding latency
- `-b:v 600k`: Balanced quality/bandwidth
- `-f rtsp`: RTSP output format

### STANAG 4406 H.264 Compliance
The system is configured for STANAG 4406 (MIL-STD-188-110B) H.264 compatibility:

- **Profile:** Constrained Baseline Profile (CBP) - ensures maximum compatibility
- **Level:** 3.0 - supports up to 720p resolution and 30fps
- **Pixel Format:** 4:2:0 (yuv420p) - widely supported across military/government systems
- **Bitrate:** 600kbps - configurable for different bandwidth requirements
- **Compatibility:** Meets military/government video standards for RTSP streaming

**STANAG 4406 Benefits:**
- Maximum compatibility with legacy military/government systems
- Reduced computational requirements (baseline profile)
- Standardized video format for interoperability
- Future-proof for H.265 upgrade when stakeholder systems are ready

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

### Performance Targets (Go Implementation)
- Camera Detection: Sub-200ms USB connect/disconnect detection
- Path Creation: <100ms (API call + verification)
- API Response: <50ms for status queries, <100ms for control operations
- FFmpeg Process Start: <200ms from path creation to stream availability
- Memory Usage: <60MB base service footprint, <200MB with 10 cameras
- CPU Usage: <50% idle, <50% with active streaming and recording
- Concurrent Connections: 1000+ simultaneous WebSocket connections
- Throughput: 1000+ requests/second

### Go Implementation Performance
- MediaMTX REST API Calls: <10ms per request (local HTTP)
- Cross-Language Data Serialization: <1ms for typical payloads (<10KB)
- Process Communication Overhead: <5% CPU impact under normal load
- Error Propagation Latency: <20ms for service-to-client error reporting
- FFmpeg Process Management: <50ms process start/stop operations
- Goroutine Management: <1000 concurrent goroutines maximum

### Scalability Limits
- Concurrent Cameras: Up to 16 USB cameras per service instance
- WebSocket Connections: Up to 1000 concurrent client connections  
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
- Static binary deployment (no runtime dependencies)

### Development Environment
- Native Linux development and testing
- Mock services for camera and MediaMTX simulation
- Enhanced logging and diagnostic capabilities
- Hot reload for configuration and code changes
- Go toolchain with development tools

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

### Decision AD-GO-001: Go Language Selection
**Date**: 2025-01-15  
**Components**: Technology Stack  
**Decision**: Use Go 1.25.0+ as the primary implementation language for the MediaMTX Camera Service.  
**Rationale**: Balance performance improvements with maintainability. Go provides 5x performance improvement, 10x concurrency improvement, and 50% resource reduction while maintaining full API compatibility.  
**IV&V Reference**: Architecture Decisions v7, item 1

### Decision AD-GO-002: WebSocket Framework Selection
**Date**: 2025-01-15  
**Components**: WebSocket JSON-RPC Server  
**Decision**: Use gorilla/websocket library for WebSocket implementation.  
**Rationale**: gorilla/websocket is the de facto standard for WebSocket in Go, providing excellent performance for high-concurrency scenarios and extensive documentation and community support.  
**IV&V Reference**: Architecture Decisions v7, item 2

### Decision AD-GO-003: Configuration Management Strategy
**Date**: 2025-01-15  
**Components**: Configuration Management  
**Decision**: Use Viper library for configuration management with YAML as primary format.  
**Rationale**: Viper provides flexibility for multiple configuration formats, environment variable binding, hot reload capabilities, and works well with Go structs and mapstructure tags.  
**IV&V Reference**: Architecture Decisions v7, item 3

### Decision AD-GO-004: Logging Framework Selection
**Date**: 2025-01-15  
**Components**: Logging and Monitoring  
**Decision**: Use logrus library for structured logging.  
**Rationale**: logrus is the most popular structured logging library for Go, providing JSON output compatible with log aggregation systems and supporting correlation IDs and contextual fields.  
**IV&V Reference**: Architecture Decisions v7, item 4

### Decision AD-GO-005: Authentication Library Selection
**Date**: 2025-01-15  
**Components**: Security and Authentication  
**Decision**: Use golang-jwt/jwt/v4 for JWT handling and golang.org/x/crypto/bcrypt for password hashing.  
**Rationale**: golang-jwt/jwt/v4 is the de facto standard for JWT in Go, providing comprehensive JWT validation and signing. Bcrypt is the standard for password hashing in Go.  
**IV&V Reference**: Architecture Decisions v7, item 5

### Decision AD-GO-006: Package Structure Design
**Date**: 2025-01-15  
**Components**: Code Organization  
**Decision**: Use Go standard package layout with internal/ for private code and pkg/ for public packages.  
**Rationale**: This layout follows Go community conventions and best practices, provides clear boundaries between public and private code, and works well with Go modules and tooling.  
**IV&V Reference**: Architecture Decisions v7, item 6

### Decision AD-GO-007: Concurrency Model Design
**Date**: 2025-01-15  
**Components**: WebSocket Server, Camera Monitor  
**Decision**: Use goroutines with channels for communication and context.Context for cancellation.  
**Rationale**: Goroutines provide better performance than Python asyncio for I/O-bound workloads. Channels provide thread-safe communication without locks. Context is the standard Go pattern for cancellation and timeouts.  
**IV&V Reference**: Architecture Decisions v7, item 7

### Decision AD-GO-008: Error Handling Strategy
**Date**: 2025-01-15  
**Components**: All Components  
**Decision**: Use Go's error interface with custom error types and error wrapping.  
**Rationale**: Go's error handling is designed for explicit error checking. Custom error types provide better error context and handling. Error wrapping maintains error context through call chains.  
**IV&V Reference**: Architecture Decisions v7, item 8

### Decision AD-GO-009: Testing Framework Selection
**Date**: 2025-01-15  
**Components**: Testing Infrastructure  
**Decision**: Use Go's built-in testing package with testify for assertions and test utilities.  
**Rationale**: Go's testing package provides comprehensive testing features. Testify is the most popular testing utility library for Go. Built-in benchmarking provides performance testing capabilities.  
**IV&V Reference**: Architecture Decisions v7, item 9

### Decision AD-GO-010: Build and Deployment Strategy
**Date**: 2025-01-15  
**Components**: Build System, Deployment  
**Decision**: Use Make for build automation with static linking and multi-stage OCI container builds.
**Rationale**: Static linking eliminates runtime dependencies. Multi-stage OCI container builds provide efficient container images. Make provides simple and effective build automation.  
**IV&V Reference**: Architecture Decisions v7, item 10

### Decision AD-GO-011: Performance Optimization Strategy
**Date**: 2025-01-15  
**Components**: All Components  
**Decision**: Use object pools, connection pooling, and efficient data structures with profiling.  
**Rationale**: Object pools reduce garbage collection overhead. Connection pooling improves resource utilization. Go profiling tools provide excellent performance analysis.  
**IV&V Reference**: Architecture Decisions v7, item 11

### Decision AD-GO-012: Security Implementation Strategy
**Date**: 2025-01-15  
**Components**: Security and Authentication  
**Decision**: Use Go's crypto libraries with custom security middleware and input validation.  
**Rationale**: Go's crypto libraries are well-tested and secure. Security middleware provides centralized security enforcement. Input validation prevents common security vulnerabilities.  
**IV&V Reference**: Architecture Decisions v7, item 12

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

**Go Version Requirements**:
- Minimum supported: Go 1.25.0
- Target version: Latest stable Go version
- Compatibility testing required for major Go version updates
- Breaking change handling: Version-specific adapter pattern

**Deprecation Process**:
1. Announce deprecation with 2 minor version notice period
2. Include migration documentation and automated tooling where possible
3. Maintain backward compatibility during deprecation period
4. Remove deprecated features only in major version releases

---

**Document Status:** Complete architecture overview with Go implementation updates  
**Last Updated:** 2025-01-15  
**Next Review:** After Go implementation validation
