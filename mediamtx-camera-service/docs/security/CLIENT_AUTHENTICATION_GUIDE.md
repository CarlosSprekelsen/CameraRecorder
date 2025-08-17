# Client Authentication Guide - MediaMTX Camera Service

**Version:** 2.0  
**Status:** Updated for E3 Sprint  
**Epic:** E3 Client API & SDK Ecosystem  

## Overview

This guide provides comprehensive information about authenticating with the MediaMTX Camera Service using JWT tokens and API keys. It covers setup, implementation, best practices, and troubleshooting for all client types.

## Authentication Methods

### JWT Token Authentication

JWT (JSON Web Token) authentication provides secure, stateless authentication for clients.

#### How JWT Authentication Works

1. **Token Generation**: Server generates JWT tokens with user claims
2. **Token Transmission**: Client sends token in JSON-RPC authenticate method
3. **Token Validation**: Server validates token signature and claims
4. **Access Control**: Server grants access based on token validity

#### JWT Token Structure

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "user123",
    "role": "admin",
    "iat": 1640995200,
    "exp": 1641081600
  },
  "signature": "HMACSHA256(base64UrlEncode(header) + '.' + base64UrlEncode(payload), secret)"
}
```

#### JWT Claims

| Claim | Description | Required |
|-------|-------------|----------|
| `user_id` | Subject (user ID) | Yes |
| `role` | User role/permissions | Yes |
| `iat` | Issued at timestamp | Yes |
| `exp` | Expiration timestamp | Yes |

### API Key Authentication

API key authentication provides simple, direct access for automated clients.

#### How API Key Authentication Works

1. **Key Generation**: Server generates unique API keys
2. **Key Transmission**: Client sends key in JSON-RPC authenticate method
3. **Key Validation**: Server validates key against stored keys
4. **Access Control**: Server grants access based on key permissions

#### API Key Structure

```
camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
```

## Client Implementation

### Python Client

#### JWT Authentication

```python
import asyncio
from examples.python.camera_client import CameraClient

async def main():
    # Create client with JWT authentication
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    )
    
    try:
        # Connect and authenticate automatically
        await client.connect()
        
        # Use the client
        cameras = await client.get_camera_list()
        print(f"Found {len(cameras)} cameras")
        
    finally:
        await client.disconnect()

asyncio.run(main())
```

#### API Key Authentication

```python
import asyncio
from examples.python.camera_client import CameraClient

async def main():
    # Create client with API key authentication
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="api_key",
        api_key="camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    )
    
    try:
        # Connect and authenticate automatically
        await client.connect()
        
        # Use the client
        cameras = await client.get_camera_list()
        print(f"Found {len(cameras)} cameras")
        
    finally:
        await client.disconnect()

asyncio.run(main())
```

### JavaScript Client

#### JWT Authentication

```javascript
import { CameraClient } from './examples/javascript/camera_client.js';

async function main() {
    // Create client with JWT authentication
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'jwt',
        authToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
    });
    
    try {
        // Connect and authenticate automatically
        await client.connect();
        
        // Use the client
        const cameras = await client.getCameraList();
        console.log(`Found ${cameras.length} cameras`);
        
    } finally {
        await client.disconnect();
    }
}

main().catch(console.error);
```

#### API Key Authentication

```javascript
import { CameraClient } from './examples/javascript/camera_client.js';

async function main() {
    // Create client with API key authentication
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'api_key',
        apiKey: 'camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef'
    });
    
    try {
        // Connect and authenticate automatically
        await client.connect();
        
        // Use the client
        const cameras = await client.getCameraList();
        console.log(`Found ${cameras.length} cameras`);
        
    } finally {
        await client.disconnect();
    }
}

main().catch(console.error);
```

### Browser Client

#### JWT Authentication

```html
<!DOCTYPE html>
<html>
<head>
    <title>MediaMTX Camera Service</title>
