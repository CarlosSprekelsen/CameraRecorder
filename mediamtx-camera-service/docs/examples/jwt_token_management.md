# JWT Token Management Guide

**Version:** 1.0  
**Status:** Production Ready  
**Epic:** E3 Client API & SDK Ecosystem  

## Overview

This guide provides comprehensive information about JWT token management for the MediaMTX Camera Service, including token generation, validation, refresh, and best practices.

## JWT Token Structure

### Token Components

JWT tokens consist of three parts separated by dots:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlcjEyMyIsInJvbGUiOiJhZG1pbiIsImlhdCI6MTY0MDk5NTIwMCwiZXhwIjoxNjQxMDgxNjAwfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

1. **Header**: Algorithm and token type
2. **Payload**: Claims (user data)
3. **Signature**: Verification signature

### Header

```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### Payload Claims

```json
{
  "user_id": "user123",
  "role": "admin",
  "iat": 1640995200,
  "exp": 1641081600
}
```

| Claim | Description | Required | Example |
|-------|-------------|----------|---------|
| `user_id` | Unique user identifier | Yes | `"user123"` |
| `role` | User role/permissions | Yes | `"admin"`, `"operator"`, `"viewer"` |
| `iat` | Issued at timestamp | Yes | `1640995200` |
| `exp` | Expiration timestamp | Yes | `1641081600` |

## Token Generation

### Python Token Generation

```python
import jwt
import time
import os

def generate_jwt_token(user_id: str, role: str, secret_key: str, expiry_hours: int = 24) -> str:
    """
    Generate a JWT token for MediaMTX Camera Service authentication.
    
    Args:
        user_id: Unique user identifier
        role: User role (admin, operator, viewer)
        secret_key: Secret key for signing
        expiry_hours: Token expiry in hours (default: 24)
    
    Returns:
        JWT token string
    """
    payload = {
        "user_id": user_id,
        "role": role,
        "iat": int(time.time()),
        "exp": int(time.time()) + (expiry_hours * 60 * 60)
    }
    
    token = jwt.encode(payload, secret_key, algorithm="HS256")
    return token

# Example usage
secret_key = os.environ.get("CAMERA_SERVICE_JWT_SECRET", "dev-secret-change-me")
token = generate_jwt_token("user123", "admin", secret_key, 24)
print(f"Generated JWT token: {token}")
```

### Node.js Token Generation

```javascript
const jwt = require('jsonwebtoken');

function generateJwtToken(userId, role, secretKey, expiryHours = 24) {
    /**
     * Generate a JWT token for MediaMTX Camera Service authentication.
     * 
     * @param {string} userId - Unique user identifier
     * @param {string} role - User role (admin, operator, viewer)
     * @param {string} secretKey - Secret key for signing
     * @param {number} expiryHours - Token expiry in hours (default: 24)
     * @returns {string} JWT token string
     */
    const payload = {
        user_id: userId,
        role: role,
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + (expiryHours * 60 * 60)
    };
    
    const token = jwt.sign(payload, secretKey, { algorithm: 'HS256' });
    return token;
}

// Example usage
const secretKey = process.env.CAMERA_SERVICE_JWT_SECRET || 'dev-secret-change-me';
const token = generateJwtToken('user123', 'admin', secretKey, 24);
console.log(`Generated JWT token: ${token}`);
```

### Command Line Token Generation

```bash
#!/bin/bash
# Generate JWT token using Python

USER_ID="user123"
ROLE="admin"
SECRET_KEY="your-secret-key"
EXPIRY_HOURS=24

python3 -c "
import jwt
import time
import sys

payload = {
    'user_id': '$USER_ID',
    'role': '$ROLE',
    'iat': int(time.time()),
    'exp': int(time.time()) + ($EXPIRY_HOURS * 60 * 60)
}

token = jwt.encode(payload, '$SECRET_KEY', algorithm='HS256')
print(token)
"
```

## Token Validation

### Python Token Validation

```python
import jwt
import time
from typing import Optional, Dict, Any

