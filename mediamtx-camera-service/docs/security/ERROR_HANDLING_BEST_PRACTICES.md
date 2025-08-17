# Error Handling Best Practices

**Version:** 1.0  
**Status:** Production Ready  
**Epic:** E3 Client API & SDK Ecosystem  

## Overview

This guide provides comprehensive error handling best practices for the MediaMTX Camera Service, including error codes, recovery strategies, and implementation examples for all client types.

## Error Code Reference

### Authentication Errors

| Code | Message | Description | Recovery Strategy |
|------|---------|-------------|-------------------|
| -32001 | Authentication required | No auth token provided | Provide JWT token or API key |
| -32001 | Authentication failed | Invalid or expired token | Generate new token or check expiry |
| -32003 | Insufficient permissions | Role does not have required permissions | Use token with higher role |

### Connection Errors

| Code | Message | Description | Recovery Strategy |
|------|---------|-------------|-------------------|
| -32000 | Connection failed | WebSocket connection failed | Check network connectivity and retry |
| -32000 | Connection timeout | Connection attempt timed out | Increase timeout or check server status |
| -32000 | Connection closed | WebSocket connection was closed | Implement reconnection logic |

### Rate Limiting Errors

| Code | Message | Description | Recovery Strategy |
|------|---------|-------------|-------------------|
| -32002 | Rate limit exceeded | Client exceeded request rate limit | Implement exponential backoff |
| -32002 | Too many connections | Maximum connections reached | Close unused connections |

### Camera Operation Errors

| Code | Message | Description | Recovery Strategy |
|------|---------|-------------|-------------------|
| -32004 | Camera not found | Camera device not found | Check device path and retry |
| -32005 | Camera busy | Camera is in use by another client | Wait and retry with backoff |
| -32006 | Operation failed | Camera operation failed | Check camera status and retry |

### MediaMTX Errors

| Code | Message | Description | Recovery Strategy |
|------|---------|-------------|-------------------|
| -32007 | MediaMTX error | MediaMTX operation failed | Check MediaMTX logs and retry |
| -32008 | Recording failed | Recording operation failed | Check disk space and permissions |
| -32009 | Snapshot failed | Snapshot operation failed | Check camera availability |

## Error Handling Patterns

### 1. Retry with Exponential Backoff

```python
import asyncio
import random
from typing import Callable, Any

async def retry_with_backoff(
    operation: Callable,
    max_retries: int = 3,
    base_delay: float = 1.0,
    max_delay: float = 60.0,
    exponential_base: float = 2.0,
    jitter: bool = True
) -> Any:
    """
    Retry operation with exponential backoff.
    
    Args:
        operation: Async function to retry
        max_retries: Maximum number of retries
        base_delay: Base delay in seconds
        max_delay: Maximum delay in seconds
        exponential_base: Exponential base for backoff
        jitter: Add random jitter to delay
    
    Returns:
        Operation result
    
    Raises:
        Last exception if all retries fail
    """
    last_exception = None
    
    for attempt in range(max_retries + 1):
        try:
            return await operation()
        except Exception as e:
            last_exception = e
            
            if attempt == max_retries:
                raise last_exception
            
            # Calculate delay with exponential backoff
            delay = min(base_delay * (exponential_base ** attempt), max_delay)
            
            # Add jitter to prevent thundering herd
            if jitter:
                delay = delay * (0.5 + random.random() * 0.5)
            
            print(f"Attempt {attempt + 1} failed: {e}. Retrying in {delay:.2f} seconds...")
            await asyncio.sleep(delay)
    
    raise last_exception

# Example usage
async def camera_operation():
    # Simulate camera operation that might fail
    if random.random() < 0.7:  # 70% failure rate
        raise ConnectionError("Camera connection failed")
    return "Operation successful"

# Retry with backoff
try:
    result = await retry_with_backoff(camera_operation, max_retries=3)
    print(f"Success: {result}")
except Exception as e:
    print(f"Failed after retries: {e}")
```

### 2. Circuit Breaker Pattern

