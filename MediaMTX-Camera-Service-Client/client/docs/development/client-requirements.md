# Client Application Requirements Document

**Version:** 1.1  
**Authors:** System Architect  
**Date:** 2025-08-04  
**Status:** Draft  
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
- `list_recordings` - Enumerate available recording files with metadata
- `list_snapshots` - Enumerate available snapshot files with metadata
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
  - API Contract: JSON-RPC `start_recording` without a `duration` parameter SHALL start an unlimited recording session which continues until `stop_recording` is invoked.
  - Alternative: When `duration_mode` is "unlimited", the `duration_value` parameter MUST be omitted.
  - Service Behavior: Service SHALL maintain the session until explicit stop; intermediate status updates MAY be emitted by service as notifications.
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

### F4: File Browsing Interface Requirements (NEW)

#### F4.1: File List Display
- **F4.1.1:** The application SHALL display paginated list of available recordings and snapshots
- **F4.1.2:** The application SHALL show file metadata (filename, size, timestamp, duration for videos)
- **F4.1.3:** The application SHALL implement pagination controls with configurable limits (10, 20, 50 files per page)
- **F4.1.4:** The application SHALL provide file filtering and sorting options (date, size, name, duration)
- **F4.1.5:** The application SHALL update file lists in real-time when new files are created

#### F4.2: File Metadata Display
- **F4.2.1:** The application SHALL display primary metadata fields prominently (filename, size, date, duration)
- **F4.2.2:** The application SHALL provide expandable secondary metadata fields (path, format, resolution, frame rate)
- **F4.2.3:** The application SHALL format file sizes in human-readable format (KB, MB, GB)
- **F4.2.4:** The application SHALL format timestamps in user's local timezone (YYYY-MM-DD HH:MM:SS)
- **F4.2.5:** The application SHALL display duration for video files in MM:SS format

### F5: File Download Functionality Requirements (NEW)

#### F5.1: Secure Download
- **F5.1.1:** The application SHALL support secure HTTPS download via `/files/recordings/` and `/files/snapshots/` endpoints
- **F5.1.2:** The application SHALL require authentication for all file download requests
- **F5.1.3:** The application SHALL use existing WebSocket authentication session for download authorization
- **F5.1.4:** The application SHALL handle authentication failures gracefully

#### F5.2: Download Management
- **F5.2.1:** The application SHALL provide progress indication for large file downloads
- **F5.2.2:** The application SHALL integrate with browser download mechanism with proper filename preservation
- **F5.2.3:** The application SHALL handle download failures and network timeout errors (30-minute timeout)
- **F5.2.4:** The application SHALL support bulk download operations for up to 10 files simultaneously
- **F5.2.5:** The application SHALL provide download queue management and progress tracking

### F6: File Management UI Requirements (MVP SCOPE)

#### F6.1: Basic File Interface (MVP - Sprint 3)
- **F6.1.1:** The application SHALL provide separate tabs/sections for recordings and snapshots
- **F6.1.2:** The application SHALL display file metadata prominently (filename, size, date, duration)
- **F6.1.3:** The application SHALL implement basic pagination controls (25 items per page default)
- **F6.1.4:** The application SHALL ensure responsive design for mobile file browsing
- **F6.1.5:** The application SHALL provide file download functionality via HTTPS endpoints

#### F6.2: Advanced File Management (Phase 4 - Deferred)
- **F6.2.1:** The application SHALL offer file preview capabilities where supported (images, video thumbnails)
- **F6.2.2:** The application SHALL provide file metadata detailed view with expandable sections
- **F6.2.3:** The application SHALL implement search and filter functionality for file discovery
- **F6.2.4:** The application SHALL support bulk download operations for up to 10 selected files
- **F6.2.5:** The application SHALL support bulk delete operations for up to 10 selected files
- **F6.2.6:** The application SHALL provide clear progress indication for bulk operations
- **F6.2.7:** The application SHALL confirm bulk delete operations with user before execution
- **F6.2.8:** The application SHALL handle partial failures gracefully in bulk operations

#### F6.3: Caching and Performance (Phase 4 - Deferred)
- **F6.3.1:** The application SHALL implement 5-minute client-side cache for file metadata
- **F6.3.2:** The application SHALL invalidate cache when new files are created or deleted
- **F6.3.3:** The application SHALL provide offline file list viewing capability
- **F6.3.4:** The application SHALL optimize performance for large file lists with advanced pagination
- **F3.2.2:** The application SHALL display recording duration selector interface
- **F3.2.3:** The application SHALL show recording progress and elapsed time
- **F3.2.4:** The application SHALL provide emergency stop functionality
- **F3.2.5:** Operator permissions SHALL be required to invoke `start_recording`, `stop_recording`, and `take_snapshot`
  - API Contract: Protected JSON-RPC methods SHALL require a valid JWT with role=operator.
  - Token Transport: The JWT SHALL be provided via JSON-RPC `authenticate` method prior to using protected methods.
    - `authenticate` request: `{ jsonrpc: "2.0", method: "authenticate", params: { auth_token: string } }`
    - On success, the server SHALL associate the client connection with the authenticated user and role for the session.
  - Error Handling: Missing, invalid, or expired tokens SHALL result in JSON-RPC error with code -32004 (authentication required) and a meaningful message.
- **F3.2.6:** The application SHALL handle token expiration by re-authenticating before retrying protected operations.
- **F3.2.7:** The application SHALL implement client-side JWT authentication service with:
  - Token validation and expiry checking
  - Automatic token refresh before expiry
  - Role-based permission checking
  - Integration with WebSocket service for protected operations

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
- **W2.4:** File browsing interface with pagination and sorting
- **W2.5:** HTTPS file download via `/files/recordings/` and `/files/snapshots/` endpoints

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
- **N1.1:** Application startup time SHALL be under 3 seconds
- **N1.2:** Camera list refresh SHALL complete within 1 second
- **N1.3:** Photo capture response SHALL be under 2 seconds
- **N1.4:** Video recording start SHALL begin within 2 seconds
- **N1.5:** UI interactions SHALL provide immediate feedback (200ms)

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
4. Basic video recording (unlimited and timed duration)
5. File browsing interface for snapshots and recordings (basic)
6. File download capabilities via HTTPS endpoints
7. Basic file metadata display
8. Real-time WebSocket updates with polling fallback
9. PWA with responsive design

### Phase 2: Enhanced Features
1. Settings and configuration management
2. Recording progress and status indicators
3. Enhanced error handling and recovery
4. Basic file management improvements

### Phase 3: Advanced Features
1. Camera preview integration
2. Background recording (Android)
3. Advanced file management capabilities
4. Cloud storage integration

### Phase 4: Advanced File Management (Deferred)
1. File preview capabilities (images, video thumbnails)
2. Advanced file metadata display with expandable sections
3. Search and filter functionality for file discovery
4. Bulk download operations (up to 10 files)
5. Bulk delete operations with confirmation
6. Advanced caching and performance optimization
7. Offline file list viewing capability

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

**Document Status:** Updated with MVP file management requirements  
**Next Steps:** Client development team assignment and Phase 1A implementation