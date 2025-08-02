# Architecture Overview

## Status

**Architecture Decisions v6: APPROVED**  
All decisions align with industry standards, balance modernity with maintainability, and support concrete operational goals without overengineering.

---

## System Design

The MediaMTX Camera Service is a lightweight, robust wrapper around MediaMTX, providing:

1. **Real-time USB camera discovery and monitoring** (hybrid udev + polling)
2. **WebSocket JSON-RPC 2.0 API** with method-level versioning and structured deprecation
3. **Dynamic MediaMTX configuration management** (YAML + env overrides)
4. **Streaming, recording, and snapshot coordination**
5. **Resilient error recovery and health monitoring** (circuit breaker, exponential backoff)
6. **Secure access control and authentication**

## Component Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Client Applications                      │
│            (Web browsers, mobile apps, etc.)               │
└─────────────────────┬──────────────────────────────────────┘
                      │ WebSocket JSON-RPC 2.0
                      │ (Authentication: JWT/API Keys)
┌─────────────────────▼──────────────────────────────────────┐
│                Camera Service (Python)                     │
├─────────────────────────────────────────────────────────────┤
│            WebSocket JSON-RPC Server                      │
│     • Client connection management                         │
│     • JSON-RPC 2.0 protocol handling                      │
│     • Method-level API versioning & deprecation           │
│     • Real-time notifications                             │
│     • Authentication and authorization                     │
├─────────────────────────────────────────────────────────────┤
│             Camera Discovery Monitor                      │
│     • Hybrid udev events + polling fallback               │
│     • Configurable polling interval                       │
│     • v4l2 capability detection                           │
│     • Camera status tracking                              │
│         - Camera status object fields:                    │
│             • device: Camera device path                  │
│             • status: Connection status                   │
│             • name: Camera display name                   │
│             • resolution: Current resolution setting      │
│             • fps: Current frame rate                     │
│             • streams: Available stream URLs              │
│             • capabilities: Device capabilities           │
│             • metrics:                                    │
│                 - bytes_sent (int): Total bytes sent by camera stream.      │
│                 - readers (int): Active stream consumers.                   │
│                 - uptime (int): Seconds since last connect.                 │
│             <!-- This change aligns architecture with API contract and resolves roadmap.md IV&V blocker. -->
├─────────────────────────────────────────────────────────────┤
│              MediaMTX Controller                          │
│     • REST API client                                     │
│     • Dynamic stream management                           │
│     • Recording coordination                              │
│     • Health monitoring and recovery                      │
├─────────────────────────────────────────────────────────────┤
│               Health & Monitoring                         │
│     • Service health checks (REST, circuit breaker)       │
│     • Resource usage monitoring                           │
│     • Error tracking and recovery (exponential backoff)   │
│     • Configuration hot reload                            │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP REST API
                      │ (Health checks & recovery)
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

## Data Flow

### Camera Discovery Flow
1. **Monitor** detects USB camera connection via udev or polling (hybrid, configurable)
2. **Detector** probes camera capabilities using v4l2-ctl
3. **Controller** creates MediaMTX stream configuration (YAML, env overrides)
4. **Health Monitor** verifies MediaMTX accepts configuration
5. **Server** broadcasts camera status notification to authenticated clients
6. **Recovery Handler** manages connection failures and retries

### Streaming Flow  
1. **Client** requests stream via authenticated JSON-RPC call (method-level versioning)
2. **Authorization** validates client permissions for camera access
3. **Controller** configures MediaMTX path with FFmpeg source
4. **Health Monitor** verifies stream establishment
5. **MediaMTX** starts camera capture and encoding
6. **Client** accesses stream via secure URL (with token if required)

### Recording Flow
1. **Client** requests recording start via authenticated JSON-RPC
2. **Authorization** validates recording permissions
3. **Controller** enables recording in MediaMTX configuration
4. **Health Monitor** tracks recording status and disk usage
5. **MediaMTX** captures video to file with metadata
6. **Server** notifies client when recording completes or fails

