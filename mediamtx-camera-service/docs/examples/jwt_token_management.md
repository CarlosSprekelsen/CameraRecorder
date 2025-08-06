# JWT Token Management Examples - MediaMTX Camera Service

## Overview

This guide provides comprehensive examples for managing JWT tokens with the MediaMTX Camera Service, including token generation, validation, refresh, and security best practices.

## Token Generation

### Basic JWT Token Generation

```python
import jwt
import datetime
import secrets

def generate_jwt_token(user_id, user_name, role, secret_key, expiration_hours=24):
    """
    Generate a JWT token for camera service authentication.
    
    Args:
        user_id: Unique user identifier
        user_name: User display name
        role: User role (admin, user, etc.)
        secret_key: Secret key for signing
        expiration_hours: Token expiration in hours
        
    Returns:
        JWT token string
    """
    payload = {
        'sub': user_id,
        'name': user_name,
        'role': role,
        'iat': datetime.datetime.utcnow(),
        'exp': datetime.datetime.utcnow() + datetime.timedelta(hours=expiration_hours),
        'jti': secrets.token_urlsafe(16)  # JWT ID for uniqueness
    }
    
    token = jwt.encode(payload, secret_key, algorithm='HS256')
    return token

# Usage example
secret_key = "your-secret-key-here"
token = generate_jwt_token(
    user_id="user123",
    user_name="John Doe",
    role="admin",
    secret_key=secret_key,
    expiration_hours=24
)
print(f"Generated JWT token: {token}")
```

### Advanced Token Generation with Claims

```python
import jwt
import datetime
import secrets
from typing import Dict, Any, Optional

def generate_advanced_jwt_token(
    user_id: str,
    user_name: str,
    role: str,
    secret_key: str,
    expiration_hours: int = 24,
    additional_claims: Optional[Dict[str, Any]] = None,
    issuer: str = "camera-service",
    audience: str = "camera-clients"
) -> str:
    """
    Generate an advanced JWT token with additional claims and metadata.
    
    Args:
        user_id: Unique user identifier
        user_name: User display name
        role: User role
        secret_key: Secret key for signing
        expiration_hours: Token expiration in hours
        additional_claims: Additional custom claims
        issuer: Token issuer
        audience: Token audience
        
    Returns:
        JWT token string
    """
    now = datetime.datetime.utcnow()
    
    payload = {
        # Standard JWT claims
        'sub': user_id,
        'name': user_name,
        'role': role,
        'iat': now,
        'exp': now + datetime.timedelta(hours=expiration_hours),
        'jti': secrets.token_urlsafe(16),
        'iss': issuer,
        'aud': audience,
        
        # Custom claims
        'permissions': get_permissions_for_role(role),
        'camera_access': get_camera_access_for_role(role),
        'session_id': secrets.token_urlsafe(16)
    }
    
    # Add additional claims if provided
    if additional_claims:
        payload.update(additional_claims)
    
    token = jwt.encode(payload, secret_key, algorithm='HS256')
    return token

def get_permissions_for_role(role: str) -> list:
    """Get permissions for a given role."""
    permissions = {
        'admin': ['camera:read', 'camera:write', 'camera:delete', 'system:admin'],
        'user': ['camera:read', 'camera:write'],
        'viewer': ['camera:read']
    }
    return permissions.get(role, [])

def get_camera_access_for_role(role: str) -> list:
    """Get camera access for a given role."""
    access = {
        'admin': ['*'],  # All cameras
        'user': ['camera1', 'camera2'],
        'viewer': ['camera1']
    }
    return access.get(role, [])

# Usage example
token = generate_advanced_jwt_token(
    user_id="user123",
    user_name="John Doe",
    role="admin",
    secret_key="your-secret-key",
    additional_claims={
        'department': 'IT',
        'location': 'HQ'
    }
)
print(f"Advanced JWT token: {token}")
```

## Token Validation

### Basic Token Validation

