# Client Port Configuration Guide

## Critical Port Configuration Issue

**IMPORTANT**: The MediaMTX Camera Service uses different ports in development vs production environments. All client examples have been updated to use the production default port, but you must specify the correct port for your environment.

## Port Configuration by Environment

### Production/Default Configuration
- **Port**: `8002`
- **Config File**: `config/default.yaml`
- **Usage**: Production deployments, systemd services, Docker containers

### Development Configuration  
- **Port**: `8080`
- **Config File**: `config/development.yaml`
- **Usage**: Local development, testing, debugging

## Client Examples Updated

All client examples have been updated to use port `8002` as the default (production configuration):

### Python Client
```bash
# Production (default)
python camera_client.py --host localhost --port 8002 --auth-type jwt --token your_token

# Development
python camera_client.py --host localhost --port 8080 --auth-type jwt --token your_token
```

### JavaScript Client
```bash
# Production (default)
node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_token

# Development
node camera_client.js --host localhost --port 8080 --auth-type jwt --token your_token
```

### CLI Client
```bash
# Production (default)
python camera_cli.py --host localhost --port 8002 --auth-type jwt --token your_token list

# Development
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_token list
```

### Browser Client
- **Default Port**: `8002` (production)
- **Development**: Change to `8080` in the UI

## SDK Requirements

### Python SDK
- **Python**: >= 3.8
- **Dependencies**: websockets, asyncio, typing-extensions

### JavaScript SDK
- **Node.js**: >= 12.0.0 (tested with Node.js 12.22.9)
- **TypeScript**: 4.5.5 (compatible with Node.js 12)
- **Dependencies**: ws, uuid

**Note**: The JavaScript SDK has been configured to be compatible with Node.js 12+ by using TypeScript 4.5.5 and compatible dependency versions.

## Configuration Files

### Default Configuration (`config/default.yaml`)
```yaml
server:
  host: "0.0.0.0"
  port: 8002  # Production default
  websocket_path: "/ws"
  max_connections: 100
```

### Development Configuration (`config/development.yaml`)
```yaml
server:
  host: "0.0.0.0"
  port: 8080  # Development default
  websocket_path: "/ws"
  max_connections: 100
```

## Troubleshooting Connection Issues

### Common Error Messages
```
❌ Camera service error: Failed to connect after 3 attempts: [Errno 111] Connect call failed ('127.0.0.1', 8002)
```

### Solutions
1. **Check which configuration is being used**:
   ```bash
   # Check if service is running on port 8002 (production)
   netstat -tlnp | grep 8002
   
   # Check if service is running on port 8080 (development)
   netstat -tlnp | grep 8080
   ```

2. **Verify service startup**:
   ```bash
   # Check service status
   sudo systemctl status camera-service
   
   # Check logs for port binding
   sudo journalctl -u camera-service | grep "Starting WebSocket"
   ```

3. **Use correct port in client**:
   ```bash
   # For development environment
   python camera_client.py --port 8080 --auth-type jwt --token your_token
   
   # For production environment  
   python camera_client.py --port 8002 --auth-type jwt --token your_token
   ```

## Environment Detection

The service automatically detects the environment based on the configuration file used:

- **Production**: Uses `config/default.yaml` (port 8002)
- **Development**: Uses `config/development.yaml` (port 8080)

## Migration Guide

If you have existing scripts or documentation using port 8080:

1. **For Production Deployments**: Update to use port 8002
2. **For Development**: Keep using port 8080 or update to use port 8002
3. **For Documentation**: Update examples to show both ports with clear labels

## SDK Configuration

The SDK packages also need to be configured with the correct port:

### Python SDK
```python
from mediamtx_camera_sdk import CameraClient

# Production
client = CameraClient(host="localhost", port=8002)

# Development  
client = CameraClient(host="localhost", port=8080)
```

### JavaScript SDK
```javascript
const { CameraClient } = require('@mediamtx/camera-sdk');

// Production
const client = new CameraClient({ host: 'localhost', port: 8002 });

// Development
const client = new CameraClient({ host: 'localhost', port: 8080 });
```

## Related Documentation

- [Installation Guide](deployment/INSTALLATION_GUIDE.md)
- [API Reference](api/README.md)
- [Development Setup](development/SETUP.md)