</head>
<body>
    <div id="connection-panel">
        <input type="text" id="host" placeholder="Host" value="localhost">
        <input type="number" id="port" placeholder="Port" value="8080">
        <select id="authType">
            <option value="jwt">JWT Token</option>
            <option value="api_key">API Key</option>
        </select>
        <input type="text" id="authToken" placeholder="Authentication Token">
        <button onclick="connect()">Connect</button>
    </div>
    
    <script>
        let websocket;
        
        async function connect() {
            const host = document.getElementById('host').value;
            const port = document.getElementById('port').value;
            const authType = document.getElementById('authType').value;
            const authToken = document.getElementById('authToken').value;
            
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${host}:${port}/ws`;
            
            websocket = new WebSocket(wsUrl);
            
            websocket.onopen = async function() {
                console.log('Connected to camera service');
                
                // Authenticate
                const response = await sendRequest('authenticate', {
                    token: authToken,
                    auth_type: authType
                });
                
                if (response.authenticated) {
                    console.log('Authentication successful');
                    // Use the service
                    const cameras = await sendRequest('get_camera_list');
                    console.log('Cameras:', cameras);
                } else {
                    console.error('Authentication failed:', response.error);
                }
            };
            
            websocket.onmessage = function(event) {
                const data = JSON.parse(event.data);
                console.log('Received:', data);
            };
        }
        
        function sendRequest(method, params = {}) {
            return new Promise((resolve, reject) => {
                const id = Math.floor(Math.random() * 1000000);
                const request = {
                    jsonrpc: '2.0',
                    id: id,
                    method: method,
                    params: params
                };
                
                websocket.send(JSON.stringify(request));
                
                // Handle response
                const originalOnMessage = websocket.onmessage;
                websocket.onmessage = function(event) {
                    const data = JSON.parse(event.data);
                    if (data.id === id) {
                        websocket.onmessage = originalOnMessage;
                        if (data.error) {
                            reject(new Error(data.error.message));
                        } else {
                            resolve(data.result);
                        }
                    }
                };
            });
        }
    </script>
</body>
</html>
```

### CLI Client

#### JWT Authentication

```bash
# List cameras with JWT authentication
python examples/cli/camera_cli.py \
    --host localhost \
    --port 8080 \
    --auth-type jwt \
    --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
    list

# Take snapshot with JWT authentication
python examples/cli/camera_cli.py \
    --host localhost \
    --port 8080 \
    --auth-type jwt \
    --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
    snapshot /dev/video0
```

#### API Key Authentication

```bash
# List cameras with API key authentication
python examples/cli/camera_cli.py \
    --host localhost \
    --port 8080 \
    --auth-type api_key \
    --key "camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" \
    list

# Take snapshot with API key authentication
python examples/cli/camera_cli.py \
    --host localhost \
    --port 8080 \
    --auth-type api_key \
    --key "camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" \
    snapshot /dev/video0
```

## SDK Usage

### Python SDK

```python
# Install the SDK
cd sdk/python
pip install -e .

# Use the SDK
import asyncio
from mediamtx_camera_sdk import CameraClient

async def main():
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="your_jwt_token_here"
    )
    
    await client.connect()
    cameras = await client.get_camera_list()
    await client.disconnect()

asyncio.run(main())
```

### JavaScript SDK

```javascript
// Install the SDK
cd sdk/javascript
npm install
npm run build

// Use the SDK
import { CameraClient } from './dist/index.js';

const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    authType: 'jwt',
    authToken: 'your_jwt_token_here'
});

client.connect()
    .then(() => client.getCameraList())
    .then(cameras => console.log('Cameras:', cameras))
    .catch(console.error);
```

## Token Generation

### JWT Token Generation

JWT tokens are generated by the server. Here are examples of how to generate them:

#### Using Python

```python
import jwt
import time

# Generate JWT token
secret_key = "your-secret-key"
payload = {
    "user_id": "user123",
    "role": "admin",
    "iat": int(time.time()),
    "exp": int(time.time()) + (24 * 60 * 60)  # 24 hours
}

token = jwt.encode(payload, secret_key, algorithm="HS256")
print(f"JWT Token: {token}")
```

#### Using Node.js

```javascript
const jwt = require('jsonwebtoken');

// Generate JWT token
const secretKey = 'your-secret-key';
const payload = {
    user_id: 'user123',
    role: 'admin',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24 hours
};

const token = jwt.sign(payload, secretKey, { algorithm: 'HS256' });
console.log(`JWT Token: ${token}`);
```

### API Key Generation

API keys are generated by the server. Here's how to generate them:

#### Using curl

```bash
# Generate API key
curl -X POST http://localhost:8002/api/keys \
    -H "Content-Type: application/json" \
    -d '{"name": "Service Key", "role": "operator"}'

# Response will include the API key (only shown once)
{
    "key_id": "key_1234567890abcdef",
    "api_key": "camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
    "name": "Service Key",
    "role": "operator",
    "created_at": "2025-01-15T10:30:00Z"
}
```

## Authentication Flow

### Step-by-Step Process

1. **Client Connection**
   ```javascript
   // Connect to WebSocket
   const websocket = new WebSocket('ws://localhost:8080/ws');
   ```

2. **Authentication Request**
   ```javascript
   // Send authentication request
   const authRequest = {
       jsonrpc: '2.0',
       id: 1,
       method: 'authenticate',
       params: {
           token: 'your_jwt_token_or_api_key',
           auth_type: 'jwt' // or 'api_key'
       }
   };
   websocket.send(JSON.stringify(authRequest));
   ```

3. **Authentication Response**
   ```javascript
   // Handle authentication response
   websocket.onmessage = function(event) {
       const response = JSON.parse(event.data);
       if (response.id === 1) {
           if (response.result.authenticated) {
               console.log('Authentication successful');
               // Proceed with API calls
           } else {
               console.error('Authentication failed:', response.result.error);
           }
       }
   };
   ```

4. **API Usage**
   ```javascript
   // After successful authentication, use API methods
   const cameraRequest = {
       jsonrpc: '2.0',
       id: 2,
       method: 'get_camera_list',
       params: {}
   };
   websocket.send(JSON.stringify(cameraRequest));
   ```

## Error Handling

### Common Authentication Errors

| Error Code | Message | Description | Resolution |
|------------|---------|-------------|------------|
| -32001 | Authentication required | No auth token provided | Provide JWT token or API key |
| -32001 | Authentication failed | Invalid or expired token | Generate new token or check expiry |
| -32003 | Insufficient permissions | Role does not have required permissions | Use token with higher role |

### Error Handling Examples

#### Python

```python
from examples.python.camera_client import CameraClient, AuthenticationError, ConnectionError

async def main():
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="invalid_token"
    )
    
    try:
        await client.connect()
    except AuthenticationError as e:
        print(f"Authentication failed: {e}")
    except ConnectionError as e:
        print(f"Connection failed: {e}")
    except Exception as e:
        print(f"Unexpected error: {e}")

asyncio.run(main())
```

#### JavaScript

```javascript
import { CameraClient, AuthenticationError, ConnectionError } from './examples/javascript/camera_client.js';

async function main() {
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'jwt',
        authToken: 'invalid_token'
    });
    
    try {
        await client.connect();
    } catch (error) {
        if (error instanceof AuthenticationError) {
            console.error(`Authentication failed: ${error.message}`);
        } else if (error instanceof ConnectionError) {
            console.error(`Connection failed: ${error.message}`);
        } else {
            console.error(`Unexpected error: ${error.message}`);
        }
    }
}

main();
```

## Security Best Practices

### JWT Security

1. **Use Strong Secret Keys**
   ```bash
   # Generate secure secret key
   openssl rand -base64 32
   ```

2. **Set Appropriate Expiry**
   ```python
   # Set reasonable expiry time
   exp = int(time.time()) + (24 * 60 * 60)  # 24 hours
   ```

3. **Validate Token Claims**
   ```python
   # Always validate token claims
   if payload.get('exp', 0) < time.time():
       raise AuthenticationError("Token expired")
   ```

### API Key Security

1. **Secure Key Storage**
   ```bash
   # Set secure permissions on API key file
   chmod 600 /etc/camera-service/api-keys.json
   chown camera-service:camera-service /etc/camera-service/api-keys.json
   ```

2. **Key Rotation**
   ```bash
   # Rotate API keys regularly
   curl -X POST http://localhost:8002/api/keys/{key_id}/rotate
   ```

3. **Access Control**
   ```bash
   # Use different keys for different services
   # Monitor key usage and revoke unused keys
   curl -X DELETE http://localhost:8002/api/keys/{key_id}
   ```

### Network Security

1. **Use HTTPS/WSS**
   ```javascript
   // Use secure WebSocket connection
   const websocket = new WebSocket('wss://localhost:8080/ws');
   ```

2. **Network Policies**
   ```bash
   # Restrict access to authentication endpoints
   # Use firewall rules to limit connections
   ```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if the camera service is running
   - Verify host and port configuration
   - Check firewall settings

2. **Authentication Failed**
   - Verify token format and content
   - Check token expiry
   - Ensure correct auth_type is specified

3. **Permission Denied**
   - Check user role permissions
   - Verify token contains correct role
   - Contact administrator for role assignment

### Debug Mode

Enable debug logging to troubleshoot authentication issues:

#### Python

```python
import logging
logging.basicConfig(level=logging.DEBUG)

client = CameraClient(
    host="localhost",
    port=8080,
    auth_type="jwt",
    auth_token="your_token"
)
```

#### JavaScript

```javascript
// Enable debug logging
const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    authType: 'jwt',
    authToken: 'your_token'
});

// Debug WebSocket messages
websocket.onmessage = function(event) {
    console.log('Received:', event.data);
};
```

## Support

For additional support and questions:

- **Documentation**: See `docs/api/` for API reference
- **Examples**: Check `examples/` directory for working examples
- **Issues**: Report problems via GitHub Issues
- **Email**: Contact team@mediamtx-camera-service.com 