```python
import jwt
from datetime import datetime
from typing import Dict, Any, Optional

def validate_jwt_token(token: str, secret_key: str) -> Optional[Dict[str, Any]]:
    """
    Validate a JWT token and return the payload if valid.
    
    Args:
        token: JWT token string
        secret_key: Secret key for verification
        
    Returns:
        Token payload if valid, None if invalid
    """
    try:
        payload = jwt.decode(token, secret_key, algorithms=['HS256'])
        return payload
    except jwt.ExpiredSignatureError:
        print("❌ Token has expired")
        return None
    except jwt.InvalidTokenError as e:
        print(f"❌ Invalid token: {e}")
        return None
    except Exception as e:
        print(f"❌ Token validation error: {e}")
        return None

# Usage example
payload = validate_jwt_token(token, secret_key)
if payload:
    print(f"✅ Token valid for user: {payload['sub']}")
    print(f"   Role: {payload['role']}")
    print(f"   Expires: {datetime.fromtimestamp(payload['exp'])}")
```

### Advanced Token Validation with Custom Checks

```python
import jwt
from datetime import datetime
from typing import Dict, Any, Optional, Tuple

def validate_advanced_jwt_token(
    token: str,
    secret_key: str,
    required_role: Optional[str] = None,
    required_permissions: Optional[list] = None
) -> Tuple[bool, Optional[Dict[str, Any]], Optional[str]]:
    """
    Advanced JWT token validation with role and permission checks.
    
    Args:
        token: JWT token string
        secret_key: Secret key for verification
        required_role: Required role for access
        required_permissions: Required permissions for access
        
    Returns:
        Tuple of (is_valid, payload, error_message)
    """
    try:
        # Decode token
        payload = jwt.decode(token, secret_key, algorithms=['HS256'])
        
        # Check expiration
        if datetime.fromtimestamp(payload['exp']) < datetime.utcnow():
            return False, None, "Token has expired"
        
        # Check required role
        if required_role and payload.get('role') != required_role:
            return False, None, f"Required role '{required_role}' not found"
        
        # Check required permissions
        if required_permissions:
            user_permissions = payload.get('permissions', [])
            missing_permissions = [perm for perm in required_permissions if perm not in user_permissions]
            if missing_permissions:
                return False, None, f"Missing permissions: {missing_permissions}"
        
        return True, payload, None
        
    except jwt.ExpiredSignatureError:
        return False, None, "Token has expired"
    except jwt.InvalidTokenError as e:
        return False, None, f"Invalid token: {e}"
    except Exception as e:
        return False, None, f"Token validation error: {e}"

# Usage examples
# Basic validation
is_valid, payload, error = validate_advanced_jwt_token(token, secret_key)
if is_valid:
    print(f"✅ Token valid: {payload['sub']}")
else:
    print(f"❌ Token invalid: {error}")

# Role-based validation
is_valid, payload, error = validate_advanced_jwt_token(
    token, secret_key, required_role="admin"
)
if is_valid:
    print("✅ Admin access granted")
else:
    print(f"❌ Admin access denied: {error}")

# Permission-based validation
is_valid, payload, error = validate_advanced_jwt_token(
    token, secret_key, required_permissions=["camera:read", "camera:write"]
)
if is_valid:
    print("✅ Required permissions granted")
else:
    print(f"❌ Insufficient permissions: {error}")
```

## Token Refresh

### Automatic Token Refresh

