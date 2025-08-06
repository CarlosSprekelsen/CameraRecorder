# API Key Setup Guide - MediaMTX Camera Service

## Overview

This guide provides comprehensive information about setting up and managing API keys for the MediaMTX Camera Service. It covers generation, validation, security best practices, and integration examples.

## API Key Fundamentals

### What are API Keys?

API keys are secure, unique identifiers that provide access to the camera service without requiring user authentication. They are ideal for:

- Automated scripts and applications
- Service-to-service communication
- Long-running processes
- Integration with third-party systems

### API Key Structure

```
camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
```

**Components:**
- **Prefix**: `camera_service_sk_` (16 characters)
- **Key**: 64-character random string
- **Total Length**: 80 characters

### API Key Format Validation

```python
import re

def validate_api_key_format(api_key: str) -> bool:
    """
    Validate API key format.
    
    Args:
        api_key: API key string to validate
        
    Returns:
        True if valid format, False otherwise
    """
    pattern = r'^camera_service_sk_[a-zA-Z0-9]{64}$'
    return bool(re.match(pattern, api_key))

# Usage example
api_key = "camera_service_sk_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
if validate_api_key_format(api_key):
    print("✅ Valid API key format")
else:
    print("❌ Invalid API key format")
```

## API Key Generation

### Basic API Key Generation

```python
import secrets
import string

def generate_api_key() -> str:
    """
    Generate a secure API key for camera service.
    
    Returns:
        API key string
    """
    # Generate 64 random characters
    alphabet = string.ascii_letters + string.digits
    random_part = ''.join(secrets.choice(alphabet) for _ in range(64))
    
    # Combine with prefix
    api_key = f"camera_service_sk_{random_part}"
    
    return api_key

# Usage example
api_key = generate_api_key()
print(f"Generated API key: {api_key}")
```

### Advanced API Key Generation with Metadata

```python
import secrets
import string
import hashlib
import time
from typing import Dict, Any, Optional

class APIKeyGenerator:
    """Advanced API key generator with metadata tracking."""
    
    def __init__(self):
        self.generated_keys = {}
    
    def generate_api_key_with_metadata(
        self,
        name: str,
        description: str = "",
        permissions: Optional[list] = None,
        expires_at: Optional[int] = None
    ) -> Dict[str, Any]:
        """
        Generate API key with metadata.
        
        Args:
            name: Key name for identification
            description: Key description
            permissions: List of permissions
            expires_at: Expiration timestamp (optional)
            
        Returns:
            Dictionary with key and metadata
        """
        # Generate the key
        api_key = generate_api_key()
        
        # Create metadata
        metadata = {
            'key': api_key,
            'name': name,
            'description': description,
            'created_at': int(time.time()),
            'permissions': permissions or ['camera:read', 'camera:write'],
            'expires_at': expires_at,
            'key_hash': hashlib.sha256(api_key.encode()).hexdigest(),
            'status': 'active'
        }
        
        # Store metadata
        self.generated_keys[api_key] = metadata
        
        return metadata
    
    def list_generated_keys(self) -> list:
        """List all generated keys with metadata."""
        return list(self.generated_keys.values())
    
    def get_key_metadata(self, api_key: str) -> Optional[Dict[str, Any]]:
        """Get metadata for a specific key."""
        return self.generated_keys.get(api_key)
    
    def revoke_key(self, api_key: str) -> bool:
        """Revoke an API key."""
        if api_key in self.generated_keys:
            self.generated_keys[api_key]['status'] = 'revoked'
            return True
        return False

# Usage example
generator = APIKeyGenerator()

# Generate key with metadata
key_data = generator.generate_api_key_with_metadata(
    name="production-camera-key",
    description="API key for production camera operations",
    permissions=['camera:read', 'camera:write', 'camera:delete'],
    expires_at=int(time.time()) + (30 * 24 * 3600)  # 30 days
)

print(f"Generated key: {key_data['key']}")
print(f"Permissions: {key_data['permissions']}")
```

### Batch API Key Generation

