# JWT Token Authentication Guide

**Version:** 1.0  
**Date:** 2025-09-27  
**Status:** JWT Token Authentication Documentation  
**Document Type:** Authentication Guide

---

## 1. Overview

The MediaMTX Camera Service uses JWT (JSON Web Token) authentication for secure access control. This guide explains how to generate, use, and manage JWT tokens for testing and development.

## 2. JWT Token Structure

### 2.1 Token Format
JWT tokens consist of three parts separated by dots:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdF9hZG1pbiIsInJvbGUiOiJhZG1pbiIsImlhdCI6MTc1ODk5NTgxNywiZXhwIjoxNzU5MjU1MDE3fQ.w4q00ieqoKDq_NE9lCIJ7wbV3JlwBWQ_q3T7HgEz9IA
```

### 2.2 Token Claims
Each JWT token contains the following claims:
- `user_id`: User identifier (e.g., "test_admin")
- `role`: User role ("viewer", "operator", "admin")
- `iat`: Issued at timestamp (Unix timestamp)
- `exp`: Expiration timestamp (Unix timestamp)

### 2.3 Algorithm
- **Signing Algorithm**: HS256 (HMAC with SHA-256)
- **Secret Key**: Configurable via server configuration
- **Default Secret**: "edge-device-secret-key-change-in-production"

## 3. Token Generation

### 3.1 Using the JWT Generator Tool

Generate JWT tokens using the built-in Go tool:

```bash
cd /home/carlossprekelsen/CameraRecorder/mediamtx-camera-service-go
go run cmd/jwt-generator/main.go --role admin --expiry-hours 72 --format json
```

### 3.2 Available Options

- `--role`: User role (viewer, operator, admin)
- `--expiry-hours`: Token expiry in hours (default: 48)
- `--secret-key`: JWT secret key (default: from config)
- `--user-id`: Custom user ID (default: test_<role>)
- `--format`: Output format (token, json)

### 3.3 Example Commands

```bash
# Generate admin token with 72h expiry
go run cmd/jwt-generator/main.go --role admin --expiry-hours 72

# Generate viewer token with custom user ID
go run cmd/jwt-generator/main.go --role viewer --user-id "john_doe" --expiry-hours 24

# Generate operator token with JSON output
go run cmd/jwt-generator/main.go --role operator --format json
```

## 4. Token Usage

### 4.1 Authentication Flow

1. **Connect** to WebSocket endpoint: `ws://localhost:8002/ws`
2. **Authenticate** using the `authenticate` method:

```json
{
  "jsonrpc": "2.0",
  "method": "authenticate",
  "params": {
    "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "id": 1
}
```

3. **Use authenticated session** for all subsequent requests

### 4.2 Role Permissions

| Role     | Permissions |
|----------|-------------|
| **viewer** | Read-only access to camera feeds, recordings, and basic information |
| **operator** | Viewer permissions + camera control, snapshots, recording management |
| **admin** | Full system access including configuration and user management |

## 5. Token Management

### 5.1 Token Lifecycle

1. **Generation**: Create token with specified role and expiry
2. **Validation**: Server validates token signature and expiration
3. **Session**: Establish authenticated session after successful validation
4. **Expiration**: Token automatically expires at specified time

### 5.2 Security Best Practices

- **Short Expiry**: Use appropriate expiry times (24-72 hours for testing)
- **Secure Storage**: Store tokens securely, never commit to version control
- **Rotation**: Regenerate tokens regularly for production use
- **Validation**: Always validate tokens on the server side

## 6. Development and Testing

### 6.1 Test Token Generation

Use the deployment script to generate test tokens:

```bash
cd /home/carlossprekelsen/CameraRecorder/mediamtx-camera-service-go/deployment/scripts
./generate-jwt-tokens.sh --expiry-hours 72
```

This generates tokens for all roles and creates environment files.

### 6.2 Environment Variables

Test tokens are automatically exported to environment variables:

```bash
export TEST_VIEWER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export TEST_OPERATOR_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export TEST_ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 6.3 Client Integration

Load tokens in client tests:

```bash
cd /home/carlossprekelsen/CameraRecorder/MediaMTX-Camera-Service-Client/client
source .test_env
npm run test:integration
```

## 7. Troubleshooting

### 7.1 Common Issues

**Token Expired**
- Error: "token has expired"
- Solution: Generate new token with longer expiry

**Invalid Signature**
- Error: "failed to validate JWT token"
- Solution: Ensure token was generated with correct secret key

**Missing Claims**
- Error: "missing required field"
- Solution: Verify token contains all required claims (user_id, role, iat, exp)

### 7.2 Debug Information

Enable verbose logging to debug authentication issues:

```bash
# Check server logs
journalctl -u camera-service -f

# Validate token manually
go run cmd/jwt-generator/main.go --role admin --format json
```

## 8. Configuration

### 8.1 Server Configuration

JWT settings in `/opt/camera-service/config/default.yaml`:

```yaml
security:
  jwt_secret_key: "edge-device-secret-key-change-in-production"
  jwt_expiry_hours: 48
```

### 8.2 Client Configuration

Environment variables in `.test_env`:

```bash
export CAMERA_SERVICE_HOST=localhost
export CAMERA_SERVICE_PORT=8002
export CAMERA_SERVICE_WS_PATH=/ws
export TEST_ADMIN_TOKEN="your-jwt-token-here"
```

---

## 9. Migration from API Keys

**Note**: The server previously supported API key authentication, but now exclusively uses JWT tokens. If you have existing API key documentation or scripts, they should be updated to use JWT tokens instead.

### 9.1 Key Differences

| Aspect | API Keys | JWT Tokens |
|--------|----------|------------|
| **Storage** | Server-side database | Self-contained |
| **Validation** | Database lookup | Cryptographic signature |
| **Expiry** | Database tracking | Embedded in token |
| **Revocation** | Immediate via database | Wait for expiration |

### 9.2 Migration Steps

1. Update authentication scripts to generate JWT tokens
2. Modify client code to use JWT tokens instead of API keys
3. Update documentation to reflect JWT authentication
4. Test authentication flow with JWT tokens
