# Epic E6: Server Recording and Snapshot File Management Infrastructure

**Duration:** 4 days  
**Team:** Server Development Team  
**Goal:** Add file management API methods and download endpoints for camera recordings and snapshots to existing MediaMTX Camera Service server  
**Status:** 📋 PLANNED  
**Prerequisites:** MediaMTX Camera Service operational (✅ COMPLETE per roadmap.md E5)  

---

## Epic Overview

Add complete file management capabilities to the existing production-ready MediaMTX Camera Service server. This includes both JSON-RPC API methods for listing files and HTTP endpoints for downloading files, enabling clients to discover and access stored recordings and snapshots.

**Current Server Status (per roadmap.md):**
- ✅ E5: Deployment & Operations Strategy COMPLETE  
- ✅ Production deployment authorized
- ✅ Server operational on port 8002 with SSL on port 443
- ✅ Health endpoints operational on port 8003
- ✅ Nginx reverse proxy configured
- ✅ Recording path: `/opt/camera-service/recordings`
- ✅ Snapshot path: `/opt/camera-service/snapshots`

**Server Enhancement Target:**
- Server provides JSON-RPC API methods: `list_recordings`, `list_snapshots`
- Server provides HTTP file download endpoints: `/files/recordings/`, `/files/snapshots/`  
- Server maintains all existing functionality

---

## Epic Stories and Tasks

### S6.1: Create Server File Download Requirements
**Priority:** Critical  
**Estimated Effort:** 0.5 days  

#### Tasks:
- **S6.1.1:** Define server JSON-RPC file listing API requirements (`list_recordings`, `list_snapshots`)
- **S6.1.2:** Define server recording file download endpoint requirements
- **S6.1.3:** Define server snapshot file download endpoint requirements  
- **S6.1.4:** Define server nginx routing updates for `/files/` endpoints
- **S6.1.5:** Add requirements to `docs/requirements/requirements-baseline.md` (REQ-FUNC-008 through REQ-FUNC-012)
- **S6.1.6:** Validate requirements against existing server architecture

#### Acceptance Criteria:
- Server file management API requirements documented in requirements baseline
- Requirements specify `list_recordings` and `list_snapshots` JSON-RPC method functionality
- Requirements specify `/files/recordings/` and `/files/snapshots/` endpoint functionality  
- Requirements specify nginx routing updates for file endpoints
- Requirements approved and ready for implementation

### S6.2: Recording and Snapshot File Management API
**Priority:** Critical  
**Estimated Effort:** 1.5 days  

#### Tasks:
- **S6.2.1:** Implement JSON-RPC `list_recordings` method to enumerate available recording files
- **S6.2.2:** Implement JSON-RPC `list_snapshots` method to enumerate available snapshot files
- **S6.2.3:** Add file metadata support (filename, size, timestamp, duration for videos)
- **S6.2.4:** Implement proper error handling for directory access and file enumeration
- **S6.2.5:** Add pagination support for large file lists
- **S6.2.6:** Update JSON-RPC API documentation with new file management methods

#### Acceptance Criteria:
- `list_recordings` returns array of recording files with metadata (filename, size, timestamp, duration)
- `list_snapshots` returns array of snapshot files with metadata (filename, size, timestamp)
- File lists include download URLs or file identifiers for client download
- Proper error handling for empty directories, permission issues, and file access
- API methods follow existing JSON-RPC 2.0 patterns and authentication requirements

### S6.3: HTTP File Download Endpoints
**Priority:** Critical  
**Estimated Effort:** 1 day  

#### Tasks:
- **S6.3.1:** Implement server `/files/recordings/` endpoint with proper MIME type detection
- **S6.3.2:** Implement server `/files/snapshots/` endpoint with image format support
- **S6.3.3:** Add server 404 handling for non-existent files with appropriate error responses
- **S6.3.4:** Configure server Content-Disposition headers for proper file downloads
- **S6.3.5:** Implement server file access logging for security and troubleshooting
- **S6.3.6:** Validate server download functionality across different file formats

#### Acceptance Criteria:
- Server recording files downloadable with correct MIME types and filenames
- Server snapshot files downloadable with proper image format handling
- Server graceful 404 responses for missing files
- Server download operations logged for security audit trail

### S6.4: Update Existing Nginx Configuration
**Priority:** Critical  
**Estimated Effort:** 0.5 days  

#### Tasks:
- **S6.4.1:** Add `/files/recordings/` location block to existing nginx configuration  
- **S6.4.2:** Add `/files/snapshots/` location block to existing nginx configuration
- **S6.4.3:** Verify existing WebSocket `/api/ws` routing continues working (port 8002)
- **S6.4.4:** Verify existing health endpoint `/health/` routing continues working (port 8003)
- **S6.4.5:** Test nginx configuration reload with new file endpoints
- **S6.4.6:** Validate SSL/HTTPS continues working for all endpoints