```python
import csv
from typing import List, Dict

def generate_batch_api_keys(count: int, base_name: str = "camera-key") -> List[Dict[str, Any]]:
    """
    Generate multiple API keys for batch operations.
    
    Args:
        count: Number of keys to generate
        base_name: Base name for key identification
        
    Returns:
        List of key metadata dictionaries
    """
    keys = []
    generator = APIKeyGenerator()
    
    for i in range(count):
        key_data = generator.generate_api_key_with_metadata(
            name=f"{base_name}-{i+1:03d}",
            description=f"Batch generated key {i+1}",
            permissions=['camera:read', 'camera:write']
        )
        keys.append(key_data)
    
    return keys

def export_keys_to_csv(keys: List[Dict[str, Any]], filename: str):
    """Export generated keys to CSV file."""
    with open(filename, 'w', newline='') as csvfile:
        fieldnames = ['name', 'key', 'description', 'permissions', 'created_at', 'expires_at']
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        
        writer.writeheader()
        for key_data in keys:
            writer.writerow({
                'name': key_data['name'],
                'key': key_data['key'],
                'description': key_data['description'],
                'permissions': ','.join(key_data['permissions']),
                'created_at': key_data['created_at'],
                'expires_at': key_data['expires_at'] or ''
            })

# Usage example
keys = generate_batch_api_keys(5, "test-key")
export_keys_to_csv(keys, "generated_api_keys.csv")
print(f"Generated {len(keys)} API keys and exported to CSV")
```

## API Key Management

### Secure API Key Storage

```python
import os
import keyring
import json
from typing import Optional, Dict, Any

class SecureAPIKeyStorage:
    """Secure storage for API keys using multiple methods."""
    
    def __init__(self, service_name: str = "camera-service"):
        self.service_name = service_name
    
    def store_key_env(self, key_name: str, api_key: str, env_var: str = None):
        """Store API key in environment variable."""
        if env_var is None:
            env_var = f"CAMERA_SERVICE_API_KEY_{key_name.upper()}"
        
        os.environ[env_var] = api_key
        print(f"✅ API key stored in environment variable: {env_var}")
    
    def get_key_env(self, key_name: str, env_var: str = None) -> Optional[str]:
        """Get API key from environment variable."""
        if env_var is None:
            env_var = f"CAMERA_SERVICE_API_KEY_{key_name.upper()}"
        
        return os.environ.get(env_var)
    
    def store_key_keyring(self, key_name: str, api_key: str):
        """Store API key securely using keyring."""
        try:
            keyring.set_password(self.service_name, key_name, api_key)
            print(f"✅ API key stored securely: {key_name}")
        except Exception as e:
            print(f"❌ Failed to store API key: {e}")
    
    def get_key_keyring(self, key_name: str) -> Optional[str]:
        """Get API key from keyring."""
        try:
            return keyring.get_password(self.service_name, key_name)
        except Exception as e:
            print(f"❌ Failed to retrieve API key: {e}")
            return None
    
    def store_key_file(self, key_name: str, api_key: str, filepath: str = None):
        """Store API key in file with restricted permissions."""
        import stat
        
        if filepath is None:
            filepath = os.path.expanduser(f"~/.camera_service_keys/{key_name}")
        
        # Create directory if it doesn't exist
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        
        try:
            with open(filepath, 'w') as f:
                f.write(api_key)
            
            # Set restrictive permissions (owner read/write only)
            os.chmod(filepath, stat.S_IRUSR | stat.S_IWUSR)
            print(f"✅ API key stored in file: {filepath}")
            
        except Exception as e:
            print(f"❌ Failed to store API key in file: {e}")
    
    def get_key_file(self, key_name: str, filepath: str = None) -> Optional[str]:
        """Get API key from file."""
        if filepath is None:
            filepath = os.path.expanduser(f"~/.camera_service_keys/{key_name}")
        
        try:
            if os.path.exists(filepath):
                with open(filepath, 'r') as f:
                    return f.read().strip()
        except Exception as e:
            print(f"❌ Failed to read API key from file: {e}")
        
        return None
    
    def list_stored_keys(self) -> Dict[str, str]:
        """List all stored API keys."""
        keys = {}
        
        # Check environment variables
        for env_var, value in os.environ.items():
            if env_var.startswith('CAMERA_SERVICE_API_KEY_'):
                key_name = env_var.replace('CAMERA_SERVICE_API_KEY_', '').lower()
                keys[key_name] = f"env:{env_var}"
        
        # Check keyring
        try:
            import keyring
            # This is a simplified version - in practice you'd need to list all keys
            pass
        except Exception:
            pass
        
        # Check files
        key_dir = os.path.expanduser("~/.camera_service_keys")
        if os.path.exists(key_dir):
            for filename in os.listdir(key_dir):
                if filename.endswith('.key'):
                    key_name = filename[:-4]
                    keys[key_name] = f"file:{os.path.join(key_dir, filename)}"
        
        return keys
    
    def clear_key(self, key_name: str):
        """Clear stored API key."""
        try:
            # Clear from keyring
            keyring.delete_password(self.service_name, key_name)
            
            # Clear from environment
            env_var = f"CAMERA_SERVICE_API_KEY_{key_name.upper()}"
            if env_var in os.environ:
                del os.environ[env_var]
            
            # Clear from file
            filepath = os.path.expanduser(f"~/.camera_service_keys/{key_name}")
            if os.path.exists(filepath):
                os.remove(filepath)
            
            print(f"✅ API key cleared: {key_name}")
            
        except Exception as e:
            print(f"❌ Failed to clear API key: {e}")

# Usage example
storage = SecureAPIKeyStorage()

# Store API key securely
api_key = generate_api_key()
storage.store_key_keyring("production", api_key)
storage.store_key_env("production", api_key)
storage.store_key_file("production", api_key)

# List stored keys
stored_keys = storage.list_stored_keys()
print(f"Stored keys: {stored_keys}")
```

