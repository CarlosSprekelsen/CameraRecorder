# Security Module Architecture

## Overview
The Security module provides comprehensive authentication, authorization, input validation, and security enforcement for the MediaMTX Camera Service Go implementation. It implements JWT-based authentication, role-based access control (RBAC), rate limiting, session management, and input sanitization following security best practices.

## Core Components

### 1. **JWTHandler** (Authentication Core)
**Role**: Central JWT token generation, validation, and rate limiting
**Location**: `jwt_handler.go`
**Responsibilities**:
- JWT token generation with configurable expiry
- Token validation with algorithm restriction (HS256)
- Rate limiting with per-client tracking
- Client cleanup and statistics

**Key Methods**:
```go
// Token operations
GenerateToken(userID, role string, expiryHours int) (string, error)
ValidateToken(tokenString string) (*JWTClaims, error)
IsTokenExpired(tokenString string) bool

// Rate limiting
CheckRateLimit(clientID string) bool
RecordRequest(clientID string)
SetRateLimit(limit int64, window time.Duration)
CleanupExpiredClients(maxInactive time.Duration)
```

### 2. **PermissionChecker** (Role-Based Access Control)
**Role**: Manages method-level permissions and role hierarchy
**Location**: `role_manager.go`
**Responsibilities**:
- Role hierarchy enforcement (Viewer < Operator < Admin)
- Method permission matrix management
- Dynamic permission configuration
- Role validation and conversion

**Key Methods**:
```go
// Permission checking
HasPermission(userRole Role, method string) bool
GetRequiredRole(method string) Role
ValidateRole(roleString string) (Role, error)

// Permission management
AddMethodPermission(method string, requiredRole Role) error
RemoveMethodPermission(method string) error
GetMethodPermissions() map[string]string
```

### 3. **SessionManager** (Session Lifecycle)
**Role**: Manages user sessions with automatic cleanup
**Location**: `session_manager.go`
**Responsibilities**:
- Session creation and validation
- Automatic session cleanup
- Activity tracking and timeout management
- Session statistics and user management

**Key Methods**:
```go
// Session operations
CreateSession(userID string, role Role) (*Session, error)
ValidateSession(sessionID string) (*Session, error)
UpdateActivity(sessionID string)
InvalidateUserSessions(userID string) error

// Management
GetSessionStats() map[string]interface{}
GetSessionByUserID(userID string) []*Session
Stop() // Graceful shutdown
```

### 4. **EnhancedRateLimiter** (Advanced Rate Limiting)
**Role**: Provides per-method rate limiting with DDoS protection
**Location**: `rate_limiter.go`
**Responsibilities**:
- Method-specific rate limits
- Global rate limiting
- DDoS protection with client blocking
- Automatic cleanup of old client data

**Key Methods**:
```go
// Rate limiting
CheckLimit(method, clientID string) error
SetMethodRateLimit(method string, config *RateLimitConfig)
ResetClientLimits(clientID string)

// Statistics and management
GetClientStats(clientID string) map[string]interface{}
GetMethodStats(method string) map[string]interface{}
GetGlobalStats() map[string]interface{}
CleanupOldClients(maxAge time.Duration)
```

### 5. **InputValidator** (Input Sanitization)
**Role**: Comprehensive input validation and sanitization
**Location**: `input_validator.go`
**Responsibilities**:
- Camera ID format validation
- Parameter type and range validation
- Path traversal protection
- String sanitization and security checks

**Key Methods**:
```go
// Validation methods
ValidateCameraID(cameraID string) *ValidationResult
ValidateDevicePath(devicePath interface{}) *ValidationResult
ValidateFilename(filename interface{}) *ValidationResult
ValidateRecordingOptions(options map[string]interface{}) *ValidationResult

// Sanitization
SanitizeString(input string) string
SanitizeMap(input map[string]interface{}) map[string]interface{}
```

### 6. **AuthMiddleware** (Authentication Enforcement)
**Role**: Decorates method handlers with authentication requirements
**Location**: `middleware.go`
**Responsibilities**:
- Authentication requirement enforcement
- Client connection validation
- Security event logging

**Key Methods**:
```go
RequireAuth(handler MethodHandler) MethodHandler
```

### 7. **RBACMiddleware** (Authorization Enforcement)
**Role**: Decorates method handlers with role-based access control
**Location**: `middleware.go`
**Responsibilities**:
- Role-based access control enforcement
- Permission validation
- Security audit logging

**Key Methods**:
```go
RequireRole(requiredRole Role, handler MethodHandler) MethodHandler
```

### 8. **SecureMethodRegistry** (Method Security Orchestration)
**Role**: Central registry for securing API methods
**Location**: `middleware.go`
**Responsibilities**:
- Method registration with automatic security enforcement
- Security decorator composition
- Method security information tracking

**Key Methods**:
```go
RegisterMethod(methodName string, handler MethodHandler, requiredRole Role)
GetMethod(methodName string) (MethodHandler, bool)
GetAllMethods() []string
GetMethodSecurityInfo(methodName string) map[string]interface{}
```

### 9. **ConfigAdapter** (Configuration Bridge)
**Role**: Bridges centralized configuration with security components
**Location**: `config_adapter.go`
**Responsibilities**:
- Adapts centralized config to security module needs
- Provides fallback defaults
- Creates security-specific configurations