```python
import jwt
import datetime
import asyncio
from typing import Optional

class JWTTokenManager:
    """Manages JWT token lifecycle including refresh."""
    
    def __init__(self, secret_key: str, refresh_threshold_minutes: int = 60):
        self.secret_key = secret_key
        self.refresh_threshold_minutes = refresh_threshold_minutes
        self.current_token = None
        self.token_payload = None
    
    def set_token(self, token: str) -> bool:
        """Set the current token and validate it."""
        try:
            payload = jwt.decode(token, self.secret_key, algorithms=['HS256'])
            self.current_token = token
            self.token_payload = payload
            return True
        except jwt.InvalidTokenError:
            return False
    
    def is_token_expired(self) -> bool:
        """Check if current token is expired."""
        if not self.token_payload:
            return True
        
        return datetime.fromtimestamp(self.token_payload['exp']) < datetime.utcnow()
    
    def is_token_expiring_soon(self) -> bool:
        """Check if token expires within the refresh threshold."""
        if not self.token_payload:
            return True
        
        exp_time = datetime.fromtimestamp(self.token_payload['exp'])
        threshold_time = datetime.utcnow() + datetime.timedelta(minutes=self.refresh_threshold_minutes)
        
        return exp_time < threshold_time
    
    def refresh_token(self) -> Optional[str]:
        """Refresh the current token if it's expiring soon."""
        if not self.token_payload:
            return None
        
        if not self.is_token_expiring_soon():
            return self.current_token
        
        try:
            # Create new token with same claims but new expiration
            new_payload = self.token_payload.copy()
            new_payload['iat'] = datetime.datetime.utcnow()
            new_payload['exp'] = datetime.datetime.utcnow() + datetime.timedelta(hours=24)
            new_payload['jti'] = secrets.token_urlsafe(16)  # New JWT ID
            
            new_token = jwt.encode(new_payload, self.secret_key, algorithm='HS256')
            
            # Update current token
            self.current_token = new_token
            self.token_payload = new_payload
            
            return new_token
            
        except Exception as e:
            print(f"Token refresh failed: {e}")
            return None
    
    def get_valid_token(self) -> Optional[str]:
        """Get a valid token, refreshing if necessary."""
        if self.is_token_expired():
            return None
        
        if self.is_token_expiring_soon():
            return self.refresh_token()
        
        return self.current_token

# Usage example
token_manager = JWTTokenManager(secret_key, refresh_threshold_minutes=60)
token_manager.set_token(token)

# Get valid token (refreshes if needed)
valid_token = token_manager.get_valid_token()
if valid_token:
    print(f"✅ Valid token: {valid_token[:50]}...")
else:
    print("❌ No valid token available")
```

### Async Token Refresh with Client Integration

```python
import asyncio
import jwt
import datetime
from typing import Optional

class AsyncJWTTokenManager:
    """Async JWT token manager with automatic refresh."""
    
    def __init__(self, client, secret_key: str, refresh_threshold_minutes: int = 60):
        self.client = client
        self.secret_key = secret_key
        self.refresh_threshold_minutes = refresh_threshold_minutes
        self.current_token = None
        self.token_payload = None
        self._refresh_task = None
    
    async def set_token(self, token: str) -> bool:
        """Set the current token and validate it."""
        try:
            payload = jwt.decode(token, self.secret_key, algorithms=['HS256'])
            self.current_token = token
            self.token_payload = payload
            return True
        except jwt.InvalidTokenError:
            return False
    
    async def refresh_token_from_server(self) -> Optional[str]:
        """Request a new token from the server."""
        try:
            # This would be a server endpoint for token refresh
            new_token = await self.client.refresh_token()
            
            if new_token:
                # Validate the new token
                payload = jwt.decode(new_token, self.secret_key, algorithms=['HS256'])
                self.current_token = new_token
                self.token_payload = payload
                return new_token
            
        except Exception as e:
            print(f"Server token refresh failed: {e}")
        
        return None
    
    async def ensure_valid_token(self) -> Optional[str]:
        """Ensure we have a valid token, refreshing if necessary."""
        if self.is_token_expired():
            return await self.refresh_token_from_server()
        
        if self.is_token_expiring_soon():
            return await self.refresh_token_from_server()
        
        return self.current_token
    
    def is_token_expired(self) -> bool:
        """Check if current token is expired."""
        if not self.token_payload:
            return True
        
        return datetime.fromtimestamp(self.token_payload['exp']) < datetime.utcnow()
    
    def is_token_expiring_soon(self) -> bool:
        """Check if token expires within the refresh threshold."""
        if not self.token_payload:
            return True
        
        exp_time = datetime.fromtimestamp(self.token_payload['exp'])
        threshold_time = datetime.utcnow() + datetime.timedelta(minutes=self.refresh_threshold_minutes)
        
        return exp_time < threshold_time
    
    async def start_auto_refresh(self):
        """Start automatic token refresh in background."""
        async def auto_refresh_loop():
            while True:
                try:
                    if self.is_token_expiring_soon():
                        await self.refresh_token_from_server()
                    
                    # Check every 5 minutes
                    await asyncio.sleep(300)
                    
                except Exception as e:
                    print(f"Auto refresh error: {e}")
                    await asyncio.sleep(60)  # Wait before retry
        
        self._refresh_task = asyncio.create_task(auto_refresh_loop())
    
    async def stop_auto_refresh(self):
        """Stop automatic token refresh."""
        if self._refresh_task:
            self._refresh_task.cancel()
            try:
                await self._refresh_task
            except asyncio.CancelledError:
                pass

# Usage example
async def main():
    # Create token manager
    token_manager = AsyncJWTTokenManager(client, secret_key)
    await token_manager.set_token(token)
    
    # Start auto refresh
    await token_manager.start_auto_refresh()
    
    try:
        # Use token manager with client
        valid_token = await token_manager.ensure_valid_token()
        if valid_token:
            # Update client with new token
            client.auth_token = valid_token
            print("✅ Client updated with valid token")
        
        # Continue with camera operations...
        
    finally:
        await token_manager.stop_auto_refresh()

# Run the example
# asyncio.run(main())
```