### API Key Validation and Security

```python
import hashlib
import time
from typing import Dict, Any, List, Optional

class APIKeyValidator:
    """Comprehensive API key validation and security checks."""
    
    def __init__(self):
        self.blacklisted_keys = set()
        self.key_usage_count = {}
        self.key_permissions = {}
    
    def validate_api_key(
        self,
        api_key: str,
        required_permissions: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        Validate API key with comprehensive security checks.
        
        Args:
            api_key: API key to validate
            required_permissions: Required permissions for access
            
        Returns:
            Dictionary with validation results
        """
        result = {
            'valid': False,
            'errors': [],
            'warnings': [],
            'permissions': []
        }
        
        # Check format
        if not validate_api_key_format(api_key):
            result['errors'].append("Invalid API key format")
            return result
        
        # Check if blacklisted
        key_hash = hashlib.sha256(api_key.encode()).hexdigest()
        if key_hash in self.blacklisted_keys:
            result['errors'].append("API key is blacklisted")
            return result
        
        # Check expiration (if applicable)
        if self._is_key_expired(api_key):
            result['errors'].append("API key has expired")
            return result
        
        # Check permissions
        permissions = self.key_permissions.get(api_key, [])
        result['permissions'] = permissions
        
        if required_permissions:
            missing_permissions = [perm for perm in required_permissions if perm not in permissions]
            if missing_permissions:
                result['errors'].append(f"Missing permissions: {missing_permissions}")
        
        # Track usage
        self._track_key_usage(key_hash)
        
        # Check for suspicious usage patterns
        if self._check_suspicious_usage(key_hash):
            result['warnings'].append("Suspicious usage patterns detected")
        
        if not result['errors']:
            result['valid'] = True
        
        return result
    
    def _is_key_expired(self, api_key: str) -> bool:
        """Check if API key has expired."""
        # This would check against stored expiration data
        # For now, return False (no expiration)
        return False
    
    def _track_key_usage(self, key_hash: str):
        """Track API key usage for security monitoring."""
        if key_hash not in self.key_usage_count:
            self.key_usage_count[key_hash] = 0
        
        self.key_usage_count[key_hash] += 1
        
        # Alert if key is used excessively
        if self.key_usage_count[key_hash] > 10000:
            print(f"⚠️ Warning: API key used {self.key_usage_count[key_hash]} times")
    
    def _check_suspicious_usage(self, key_hash: str) -> bool:
        """Check for suspicious usage patterns."""
        usage_count = self.key_usage_count.get(key_hash, 0)
        
        # Check for excessive usage in short time
        if usage_count > 1000:  # More than 1000 requests
            return True
        
        return False
    
    def blacklist_key(self, api_key: str):
        """Add API key to blacklist."""
        key_hash = hashlib.sha256(api_key.encode()).hexdigest()
        self.blacklisted_keys.add(key_hash)
        print(f"✅ API key blacklisted")
    
    def set_key_permissions(self, api_key: str, permissions: List[str]):
        """Set permissions for an API key."""
        self.key_permissions[api_key] = permissions
    
    def get_key_usage_stats(self) -> Dict[str, int]:
        """Get API key usage statistics."""
        return self.key_usage_count.copy()

# Usage example
validator = APIKeyValidator()

# Set permissions for a key
validator.set_key_permissions(api_key, ['camera:read', 'camera:write'])

# Validate key
result = validator.validate_api_key(api_key, required_permissions=['camera:read'])
if result['valid']:
    print("✅ API key is valid")
    print(f"   Permissions: {result['permissions']}")
else:
    print(f"❌ API key validation failed: {result['errors']}")
```

