# MediaMTX Camera Service JavaScript Client Example

This directory contains a JavaScript/Node.js client example for the MediaMTX Camera Service.

## Features

- WebSocket JSON-RPC 2.0 communication
- JWT and API Key authentication
- Camera discovery and control
- Snapshot and recording operations
- Real-time status notifications
- Comprehensive error handling
- Retry logic and connection recovery

## Requirements

- Node.js >= 12.0.0 (tested with v12.22.9)
- npm or yarn package manager

## Installation

1. Navigate to the examples/javascript directory:
   ```bash
   cd examples/javascript
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Verify installation:
   ```bash
   node camera_client.js --help
   ```

## Usage

### Basic Usage

```bash
# Connect with JWT authentication
node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_jwt_token

# Connect with API key authentication
node camera_client.js --host localhost --port 8002 --auth-type api_key --key your_api_key
```

### Command Line Options

- `--host`: Server hostname (default: localhost)
- `--port`: Server port (default: 8002)
- `--auth-type`: Authentication type: `jwt` or `api_key` (default: jwt)
- `--token`: JWT token for authentication
- `--key`: API key for authentication
- `--ssl`: Use SSL/TLS connection (default: false)
- `--help`: Show help information

### Example Scripts

```bash
# List all cameras
node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_token --list-cameras

# Get camera status
node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_token --camera /dev/video0 --status

# Take a snapshot
node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_token --camera /dev/video0 --snapshot

# Start recording
node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_token --camera /dev/video0 --start-recording

# Stop recording
node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_token --camera /dev/video0 --stop-recording
```

## Programmatic Usage

```javascript
const CameraClient = require('./camera_client.js');

async function main() {
    const client = new CameraClient({
        host: 'localhost',
        port: 8002,
        authType: 'jwt',
        authToken: 'your_jwt_token',
        useSsl: false,
        maxRetries: 3,
        retryDelay: 1.0
    });

    try {
        await client.connect();
        console.log('Connected successfully');

        // List cameras
        const cameras = await client.getCameraList();
        console.log('Available cameras:', cameras);

        // Get camera status
        const status = await client.getCameraStatus('/dev/video0');
        console.log('Camera status:', status);

        // Take snapshot
        const snapshot = await client.takeSnapshot('/dev/video0');
        console.log('Snapshot taken:', snapshot);

    } catch (error) {
        console.error('Error:', error.message);
    } finally {
        await client.disconnect();
    }
}

main();
```

## Error Handling

The client includes comprehensive error handling for:

- Connection failures
- Authentication errors
- Camera not found errors
- Network timeouts
- Invalid responses

All errors are wrapped in appropriate exception classes:

- `AuthenticationError`: Authentication failures
- `ConnectionError`: Connection and network issues
- `CameraNotFoundError`: Camera device not found
- `CameraServiceError`: General service errors

## Dependencies

- `ws`: WebSocket client library (v7.5.9)
- `uuid`: UUID generation (v8.3.2)

## Node.js Compatibility

This client is compatible with Node.js v12.0.0 and later, including:
- Node.js v12.22.9 (tested)
- Node.js v14.x
- Node.js v16.x
- Node.js v18.x

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure the MediaMTX Camera Service is running on the specified host and port
2. **Authentication Failed**: Verify your JWT token or API key is valid
3. **Camera Not Found**: Check that the camera device path exists and is accessible
4. **WebSocket Error**: Ensure the service supports WebSocket connections on the specified endpoint

### Debug Mode

Enable debug logging by setting the `DEBUG` environment variable:

```bash
DEBUG=* node camera_client.js --host localhost --port 8080 --auth-type jwt --token your_token
```

## License

MIT License - see the main project license for details.
