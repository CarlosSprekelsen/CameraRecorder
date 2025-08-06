# Authentication Setup and Usage Guide

**Version:** 1.0  
**Architecture Decision:** AD-7  
**Implementation:** Sprint 1 (S6)

## Overview

The MediaMTX Camera Service implements a comprehensive authentication system supporting both JWT tokens for user sessions and API keys for service authentication as specified in Architecture Decision AD-7.

## Authentication Methods

### JWT Token Authentication

JWT (JSON Web Token) authentication is designed for user sessions and provides:
- Configurable token expiry (default: 24 hours)
- Role-based access control (viewer, operator, admin)
- HS256 algorithm with secure secret key
- Automatic token validation and permission checking

### API Key Authentication

API key authentication is designed for service-to-service communication and provides:
- Secure key generation with bcrypt hashing
- Key rotation capability
- Configurable expiry dates
- Programmatic access for automated systems

## Configuration

### Security Configuration Schema

```yaml
security:
  jwt:
    secret_key: "${JWT_SECRET_KEY}"
    expiry_hours: 24
    algorithm: "HS256"
  api_keys:
    storage_file: "${API_KEYS_FILE:/etc/camera-service/api-keys.json}"
  rate_limiting:
    max_connections: 100
    requests_per_minute: 60
```

### Environment Variables

**JWT Configuration:**
- `JWT_SECRET_KEY`: Secret key for JWT signing (required)
- `JWT_EXPIRY_HOURS`: Token expiry in hours (default: 24)

**API Key Configuration:**
- `API_KEYS_FILE`: Path to API keys storage file (default: /etc/camera-service/api-keys.json)

**Rate Limiting:**
- `MAX_CONNECTIONS`: Maximum concurrent connections (default: 100)
- `REQUESTS_PER_MINUTE`: Rate limit per client (default: 60)

## JWT Token Usage

### Token Generation

JWT tokens are generated with the following claims:
- `user_id`: Unique user identifier
- `role`: User role (viewer, operator, admin)
- `iat`: Issued at timestamp
- `exp`: Expiration timestamp

### Role Hierarchy

```
viewer (1) < operator (2) < admin (3)
```

- **viewer**: Read-only access to camera status and streams
- **operator**: Viewer permissions + camera control (snapshots, recording)
- **admin**: Full access to all features

### WebSocket Authentication

JWT tokens are passed in JSON-RPC requests:

```json
{
  "jsonrpc": "2.0",
  "method": "get_camera_list",
  "id": 1,
  "params": {
    "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### Authentication Flow

1. Client connects to WebSocket server
2. Client sends JSON-RPC request with `auth_token` in params
3. Server validates JWT token
4. Server checks user permissions for requested method
5. Server executes method if authorized

## API Key Usage

### Key Generation

API keys are generated with secure random strings and stored with bcrypt hashing:

```bash
# Generate API key (only shown once)
curl -X POST http://localhost:8002/api/keys \
  -H "Content-Type: application/json" \
  -d '{"name": "Service Key", "role": "operator"}'
```

### Key Management

**List API Keys:**
```bash
curl -X GET http://localhost:8002/api/keys
```

**Revoke API Key:**
```bash
curl -X DELETE http://localhost:8002/api/keys/{key_id}
```

**Rotate API Key:**
```bash
curl -X POST http://localhost:8002/api/keys/{key_id}/rotate
```

### API Key Authentication

API keys are passed in the same way as JWT tokens:

```json
{
  "jsonrpc": "2.0",
  "method": "take_snapshot",
  "id": 1,
  "params": {
    "auth_token": "AbC123DeF456GhI789JkL012MnO345PqR678StU901VwX234YzA567BcD890"
  }
}
```

## Rate Limiting

### Connection Limits

- Maximum 100 concurrent WebSocket connections (configurable)
- Graceful connection rejection with proper error codes
- Connection tracking and cleanup

### Request Rate Limiting

- Per-client sliding window rate limiting
- Default: 60 requests per minute per client
- Configurable limits by user role
- Rate limit exceeded responses

### Rate Limit Responses

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32002,
    "message": "Rate limit exceeded"
  },
  "id": 1
}
```

## Error Codes

### Authentication Errors