## Error Recovery and Resilience

### MediaMTX Failure Handling
- **Health Checks**: Continuous REST API monitoring (every 5s)
- **Automatic Recovery**: Service restart with exponential backoff
- **Circuit Breaker**: Graceful degradation on repeated failures
- **State Preservation**: Camera configurations cached and restored
- **Client Notification**: Real-time failure/recovery status updates
- **Graceful Degradation**: Read-only mode when MediaMTX unavailable

### Camera Disconnect Handling  
- **Event Detection**: Immediate udev notification + polling fallback
- **Stream Cleanup**: Automatic MediaMTX path removal
- **Client Updates**: Real-time disconnect notifications
- **Reconnection Logic**: Automatic stream restoration on reconnect
- **Timeout Management**: Configurable detection and cleanup timeouts

### Network and System Failures
- **WebSocket Resilience**: Connection keepalive and automatic reconnection
- **Configuration Backup**: Versioned config with rollback capability
- **Resource Monitoring**: Disk space, memory, and CPU usage tracking
- **Rate Limiting**: Protection against API abuse and resource exhaustion

## Security Model

### API Authentication
- **JWT Tokens**: Stateless authentication with configurable expiry
- **API Keys**: Long-term service authentication for trusted clients  
- **mTLS Support**: Mutual certificate authentication for high-security deployments
- **Role-Based Access**: Different permission levels (viewer, operator, admin)

### Stream Access Control
- **Signed URLs**: Time-limited access tokens for stream endpoints
- **IP Restrictions**: Optional client IP address allowlisting
- **Camera Permissions**: Per-camera access control lists
- **Session Management**: Active session tracking and termination

### System Security
- **Process Isolation**: Non-root service user with minimal privileges
- **File Permissions**: Restricted access to recordings and configuration
- **Network Binding**: Configurable interface binding (localhost vs public)
- **Audit Logging**: Security event logging and monitoring

## Configuration Management

### Hot Reload Capability
- **Runtime Updates**: Configuration changes without service restart
- **Validation**: Schema validation before applying changes
- **Rollback**: Automatic revert on invalid configuration
- **Change Notification**: Real-time config update broadcasts

### Versioned Configuration
- **YAML Primary Configuration**: Human-readable, MediaMTX-consistent
- **Environment Variable Overrides**: For CI/CD and container deployments
- **Schema Validation**: On configuration load
- **Change Tracking**: Configuration history with timestamps
- **Backup Strategy**: Automatic configuration backups
- **Migration Support**: Smooth upgrades between config versions
- **Environment Overrides**: Development vs production config separation

## Health Monitoring and Observability

### Service Health Endpoints
- **Liveness Check**: `/health/alive` - Basic service responsiveness
- **Readiness Check**: `/health/ready` - Full system operational status  
- **MediaMTX Status**: `/health/mediamtx` - Backend service connectivity
- **Camera Status**: `/health/cameras` - Connected camera summary

### Metrics and Monitoring
- **Performance Metrics**: Response times, throughput, error rates
- **Resource Usage**: CPU, memory, disk space, network bandwidth
- **Camera Metrics**: Connection count, stream quality, failure rates
- **Integration Metrics**: MediaMTX API response times and error rates

### Logging and Diagnostics
- **Structured Logging**: JSON format with correlation IDs
- **Log Levels**: Configurable verbosity (ERROR, WARN, INFO, DEBUG)
- **Log Rotation**: Size and time-based rotation with compression
- **Diagnostic Tools**: Built-in troubleshooting and debug endpoints

## Extensibility and Future Enhancements

### Camera Source Extensions
- **Network Cameras**: RTSP/HTTP camera integration
- **Virtual Cameras**: Software-defined camera sources
- **Mobile Devices**: Smartphone camera integration via app
- **Plugin Architecture**: Loadable camera driver modules