```python
import asyncio
import time
from enum import Enum
from typing import Callable, Any

class CircuitState(Enum):
    CLOSED = "closed"      # Normal operation
    OPEN = "open"          # Circuit is open, failing fast
    HALF_OPEN = "half_open"  # Testing if service is back

class CircuitBreaker:
    def __init__(
        self,
        failure_threshold: int = 5,
        recovery_timeout: float = 60.0,
        expected_exception: type = Exception
    ):
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.expected_exception = expected_exception
        
        self.state = CircuitState.CLOSED
        self.failure_count = 0
        self.last_failure_time = None
    
    async def call(self, operation: Callable) -> Any:
        """Execute operation with circuit breaker protection."""
        if self.state == CircuitState.OPEN:
            if time.time() - self.last_failure_time > self.recovery_timeout:
                self.state = CircuitState.HALF_OPEN
            else:
                raise Exception("Circuit breaker is OPEN")
        
        try:
            result = await operation()
            self._on_success()
            return result
        except self.expected_exception as e:
            self._on_failure()
            raise e
    
    def _on_success(self):
        """Handle successful operation."""
        self.failure_count = 0
        self.state = CircuitState.CLOSED
    
    def _on_failure(self):
        """Handle failed operation."""
        self.failure_count += 1
        self.last_failure_time = time.time()
        
        if self.failure_count >= self.failure_threshold:
            self.state = CircuitState.OPEN

# Example usage
circuit_breaker = CircuitBreaker(failure_threshold=3, recovery_timeout=30.0)

async def unreliable_operation():
    if random.random() < 0.8:  # 80% failure rate
        raise ConnectionError("Service unavailable")
    return "Success"

try:
    result = await circuit_breaker.call(unreliable_operation)
    print(f"Result: {result}")
except Exception as e:
    print(f"Circuit breaker error: {e}")
```

### 3. Graceful Degradation

```python
from typing import Optional, Dict, Any
import asyncio

class CameraServiceClient:
    def __init__(self):
        self.primary_client = None
        self.fallback_client = None
        self.degraded_mode = False
    
    async def get_camera_list(self) -> list:
        """Get camera list with graceful degradation."""
        try:
            if not self.degraded_mode:
                return await self._get_camera_list_primary()
        except Exception as e:
            print(f"Primary service failed: {e}")
            self.degraded_mode = True
        
        # Fallback to degraded mode
        return await self._get_camera_list_fallback()
    
    async def _get_camera_list_primary(self) -> list:
        """Primary camera list implementation."""
        # Full implementation
        pass
    
    async def _get_camera_list_fallback(self) -> list:
        """Fallback camera list implementation."""
        # Simplified implementation with cached data
        return [
            {"device_path": "/dev/video0", "name": "Camera 1", "status": "unknown"},
            {"device_path": "/dev/video1", "name": "Camera 2", "status": "unknown"}
        ]
    
    async def take_snapshot(self, device_path: str) -> Optional[Dict[str, Any]]:
        """Take snapshot with graceful degradation."""
        try:
            if not self.degraded_mode:
                return await self._take_snapshot_primary(device_path)
        except Exception as e:
            print(f"Primary snapshot failed: {e}")
            self.degraded_mode = True
        
        # Fallback: return placeholder
        return {
            "device_path": device_path,
            "filename": f"snapshot_{device_path.replace('/', '_')}.jpg",
            "timestamp": time.time(),
            "degraded": True
        }
```

## Client-Specific Error Handling

### Python Client Error Handling

```python
import asyncio
from examples.python.camera_client import (
    CameraClient,
    CameraServiceError,
    AuthenticationError,
    ConnectionError,
    CameraNotFoundError,
    MediaMTXError
)

class RobustCameraClient:
    def __init__(self, **kwargs):
        self.client = CameraClient(**kwargs)
        self.max_retries = 3
        self.retry_delay = 1.0
    
    async def connect_with_retry(self):
        """Connect with automatic retry."""
        for attempt in range(self.max_retries):
            try:
                await self.client.connect()
                return
            except ConnectionError as e:
                if attempt == self.max_retries - 1:
                    raise
                print(f"Connection attempt {attempt + 1} failed: {e}")
                await asyncio.sleep(self.retry_delay * (attempt + 1))
    
    async def safe_camera_operation(self, operation, *args, **kwargs):
        """Execute camera operation with error handling."""
        try:
            return await operation(*args, **kwargs)
        except AuthenticationError as e:
            print(f"Authentication error: {e}")
            # Re-authenticate or refresh token
            raise
        except CameraNotFoundError as e:
            print(f"Camera not found: {e}")
            # Log and return empty result
            return None
        except MediaMTXError as e:
            print(f"MediaMTX error: {e}")
            # Retry with backoff
            return await self._retry_operation(operation, *args, **kwargs)
        except ConnectionError as e:
            print(f"Connection error: {e}")
            # Attempt reconnection
            await self._reconnect()
            return await operation(*args, **kwargs)
        except Exception as e:
            print(f"Unexpected error: {e}")
            raise
    
    async def _retry_operation(self, operation, *args, **kwargs):
        """Retry operation with exponential backoff."""
        return await retry_with_backoff(
            lambda: operation(*args, **kwargs),
            max_retries=3
        )
    
    async def _reconnect(self):
        """Reconnect to the service."""
        try:
            await self.client.disconnect()
        except:
            pass
        
        await self.connect_with_retry()

# Example usage
async def main():
    client = RobustCameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="your_token"
    )
    
    try:
        await client.connect_with_retry()
        
        # Safe camera operations
        cameras = await client.safe_camera_operation(client.client.get_camera_list)
        if cameras:
            snapshot = await client.safe_camera_operation(
                client.client.take_snapshot,
                cameras[0].device_path
            )
            print(f"Snapshot: {snapshot}")
    
    except Exception as e:
        print(f"Client error: {e}")

asyncio.run(main())
```