| Code | Message | Description |
|------|---------|-------------|
| -32001 | Authentication required | No auth token provided |
| -32001 | Authentication failed | Invalid or expired token |
| -32003 | Insufficient permissions | Role does not have required permissions |

### Rate Limiting Errors

| Code | Message | Description |
|------|---------|-------------|
| -32002 | Rate limit exceeded | Client exceeded request rate limit |

## Security Best Practices

### JWT Security

1. **Use Strong Secret Keys**
   ```bash
   # Generate secure secret key
   openssl rand -base64 32
   ```

2. **Set Appropriate Expiry**
   ```yaml
   security:
     jwt:
       expiry_hours: 24  # Adjust based on security requirements
   ```

3. **Rotate Secret Keys**
   - Change JWT secret key periodically
   - Implement key rotation strategy
   - Monitor for token compromise

### API Key Security

1. **Secure Key Storage**
   ```bash
   # Set secure permissions on API key file
   chmod 600 /etc/camera-service/api-keys.json
   chown camera-service:camera-service /etc/camera-service/api-keys.json
   ```

2. **Key Rotation**
   - Rotate API keys regularly
   - Implement automated rotation
   - Monitor key usage patterns

3. **Access Control**
   - Limit API key permissions to minimum required
   - Use different keys for different services
   - Monitor key usage and revoke unused keys

### Network Security

1. **Use HTTPS/WSS**
   ```yaml
   security:
     ssl:
       enabled: true
       cert_file: "/path/to/cert.pem"
       key_file: "/path/to/key.pem"
   ```

2. **Network Policies**
   - Restrict access to authentication endpoints
   - Use firewall rules to limit connections
   - Monitor network traffic patterns

## Implementation Examples

### Python Client Example

```python
import jwt
import websockets
import json

# Generate JWT token
secret_key = "your-secret-key"
token = jwt.encode(
    {
        "user_id": "user123",
        "role": "operator",
        "iat": int(time.time()),
        "exp": int(time.time()) + 3600
    },
    secret_key,
    algorithm="HS256"
)

# Connect to WebSocket
async with websockets.connect("ws://localhost:8002/ws") as websocket:
    # Send authenticated request
    request = {
        "jsonrpc": "2.0",
        "method": "get_camera_list",
        "id": 1,
        "params": {"auth_token": token}
    }
    
    await websocket.send(json.dumps(request))
    response = await websocket.recv()
    print(json.loads(response))
```

### JavaScript Client Example

```javascript
// Generate JWT token (client-side)
const token = jwt.sign(
    {
        user_id: "user123",
        role: "viewer",
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + 3600
    },
    secretKey,
    { algorithm: "HS256" }
);

// Connect to WebSocket
const ws = new WebSocket("ws://localhost:8002/ws");

ws.onopen = function() {
    // Send authenticated request
    const request = {
        jsonrpc: "2.0",
        method: "get_camera_status",
        id: 1,
        params: { auth_token: token }
    };
    
    ws.send(JSON.stringify(request));
};

ws.onmessage = function(event) {
    const response = JSON.parse(event.data);
    console.log(response);
};
```

## Troubleshooting

### Common Authentication Issues

1. **Invalid JWT Token**
   - Check token format and signature
   - Verify token hasn't expired
   - Ensure correct secret key is used

2. **Permission Denied**
   - Verify user role has required permissions
   - Check role hierarchy for method access
   - Review method-specific permission requirements

3. **Rate Limit Exceeded**
   - Reduce request frequency
   - Implement client-side rate limiting
   - Consider upgrading user role for higher limits

### Debug Authentication

Enable debug logging to troubleshoot authentication issues:

```bash
# Set log level to DEBUG
export LOG_LEVEL=DEBUG

# Check authentication logs
tail -f /opt/camera-service/logs/camera-service.log | grep -i auth
```

## Monitoring and Auditing

### Authentication Metrics

Monitor authentication patterns:
- Successful vs failed authentications
- Token expiry patterns
- Rate limiting triggers
- Permission denied events

### Security Auditing

Implement security auditing:
- Log all authentication attempts
- Monitor for suspicious patterns
- Track API key usage
- Audit permission changes

## Version History

- **v1.0**: Initial implementation with JWT and API key authentication 