def validate_jwt_token(token: str, secret_key: str) -> Optional[Dict[str, Any]]:
    """
    Validate a JWT token and extract claims.
    
    Args:
        token: JWT token string
        secret_key: Secret key for verification
    
    Returns:
        Token payload if valid, None if invalid
    """
    try:
        payload = jwt.decode(token, secret_key, algorithms=["HS256"])
        
        # Validate required claims
        required_fields = ["user_id", "role", "iat", "exp"]
        for field in required_fields:
            if field not in payload:
                print(f"Missing required field: {field}")
                return None
        
        # Validate role
        valid_roles = ["viewer", "operator", "admin"]
        if payload["role"] not in valid_roles:
            print(f"Invalid role: {payload['role']}")
            return None
        
        # Check expiration
        current_time = int(time.time())
        if payload["exp"] < current_time:
            print("Token has expired")
            return None
        
        return payload
        
    except jwt.ExpiredSignatureError:
        print("Token has expired")
        return None
    except jwt.InvalidTokenError as e:
        print(f"Invalid token: {e}")
        return None
    except Exception as e:
        print(f"Token validation error: {e}")
        return None

# Example usage
secret_key = "your-secret-key"
token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

payload = validate_jwt_token(token, secret_key)
if payload:
    print(f"Valid token for user: {payload['user_id']}")
    print(f"Role: {payload['role']}")
    print(f"Expires at: {payload['exp']}")
else:
    print("Invalid token")
```

### Node.js Token Validation

```javascript
const jwt = require('jsonwebtoken');

function validateJwtToken(token, secretKey) {
    /**
     * Validate a JWT token and extract claims.
     * 
     * @param {string} token - JWT token string
     * @param {string} secretKey - Secret key for verification
     * @returns {object|null} Token payload if valid, null if invalid
     */
    try {
        const payload = jwt.verify(token, secretKey, { algorithms: ['HS256'] });
        
        // Validate required claims
        const requiredFields = ['user_id', 'role', 'iat', 'exp'];
        for (const field of requiredFields) {
            if (!(field in payload)) {
                console.log(`Missing required field: ${field}`);
                return null;
            }
        }
        
        // Validate role
        const validRoles = ['viewer', 'operator', 'admin'];
        if (!validRoles.includes(payload.role)) {
            console.log(`Invalid role: ${payload.role}`);
            return null;
        }
        
        return payload;
        
    } catch (error) {
        if (error.name === 'TokenExpiredError') {
            console.log('Token has expired');
        } else if (error.name === 'JsonWebTokenError') {
            console.log(`Invalid token: ${error.message}`);
        } else {
            console.log(`Token validation error: ${error.message}`);
        }
        return null;
    }
}

// Example usage
const secretKey = 'your-secret-key';
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';

const payload = validateJwtToken(token, secretKey);
if (payload) {
    console.log(`Valid token for user: ${payload.user_id}`);
    console.log(`Role: ${payload.role}`);
    console.log(`Expires at: ${payload.exp}`);
} else {
    console.log('Invalid token');
}
```

## Token Refresh

### Python Token Refresh

```python
import jwt
import time
from typing import Optional

def refresh_jwt_token(current_token: str, secret_key: str, expiry_hours: int = 24) -> Optional[str]:
    """
    Refresh a JWT token before expiration.
    
    Args:
        current_token: Current JWT token
        secret_key: Secret key for signing
        expiry_hours: New token expiry in hours
    
    Returns:
        New JWT token if refresh needed, None if current token is still valid
    """
    try:
        # Decode current token without verification to extract claims
        payload = jwt.decode(current_token, options={"verify_signature": False})
        
        # Check if token expires soon (within 1 hour)
        current_time = int(time.time())
        expires_in = payload["exp"] - current_time
        
        if expires_in < 3600:  # Less than 1 hour remaining
            # Generate new token with same claims but new expiry
            new_payload = {
                "user_id": payload["user_id"],
                "role": payload["role"],
                "iat": current_time,
                "exp": current_time + (expiry_hours * 60 * 60)
            }
            
            new_token = jwt.encode(new_payload, secret_key, algorithm="HS256")
            return new_token
        else:
            print(f"Token still valid for {expires_in} seconds")
            return None
            
    except jwt.InvalidTokenError:
        print("Invalid token for refresh")
        return None
    except Exception as e:
        print(f"Token refresh error: {e}")
        return None

# Example usage
secret_key = "your-secret-key"
current_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

new_token = refresh_jwt_token(current_token, secret_key, 24)
if new_token:
    print(f"New token generated: {new_token}")
else:
    print("Token refresh not needed")
```

### Node.js Token Refresh

```javascript
const jwt = require('jsonwebtoken');