## Client Integration

### Python Client Integration

```python
import os
from examples.python.camera_client import CameraClient

class APIKeyCameraClient:
    """Camera client with API key management."""
    
    def __init__(self, host: str, port: int):
        self.host = host
        self.port = port
        self.client = None
        self.storage = SecureAPIKeyStorage()
        self.validator = APIKeyValidator()
    
    def load_api_key(self, key_name: str) -> Optional[str]:
        """Load API key from secure storage."""
        # Try keyring first
        api_key = self.storage.get_key_keyring(key_name)
        if api_key:
            return api_key
        
        # Try environment variable
        api_key = self.storage.get_key_env(key_name)
        if api_key:
            return api_key
        
        # Try file
        api_key = self.storage.get_key_file(key_name)
        if api_key:
            return api_key
        
        return None
    
    async def connect_with_api_key(self, key_name: str) -> bool:
        """Connect to camera service using API key."""
        api_key = self.load_api_key(key_name)
        if not api_key:
            print(f"❌ API key not found: {key_name}")
            return False
        
        # Validate API key
        result = self.validator.validate_api_key(api_key)
        if not result['valid']:
            print(f"❌ API key validation failed: {result['errors']}")
            return False
        
        # Create client
        self.client = CameraClient(
            host=self.host,
            port=self.port,
            auth_type="api_key",
            api_key=api_key
        )
        
        try:
            await self.client.connect()
            print("✅ Connected with API key")
            return True
        except Exception as e:
            print(f"❌ Connection failed: {e}")
            return False
    
    async def camera_operation(self, operation, *args, **kwargs):
        """Perform camera operation with API key authentication."""
        if not self.client or not self.client.connected:
            print("❌ Not connected to camera service")
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
    api_client = APIKeyCameraClient("localhost", 8080)
    
    # Connect using stored API key
    if await api_client.connect_with_api_key("production"):
        # Perform operations
        cameras = await api_client.camera_operation("list")
        if cameras:
            print(f"📹 Found {len(cameras)} cameras")

# Run the example
# asyncio.run(main())
```

### JavaScript/Node.js Integration

```javascript
const { CameraClient } = require('./examples/javascript/camera_client.js');

class APIKeyCameraClient {
    constructor(host, port) {
        this.host = host;
        this.port = port;
        this.client = null;
    }
    
    loadAPIKey(keyName) {
        // Try environment variable
        const envKey = process.env[`CAMERA_SERVICE_API_KEY_${keyName.toUpperCase()}`];
        if (envKey) {
            return envKey;
        }
        
        // Try from file (simplified)
        const fs = require('fs');
        const path = require('path');
        const keyPath = path.join(process.env.HOME, '.camera_service_keys', keyName);
        
        try {
            if (fs.existsSync(keyPath)) {
                return fs.readFileSync(keyPath, 'utf8').trim();
            }
        } catch (error) {
            console.error(`Failed to read API key from file: ${error.message}`);
        }
        
        return null;
    }
    
    async connectWithAPIKey(keyName) {
        const apiKey = this.loadAPIKey(keyName);
        if (!apiKey) {
            console.error(`❌ API key not found: ${keyName}`);
            return false;
        }
        
        // Validate API key format
        if (!this.validateAPIKeyFormat(apiKey)) {
            console.error("❌ Invalid API key format");
            return false;
        }
        
        // Create client
        this.client = new CameraClient({
            host: this.host,
            port: this.port,
            authType: 'api_key',
            apiKey: apiKey
        });
        
        try {
            await this.client.connect();
            console.log("✅ Connected with API key");
            return true;
        } catch (error) {
            console.error(`❌ Connection failed: ${error.message}`);
            return false;
        }
    }
    
    validateAPIKeyFormat(apiKey) {
        const pattern = /^camera_service_sk_[a-zA-Z0-9]{64}$/;
        return pattern.test(apiKey);
    }
    
    async cameraOperation(operation, ...args) {
        if (!this.client || !this.client.connected) {
            console.error("❌ Not connected to camera service");
            return null;
        }
        
        try {
            switch (operation) {
                case "list":
                    return await this.client.getCameraList();
                case "snapshot":
                    return await this.client.takeSnapshot(...args);
                case "record":
                    return await this.client.startRecording(...args);
                default:
                    throw new Error(`Unknown operation: ${operation}`);
            }
        } catch (error) {
            console.error(`❌ Operation failed: ${error.message}`);
            return null;
        }
    }
}

// Usage example
async function main() {
    const apiClient = new APIKeyCameraClient("localhost", 8080);
    
    // Connect using stored API key
    if (await apiClient.connectWithAPIKey("production")) {
        // Perform operations
        const cameras = await apiClient.cameraOperation("list");
        if (cameras) {
            console.log(`📹 Found ${cameras.length} cameras`);
        }
    }
}

// Run the example
// main().catch(console.error);
```