#### Acceptance Criteria:
- Server nginx configuration updated with file download locations
- Existing server routing preserved (WebSocket on 8002, health on 8003)
- Server SSL/HTTPS operational for all endpoints including file downloads
- Server nginx configuration reload successful
- All existing server functionality verified operational

### S6.5: Update Existing Installation Procedures
**Priority:** Medium  
**Estimated Effort:** 0.5 days  

#### Tasks:
- **S6.5.1:** Update existing server installation script (`deployment/scripts/install.sh`) for file endpoint support
- **S6.5.2:** Verify recording and snapshot directory permissions (`/opt/camera-service/recordings`, `/opt/camera-service/snapshots`)
- **S6.5.3:** Update server validation script (`deployment/scripts/verify_installation.sh`) to test file endpoints and API methods
- **S6.5.4:** Update server production validation (`deployment/scripts/validate_production.sh`) to include file endpoint and API method checks
- **S6.5.5:** Document new file management API methods and download endpoints in installation guide

#### Acceptance Criteria:
- Server installation procedures support file download endpoints and API methods
- Server recording/snapshot directory permissions verified correct
- Server validation scripts test both file endpoint functionality and JSON-RPC file listing methods
- Server installation documentation updated with complete file management capabilities
- All existing server installation functionality preserved

---

## Epic Dependencies

### Input Dependencies (✅ All Currently Operational):
- ✅ MediaMTX Camera Service operational on port 8002 (E5 Complete per roadmap.md)
- ✅ Server SSL/HTTPS configuration operational on port 443 with nginx
- ✅ Server WebSocket JSON-RPC API endpoints operational at `/api/ws`
- ✅ Server health monitoring endpoints operational at `/health/` (port 8003)
- ✅ Server recording directory `/opt/camera-service/recordings` configured
- ✅ Server snapshot directory `/opt/camera-service/snapshots` configured
- ✅ Server nginx reverse proxy configuration operational

### Output Dependencies:
- Server provides JSON-RPC API methods for clients to list available recording and snapshot files
- Server provides HTTP file download access to recording files
- Server provides HTTP file download access to snapshot files
- All existing server functionality preserved (WebSocket, health, SSL)

---

## Quality Gates and Validation

### Automated Validation:
- Server health endpoint connectivity tests
- Server WebSocket connection establishment tests
- Server JSON-RPC `list_recordings` method functionality tests
- Server JSON-RPC `list_snapshots` method functionality tests
- Server recording file download endpoint functionality tests
- Server snapshot file download endpoint functionality tests
- Server nginx configuration syntax validation

### Manual Validation:
- End-to-end server file management workflow testing (list → download)
- Server recording file download verification across file formats
- Server snapshot file download verification across image formats
- Server SSL certificate functionality across all endpoints
- Server performance impact assessment on existing operations

### Evidence Requirements:
- Server automated test suite execution results including API method testing
- Server file management functionality documentation
- Server performance benchmarks before/after changes
- Server security assessment of new file management capabilities

---

## STOP Comments Requiring Resolution

- **File Access Control:** Should server file downloads and API listing methods require authentication/authorization checks?
- **File Security:** Should server log all file download requests and API calls for security audit?
- **Directory Traversal:** How should server prevent directory traversal attacks on file endpoints?
- **File Caching:** Should server implement caching headers for recording/snapshot file downloads?
- **Bandwidth Limiting:** Should server implement bandwidth throttling for large file downloads?
- **API Pagination:** What pagination limits should be implemented for `list_recordings` and `list_snapshots` methods?

---

## Epic Deliverable

**Server File Management Enhancement Package** comprising:
- Server JSON-RPC API methods: `list_recordings` and `list_snapshots` for file enumeration
- Server HTTP file download endpoints operational at `/files/recordings/` and `/files/snapshots/`
- Updated server nginx configuration supporting file download routing
- Server file management with proper MIME type handling and metadata support
- Updated server installation and validation procedures including file management testing
- Preserved existing server functionality (WebSocket API on 8002, health endpoints on 8003, SSL/HTTPS on 443)

**Handoff to Client Team:** Complete server file management infrastructure allowing clients to discover stored files via API methods and download them via HTTP endpoints with documented API specifications and preserved existing functionality.

---

## Epic Timeline

### Week 7: Epic E6 Implementation
- **Day 1:** S6.1 Requirements definition and S6.2 API implementation
- **Day 2:** S6.2 API completion and S6.3 HTTP endpoints
- **Day 3:** S6.4 Nginx configuration and S6.5 Installation procedures
- **Day 4:** Integration testing, validation, and documentation

### Success Metrics:
- All 5 stories completed within 4-day timeline
- Server file management API methods operational
- Server file download endpoints functional
- Existing server functionality preserved
- Quality gates passed with automated and manual validation

---

**Epic E6 Status:** 📋 PLANNED - Ready for implementation after E5 completion