function refreshJwtToken(currentToken, secretKey, expiryHours = 24) {
    /**
     * Refresh a JWT token before expiration.
     * 
     * @param {string} currentToken - Current JWT token
     * @param {string} secretKey - Secret key for signing
     * @param {number} expiryHours - New token expiry in hours
     * @returns {string|null} New JWT token if refresh needed, null if current token is still valid
     */
    try {
        // Decode current token without verification to extract claims
        const payload = jwt.decode(currentToken, { complete: false });
        
        // Check if token expires soon (within 1 hour)
        const currentTime = Math.floor(Date.now() / 1000);
        const expiresIn = payload.exp - currentTime;
        
        if (expiresIn < 3600) { // Less than 1 hour remaining
            // Generate new token with same claims but new expiry
            const newPayload = {
                user_id: payload.user_id,
                role: payload.role,
                iat: currentTime,
                exp: currentTime + (expiryHours * 60 * 60)
            };
            
            const newToken = jwt.sign(newPayload, secretKey, { algorithm: 'HS256' });
            return newToken;
        } else {
            console.log(`Token still valid for ${expiresIn} seconds`);
            return null;
        }
        
    } catch (error) {
        console.log(`Token refresh error: ${error.message}`);
        return null;
    }
}

// Example usage
const secretKey = 'your-secret-key';
const currentToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';

const newToken = refreshJwtToken(currentToken, secretKey, 24);
if (newToken) {
    console.log(`New token generated: ${newToken}`);
} else {
    console.log('Token refresh not needed');
}
```

## Role-Based Access Control

### Role Hierarchy

```
viewer (1) < operator (2) < admin (3)
```

### Role Permissions

| Role | Permissions |
|------|-------------|
| **viewer** | Read-only access to camera status and streams |
| **operator** | Viewer permissions + camera control (snapshots, recording) |
| **admin** | Full access to all features |

### Role Validation

```python
def validate_user_role(token_payload: Dict[str, Any], required_role: str) -> bool:
    """
    Validate if user has required role for operation.
    
    Args:
        token_payload: JWT token payload
        required_role: Required role for operation
    
    Returns:
        True if user has required role, False otherwise
    """
    role_hierarchy = {
        "viewer": 1,
        "operator": 2,
        "admin": 3
    }
    
    user_role = token_payload.get("role", "viewer")
    user_level = role_hierarchy.get(user_role, 0)
    required_level = role_hierarchy.get(required_role, 0)
    
    return user_level >= required_level

# Example usage
payload = {
    "user_id": "user123",
    "role": "operator",
    "iat": 1640995200,
    "exp": 1641081600
}

# Check if user can perform admin operations
if validate_user_role(payload, "admin"):
    print("User can perform admin operations")
else:
    print("User cannot perform admin operations")

# Check if user can perform operator operations
if validate_user_role(payload, "operator"):
    print("User can perform operator operations")
else:
    print("User cannot perform operator operations")
```

## Token Security Best Practices

### 1. Secure Secret Key Generation

```bash
# Generate secure secret key
openssl rand -base64 32

# Or using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

### 2. Environment Variable Storage

```python
import os

# Store secret key in environment variable
secret_key = os.environ.get("CAMERA_SERVICE_JWT_SECRET")
if not secret_key:
    raise ValueError("CAMERA_SERVICE_JWT_SECRET environment variable not set")
```

### 3. Token Expiration Management

```python
import time
from datetime import datetime, timedelta

def create_token_with_expiry(user_id: str, role: str, secret_key: str, 
                           expiry_hours: int = 24) -> str:
    """Create token with explicit expiry time."""
    now = datetime.utcnow()
    expiry = now + timedelta(hours=expiry_hours)
    
    payload = {
        "user_id": user_id,
        "role": role,
        "iat": int(now.timestamp()),
        "exp": int(expiry.timestamp())
    }
    
    return jwt.encode(payload, secret_key, algorithm="HS256")

def is_token_expiring_soon(token: str, hours_threshold: int = 1) -> bool:
    """Check if token expires within specified hours."""
    try:
        payload = jwt.decode(token, options={"verify_signature": False})
        current_time = int(time.time())
        expires_in = payload["exp"] - current_time
        return expires_in < (hours_threshold * 3600)
    except:
        return True
```

### 4. Token Rotation