## Security Best Practices

### 1. API Key Generation
- Use cryptographically secure random generation
- Include sufficient entropy (64 characters)
- Use consistent format with prefix
- Generate unique keys for each application

### 2. API Key Storage
- Never hardcode keys in source code
- Use secure storage methods (keyring, environment variables)
- Implement proper file permissions
- Encrypt keys at rest when possible

### 3. API Key Validation
- Always validate key format
- Check key permissions
- Implement key expiration
- Monitor key usage patterns

### 4. API Key Security
- Rotate keys regularly
- Implement key blacklisting
- Monitor for suspicious usage
- Use HTTPS for all communications

### 5. Access Control
- Implement least-privilege access
- Use role-based permissions
- Monitor access patterns
- Implement rate limiting

## Troubleshooting

### Common Issues

#### 1. "Invalid API Key" Error
```bash
# Check API key format
echo $API_KEY | grep -E "^camera_service_sk_[a-zA-Z0-9]{64}$"

# Verify API key with server
curl -H "X-API-Key: $API_KEY" http://localhost:8080/health
```

#### 2. "API Key Not Found" Error
```bash
# Check environment variables
env | grep CAMERA_SERVICE_API_KEY

# Check keyring
python -c "import keyring; print(keyring.get_password('camera-service', 'your-key-name'))"

# Check file storage
ls -la ~/.camera_service_keys/
```

#### 3. "Permission Denied" Error
```python
# Check API key permissions
validator = APIKeyValidator()
result = validator.validate_api_key(api_key, required_permissions=['camera:read'])
print(f"Permissions: {result['permissions']}")
```

### Debug API Keys

```python
# Enable debug logging
import logging
logging.basicConfig(level=logging.DEBUG)

# Test API key validation
def debug_api_key(api_key):
    print(f"API Key: {api_key}")
    print(f"Length: {len(api_key)}")
    print(f"Format valid: {validate_api_key_format(api_key)}")
    
    # Test with validator
    validator = APIKeyValidator()
    result = validator.validate_api_key(api_key)
    print(f"Validation result: {result}")

# Usage
debug_api_key("your_api_key_here")
```

## Best Practices Summary

### 1. Key Generation
- Use secure random generation
- Include sufficient entropy
- Use consistent naming
- Generate unique keys per application

### 2. Key Storage
- Use secure storage methods
- Never hardcode in source
- Implement proper permissions
- Encrypt when possible

### 3. Key Validation
- Always validate format
- Check permissions
- Monitor usage
- Implement expiration

### 4. Security Monitoring
- Track key usage
- Monitor for abuse
- Implement blacklisting
- Regular key rotation

### 5. Access Control
- Least privilege access
- Role-based permissions
- Rate limiting
- Audit logging

---

**Version:** 1.0  
**Last Updated:** 2025-08-06  
**Compatibility:** All client types (Python, JavaScript, Browser, CLI) 