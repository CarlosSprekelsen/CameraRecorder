# Client Application Requirements Document

**Version:** 1.1  
**Authors:** System Architect  
**Date:** 2025-08-04  
**Status:** Approved  
**Related Epic:** Client Applications Development

---

## Purpose

This document specifies the functional and non-functional requirements for the client applications that will interface with the MediaMTX Camera Service. The client applications consist of a Web interface accessible via web browser and an Android APK application, both designed for TWT and TT camera operations.

## Scope

The client applications will provide camera control functionality by communicating with the MediaMTX Camera Service via the existing WebSocket JSON-RPC 2.0 API. This document covers both platforms while ensuring consistency in functionality and user experience.

---

## Architecture Integration

### Service Integration
- **Communication Protocol:** WebSocket JSON-RPC 2.0 as defined in `docs/api/json-rpc-methods.md`
- **Connection Endpoint:** `ws://[service-host]:8002/ws`
- **Authentication:** JWT token-based authentication as per service security model
- **Real-time Notifications:** Subscribe to camera status updates and recording events

### Supported Service Methods
- `get_camera_list` - Enumerate available cameras
- `get_camera_status` - Get specific camera status
- `take_snapshot` - Capture still images
- `start_recording` - Begin video recording
- `stop_recording` - End video recording
- Real-time notifications for camera and recording status updates

---

## Functional Requirements

### F1: Camera Interface Requirements

#### F1.1: Photo Capture
- **F1.1.1:** The application SHALL allow users to take photos using available cameras
- **F1.1.2:** The application SHALL use the service's `take_snapshot` JSON-RPC method
- **F1.1.3:** The application SHALL display a preview of captured photos
- **F1.1.4:** The application SHALL handle photo capture errors gracefully with user feedback

#### F1.2: Video Recording
- **F1.2.1:** The application SHALL allow users to record videos using available cameras
- **F1.2.2:** The application SHALL support unlimited duration recording mode
  - API Contract: JSON-RPC `start_recording` without a `duration` parameter SHALL start unlimited recording
  - Service Behavior: Service SHALL maintain session until explicit stop_recording call
- **F1.2.3:** The application SHALL support timed recording with user-specified duration in seconds, minutes, or hours
  - API Contract: JSON-RPC `start_recording` accepts one of the following mutually exclusive parameter sets:
    - `{ device: string, duration_seconds: integer (1-3600) }`
    - `{ device: string, duration_minutes: integer (1-1440) }`
    - `{ device: string, duration_hours: integer (1-24) }`
  - Service Behavior: Service SHALL automatically stop the recording once the specified duration elapses and SHOULD emit a completion notification.
- **F1.2.4:** The application SHALL allow users to manually stop video recording
- **F1.2.5:** The application SHALL handle recording session management via service API

#### F1.3: Recording Management
- **F1.3.1:** The application SHALL automatically create new video files when maximum file size is reached (handled by service)
- **F1.3.2:** The application SHALL display recording status and elapsed time in real-time
- **F1.3.3:** The application SHALL notify users when video recording is completed
- **F1.3.4:** The application SHALL provide visual indicators for active recording state

### F2: File Management Requirements

#### F2.1: Metadata Management
- **F2.1.1:** The application SHALL ensure photos and videos include location metadata (when available)
- **F2.1.2:** The application SHALL ensure photos and videos include timestamp metadata
- **F2.1.3:** The application SHALL request device location permissions appropriately

#### F2.2: File Naming Convention
- **F2.2.1:** The application SHALL use default naming format: `[datetime]_[unique_id].[extension]`
- **F2.2.2:** DateTime format SHALL be: `YYYY-MM-DD_HH-MM-SS`
- **F2.2.3:** Unique ID SHALL be a 6-character alphanumeric string
- **F2.2.4:** Examples: `2025-08-04_14-30-15_ABC123.jpg`, `2025-08-04_14-30-15_XYZ789.mp4`

#### F2.3: Storage Configuration
- **F2.3.1:** The application SHALL store media files in a user-configurable default folder
- **F2.3.2:** The application SHALL provide folder selection interface
- **F2.3.3:** The application SHALL validate storage permissions and available space
- **F2.3.4:** Default storage location SHALL be platform-appropriate:
  - Android: `/storage/emulated/0/DCIM/TWT_Camera/`
  - Web: Browser downloads folder with filename prompt

### F3: User Interface Requirements

