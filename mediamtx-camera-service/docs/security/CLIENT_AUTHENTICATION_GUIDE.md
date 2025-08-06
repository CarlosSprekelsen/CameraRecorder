# Client Authentication Guide - MediaMTX Camera Service

## Overview

This guide provides comprehensive information about authenticating with the MediaMTX Camera Service using JWT tokens and API keys. It covers setup, implementation, best practices, and troubleshooting for all client types.

## Authentication Methods

### JWT Token Authentication

JWT (JSON Web Token) authentication provides secure, stateless authentication for clients.

#### How JWT Authentication Works

1. **Token Generation**: Server generates JWT tokens with user claims
2. **Token Transmission**: Client sends token in Authorization header
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
    "sub": "user123",
    "name": "John Doe",
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
| `sub` | Subject (user ID) | Yes |
| `name` | User display name | No |
| `role` | User role/permissions | Yes |
| `iat` | Issued at timestamp | Yes |
| `exp` | Expiration timestamp | Yes |

### API Key Authentication

API key authentication provides simple, direct access for automated clients.

#### How API Key Authentication Works

1. **Key Generation**: Server generates unique API keys
2. **Key Transmission**: Client sends key in X-API-Key header
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
from examples.python.camera_client import CameraClient

# Create client with JWT authentication
client = CameraClient(
    host="localhost",
    port=8080,
    auth_type="jwt",
    auth_token="your_jwt_token_here"
)

# Connect and use
await client.connect()
cameras = await client.get_camera_list()
```

#### API Key Authentication

```python
# Create client with API key authentication
client = CameraClient(
    host="localhost",
    port=8080,
    auth_type="api_key",
    api_key="your_api_key_here"
)

# Connect and use
await client.connect()
cameras = await client.get_camera_list()
```

### JavaScript/Node.js Client

#### JWT Authentication

```javascript
const { CameraClient } = require('./examples/javascript/camera_client.js');

// Create client with JWT authentication
const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    authType: 'jwt',
    authToken: 'your_jwt_token_here'
});

// Connect and use
await client.connect();
const cameras = await client.getCameraList();
```

#### API Key Authentication

```javascript
// Create client with API key authentication
const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    authType: 'api_key',
    apiKey: 'your_api_key_here'
});

// Connect and use
await client.connect();
const cameras = await client.getCameraList();
```

### Browser Client

#### JWT Authentication

```javascript
// Browser client with JWT authentication
const client = new CameraServiceClient();

// Set authentication in connection form
document.getElementById('authType').value = 'jwt';
document.getElementById('authToken').value = 'your_jwt_token_here';

// Connect
await client.connect();
```

#### API Key Authentication

```javascript
// Browser client with API key authentication
const client = new CameraServiceClient();

// Set authentication in connection form
document.getElementById('authType').value = 'api_key';
document.getElementById('authToken').value = 'your_api_key_here';

// Connect
await client.connect();
```

### CLI Tool

#### JWT Authentication

```bash
# Use JWT token authentication
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list
```

#### API Key Authentication

```bash
# Use API key authentication
python camera_cli.py --host localhost --port 8080 --auth-type api_key --key your_api_key list
```

## Authentication Headers

### JWT Token Headers

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwicm9sZSI6ImFkbWluIiwiaWF0IjoxNjQwOTk1MjAwLCJleHAiOjE2NDEwODE2MDB9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

### API Key Headers

```http
X-API-Key: camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
```

## Token Management

### JWT Token Lifecycle

#### 1. Token Generation

```python
import jwt
import datetime

# Generate JWT token
payload = {
    'sub': 'user123',
    'name': 'John Doe',
    'role': 'admin',
    'iat': datetime.datetime.utcnow(),
    'exp': datetime.datetime.utcnow() + datetime.timedelta(hours=24)
}

token = jwt.encode(payload, 'your_secret_key', algorithm='HS256')
print(f"JWT Token: {token}")
```

#### 2. Token Validation

```python
import jwt
from datetime import datetime

# Validate JWT token
try:
    payload = jwt.decode(token, 'your_secret_key', algorithms=['HS256'])
    print(f"Token valid for user: {payload['sub']}")
    print(f"Expires at: {datetime.fromtimestamp(payload['exp'])}")
except jwt.ExpiredSignatureError:
    print("Token has expired")
except jwt.InvalidTokenError:
    print("Invalid token")
```

#### 3. Token Refresh

```python
import jwt
import datetime