### JavaScript Client Error Handling

```javascript
import { CameraClient, AuthenticationError, ConnectionError, CameraNotFoundError, MediaMTXError } from './examples/javascript/camera_client.js';

class RobustCameraClient {
    constructor(config) {
        this.client = new CameraClient(config);
        this.maxRetries = 3;
        this.retryDelay = 1000;
    }
    
    async connectWithRetry() {
        /** Connect with automatic retry. */
        for (let attempt = 0; attempt < this.maxRetries; attempt++) {
            try {
                await this.client.connect();
                return;
            } catch (error) {
                if (error instanceof ConnectionError) {
                    if (attempt === this.maxRetries - 1) {
                        throw error;
                    }
                    console.log(`Connection attempt ${attempt + 1} failed: ${error.message}`);
                    await this.sleep(this.retryDelay * (attempt + 1));
                } else {
                    throw error;
                }
            }
        }
    }
    
    async safeCameraOperation(operation, ...args) {
        /** Execute camera operation with error handling. */
        try {
            return await operation.apply(this.client, args);
        } catch (error) {
            if (error instanceof AuthenticationError) {
                console.log(`Authentication error: ${error.message}`);
                // Re-authenticate or refresh token
                throw error;
            } else if (error instanceof CameraNotFoundError) {
                console.log(`Camera not found: ${error.message}`);
                // Log and return empty result
                return null;
            } else if (error instanceof MediaMTXError) {
                console.log(`MediaMTX error: ${error.message}`);
                // Retry with backoff
                return await this.retryOperation(operation, ...args);
            } else if (error instanceof ConnectionError) {
                console.log(`Connection error: ${error.message}`);
                // Attempt reconnection
                await this.reconnect();
                return await operation.apply(this.client, args);
            } else {
                console.log(`Unexpected error: ${error.message}`);
                throw error;
            }
        }
    }
    
    async retryOperation(operation, ...args) {
        /** Retry operation with exponential backoff. */
        return await this.retryWithBackoff(
            () => operation.apply(this.client, args),
            3
        );
    }
    
    async reconnect() {
        /** Reconnect to the service. */
        try {
            await this.client.disconnect();
        } catch (error) {
            // Ignore disconnect errors
        }
        
        await this.connectWithRetry();
    }
    
    async retryWithBackoff(operation, maxRetries = 3) {
        /** Retry operation with exponential backoff. */
        let lastError;
        
        for (let attempt = 0; attempt <= maxRetries; attempt++) {
            try {
                return await operation();
            } catch (error) {
                lastError = error;
                
                if (attempt === maxRetries) {
                    throw lastError;
                }
                
                const delay = Math.min(1000 * Math.pow(2, attempt), 60000);
                console.log(`Attempt ${attempt + 1} failed: ${error.message}. Retrying in ${delay}ms...`);
                await this.sleep(delay);
            }
        }
        
        throw lastError;
    }
    
    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

// Example usage
async function main() {
    const client = new RobustCameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'jwt',
        authToken: 'your_token'
    });
    
    try {
        await client.connectWithRetry();
        
        // Safe camera operations
        const cameras = await client.safeCameraOperation(client.client.getCameraList);
        if (cameras && cameras.length > 0) {
            const snapshot = await client.safeCameraOperation(
                client.client.takeSnapshot,
                cameras[0].devicePath
            );
            console.log(`Snapshot: ${snapshot}`);
        }
        
    } catch (error) {
        console.error(`Client error: ${error.message}`);
    }
}

main();
```

