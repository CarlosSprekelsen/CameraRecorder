# MediaMTX Camera Service Architecture

**Version:** 2.0  
**Date:** 2025-01-15  
**Status:** Language-Agnostic Architecture  
**Related Epic/Story:** Architecture Standardization  

## Table of Contents

1. [System Overview](#system-overview)
2. [Architectural Principles](#architectural-principles)
3. [Component Architecture](#component-architecture)
4. [Security Architecture](#security-architecture)
5. [Logging Architecture](#logging-architecture)
6. [Testing Architecture](#testing-architecture)
7. [Performance Targets](#performance-targets)
8. [API Contract](#api-contract)

---

## System Overview

The MediaMTX Camera Service is a high-performance camera management system providing:

1. **Real-time USB camera discovery and monitoring**
2. **WebSocket JSON-RPC 2.0 API** (1000+ concurrent connections)
3. **Dynamic MediaMTX configuration management** (100ms response time)
4. **Streaming, recording, and snapshot coordination**
5. **External stream discovery and management** (STANAG 4609 UAV streams and network-based sources)
6. **Resilient error recovery and health monitoring**
7. **Secure access control and authentication**

### System Goals
- **Performance**: High-performance camera service with real-time capabilities
- **Resource Usage**: Efficient memory and CPU utilization, power efficient
- **Compatibility**: Standards-compliant API with broad client support
- **Risk Management**: Working software first, integration incrementally

### Success Criteria
- Camera detection <200ms latency
- WebSocket server handles 1000+ concurrent connections
- Memory usage <60MB base, <200MB with 10 cameras
- 1000+ concurrent WebSocket connections supported
- **Working Service**: Fully functional camera service
- **Basic Integration**: Added incrementally when platform systems exist (i.e. UAVs)

---

## Architectural Principles

### 1. Single Source of Truth Architecture

The system implements a **single source of truth** pattern where MediaMTX Controller is the complete business logic layer, with WebSocket server being a thin protocol layer that delegates all operations.

#### Architectural Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                             │
│  • Works with camera identifiers (camera0, camera1)         │
│  • No knowledge of internal device paths                    │
│  • Clean, abstract API interface                           │
│  • Hardware-independent client code                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                WEBSOCKET PROTOCOL LAYER                     │
│  • JSON-RPC 2.0 protocol handling                          │
│  • Authentication and authorization                         │
│  • Request/response formatting                             │
│  • NO business logic - delegates to MediaMTX               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                MEDIAMTX CONTROLLER                          │
│              (SINGLE SOURCE OF TRUTH)                      │
│  • API abstraction layer (camera0 ↔ /dev/video0)          │
│  • All camera operations                                   │
│  • All business logic                                      │
│  • Orchestrates all sub-components                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  HARDWARE LAYER                             │
│  • Camera discovery monitor                                │
│  • MediaMTX path manager                                   │
│  • Device path operations (/dev/video0, /dev/video1)      │
│  • Hardware-specific operations                            │
└─────────────────────────────────────────────────────────────┘
```

### 2. Strict Separation of Concerns

The architecture enforces strict separation of concerns with clear boundaries and responsibilities.

#### WebSocket Server Dependencies
- **ONLY** depends on MediaMTX Controller for business logic
- **NO** direct camera monitor access
- **NO** direct file system access
- **NO** direct MediaMTX server access
- **ONLY** protocol handling and security

#### MediaMTX Controller Dependencies
- **Orchestrates** all business logic components
- **Single source of truth** for all camera operations
- **Manages** all sub-components (Camera Monitor, Path Manager, etc.)
- **Provides** complete API abstraction layer

### 3. Architecture Enforcement Rules

#### **RULE 1: WebSocket Server MUST Be Thin**
- **FORBIDDEN**: Direct camera monitor access
- **FORBIDDEN**: Business logic in WebSocket methods
- **FORBIDDEN**: Camera validation in WebSocket layer
- **FORBIDDEN**: File operations in WebSocket layer
- **REQUIRED**: All operations delegate to MediaMTX Controller

#### **RULE 2: MediaMTX Controller MUST Be Complete**
- **REQUIRED**: Camera monitor integration
- **REQUIRED**: All camera discovery methods
- **REQUIRED**: All business logic operations
- **REQUIRED**: Single source of truth for camera operations
- **REQUIRED**: Proper abstraction layer (camera0 ↔ /dev/video0)

#### **RULE 3: No Direct Hardware Access**
- **FORBIDDEN**: WebSocket server accessing camera monitor directly
- **FORBIDDEN**: WebSocket server accessing file system directly
- **FORBIDDEN**: WebSocket server accessing MediaMTX server directly
- **REQUIRED**: All hardware access through MediaMTX Controller

#### **RULE 4: Clear Dependency Chain**
```
WebSocket Server → MediaMTX Controller → Camera Monitor
                → MediaMTX Controller → Path Manager
                → MediaMTX Controller → File Manager
```

#### **RULE 5: No Duplicated Logic**
- **FORBIDDEN**: Abstraction layer in both WebSocket and MediaMTX
- **FORBIDDEN**: Camera validation in multiple places
- **FORBIDDEN**: Stream URL generation in multiple places
- **REQUIRED**: Single implementation in MediaMTX Controller

### 4. Event-Driven Architecture

The event system is managed by MediaMTX Controller, not WebSocket server.

#### Event System Components
- **EventManager**: Central hub for event distribution
- **Topic-Based Filtering**: Events sent only to interested clients
- **Component Adapters**: Bridge between components and event system
- **Performance**: O(log n) scaling with client count

#### Event Topics
- **Camera Events**: Connected, disconnected, status changes
- **Recording Events**: Start, stop, status updates
- **System Events**: Health, startup, configuration changes

### 5. Stream Lifecycle Management

#### Stream Lifecycle Types
- **Recording Streams**: Long-duration video recording with file rotation
- **Viewing Streams**: Live stream viewing with auto-close after inactivity
- **Snapshot Streams**: Quick photo capture with immediate activation/deactivation

#### On-Demand Stream Activation
- **Power Efficiency**: FFmpeg processes start only when needed
- **Configuration-Driven**: MediaMTX settings control lifecycle behavior
- **Automatic Management**: Streams activate/deactivate based on demand

---

## Security Architecture

**CRITICAL**: Security is the foundation of the entire system. All components must implement proper authentication, authorization, and security controls.

### Security-First Design Principles

1. **Zero Trust Architecture**: Every request must be authenticated and authorized
2. **Defense in Depth**: Multiple security layers at every component
3. **Least Privilege**: Users get minimum required permissions
4. **Audit Everything**: All security events must be logged
5. **Secure by Default**: All components start in secure state

### Security Middleware Design

The security layer is designed to integrate seamlessly with existing systems rather than creating parallel infrastructure:

```
┌────────────────────────────────────────────────────────────┐
│                    Security Layer                          │
├─────────────────────────────────────────────────────────────┤
│            Authentication Middleware                       │
│     • JWT token validation                                │
│     • Session management                                  │
│     • Uses existing SecurityConfig                        │
├─────────────────────────────────────────────────────────────┤
│            RBAC Middleware                                │
│     • Role-based access control                           │
│     • Permission matrix enforcement                       │
│     • Integrates with existing PermissionChecker          │
├─────────────────────────────────────────────────────────────┤
│            Rate Limiting                                  │
│     • Per-method rate limits                              │
│     • DDoS protection                                     │
│     • Uses existing SecurityConfig values                 │
├─────────────────────────────────────────────────────────────┤
│            Input Validation                               │
│     • Parameter sanitization                              │
│     • Type safety enforcement                             │
│     • Centralized validation logic                        │
├─────────────────────────────────────────────────────────────┤
│            Audit Logging                                  │
│     • Security event tracking                             │
│     • Uses existing LoggingConfig                         │
│     • File rotation and retention                         │
└─────────────────────┬──────────────────────────────────────┘
                      │ Configuration Integration
┌─────────────────────▼──────────────────────────────────────┐
│            Existing Configuration System                   │
│     • SecurityConfig for rate limits and JWT settings     │
│     • LoggingConfig for audit log configuration           │
│     • No hard-coded values or parallel infrastructure    │
└─────────────────────────────────────────────────────────────┘
```

### Security Integration Principles

1. **Leverage Existing Systems**: Use `SecurityConfig`, `LoggingConfig`, and existing logger
2. **No Hard-coded Values**: All security parameters come from configuration
3. **Transparent Integration**: Security middleware works seamlessly with existing code
4. **Configuration Adapter Pattern**: Bridge between security middleware and existing config
5. **Audit Trail**: Comprehensive logging of all security events

### Security Middleware Components

- **AuthMiddleware**: Centralized authentication enforcement
- **RBACMiddleware**: Role-based access control with existing permission matrix
- **EnhancedRateLimiter**: Rate limiting using existing configuration values
- **InputValidator**: Centralized input validation and sanitization
- **SecurityAuditLogger**: Comprehensive security event logging
- **ConfigAdapter**: Bridge between security middleware and existing configuration

### Role-Based Access Control (RBAC)

#### Role Definitions

- **viewer**: Read-only access to camera status, file listings, and basic information
- **operator**: Viewer permissions + camera control operations (snapshots, recording)
- **admin**: Full access to all features including system metrics and configuration

#### Permission Matrix

| Method | viewer | operator | admin |
|--------|--------|----------|-------|
| ping | ✅ | ✅ | ✅ |
| authenticate | ✅ | ✅ | ✅ |
| get_camera_list | ✅ | ✅ | ✅ |
| get_camera_status | ✅ | ✅ | ✅ |
| get_camera_capabilities | ✅ | ✅ | ✅ |
| take_snapshot | ❌ | ✅ | ✅ |
| start_recording | ❌ | ✅ | ✅ |
| stop_recording | ❌ | ✅ | ✅ |
| list_recordings | ✅ | ✅ | ✅ |
| list_snapshots | ✅ | ✅ | ✅ |
| delete_recording | ❌ | ✅ | ✅ |
| delete_snapshot | ❌ | ✅ | ✅ |
| get_metrics | ❌ | ❌ | ✅ |
| get_storage_info | ❌ | ❌ | ✅ |
| set_retention_policy | ❌ | ❌ | ✅ |
| cleanup_old_files | ❌ | ❌ | ✅ |

### Security Enforcement Points

#### WebSocket Server Security
- **Authentication**: Every WebSocket connection must authenticate
- **Authorization**: Every method call must be authorized
- **Rate Limiting**: Per-client rate limits to prevent abuse
- **Input Validation**: All parameters must be validated and sanitized

#### MediaMTX Controller Security
- **Device Access Control**: Only authorized devices can be accessed
- **File System Security**: Secure file operations with proper permissions
- **Process Security**: FFmpeg processes run with minimal privileges
- **Network Security**: Secure communication with MediaMTX server

### Security Compliance Checklist

Before implementing any changes, verify security compliance:

#### ✅ Authentication Compliance
- [ ] All WebSocket connections require authentication
- [ ] JWT tokens are properly validated
- [ ] Session management is secure
- [ ] Authentication failures are logged

#### ✅ Authorization Compliance
- [ ] RBAC is enforced for all methods
- [ ] Permission matrix is properly implemented
- [ ] Role escalation is prevented
- [ ] Authorization failures are logged

#### ✅ Input Validation Compliance
- [ ] All parameters are validated
- [ ] SQL injection prevention
- [ ] Path traversal prevention
- [ ] XSS prevention

#### ✅ Audit Logging Compliance
- [ ] All security events are logged
- [ ] Logs include user context
- [ ] Logs are tamper-proof
- [ ] Log retention policies are enforced

---

## Logging Architecture

### Logger Factory Pattern

The system implements a **Logger Factory Pattern** to ensure consistent logging configuration across all components.

#### Architecture Principles

1. **Centralized Logger Creation**: Single LoggerFactory responsible for creating all logger instances
2. **Configuration-Driven**: Factory respects global configuration settings
3. **No Direct Logger Creation**: Modules request loggers from factory, don't create directly
4. **Test Isolation**: Test configuration affects factory behavior
5. **Language Agnostic**: Pattern works in any programming language

#### Logger Factory Design

```
┌─────────────────────────────────────────────────────────────┐
│                    LOGGER FACTORY                           │
│  • Centralized logger creation                             │
│  • Respects global configuration                           │
│  • Provides configured logger instances                    │
│  • Supports test configuration override                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                GLOBAL CONFIGURATION                        │
│  • Log level settings                                      │
│  • Output format configuration                             │
│  • File/console output settings                            │
│  • Test-specific overrides                                 │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    MODULE LOGGERS                          │
│  • WebSocket Server Logger                                 │
│  • MediaMTX Controller Logger                              │
│  • Camera Monitor Logger                                   │
│  • Security Middleware Logger                              │
│  • All components use factory-created loggers              │
└─────────────────────────────────────────────────────────────┘
```

#### Logger Factory Interface

**Core Responsibilities:**
- **Create Logger**: `GetLogger(componentName)` - Returns configured logger instance
- **Configuration Management**: Respects global logging configuration
- **Test Support**: Allows test-specific configuration overrides
- **Consistency**: Ensures all loggers use same configuration

**Configuration Integration:**
- **Global Settings**: Log level, format, output destinations
- **Component Identification**: Each logger tagged with component name
- **Test Overrides**: Test fixtures can override configuration
- **Runtime Changes**: Configuration changes affect all new loggers

#### Benefits

1. **Configuration Consistency**: All loggers use same configuration
2. **Test Isolation**: Test configuration affects all loggers
3. **Maintainability**: Single point of logging configuration
4. **No Global State**: No singleton pattern complexity
5. **Easy Testing**: Simple to inject test loggers
6. **Language Agnostic**: Works in any programming language

#### Implementation Requirements

**Logger Factory:**
- Must implement `GetLogger(componentName)` method
- Must respect global configuration settings
- Must support test configuration overrides
- Must provide consistent logger instances

**Module Integration:**
- Modules must use factory to get logger instances
- Modules must not create loggers directly
- Modules must pass component name to factory
- Modules must not store logger configuration

**Test Integration:**
- Test fixtures must configure factory before use
- Test fixtures must use factory-created loggers
- Test configuration must override global settings
- Test isolation must be maintained

#### Security Integration

**Audit Logging:**
- All security events logged through factory-created loggers
- Consistent log format across all components
- User context included in security logs
- Log retention policies enforced globally

**Compliance:**
- All security events are logged
- Logs include user context and correlation IDs
- Logs are tamper-proof and auditable
- Log retention policies are enforced


## Component Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Client Applications                      │
│            (Web browsers, mobile apps, etc.)               │
│  • Use camera identifiers (camera0, camera1)               │
│  • Hardware-independent interface                          │
└─────────────────────┬──────────────────────────────────────┘
                      │ WebSocket JSON-RPC 2.0
┌─────────────────────▼──────────────────────────────────────┐
│            WebSocket JSON-RPC Server                       │
│                 (THIN PROTOCOL LAYER)                     │
│     • Client connection management (1000+ concurrent)     │
│     • JSON-RPC 2.0 protocol handling                      │
│     • Real-time notifications (<20ms latency)             │
│     • Authentication and authorization                     │
│     • Security middleware with RBAC enforcement            │
│     • Rate limiting and DDoS protection                   │
│     • NO business logic - delegates to MediaMTX           │
└─────────────────────┬──────────────────────────────────────┘
                      │ Delegates to
┌─────────────────────▼──────────────────────────────────────┐
│                MediaMTX Controller                         │
│              (COMPLETE BUSINESS LOGIC LAYER)              │
│     • Camera discovery integration                         │
│     • API abstraction layer (camera0 ↔ /dev/video0)       │
│     • All camera operations (recording, snapshots, etc.)  │
│     • Stream management and lifecycle                     │
│     • File operations and storage management              │
│     • Event management and notifications                  │
│     • Single source of truth for all camera operations   │
├─────────────────────────────────────────────────────────────┤
│             Camera Discovery Monitor                       │
│     • USB camera detection (<200ms)                       │
│     • Camera status tracking                              │
│     • Hot-plug event handling                             │
│     • Concurrent monitoring                               │
│     • Internal device path management (/dev/video*)       │
├─────────────────────────────────────────────────────────────┤
│         External Stream Discovery                          │
│     • UAV stream discovery                                │
│     • Network range scanning                              │
│     • RTSP stream validation and health monitoring        │
│     • On-demand and periodic discovery modes              │
│     • STANAG 4609 compliance for military UAVs            │
├─────────────────────────────────────────────────────────────┤
│            MediaMTX Path Manager                           │
│     • Dynamic path creation via REST API                  │
│     • FFmpeg command generation                           │
│     • Path lifecycle management                           │
│     • Error handling and recovery                         │
│     • Internal device path operations                     │
├─────────────────────────────────────────────────────────────┤
│               Health & Monitoring                          │
│     • Service health checks                               │
│     • Resource usage monitoring                           │
│     • Error tracking and recovery                         │
│     • Configuration management                            │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP REST API
┌─────────────────────▼───────────────────────────────────────┐
│                   MediaMTX Server                          │
├─────────────────────────────────────────────────────────────┤
│                Media Processing                           │
│     • RTSP/WebRTC/HLS streaming                           │
│     • FFmpeg process management                           │
│     • Multi-protocol support                              │
│     • Recording and snapshot generation                   │
│     • Internal device path operations                     │
└─────────────────────┬───────────────────────────────────────┘
                      │ FFmpeg Processes
┌─────────────────────▼───────────────────────────────────────┐
│                 USB Cameras                                 │
│         /dev/video0, /dev/video1, etc.                     │
│  • Hardware layer (internal only)                          │
│  • Not exposed to clients                                  │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### WebSocket JSON-RPC Server - THIN PROTOCOL LAYER
- **SECURITY FIRST**: Authentication, authorization, and input validation for every request
- **ONLY** client connection management and authentication (1000+ concurrent)
- **ONLY** JSON-RPC 2.0 protocol implementation
- **ONLY** real-time event notifications (<20ms latency)
- **ONLY** session management and authorization
- **ONLY** rate limiting and DDoS protection
- **NO business logic** - all operations delegate to MediaMTX Controller
- **NO camera operations** - MediaMTX Controller handles all camera logic
- **NO file operations** - MediaMTX Controller handles all file logic
- **NO stream operations** - MediaMTX Controller handles all stream logic

#### MediaMTX Controller - COMPLETE BUSINESS LOGIC LAYER
- **Single source of truth** for all camera operations
- Camera discovery integration and management
- API abstraction layer (camera0 ↔ /dev/video0 mapping)
- All camera operations (recording, snapshots, streaming)
- Stream management and lifecycle coordination
- File operations and storage management
- Event management and notifications
- Orchestrates all sub-components (Camera Monitor, Path Manager, etc.)

#### Camera Discovery Monitor - HARDWARE ABSTRACTION LAYER
- USB camera detection via V4L2 (<200ms)
- Hot-plug event handling
- Concurrent monitoring of multiple cameras
- Camera capability probing and status tracking
- **Internal to MediaMTX Controller** - not directly accessible

#### External Stream Discovery - NETWORK ABSTRACTION LAYER
- UAV stream discovery with configurable parameters
- Network range scanning and IP enumeration
- RTSP stream validation and health monitoring
- On-demand and periodic discovery modes
- STANAG 4609 compliance for military UAVs
- **Internal to MediaMTX Controller** - not directly accessible

#### MediaMTX Path Manager - STREAM INFRASTRUCTURE LAYER
- Dynamic path creation via MediaMTX REST API
- FFmpeg command generation and management
- Path lifecycle management (create, update, delete)
- Error handling and automatic recovery
- **Internal to MediaMTX Controller** - not directly accessible

#### Health & Monitoring - OBSERVABILITY LAYER
- Structured logging with correlation IDs
- Service health monitoring and reporting
- Resource usage tracking and alerts
- Configuration management with hot-reload

---

## Testing Architecture

### Single Systemd-Managed MediaMTX Instance

**Decision**: All tests MUST use the single systemd-managed MediaMTX service instance.

#### Service Configuration
- MediaMTX service managed by systemd
- Fixed ports: API (9997), RTSP (8554), WebRTC (8889), HLS (8888)
- Tests verify service availability before execution
- No test-specific MediaMTX instances

#### Test Integration Requirements
- Tests must check MediaMTX service status before execution
- Tests must wait for MediaMTX API readiness
- Tests must use real MediaMTX service, not mocks
- Tests must clean up after execution

#### Port Configuration
- **API Port**: 9997 (fixed systemd service port)
- **RTSP Port**: 8554 (fixed systemd service port)
- **WebRTC Port**: 8889 (fixed systemd service port)
- **HLS Port**: 8888 (fixed systemd service port)

---

## Performance Targets

### Response Time Targets
- **Camera Detection**: <200ms latency
- **WebSocket Response**: <50ms for JSON-RPC methods
- **Stream Activation**: <3s for on-demand activation
- **Snapshot Capture**: <0.5s (Tier 1), <3s (Tier 2), <5s (Tier 3)
- **External Stream Discovery**: <30s for network scan completion

### Concurrency Targets
- **WebSocket Connections**: 1000+ concurrent connections
- **Camera Monitoring**: 10+ cameras with concurrent monitoring
- **FFmpeg Processes**: 10+ concurrent FFmpeg processes
- **External Stream Discovery**: 5+ concurrent network scans

### Resource Usage Targets
- **Memory Usage**: <60MB base, <200MB with 10 cameras
- **CPU Usage**: <20% idle, <80% under load
- **Network**: <100Mbps per camera stream

---

## API Contract

The MediaMTX Camera Service implements a comprehensive JSON-RPC 2.0 API over WebSocket connections. **All API methods are implemented by MediaMTX Controller, with WebSocket server providing only protocol handling.**

### API Documentation Reference

**Complete API Specification**: See `docs/api/json_rpc_methods.md` for the complete, accurate, and up-to-date API documentation including:
- All available methods with detailed parameters and responses
- Authentication and authorization requirements
- Error codes and response formats
- Real-time notifications and event subscriptions
- External stream discovery methods
- File management operations
- System management and monitoring

### Architecture Compliance

- **WebSocket Server**: Thin protocol layer, delegates all operations to MediaMTX Controller
- **MediaMTX Controller**: Implements all business logic and camera operations
- **No Direct Access**: WebSocket server cannot access camera monitor, file system, or MediaMTX server directly
- **API Ground Truth**: API documentation (`docs/api/json_rpc_methods.md`) is the authoritative source for all API specifications

---

## Architecture Compliance Checklist

Before implementing any changes, verify compliance with these architecture rules:

### ✅ Security Compliance (CRITICAL)
- [ ] All WebSocket connections require authentication
- [ ] RBAC is enforced for all methods
- [ ] Input validation is implemented for all parameters
- [ ] Rate limiting is active for all clients
- [ ] Security events are logged and audited
- [ ] No hard-coded security values
- [ ] Security configuration is externalized

### ✅ WebSocket Server Compliance
- [ ] WebSocket server has ONLY MediaMTX Controller as business logic dependency
- [ ] NO direct camera monitor access
- [ ] NO business logic in WebSocket methods
- [ ] NO camera validation in WebSocket layer
- [ ] NO file operations in WebSocket layer
- [ ] ALL operations delegate to MediaMTX Controller
- [ ] Security middleware is properly integrated

### ✅ MediaMTX Controller Compliance
- [ ] MediaMTX Controller has camera monitor integration
- [ ] MediaMTX Controller implements all camera discovery methods
- [ ] MediaMTX Controller implements all business logic operations
- [ ] MediaMTX Controller is single source of truth for camera operations
- [ ] MediaMTX Controller has proper abstraction layer (camera0 ↔ /dev/video0)
- [ ] MediaMTX Controller validates all camera operations
- [ ] MediaMTX Controller enforces device access controls

### ✅ No Direct Hardware Access
- [ ] WebSocket server does NOT access camera monitor directly
- [ ] WebSocket server does NOT access file system directly
- [ ] WebSocket server does NOT access MediaMTX server directly
- [ ] ALL hardware access goes through MediaMTX Controller

### ✅ No Duplicated Logic
- [ ] Abstraction layer exists ONLY in MediaMTX Controller
- [ ] Camera validation exists ONLY in MediaMTX Controller
- [ ] Stream URL generation exists ONLY in MediaMTX Controller
- [ ] NO duplicated business logic between layers
- [ ] Security logic is centralized and not duplicated

### ✅ Logging Architecture Compliance
- [ ] All modules use LoggerFactory to create logger instances
- [ ] NO direct logger creation in modules
- [ ] Test configuration affects all loggers
- [ ] Logger configuration is centralized
- [ ] Security events are logged through factory-created loggers

---

**Document Status**: Architecture guide
**Next Review**: Before any architectural changes