## Token Security

### Secure Token Storage

```python
import os
import keyring
import json
from typing import Optional

class SecureTokenStorage:
    """Secure token storage using environment variables and keyring."""
    
    def __init__(self, service_name: str = "camera-service"):
        self.service_name = service_name
    
    def store_token_env(self, token: str, env_var: str = "CAMERA_SERVICE_JWT_TOKEN"):
        """Store token in environment variable."""
        os.environ[env_var] = token
        print(f"✅ Token stored in environment variable: {env_var}")
    
    def get_token_env(self, env_var: str = "CAMERA_SERVICE_JWT_TOKEN") -> Optional[str]:
        """Get token from environment variable."""
        return os.environ.get(env_var)
    
    def store_token_keyring(self, token: str, username: str = "default"):
        """Store token securely using keyring."""
        try:
            keyring.set_password(self.service_name, username, token)
            print(f"✅ Token stored securely for user: {username}")
        except Exception as e:
            print(f"❌ Failed to store token securely: {e}")
    
    def get_token_keyring(self, username: str = "default") -> Optional[str]:
        """Get token from keyring."""
        try:
            return keyring.get_password(self.service_name, username)
        except Exception as e:
            print(f"❌ Failed to retrieve token: {e}")
            return None
    
    def store_token_file(self, token: str, filepath: str = "~/.camera_service_token"):
        """Store token in file with restricted permissions."""
        import stat
        
        filepath = os.path.expanduser(filepath)
        
        try:
            with open(filepath, 'w') as f:
                f.write(token)
            
            # Set restrictive permissions (owner read/write only)
            os.chmod(filepath, stat.S_IRUSR | stat.S_IWUSR)
            print(f"✅ Token stored in file: {filepath}")
            
        except Exception as e:
            print(f"❌ Failed to store token in file: {e}")
    
    def get_token_file(self, filepath: str = "~/.camera_service_token") -> Optional[str]:
        """Get token from file."""
        filepath = os.path.expanduser(filepath)
        
        try:
            if os.path.exists(filepath):
                with open(filepath, 'r') as f:
                    return f.read().strip()
        except Exception as e:
            print(f"❌ Failed to read token from file: {e}")
        
        return None
    
    def clear_token(self, username: str = "default"):
        """Clear stored token."""
        try:
            # Clear from keyring
            keyring.delete_password(self.service_name, username)
            
            # Clear from environment
            if "CAMERA_SERVICE_JWT_TOKEN" in os.environ:
                del os.environ["CAMERA_SERVICE_JWT_TOKEN"]
            
            # Clear from file
            filepath = os.path.expanduser("~/.camera_service_token")
            if os.path.exists(filepath):
                os.remove(filepath)
            
            print("✅ Token cleared from all storage locations")
            
        except Exception as e:
            print(f"❌ Failed to clear token: {e}")

# Usage example
storage = SecureTokenStorage()

# Store token securely
storage.store_token_keyring(token, "john.doe")
storage.store_token_env(token)
storage.store_token_file(token)

# Retrieve token
retrieved_token = storage.get_token_keyring("john.doe")
if retrieved_token:
    print("✅ Token retrieved from secure storage")
```