# Refresh JWT token before expiration
def refresh_token(current_token):
    try:
        payload = jwt.decode(current_token, 'your_secret_key', algorithms=['HS256'])
        
        # Check if token expires soon (within 1 hour)
        if payload['exp'] - datetime.datetime.utcnow().timestamp() < 3600:
            # Generate new token
            new_payload = {
                'sub': payload['sub'],
                'name': payload['name'],
                'role': payload['role'],
                'iat': datetime.datetime.utcnow(),
                'exp': datetime.datetime.utcnow() + datetime.timedelta(hours=24)
            }
            return jwt.encode(new_payload, 'your_secret_key', algorithm='HS256')
        
        return current_token
    except jwt.InvalidTokenError:
        raise Exception("Invalid token")
```

### API Key Management

#### 1. Key Generation

```python
import secrets
import string

# Generate secure API key
def generate_api_key():
    alphabet = string.ascii_letters + string.digits
    key = 'camera_service_sk_' + ''.join(secrets.choice(alphabet) for _ in range(64))
    return key

api_key = generate_api_key()
print(f"API Key: {api_key}")
```

#### 2. Key Validation

```python
# Validate API key format
def validate_api_key(key):
    if not key.startswith('camera_service_sk_'):
        return False
    
    if len(key) != 80:  # 16 (prefix) + 64 (key)
        return False
    
    return True

# Usage
if validate_api_key(api_key):
    print("Valid API key format")
else:
    print("Invalid API key format")
```

## Security Best Practices

### JWT Token Security

#### 1. Secure Token Storage

```python
# Store tokens securely (not in code)
import os

# Use environment variables
JWT_TOKEN = os.environ.get('CAMERA_SERVICE_JWT_TOKEN')

# Or use secure storage
import keyring
JWT_TOKEN = keyring.get_password("camera_service", "jwt_token")
```

#### 2. Token Expiration

```python
# Always check token expiration
import jwt
from datetime import datetime

def is_token_expired(token):
    try:
        payload = jwt.decode(token, 'your_secret_key', algorithms=['HS256'])
        return datetime.fromtimestamp(payload['exp']) < datetime.utcnow()
    except:
        return True

# Usage
if is_token_expired(JWT_TOKEN):
    print("Token expired, need to refresh")
```

#### 3. Token Rotation

```python
# Implement token rotation
def rotate_token(client):
    try:
        # Request new token from server
        new_token = await client.refresh_token()
        
        # Update stored token
        os.environ['CAMERA_SERVICE_JWT_TOKEN'] = new_token
        
        return new_token
    except Exception as e:
        print(f"Token rotation failed: {e}")
        return None
```

### API Key Security

#### 1. Secure Key Storage

```python
# Store API keys securely
import os

# Use environment variables
API_KEY = os.environ.get('CAMERA_SERVICE_API_KEY')

# Or use secure storage
import keyring
API_KEY = keyring.get_password("camera_service", "api_key")
```

#### 2. Key Rotation

```python
# Implement API key rotation
def rotate_api_key(client):
    try:
        # Request new API key from server
        new_key = await client.rotate_api_key()
        
        # Update stored key
        os.environ['CAMERA_SERVICE_API_KEY'] = new_key
        
        return new_key
    except Exception as e:
        print(f"API key rotation failed: {e}")
        return None
```

#### 3. Key Permissions

```python
# Implement key permission checking
def check_key_permissions(api_key):
    # Check if key has required permissions
    permissions = get_key_permissions(api_key)
    
    required_permissions = ['camera:read', 'camera:write']
    
    for permission in required_permissions:
        if permission not in permissions:
            return False
    
    return True
```

## Error Handling

### Authentication Errors

#### 1. Invalid Token

```python
try:
    await client.connect()
except AuthenticationError as e:
    print(f"Authentication failed: {e}")
    # Handle invalid token
    if "invalid token" in str(e).lower():
        print("Token is invalid, please check your credentials")
    elif "expired" in str(e).lower():
        print("Token has expired, please refresh your token")
```

#### 2. Missing Credentials

```python
def validate_auth_config(auth_type, token=None, api_key=None):
    if auth_type == 'jwt' and not token:
        raise ValueError("JWT token required for jwt authentication")
    elif auth_type == 'api_key' and not api_key:
        raise ValueError("API key required for api_key authentication")
    
    return True

# Usage
try:
    validate_auth_config('jwt', token=JWT_TOKEN)
