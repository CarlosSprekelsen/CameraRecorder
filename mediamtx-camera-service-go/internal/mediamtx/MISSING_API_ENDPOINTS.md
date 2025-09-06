# Missing MediaMTX API Endpoints

## Overview
This document lists MediaMTX API endpoints that are available in the swagger.json specification but not currently implemented in this codebase. Each endpoint is categorized by purpose and impact on user experience.

## Authentication & Security

### `POST /v3/auth/jwks/refresh`
- **Purpose**: Manually refresh JWT JWKS (JSON Web Key Set)
- **Current Status**: ❌ Not implemented
- **User Experience Impact**: 
  - **LOW** - Only needed for JWT-based authentication
  - **Needed if**: Your application uses JWT authentication with MediaMTX
  - **Improvement**: Automatic token refresh, better security management

## Configuration Management

### `GET /v3/config/global/get`
### `PATCH /v3/config/global/patch`
- **Purpose**: Get and modify global MediaMTX configuration
- **Current Status**: ❌ Not implemented
- **User Experience Impact**:
  - **MEDIUM** - Allows runtime configuration changes
  - **Needed if**: You want to modify MediaMTX settings without restart
  - **Improvement**: Dynamic configuration, no service restarts required

### `GET /v3/config/pathdefaults/get`
### `PATCH /v3/config/pathdefaults/patch`
- **Purpose**: Get and modify default path configuration
- **Current Status**: ❌ Not implemented
- **User Experience Impact**:
  - **MEDIUM** - Centralized path configuration management
  - **Needed if**: You want to set default settings for all new paths
  - **Improvement**: Consistent path behavior, easier management

## HLS (HTTP Live Streaming) Management

### `GET /v3/hlsmuxers/list`
### `GET /v3/hlsmuxers/get/{name}`
- **Purpose**: List and get HLS muxer information
- **Current Status**: ❌ Not implemented
- **User Experience Impact**:
  - **HIGH** - Essential for HLS streaming functionality
  - **Needed if**: Your application provides HLS streaming
  - **Improvement**: Real-time streaming, better viewer experience

## Recording Management

### `DELETE /v3/recordings/deletesegment`
- **Purpose**: Delete specific recording segments
- **Current Status**: ❌ Not implemented
- **User Experience Impact**:
  - **MEDIUM** - Granular recording management
  - **Needed if**: You want to delete specific parts of recordings
  - **Improvement**: Storage management, selective cleanup

## Connection Management

### RTSP Connections
- `GET /v3/rtspconns/list`
- `GET /v3/rtspconns/get/{id}`

### RTSP Sessions
- `GET /v3/rtspsessions/list`
- `GET /v3/rtspsessions/get/{id}`
- `POST /v3/rtspsessions/kick/{id}`

### RTSPS (Secure RTSP) Connections & Sessions
- `GET /v3/rtspsconns/list`
- `GET /v3/rtspsconns/get/{id}`
- `GET /v3/rtspssessions/list`
- `GET /v3/rtspssessions/get/{id}`
- `POST /v3/rtspssessions/kick/{id}`

### RTMP Connections
- `GET /v3/rtmpconns/list`
- `GET /v3/rtmpconns/get/{id}`
- `POST /v3/rtmpconns/kick/{id}`

### RTMPS (Secure RTMP) Connections
- `GET /v3/rtmpsconns/list`
- `GET /v3/rtmpsconns/get/{id}`
- `POST /v3/rtmpsconns/kick/{id}`

### SRT Connections
- `GET /v3/srtconns/list`
- `GET /v3/srtconns/get/{id}`
- `POST /v3/srtconns/kick/{id}`

### WebRTC Sessions
- `GET /v3/webrtcsessions/list`
- `GET /v3/wertcsessions/get/{id}`
- `POST /v3/webrtcsessions/kick/{id}`

- **Purpose**: Monitor and manage all types of media connections
- **Current Status**: ❌ Not implemented
- **User Experience Impact**:
  - **HIGH** - Essential for connection monitoring and management
  - **Needed if**: You want to monitor active connections, troubleshoot issues, or manage bandwidth
  - **Improvement**: Better debugging, connection management, bandwidth control

## Implementation Priority

### High Priority (Core Functionality)
1. **HLS Management** - If your application provides streaming
 - RTSP is sufficient!!. Streaming SHLL BE COMPATIBLE WITH STANAG  **Connection Management** - For monitoring and troubleshooting

### Medium Priority (Enhanced Features)
1. **Global Configuration** - For runtime configuration changes
2. **Path Defaults** - For centralized path management
3. **Recording Segment Deletion** - For storage management

### Low Priority (Advanced Features)
1. **JWT Authentication** - Only if using JWT-based auth

## Decision Matrix

| Endpoint Category | Implementation Effort | User Value | Recommended Action |
|------------------|---------------------|------------|-------------------|
| HLS Management | Medium | High | **IMPLEMENT** if streaming needed |
| Connection Management | High | High | **IMPLEMENT** for production use |
| Global Config | Low | Medium | **IMPLEMENT** for flexibility |
| Path Defaults | Low | Medium | **IMPLEMENT** for consistency |
| Recording Segments | Low | Medium | **IMPLEMENT** for storage management |
| JWT Auth | Medium | Low | **SKIP** unless JWT required |

## Notes
- All endpoints are documented in `docs/api/swagger.json`
- Current implementation covers basic path and recording management
- Missing endpoints can be added incrementally based on requirements
- Connection management is particularly valuable for production monitoring