#### F3.1: Camera Selection
- **F3.1.1:** The application SHALL display list of available cameras from service API
- **F3.1.2:** The application SHALL show camera status (connected/disconnected)
- **F3.1.3:** The application SHALL handle camera hot-plug events via real-time notifications
- **F3.1.4:** The application SHALL provide camera switching interface

#### F3.2: Recording Controls and Security Enforcement
- **F3.2.1:** The application SHALL provide intuitive recording start/stop controls
- **F3.2.2:** The application SHALL display recording duration selector interface
- **F3.2.3:** The application SHALL show recording progress and elapsed time
- **F3.2.4:** The application SHALL provide emergency stop functionality
- **F3.2.5:** Operator permissions SHALL be required to invoke `start_recording`, `stop_recording`, and `take_snapshot`
  - API Contract: Protected JSON-RPC methods SHALL require a valid JWT with role=operator.
  - Token Transport: The JWT SHALL be provided via JSON-RPC `authenticate` method prior to using protected methods.
    - `authenticate` request: `{ jsonrpc: "2.0", method: "authenticate", params: { token: string } }`
    - On success, the server SHALL associate the client connection with the authenticated user and role for the session.
  - Error Handling: Missing, invalid, or expired tokens SHALL result in JSON-RPC error with code -32003 (authorization) and a meaningful message.
- **F3.2.6:** The application SHALL handle token expiration by re-authenticating before retrying protected operations.

#### F3.3: Settings Management
- **F3.3.1:** The application SHALL provide settings interface for:
  - Server connection configuration
  - Default storage location
  - Recording quality preferences
  - Notification preferences
- **F3.3.2:** The application SHALL validate and persist user settings
- **F3.3.3:** The application SHALL provide settings reset to defaults

---

## Platform-Specific Requirements

### Web Application (PWA)

#### W1: Web Platform Features
- **W1.1:** Browser compatibility with Chrome 90+, Firefox 88+, Safari 14+
- **W1.2:** Responsive design for desktop and mobile browsers
- **W1.3:** Progressive Web App capabilities for mobile installation
- **W1.4:** WebRTC integration for camera preview when supported

#### W2: Web File Handling
- **W2.1:** Integration with browser download mechanism
- **W2.2:** File naming preservation in downloads
- **W2.3:** Large file download handling with progress indication

### Android Application

#### A1: Android Platform Features
- **A1.1:** Target Android API level 28 (Android 9.0) minimum
- **A1.2:** Target Android API level 34 (Android 14) for compilation
- **A1.3:** Camera permissions management (CAMERA, RECORD_AUDIO)
- **A1.4:** Storage permissions management (WRITE_EXTERNAL_STORAGE, READ_EXTERNAL_STORAGE)
- **A1.5:** Location permissions management (ACCESS_FINE_LOCATION, ACCESS_COARSE_LOCATION)

#### A2: Android Integration
- **A2.1:** Integration with Android MediaStore for media file registration
- **A2.2:** Background recording capabilities with foreground service
- **A2.3:** Android notification system integration for recording status
- **A2.4:** Battery optimization exclusion guidance for users

---

## Non-Functional Requirements

### N1: Performance Requirements
- **N1.1:** Application startup time SHALL be under 3 seconds (includes service connection <1s)
- **N1.2:** Camera list refresh SHALL complete within 1 second (service API <50ms + UI rendering)
- **N1.3:** Photo capture response SHALL be under 2 seconds (service processing <100ms + file transfer)
- **N1.4:** Video recording start SHALL begin within 2 seconds (service API <100ms + MediaMTX setup)
- **N1.5:** UI interactions SHALL provide immediate feedback (200ms, excludes service calls)

### N2: Reliability Requirements
- **N2.1:** Application SHALL handle service disconnections gracefully
- **N2.2:** Application SHALL implement automatic reconnection with exponential backoff
- **N2.3:** Application SHALL preserve recording state across temporary disconnections
- **N2.4:** Application SHALL validate all user inputs and service responses

### N3: Security Requirements
- **N3.1:** Application SHALL implement secure WebSocket connections (WSS) in production
- **N3.2:** Application SHALL validate JWT tokens and handle expiration
- **N3.3:** Application SHALL not store sensitive credentials in plain text
- **N3.4:** Application SHALL implement timeout for inactive sessions

---

## Revision History

- 1.1 (2025-08-09): Clarified F1.2.2 unlimited duration API contract; specified F1.2.3 time unit semantics for timed recording; added F3.2.5/F3.2.6 security enforcement and authentication flow for protected methods.