### Token Validation and Security Checks

```python
import jwt
import hashlib
import secrets
from datetime import datetime, timedelta
from typing import Dict, Any, List, Optional

class JWTSecurityValidator:
    """Comprehensive JWT security validation."""
    
    def __init__(self, secret_key: str):
        self.secret_key = secret_key
        self.blacklisted_tokens = set()
        self.token_usage_count = {}
    
    def validate_token_security(self, token: str) -> Dict[str, Any]:
        """
        Comprehensive token security validation.
        
        Returns:
            Dictionary with validation results
        """
        result = {
            'valid': False,
            'errors': [],
            'warnings': [],
            'payload': None
        }
        
        try:
            # Decode token
            payload = jwt.decode(token, self.secret_key, algorithms=['HS256'])
            result['payload'] = payload
            
            # Check if token is blacklisted
            token_hash = hashlib.sha256(token.encode()).hexdigest()
            if token_hash in self.blacklisted_tokens:
                result['errors'].append("Token is blacklisted")
                return result
            
            # Check expiration
            if datetime.fromtimestamp(payload['exp']) < datetime.utcnow():
                result['errors'].append("Token has expired")
                return result
            
            # Check issued at time (not issued in the future)
            if datetime.fromtimestamp(payload['iat']) > datetime.utcnow():
                result['errors'].append("Token issued in the future")
                return result
            
            # Check token age (not too old)
            max_age = timedelta(days=30)
            if datetime.fromtimestamp(payload['iat']) < datetime.utcnow() - max_age:
                result['warnings'].append("Token is very old")
            
            # Check required claims
            required_claims = ['sub', 'iat', 'exp']
            for claim in required_claims:
                if claim not in payload:
                    result['errors'].append(f"Missing required claim: {claim}")
            
            # Check for suspicious patterns
            if self._check_suspicious_patterns(payload):
                result['warnings'].append("Suspicious token patterns detected")
            
            # Track token usage
            self._track_token_usage(token_hash)
            
            if not result['errors']:
                result['valid'] = True
            
        except jwt.InvalidTokenError as e:
            result['errors'].append(f"Invalid token: {e}")
        except Exception as e:
            result['errors'].append(f"Validation error: {e}")
        
        return result
    
    def _check_suspicious_patterns(self, payload: Dict[str, Any]) -> bool:
        """Check for suspicious patterns in token payload."""
        suspicious = False
        
        # Check for excessive permissions
        if 'permissions' in payload and len(payload['permissions']) > 10:
            suspicious = True
        
        # Check for unusual roles
        unusual_roles = ['superadmin', 'root', 'master']
        if payload.get('role') in unusual_roles:
            suspicious = True
        
        # Check for missing standard claims
        if 'jti' not in payload:
            suspicious = True
        
        return suspicious
    
    def _track_token_usage(self, token_hash: str):
        """Track token usage for security monitoring."""
        if token_hash not in self.token_usage_count:
            self.token_usage_count[token_hash] = 0
        
        self.token_usage_count[token_hash] += 1
        
        # Alert if token is used excessively
        if self.token_usage_count[token_hash] > 1000:
            print(f"⚠️ Warning: Token used {self.token_usage_count[token_hash]} times")
    
    def blacklist_token(self, token: str):
        """Add token to blacklist."""
        token_hash = hashlib.sha256(token.encode()).hexdigest()
        self.blacklisted_tokens.add(token_hash)
        print(f"✅ Token blacklisted")
    
    def get_token_usage_stats(self) -> Dict[str, int]:
        """Get token usage statistics."""
        return self.token_usage_count.copy()

# Usage example
validator = JWTSecurityValidator(secret_key)

# Validate token security
result = validator.validate_token_security(token)

if result['valid']:
    print("✅ Token is secure and valid")
    if result['warnings']:
        print(f"⚠️ Warnings: {result['warnings']}")
else:
    print(f"❌ Token validation failed: {result['errors']}")

# Get usage statistics
stats = validator.get_token_usage_stats()
print(f"Token usage statistics: {stats}")
```