**Key Methods**:
```go
// Configuration access
GetSecurityConfig() *config.SecurityConfig
GetRateLimitRequests() int
GetJWTSecretKey() string
GetJWTExpiryHours() int

// Configuration creation
CreateRateLimiterConfig() map[string]*RateLimitConfig
CreateAuditLoggerConfig() map[string]interface{}
```

## Data Flow

### Authentication Flow
```
Client Request
→ WebSocket Server
→ Rate Limiting Check (JWTHandler)
→ Permission Check (PermissionChecker)
→ Method Execution
→ Response
```

### Session Management Flow
```
User Login
→ Session Creation (SessionManager)
→ JWT Token Generation (JWTHandler)
→ Session Validation
→ Activity Updates
→ Automatic Cleanup
```

### Input Validation Flow
```
Client Input
→ Input Validation (InputValidator)
→ Sanitization
→ Security Checks
→ Method Handler
```

## Configuration Integration

### Centralized Config Usage
All components use centralized configuration through `ConfigAdapter`:
```go
type ConfigAdapter struct {
    securityConfig *config.SecurityConfig
    loggingConfig  *config.LoggingConfig
}
```

### Centralized Logger Usage
All components use `*logging.Logger` from centralized logging:
```go
logger.WithFields(logging.Fields{
    "component": "security",
    "action":    "authentication",
    "user_id":   userID,
}).Info("Authentication successful")
```

## Role Hierarchy

### RoleViewer (Level 1)
- **Permissions**: Read-only access to camera status and basic information
- **Methods**: ping, get_camera_list, get_camera_status, get_camera_capabilities, list_recordings, list_snapshots, get_recording_info, get_snapshot_info, get_streams, get_stream_url, get_stream_status

### RoleOperator (Level 2)
- **Permissions**: Viewer permissions plus camera control operations
- **Methods**: take_snapshot, start_recording, stop_recording, delete_recording, delete_snapshot, start_streaming, stop_streaming

### RoleAdmin (Level 3)
- **Permissions**: Full access to all features including system management
- **Methods**: get_metrics, get_status, get_server_info, get_storage_info, set_retention_policy, cleanup_old_files

## Integration Points

### External Integration
- **WebSocket Server**: Uses security components for authentication and authorization
- **HTTP API**: Uses security components for request validation
- **Main Application**: Initializes and configures security components

### Internal Integration
- **JWTHandler** provides rate limiting to **WebSocket Server**
- **PermissionChecker** validates roles for **RBACMiddleware**
- **SessionManager** manages sessions for **AuthMiddleware**
- **InputValidator** validates parameters for all method handlers
- **ConfigAdapter** provides configuration to all security components

## Security Features

### Authentication
- JWT token-based authentication with HS256 algorithm
- Configurable token expiry
- Token validation with algorithm restriction
- Secure token generation and signing

### Authorization
- Role-based access control (RBAC)
- Method-level permission enforcement
- Role hierarchy validation
- Dynamic permission management

### Rate Limiting
- Per-client rate limiting
- Method-specific rate limits
- Global rate limiting
- DDoS protection with client blocking
- Automatic cleanup of expired data

### Input Validation
- Comprehensive parameter validation
- Path traversal protection
- String sanitization
- Type and range validation
- Security-focused validation patterns

### Session Management
- Secure session creation and validation
- Automatic session cleanup
- Activity tracking
- User session management
- Graceful shutdown handling

## Architecture Benefits

1. **Separation of Concerns**: Each component has a clear, single responsibility
2. **Centralized Configuration**: All components use shared config and logger
3. **Composable Security**: Middleware can be combined for layered security
4. **Extensible Design**: Easy to add new security features or validation rules
5. **Performance Optimized**: Efficient rate limiting and session management
6. **Thread-Safe**: All components are designed for concurrent access
7. **Comprehensive Logging**: Detailed security event logging for auditing

## Testing Strategy

### Unit Tests
- All security components have comprehensive unit tests
- Mock implementations for external dependencies
- Edge case and error condition testing

### Integration Tests
- Security middleware integration with WebSocket server
- Configuration adapter integration testing
- End-to-end authentication and authorization flows

### Security Tests
- Input validation security testing
- Rate limiting effectiveness testing
- Session management security testing
- JWT token security validation

## Dependencies

### External Dependencies
- `github.com/golang-jwt/jwt/v4`: JWT token handling
- `golang.org/x/time/rate`: Advanced rate limiting
- `github.com/google/uuid`: Session ID generation

### Internal Dependencies
- `internal/logging`: Centralized logging
- `internal/config`: Centralized configuration
- `internal/websocket`: WebSocket server integration

## Performance Considerations

### Rate Limiting
- Uses efficient token bucket algorithm
- Per-client tracking with automatic cleanup
- Configurable limits per method

### Session Management
- In-memory session storage with automatic cleanup
- Efficient session lookup and validation
- Graceful shutdown with connection cleanup

### Input Validation
- Fast regex-based validation patterns
- Efficient string sanitization
- Minimal memory allocation during validation

## Security Considerations

### JWT Security
- Algorithm restriction to prevent confusion attacks
- Secure token generation and validation
- Configurable token expiry

### Input Security
- Path traversal protection
- String sanitization
- Type validation and range checking

### Rate Limiting Security
- DDoS protection with client blocking
- Per-method rate limiting
- Automatic cleanup of attack patterns

### Session Security
- Secure session ID generation
- Automatic session timeout
- User session invalidation capabilities