### N4: Usability Requirements
- **N4.1:** Application SHALL provide clear error messages and recovery guidance
- **N4.2:** Application SHALL implement consistent UI patterns across platforms
- **N4.3:** Application SHALL provide accessibility support (screen readers, keyboard navigation)
- **N4.4:** Application SHALL support offline mode with limited functionality

---

## Technical Specifications

### T1: Communication Protocol
- **Protocol:** WebSocket JSON-RPC 2.0
- **Message Format:** JSON with correlation ID support
- **Error Handling:** Standard JSON-RPC error codes plus service-specific codes
- **Heartbeat:** Ping every 30 seconds to maintain connection

### T2: Data Flow Architecture

```
┌─────────────────┐    WebSocket     ┌─────────────────┐
│ Client App      │ ◄──JSON-RPC────► │ Camera Service  │
│ (Web/Android)   │                  │                 │
├─────────────────┤                  ├─────────────────┤
│ • UI Layer      │                  │ • API Handler   │
│ • State Mgmt    │                  │ • MediaMTX Ctrl │
│ • File Mgmt     │                  │ • Camera Monitor│
│ • Settings      │                  │ • Auth/Security │
└─────────────────┘                  └─────────────────┘
```

### T3: State Management
- **Connection State:** Connected, Disconnected, Connecting, Error
- **Camera State:** Available, Recording, Capturing, Error
- **Recording State:** Idle, Recording, Stopping, Paused
- **Application State:** Settings, User Preferences, File Storage

### T4: Error Recovery Patterns
- **Connection Failures:** Automatic retry with exponential backoff (1s, 2s, 4s, 8s, max 30s)
- **Service Errors:** Display user-friendly error messages with suggested actions
- **Camera Errors:** Graceful fallback to available cameras or manual refresh
- **Storage Errors:** Alternative storage options or user guidance

---

## Implementation Priorities

### Phase 1: Core Functionality (MVP)
1. Service connection and authentication
2. Camera list and selection
3. Basic photo capture
4. Basic video recording (unlimited duration)
5. File saving with metadata

### Phase 2: Enhanced Features
1. Timed recording functionality
2. Settings and configuration management
3. Recording progress and status indicators
4. Enhanced error handling and recovery

### Phase 3: Advanced Features
1. Camera preview integration
2. Background recording (Android)
3. PWA installation (Web)
4. Advanced file management

---

## Testing Requirements

### T1: Functional Testing
- **Unit Tests:** Individual component functionality
- **Integration Tests:** Service API communication
- **User Interface Tests:** User interaction workflows
- **Cross-Platform Tests:** Feature parity verification

### T2: Non-Functional Testing
- **Performance Tests:** Response time and resource usage
- **Security Tests:** Authentication and data protection
- **Compatibility Tests:** Browser and Android version support
- **Stress Tests:** Extended recording sessions and error scenarios

---

## Documentation Requirements

### D1: User Documentation
- **Installation Guide:** Platform-specific setup instructions
- **User Manual:** Feature usage and troubleshooting
- **Quick Start Guide:** Essential functionality overview

### D2: Developer Documentation
- **API Integration Guide:** Service communication patterns
- **Architecture Documentation:** Client-side design decisions
- **Build and Deployment Guide:** Development environment setup

---

## Compliance and Standards

### Code Quality
- Follow established coding standards from `docs/development/coding-standards.md`
- Implement comprehensive error handling and logging
- Maintain professional code formatting and documentation

### Security Compliance
- Implement secure credential storage
- Validate all external inputs
- Follow platform security best practices
- Regular security dependency updates

### Architecture Compliance
- Adhere to approved service API contracts
- Maintain consistency with service architecture patterns
- Document all architectural decisions

---

## Success Criteria

The client applications are considered successful when:

1. **Functional Completeness:** All specified camera operations work reliably
2. **Integration Quality:** Seamless communication with MediaMTX Camera Service
3. **User Experience:** Intuitive interface requiring minimal user training
4. **Platform Optimization:** Native look and feel on each target platform
5. **Reliability:** Graceful handling of service outages and errors
6. **Performance:** Meets all specified response time targets
7. **Security:** Secure credential and data handling
8. **Maintainability:** Clear code structure following project standards

---

## Future Considerations

### Potential Enhancements
- Live camera preview streaming
- Multi-camera simultaneous recording
- Cloud storage integration
- Advanced video editing capabilities
- Shared recording sessions

### Platform Expansion
- iOS application development
- Desktop applications (Electron/native)
- API client libraries for third-party integration

---

**Document Status:** Ready for architecture review and development planning  
**Next Steps:** Architecture validation and Epic creation in project roadmap