### Browser Client Error Handling

```javascript
class BrowserCameraClient {
    constructor() {
        this.websocket = null;
        this.connected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
    }
    
    async connect(host, port, authToken) {
        /** Connect with error handling and reconnection. */
        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${host}:${port}/ws`;
            
            this.websocket = new WebSocket(wsUrl);
            
            return new Promise((resolve, reject) => {
                const timeout = setTimeout(() => {
                    reject(new Error('Connection timeout'));
                }, 10000);
                
                this.websocket.onopen = async () => {
                    clearTimeout(timeout);
                    this.connected = true;
                    this.reconnectAttempts = 0;
                    
                    try {
                        // Authenticate
                        const authResponse = await this.sendRequest('authenticate', {
                            token: authToken,
                            auth_type: 'jwt'
                        });
                        
                        if (authResponse.authenticated) {
                            resolve();
                        } else {
                            reject(new Error(`Authentication failed: ${authResponse.error}`));
                        }
                    } catch (error) {
                        reject(error);
                    }
                };
                
                this.websocket.onerror = (error) => {
                    clearTimeout(timeout);
                    reject(new Error(`WebSocket error: ${error.message}`));
                };
                
                this.websocket.onclose = () => {
                    clearTimeout(timeout);
                    this.connected = false;
                    
                    if (this.reconnectAttempts < this.maxReconnectAttempts) {
                        this.reconnectAttempts++;
                        console.log(`Connection closed, attempting reconnect ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
                        setTimeout(() => this.connect(host, port, authToken), 1000 * this.reconnectAttempts);
                    }
                };
            });
            
        } catch (error) {
            throw new Error(`Connection failed: ${error.message}`);
        }
    }
    
    async sendRequest(method, params = {}) {
        /** Send request with error handling. */
        if (!this.connected) {
            throw new Error('Not connected');
        }
        
        return new Promise((resolve, reject) => {
            const id = Math.floor(Math.random() * 1000000);
            const request = {
                jsonrpc: '2.0',
                id: id,
                method: method,
                params: params
            };
            
            const timeout = setTimeout(() => {
                reject(new Error(`Request timeout: ${method}`));
            }, 30000);
            
            const originalOnMessage = this.websocket.onmessage;
            this.websocket.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    if (data.id === id) {
                        clearTimeout(timeout);
                        this.websocket.onmessage = originalOnMessage;
                        
                        if (data.error) {
                            reject(new Error(data.error.message || 'Request failed'));
                        } else {
                            resolve(data.result);
                        }
                    }
                } catch (error) {
                    clearTimeout(timeout);
                    this.websocket.onmessage = originalOnMessage;
                    reject(new Error(`Invalid response: ${error.message}`));
                }
            };
            
            try {
                this.websocket.send(JSON.stringify(request));
            } catch (error) {
                clearTimeout(timeout);
                this.websocket.onmessage = originalOnMessage;
                reject(new Error(`Send failed: ${error.message}`));
            }
        });
    }
    
    async safeOperation(operation, ...args) {
        /** Execute operation with error handling. */
        try {
            return await operation(...args);
        } catch (error) {
            console.error(`Operation failed: ${error.message}`);
            
            if (error.message.includes('Authentication failed')) {
                // Handle authentication errors
                throw new Error('Authentication failed - please check your token');
            } else if (error.message.includes('Camera not found')) {
                // Handle camera not found
                return null;
            } else if (error.message.includes('Connection')) {
                // Handle connection errors
                throw new Error('Connection lost - please reconnect');
            } else {
                // Handle other errors
                throw error;
            }
        }
    }
}

// Example usage
async function main() {
    const client = new BrowserCameraClient();
    
    try {
        await client.connect('localhost', 8080, 'your_jwt_token');
        
        // Safe operations
        const cameras = await client.safeOperation(
            () => client.sendRequest('get_camera_list')
        );
        
        if (cameras && cameras.length > 0) {
            const snapshot = await client.safeOperation(
                () => client.sendRequest('take_snapshot', { device: cameras[0].device_path })
            );
            console.log('Snapshot taken:', snapshot);
        }
        
    } catch (error) {
        console.error('Error:', error.message);
        // Show user-friendly error message
        showErrorMessage(error.message);
    }
}

function showErrorMessage(message) {
    // Display error message to user
    const errorDiv = document.getElementById('error-message');
    if (errorDiv) {
        errorDiv.textContent = message;
        errorDiv.style.display = 'block';
    }
}
```

## Logging and Monitoring

### Structured Error Logging

```python
import logging
import json
from datetime import datetime
from typing import Dict, Any

class StructuredErrorLogger:
    def __init__(self, logger_name: str = "camera_service_client"):
        self.logger = logging.getLogger(logger_name)
        self.setup_logging()
    
    def setup_logging(self):
        """Setup structured logging."""
        formatter = logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
        
        handler = logging.StreamHandler()
        handler.setFormatter(formatter)
        self.logger.addHandler(handler)
        self.logger.setLevel(logging.INFO)
    
    def log_error(self, error: Exception, context: Dict[str, Any] = None):
        """Log error with structured context."""
        error_data = {
            "timestamp": datetime.utcnow().isoformat(),
            "error_type": type(error).__name__,
            "error_message": str(error),
            "context": context or {}
        }
        
        self.logger.error(json.dumps(error_data))
    
    def log_operation(self, operation: str, success: bool, duration: float, context: Dict[str, Any] = None):
        """Log operation with metrics."""
        operation_data = {
            "timestamp": datetime.utcnow().isoformat(),
            "operation": operation,
            "success": success,
            "duration_ms": duration * 1000,
            "context": context or {}
        }
        
        if success:
            self.logger.info(json.dumps(operation_data))
        else:
            self.logger.warning(json.dumps(operation_data))

# Example usage
logger = StructuredErrorLogger()

try:
    # Perform operation
    start_time = time.time()
    result = await camera_operation()
    duration = time.time() - start_time
    
    logger.log_operation("camera_operation", True, duration, {
        "device_path": "/dev/video0",
        "result": result
    })
    
except Exception as e:
    duration = time.time() - start_time
    logger.log_operation("camera_operation", False, duration, {
        "device_path": "/dev/video0",
        "error": str(e)
    })
    
    logger.log_error(e, {
        "operation": "camera_operation",
        "device_path": "/dev/video0"
    })
```

### Error Metrics Collection

```python
from collections import defaultdict
import time
from typing import Dict, List

class ErrorMetricsCollector:
    def __init__(self):
        self.error_counts = defaultdict(int)
        self.operation_times = defaultdict(list)
        self.error_timestamps = defaultdict(list)
    
    def record_error(self, error_type: str, operation: str = None):
        """Record error occurrence."""
        key = f"{error_type}:{operation}" if operation else error_type
        self.error_counts[key] += 1
        self.error_timestamps[key].append(time.time())
    
    def record_operation_time(self, operation: str, duration: float):
        """Record operation duration."""
        self.operation_times[operation].append(duration)
    
    def get_error_rate(self, error_type: str, operation: str = None, window_minutes: int = 60) -> float:
        """Calculate error rate in time window."""
        key = f"{error_type}:{operation}" if operation else error_type
        now = time.time()
        window_start = now - (window_minutes * 60)
        
        recent_errors = sum(1 for ts in self.error_timestamps[key] if ts > window_start)
        total_operations = len([t for t in self.operation_times.get(operation, []) if t > window_start])
        
        return recent_errors / total_operations if total_operations > 0 else 0
    
    def get_average_operation_time(self, operation: str) -> float:
        """Calculate average operation time."""
        times = self.operation_times.get(operation, [])
        return sum(times) / len(times) if times else 0
    
    def generate_report(self) -> Dict[str, Any]:
        """Generate error metrics report."""
        return {
            "error_counts": dict(self.error_counts),
            "average_times": {
                op: self.get_average_operation_time(op)
                for op in self.operation_times.keys()
            },
            "error_rates": {
                error: self.get_error_rate(error)
                for error in self.error_counts.keys()
            }
        }

# Example usage
metrics = ErrorMetricsCollector()

# Record metrics during operations
try:
    start_time = time.time()
    result = await camera_operation()
    duration = time.time() - start_time
    
    metrics.record_operation_time("camera_operation", duration)
    
except AuthenticationError as e:
    metrics.record_error("AuthenticationError", "camera_operation")
    raise
except ConnectionError as e:
    metrics.record_error("ConnectionError", "camera_operation")
    raise

# Generate report
report = metrics.generate_report()
print("Error metrics report:", json.dumps(report, indent=2))
```

## Recovery Strategies

### 1. Automatic Reconnection

```python
class AutoReconnectingClient:
    def __init__(self, **kwargs):
        self.client = CameraClient(**kwargs)
        self.reconnect_delay = 1.0
        self.max_reconnect_delay = 60.0
        self.current_delay = self.reconnect_delay
    
    async def ensure_connected(self):
        """Ensure client is connected, reconnect if necessary."""
        if not self.client.connected:
            await self._reconnect()
    
    async def _reconnect(self):
        """Reconnect with exponential backoff."""
        while not self.client.connected:
            try:
                await self.client.connect()
                self.current_delay = self.reconnect_delay  # Reset delay on success
                print("Reconnected successfully")
            except Exception as e:
                print(f"Reconnection failed: {e}")
                await asyncio.sleep(self.current_delay)
                self.current_delay = min(self.current_delay * 2, self.max_reconnect_delay)
    
    async def safe_operation(self, operation, *args, **kwargs):
        """Execute operation with automatic reconnection."""
        try:
            await self.ensure_connected()
            return await operation(*args, **kwargs)
        except ConnectionError:
            await self._reconnect()
            await self.ensure_connected()
            return await operation(*args, **kwargs)
```

### 2. Token Refresh

```python
class TokenRefreshingClient:
    def __init__(self, **kwargs):
        self.client = CameraClient(**kwargs)
        self.token_refresh_callback = None
    
    def set_token_refresh_callback(self, callback):
        """Set callback for token refresh."""
        self.token_refresh_callback = callback
    
    async def safe_operation(self, operation, *args, **kwargs):
        """Execute operation with token refresh."""
        try:
            return await operation(*args, **kwargs)
        except AuthenticationError:
            if self.token_refresh_callback:
                new_token = await self.token_refresh_callback()
                self.client.auth_token = new_token
                await self.client._authenticate()
                return await operation(*args, **kwargs)
            else:
                raise

# Example usage
async def refresh_token():
    """Refresh JWT token."""
    # Implement token refresh logic
    return "new_jwt_token"

client = TokenRefreshingClient(
    host="localhost",
    port=8080,
    auth_type="jwt",
    auth_token="initial_token"
)
client.set_token_refresh_callback(refresh_token)
```

### 3. Fallback Operations

```python
class FallbackCameraClient:
    def __init__(self, **kwargs):
        self.primary_client = CameraClient(**kwargs)
        self.fallback_client = None  # Could be a different service
        self.use_fallback = False
    
    async def get_camera_list(self):
        """Get camera list with fallback."""
        try:
            if not self.use_fallback:
                return await self.primary_client.get_camera_list()
        except Exception as e:
            print(f"Primary client failed: {e}")
            self.use_fallback = True
        
        # Fallback to cached or simplified data
        return await self._get_camera_list_fallback()
    
    async def _get_camera_list_fallback(self):
        """Fallback camera list implementation."""
        # Return cached or simplified camera list
        return [
            CameraInfo(
                device_path="/dev/video0",
                name="Camera 1 (Cached)",
                capabilities=["snapshot"],
                status="unknown"
            )
        ]
```

## Testing Error Handling

### Error Simulation

```python
import random
import asyncio

class ErrorSimulator:
    def __init__(self, error_rate: float = 0.3):
        self.error_rate = error_rate
        self.error_types = [
            ConnectionError("Simulated connection error"),
            AuthenticationError("Simulated authentication error"),
            CameraNotFoundError("Simulated camera not found"),
            MediaMTXError("Simulated MediaMTX error")
        ]
    
    async def simulate_operation(self, operation, *args, **kwargs):
        """Simulate operation with random errors."""
        if random.random() < self.error_rate:
            error = random.choice(self.error_types)
            raise error
        
        return await operation(*args, **kwargs)

# Example usage
simulator = ErrorSimulator(error_rate=0.5)

async def test_error_handling():
    client = RobustCameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="test_token"
    )
    
    try:
        await client.connect_with_retry()
        
        # Test with error simulation
        result = await client.safe_camera_operation(
            lambda: simulator.simulate_operation(client.client.get_camera_list)
        )
        
        print(f"Operation result: {result}")
        
    except Exception as e:
        print(f"Test completed with error: {e}")

asyncio.run(test_error_handling())
```

## Support

For additional support and questions:

- **Documentation**: See `docs/security/CLIENT_AUTHENTICATION_GUIDE.md`
- **Examples**: Check `examples/` directory for working examples
- **Issues**: Report problems via GitHub Issues
- **Email**: Contact team@mediamtx-camera-service.com
