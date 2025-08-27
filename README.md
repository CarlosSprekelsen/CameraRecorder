# MediaMTX Camera Service

A distributed video sensor management service designed for OCI-compliant container environments. This service provides real-time video source discovery, streaming, recording, and management capabilities as part of a larger multi-sensor ecosystem with centralized service discovery.

## System Overview

The MediaMTX Camera Service is an always-on containerized service that manages both USB video devices and external RTSP feeds within a coordinated sensor ecosystem. It operates as a specialized video sensor container that registers with a central service discovery aggregator and provides standardized video services to client applications.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Service Discovery Aggregator                  │
│              (Control Plane - Identity, Config, Discovery)       │
└─────────────────┬───────────────────────┬───────────────────────┘
                  │                       │
         ┌────────▼──────────┐   ┌──────▼──────┐
         │   Client Apps     │   │   Platform  │
         │ (Android/iOS/Web) │   │  Management │
         └────────┬──────────┘   └─────────────┘
                  │
         ┌────────▼──────────┐
         │    Video Sensor   │
         │    Container      │ ◄──── Hub OS (VID:PID Routing)
         │  (This Service)   │
         └────────┬──────────┘
                  │
        ┌─────────┼─────────┐
        │                   │
    ┌───▼──┐          ┌────▼────┐
    │ USB  │          │ RTSP    │
    │Video │          │ Feeds   │
    │Device│          │         │
    └──────┘          └─────────┘
```

## System Integration Model

### Service Discovery Pattern
**Three-Tier Communication Model:**
1. **Container → Aggregator**: Service registration, health reporting, capability announcement
2. **Client → Aggregator**: Service discovery, endpoint resolution
3. **Client → Container**: Direct data consumption (streaming, recording, control)

### Device Management Philosophy
**Always-On Container Principle:**
- Containers automatically detect and claim assigned video sources
- Hub OS handles VID:PID-based device routing to appropriate containers
- Containers manage their own lifecycle and report health status
- Graceful handling of device connect/disconnect events

## Architecture Components

### Video Sensor Container (This Service)
**Purpose**: Unified video source management with abstraction layer
**Responsibilities**:
- USB camera discovery via Linux udev events
- RTSP feed management via configuration
- Multi-protocol streaming through MediaMTX integration
- Recording and snapshot capabilities
- Client-agnostic API (WebSocket JSON-RPC)
- Service registration with discovery aggregator

### Service Discovery Integration
**Registration**: Announces video service capabilities and endpoints to aggregator
**Health Reporting**: Continuous status updates for coordination and monitoring  
**Resource Advertising**: Reports available capacity and device inventory
**Identity Management**: Consumes platform security tokens and identity services

### Client Application Layer
**Discovery**: Queries aggregator for available video services
**Connection**: Establishes direct connections to video containers for data streams
**Platform Agnostic**: Supports Android, iOS, web, and other client types
**Offline Behavior**: Graceful degradation during connectivity issues

## Video Source Management

### USB Device Handling
- **Device Assignment**: Receives pre-configured VID:PID assignments from Hub OS
- **Hot-Plug Support**: Real-time device connection and disconnection handling
- **Capability Detection**: Automatic discovery of camera specifications and formats
- **Hardware Constraints**: Supports up to 8 concurrent video streams per hardware limitations

### RTSP Feed Integration  
- **Configuration-Driven**: Manages only pre-configured RTSP sources via YAML configuration
- **Network Sources**: Integrates external IP cameras and video feeds
- **Authentication Support**: Handles credentials and security for external feeds  
- **Unified Abstraction**: Provides consistent API regardless of video source type

### Media Processing Pipeline
```
Video Sources → FFmpeg Capture → MediaMTX Server → Multi-Protocol Streaming
     │                                                        │
     └── USB Cameras                                         ├── RTSP
     └── RTSP Feeds                                          ├── WebRTC  
                                                             └── HLS
```

## OCI Compliance and Container Architecture

### Standards Compliance
**Container Runtime**: Compatible with containerd, CRI-O, and other OCI-compliant runtimes  
**Orchestration**: Supports CNCF-compliant orchestration platforms  
**Image Format**: Adheres to OCI image specification  
**Resource Management**: Implements proper resource limits and health checks

### Container Characteristics
**Always-On Operation**: Designed for continuous operation with automatic recovery  
**Device Integration**: Requires USB device passthrough and udev event access  
**Network Architecture**: Multi-port service supporting control and data planes  
**Resource Efficiency**: Optimized for low power consumption and small footprint

### Deployment Flexibility
**Single Container**: Self-contained service with embedded MediaMTX integration  
**Multi-Instance**: Supports load balancing across multiple containers for different device types  
**Platform Agnostic**: Runs on any Linux-based container platform with USB support

## Service Interfaces

### Control Plane (JSON-RPC over WebSocket)
**Real-time Operations**: Camera control, recording management, status monitoring  
**Event Notifications**: Device connect/disconnect, recording status, system health  
**Configuration Management**: Dynamic source configuration and capability reporting

### Data Plane (Multi-Protocol Streaming)
**RTSP**: Standard IP camera protocol for professional integrations  
**WebRTC**: Low-latency browser streaming for web applications  
**HLS**: HTTP Live Streaming for mobile and diverse client support

### Management Plane (REST API)
**File Access**: Recording downloads, snapshot retrieval  
**System Status**: Health checks, capability queries, resource utilization  
**Configuration**: Service configuration and source management

## Operational Characteristics

### Performance Profile
**Discovery Latency**: Sub-200ms device detection and registration  
**Streaming Latency**: Sub-500ms end-to-end for WebRTC, under 2s for HLS  
**Concurrent Capacity**: Supports hardware-limited concurrent streams (up to 8)  
**Resource Efficiency**: Optimized for edge computing and resource-constrained environments

### Reliability Features
**Fault Tolerance**: Automatic recovery from device disconnections and network issues  
**Health Monitoring**: Continuous self-health assessment and reporting  
**Graceful Degradation**: Maintains service availability during partial failures  
**Clean Shutdown**: Proper deregistration and resource cleanup

## Multi-Sensor Ecosystem Integration

### Container Ecosystem Role
**Specialized Service**: Handles video sensor class within broader sensor management platform  
**Peer Services**: Coordinates with serial sensor containers and other specialized services  
**Resource Coordination**: Participates in platform-wide resource management and load balancing

### Service Discovery Ecosystem
**Service Registration**: Announces capabilities and endpoints to central aggregator  
**Client Discovery**: Enables platform-agnostic client applications to find video services  
**Health Coordination**: Participates in platform-wide health and monitoring systems

### Platform Integration
**Security Model**: Integrates with platform identity and authentication systems  
**Configuration Management**: Receives configuration through platform management interfaces  
**Observability**: Provides structured logging, metrics, and tracing for platform monitoring

## Deployment Scenarios

### Edge Computing
**Single Node**: Complete video management on edge devices with local USB cameras  
**Resource Constrained**: Optimized for low power consumption and minimal footprint  
**Offline Operation**: Maintains core functionality during network connectivity issues

### Distributed Systems
**Multi-Container**: Load balancing across multiple instances for high-capacity deployments  
**Service Mesh**: Integration with CNCF service mesh technologies for advanced networking  
**Hybrid Sources**: Simultaneous management of local USB devices and remote RTSP feeds

---

**Architecture Version**: 2.0  
**Service Classification**: Video Sensor Container  
**Platform Compliance**: OCI + CNCF Standards  
**Integration Model**: Service Discovery + Direct Data Plane