## Integration Examples

### Client Integration with Token Management

```python
import asyncio
from examples.python.camera_client import CameraClient

class AuthenticatedCameraClient:
    """Camera client with integrated JWT token management."""
    
    def __init__(self, host: str, port: int, secret_key: str):
        self.host = host
        self.port = port
        self.secret_key = secret_key
        self.client = None
        self.token_manager = JWTTokenManager(secret_key)
        self.storage = SecureTokenStorage()
    
    async def authenticate(self, username: str) -> bool:
        """Authenticate using stored token or generate new one."""
        # Try to get stored token
        stored_token = self.storage.get_token_keyring(username)
        
        if stored_token:
            # Validate stored token
            if self.token_manager.set_token(stored_token):
                print(f"✅ Using stored token for {username}")
                return True
        
        # Generate new token (in real app, this would be from auth server)
        new_token = generate_jwt_token(
            user_id=username,
            user_name=username,
            role="user",
            secret_key=self.secret_key
        )
        
        if self.token_manager.set_token(new_token):
            # Store new token
            self.storage.store_token_keyring(new_token, username)
            print(f"✅ Generated and stored new token for {username}")
            return True
        
        return False
    
    async def connect(self) -> bool:
        """Connect to camera service with valid token."""
        valid_token = self.token_manager.get_valid_token()
        if not valid_token:
            print("❌ No valid token available")
            return False
        
        # Create client with token
        self.client = CameraClient(
            host=self.host,
            port=self.port,
            auth_type="jwt",
            auth_token=valid_token
        )
        
        try:
            await self.client.connect()
            print("✅ Connected to camera service")
            return True
        except Exception as e:
            print(f"❌ Connection failed: {e}")
            return False
    
    async def ensure_connection(self) -> bool:
        """Ensure we have a valid connection, refreshing token if needed."""
        if not self.client or not self.client.connected:
            return await self.connect()
        
        # Check if token needs refresh
        valid_token = self.token_manager.get_valid_token()
        if not valid_token:
            print("🔄 Token expired, reconnecting...")
            await self.client.disconnect()
            return await self.connect()
        
        return True
    
    async def camera_operation(self, operation, *args, **kwargs):
        """Perform camera operation with automatic token management."""
        if not await self.ensure_connection():
            return None
        
        try:
            if operation == "list":
                return await self.client.get_camera_list()
            elif operation == "snapshot":
                return await self.client.take_snapshot(*args, **kwargs)
            elif operation == "record":
                return await self.client.start_recording(*args, **kwargs)
            else:
                raise ValueError(f"Unknown operation: {operation}")
        except Exception as e:
            print(f"❌ Operation failed: {e}")
            return None

# Usage example
async def main():
    auth_client = AuthenticatedCameraClient("localhost", 8080, secret_key)
    
    # Authenticate
    if not await auth_client.authenticate("john.doe"):
        print("❌ Authentication failed")
        return
    
    # Connect
    if not await auth_client.connect():
        print("❌ Connection failed")
        return
    
    # Perform operations
    cameras = await auth_client.camera_operation("list")
    if cameras:
        print(f"📹 Found {len(cameras)} cameras")
    
    # Token will be automatically refreshed as needed

# Run the example
# asyncio.run(main())
```

## Best Practices Summary

### 1. Token Generation
- Use strong secret keys
- Include all required claims
- Set appropriate expiration times
- Generate unique JWT IDs

### 2. Token Validation
- Always validate token signature
- Check expiration times
- Verify required claims
- Implement role-based access control

### 3. Token Refresh
- Monitor token expiration
- Implement automatic refresh
- Handle refresh failures gracefully
- Use secure refresh endpoints

### 4. Token Security
- Store tokens securely
- Implement token blacklisting
- Monitor token usage
- Use HTTPS for all communications

### 5. Error Handling
- Handle all JWT exceptions
- Provide clear error messages
- Implement retry logic
- Log security events

---

**Version:** 1.0  
**Last Updated:** 2025-08-06  
**Compatibility:** Python 3.8+, PyJWT, keyring 