except ValueError as e:
    print(f"Authentication configuration error: {e}")
```

#### 3. Connection Errors

```python
try:
    await client.connect()
except ConnectionError as e:
    print(f"Connection failed: {e}")
    # Retry with exponential backoff
    await retry_connection(client, max_retries=3)
```

### Retry Logic

```python
import asyncio
import random

async def retry_connection(client, max_retries=3):
    for attempt in range(max_retries):
        try:
            await client.connect()
            print("Connection successful")
            return
        except Exception as e:
            if attempt < max_retries - 1:
                delay = (2 ** attempt) + random.uniform(0, 1)
                print(f"Connection failed, retrying in {delay:.2f} seconds...")
                await asyncio.sleep(delay)
            else:
                print(f"Connection failed after {max_retries} attempts")
                raise e
```

## Troubleshooting

### Common Authentication Issues

#### 1. "Invalid Token" Error

**Symptoms:**
- Authentication fails with "invalid token" message
- Connection refused after authentication attempt

**Solutions:**
```bash
# Check token format
echo $JWT_TOKEN | cut -d'.' -f2 | base64 -d | jq .

# Verify token expiration
python -c "import jwt; print(jwt.decode('$JWT_TOKEN', options={'verify_signature': False}))"
```

#### 2. "Token Expired" Error

**Symptoms:**
- Authentication fails with "token expired" message
- Previously working token suddenly fails

**Solutions:**
```python
# Refresh token
new_token = await client.refresh_token()
print(f"New token: {new_token}")

# Update environment variable
import os
os.environ['CAMERA_SERVICE_JWT_TOKEN'] = new_token
```

#### 3. "API Key Invalid" Error

**Symptoms:**
- Authentication fails with "invalid API key" message
- API key format appears correct

**Solutions:**
```bash
# Check API key format
echo $API_KEY | grep -E "^camera_service_sk_[a-zA-Z0-9]{64}$"

# Verify API key with server
curl -H "X-API-Key: $API_KEY" http://localhost:8080/health
```

#### 4. "Connection Refused" Error

**Symptoms:**
- Connection fails before authentication
- Network connectivity issues

**Solutions:**
```bash
# Check server status
curl http://localhost:8080/health

# Check network connectivity
telnet localhost 8080

# Check firewall settings
sudo ufw status
```

### Debug Authentication

#### 1. Enable Verbose Logging

```python
import logging

# Enable debug logging
logging.basicConfig(level=logging.DEBUG)

# Create client with debug info
client = CameraClient(
    host="localhost",
    port=8080,
    auth_type="jwt",
    auth_token="your_token"
)
```

#### 2. Test Authentication Separately

```python
# Test authentication without full connection
async def test_auth():
    try:
        # Test basic connectivity
        await client.ping()
        print("✅ Authentication successful")
    except AuthenticationError as e:
        print(f"❌ Authentication failed: {e}")
    except Exception as e:
        print(f"❌ Connection error: {e}")
```

#### 3. Validate Token Manually

```python
import jwt

def validate_token_manually(token):
    try:
        # Decode without verification (for debugging)
        payload = jwt.decode(token, options={'verify_signature': False})
        print(f"Token payload: {payload}")
        
        # Check expiration
        import datetime
        exp_time = datetime.datetime.fromtimestamp(payload['exp'])
        now = datetime.datetime.utcnow()
        
        if exp_time < now:
            print(f"❌ Token expired at {exp_time}")
        else:
            print(f"✅ Token valid until {exp_time}")
            
    except Exception as e:
        print(f"❌ Token validation failed: {e}")

# Usage
validate_token_manually(JWT_TOKEN)
```

## Best Practices Summary

### 1. Secure Token Storage
- Never hardcode tokens in source code
- Use environment variables or secure storage
- Implement token rotation

### 2. Error Handling
- Always handle authentication errors gracefully
- Implement retry logic with exponential backoff
- Provide clear error messages to users

### 3. Token Management
- Monitor token expiration
- Implement automatic token refresh
- Use short-lived tokens for security

### 4. API Key Security
- Generate strong, random API keys
- Implement key rotation policies
- Use least-privilege access

### 5. Network Security
- Use HTTPS/WSS for production
- Implement proper CORS policies
- Monitor authentication attempts

---

**Version:** 1.0  
**Last Updated:** 2025-08-06  
**Compatibility:** All client types (Python, JavaScript, Browser, CLI) 