### Protocol Extensions  
- **Additional Streaming**: SRT, NDI, or custom protocol support
- **Cloud Integration**: AWS Kinesis, Azure Media Services connectivity
- **Message Queues**: MQTT, RabbitMQ for enterprise integration
- **API Extensions**: GraphQL or additional REST endpoints

### Advanced Features
- **AI Integration**: Object detection, face recognition, analytics
- **Multi-Node Support**: Distributed camera service clusters
- **Cloud Storage**: Automatic recording upload to cloud providers
- **Advanced Recording**: Multi-camera synchronized recording

## Non-Functional Requirements

### Performance Targets
- **Camera Detection**: Sub-200ms USB connect/disconnect detection (hybrid method)
- **API Response**: <50ms for status queries, <100ms for control operations
- **Memory Usage**: <30MB base service footprint, <100MB with 10 cameras
- **CPU Usage**: <5% idle, <20% with active streaming and recording

### Scalability Limits
- **Concurrent Cameras**: Up to 16 USB cameras per service instance
- **WebSocket Connections**: Up to 100 concurrent client connections  
- **Recording Capacity**: Limited by available disk space and I/O bandwidth
- **Stream Bandwidth**: Limited by network capacity and MediaMTX performance

### Reliability Requirements
- **Uptime Target**: 99.9% availability excluding planned maintenance
- **Recovery Time**: <30 seconds for service restart, <10 seconds for camera reconnect
- **Data Integrity**: Zero-loss recording with atomic file operations
- **Error Rate**: <0.1% API call failure rate under normal conditions

## Technology Stack

- **Camera Service**: Python 3.10+, asyncio, websockets, aiohttp
- **Media Server**: MediaMTX (Go binary), target latest stable, minimum version pinned after compatibility confirmation
- **Camera Interface**: V4L2, FFmpeg 6.0+
- **Protocols**: WebSocket, JSON-RPC 2.0 (method-level versioning), REST, RTSP, WebRTC, HLS
- **Security**: JWT, TLS 1.3, optional mTLS
- **Deployment**: Systemd services, native Linux (Ubuntu 22.04+)
- **Monitoring**: Prometheus metrics, structured JSON logging

## Deployment Architecture

### Production Deployment
- **Service Isolation**: Dedicated system user with minimal privileges  
- **Process Management**: Systemd with automatic restart and resource limits
- **Network Configuration**: Configurable ports with firewall integration
- **File System**: Structured layout under `/opt/camera-service/`
- **Log Management**: Integration with system logging and rotation

### Development Environment
- **Native Linux**: Direct systemd service development and testing
- **Mock Services**: Camera and MediaMTX simulators for testing
- **Debug Tools**: Enhanced logging and diagnostic endpoints
- **Hot Reload**: Development-mode configuration and code reloading

---

## Architecture Decisions (Summary)

1. **MediaMTX Version Compatibility**: Target latest stable, pin minimum after compatibility confirmation, document upgrade/test process.
2. **Camera Discovery Method**: Hybrid udev + polling, configurable, with environment-based switching.
3. **Configuration Management**: YAML primary config, environment variable overrides, schema validation.
4. **Error Recovery Strategy**: Health monitoring, exponential backoff, circuit breaker, health event logging.
5. **API Versioning Strategy**: Method-level JSON-RPC versioning, structured deprecation, migration
6. **API Protocol Strategy**: WebSocket-only JSON-RPC with minimal REST health endpoints
7. **Authentication Strategy**: JWT for users, API keys for services, optional mTLS for high-security
8. **Logging Format**: Structured JSON with correlation IDs, pretty-print for development  
9. **Performance Targets**: <200ms camera detection, <50ms API calls, 16 cameras, 100 clients
10. **Resource Limits**: <100MB RAM, <20% CPU, 500MB logs with rotation, disk usage warnings