```python
class TokenManager:
    def __init__(self, secret_key: str):
        self.secret_key = secret_key
        self.current_token = None
    
    def get_valid_token(self, user_id: str, role: str) -> str:
        """Get a valid token, refreshing if necessary."""
        if not self.current_token:
            self.current_token = self.generate_token(user_id, role)
            return self.current_token
        
        # Check if current token needs refresh
        if is_token_expiring_soon(self.current_token):
            self.current_token = self.generate_token(user_id, role)
        
        return self.current_token
    
    def generate_token(self, user_id: str, role: str) -> str:
        """Generate a new token."""
        return create_token_with_expiry(user_id, role, self.secret_key, 24)

# Example usage
token_manager = TokenManager("your-secret-key")
token = token_manager.get_valid_token("user123", "admin")
```

## Error Handling

### Common JWT Errors

```python
import jwt

def handle_jwt_errors(token: str, secret_key: str):
    """Handle common JWT errors with proper error messages."""
    try:
        payload = jwt.decode(token, secret_key, algorithms=["HS256"])
        return {"valid": True, "payload": payload}
        
    except jwt.ExpiredSignatureError:
        return {"valid": False, "error": "Token has expired"}
        
    except jwt.InvalidTokenError as e:
        return {"valid": False, "error": f"Invalid token: {e}"}
        
    except jwt.InvalidSignatureError:
        return {"valid": False, "error": "Invalid token signature"}
        
    except jwt.DecodeError:
        return {"valid": False, "error": "Token could not be decoded"}
        
    except Exception as e:
        return {"valid": False, "error": f"Unexpected error: {e}"}

# Example usage
result = handle_jwt_errors(token, secret_key)
if result["valid"]:
    print("Token is valid")
    print(f"User: {result['payload']['user_id']}")
else:
    print(f"Token error: {result['error']}")
```

## Testing JWT Tokens

### Test Token Generation

```python
def test_token_generation():
    """Test JWT token generation and validation."""
    secret_key = "test-secret-key"
    user_id = "test_user"
    role = "admin"
    
    # Generate token
    token = generate_jwt_token(user_id, role, secret_key, 1)
    print(f"Generated token: {token}")
    
    # Validate token
    payload = validate_jwt_token(token, secret_key)
    if payload:
        print("✅ Token validation successful")
        print(f"User ID: {payload['user_id']}")
        print(f"Role: {payload['role']}")
    else:
        print("❌ Token validation failed")
    
    # Test role validation
    if validate_user_role(payload, "admin"):
        print("✅ Admin role validation successful")
    else:
        print("❌ Admin role validation failed")

if __name__ == "__main__":
    test_token_generation()
```

### Token Decoding (Debug)

```bash
# Decode JWT token header
echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" | base64 -d | jq .

# Decode JWT token payload
echo "eyJ1c2VyX2lkIjoidXNlcjEyMyIsInJvbGUiOiJhZG1pbiIsImlhdCI6MTY0MDk5NTIwMCwiZXhwIjoxNjQxMDgxNjAwfQ" | base64 -d | jq .

# Full token decode (without signature verification)
python3 -c "
import jwt
token = 'your-jwt-token-here'
payload = jwt.decode(token, options={'verify_signature': False})
print('Token payload:')
for key, value in payload.items():
    print(f'  {key}: {value}')
"
```

## Integration Examples

### With Camera Client

```python
import asyncio
from examples.python.camera_client import CameraClient

async def main():
    # Generate JWT token
    secret_key = "your-secret-key"
    token = generate_jwt_token("user123", "admin", secret_key, 24)
    
    # Create client with JWT authentication
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token=token
    )
    
    try:
        await client.connect()
        cameras = await client.get_camera_list()
        print(f"Found {len(cameras)} cameras")
    finally:
        await client.disconnect()

asyncio.run(main())
```

### With SDK

```python
import asyncio
from mediamtx_camera_sdk import CameraClient

async def main():
    # Generate JWT token
    secret_key = "your-secret-key"
    token = generate_jwt_token("user123", "admin", secret_key, 24)
    
    # Use SDK with JWT authentication
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token=token
    )
    
    await client.connect()
    cameras = await client.get_camera_list()
    await client.disconnect()

asyncio.run(main())
```

## Support

For additional support and questions:

- **Documentation**: See `docs/security/CLIENT_AUTHENTICATION_GUIDE.md`
- **Examples**: Check `examples/` directory for working examples
- **Issues**: Report problems via GitHub Issues
- **Email**: Contact team@mediamtx-camera-service.com 