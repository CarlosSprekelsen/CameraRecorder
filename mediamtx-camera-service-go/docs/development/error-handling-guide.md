# Error Handling Guide

**Version:** 1.0.0  
**Date:** 2025-09-28  
**Purpose:** Comprehensive error handling patterns and recovery strategies for MediaMTX Camera Service

## **🎯 OVERVIEW**

This guide provides comprehensive error handling patterns, recovery strategies, and troubleshooting procedures for the MediaMTX Camera Service API.

## **📋 ERROR CATEGORIES**

### **JSON-RPC Standard Errors**

| Code | Name | Description | Client Action |
|------|------|-------------|---------------|
| **-32700** | Parse Error | Invalid JSON | Fix JSON format |
| **-32600** | Invalid Request | Malformed request | Fix request structure |
| **-32601** | Method Not Found | Unknown method | Use correct method name |
| **-32602** | Invalid Params | Invalid parameters | Fix parameter values |
| **-32603** | Internal Error | Server internal error | Retry or contact support |

### **Custom Application Errors (-32000 to -32099)**

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| **-32001** | Authentication Failed | Invalid or expired token | Re-authenticate |
| **-32002** | Permission Denied | Insufficient permissions | Check user role |
| **-32010** | Camera Not Found | Camera device not available | Check camera status |
| **-32011** | Camera Busy | Camera in use by another client | Wait and retry |
| **-32012** | Camera Offline | Camera device disconnected | Check hardware connection |
| **-32020** | Recording Failed | Recording operation failed | Check storage space |
| **-32021** | Recording Busy | Recording already in progress | Stop current recording |
| **-32030** | Stream Not Found | Stream not available | Check stream status |
| **-32031** | Stream Busy | Stream already active | Stop current stream |
| **-32040** | Storage Full | Insufficient storage space | Free up space |
| **-32041** | File Not Found | Requested file doesn't exist | Check file path |
| **-32050** | MediaMTX Error | MediaMTX service error | Check MediaMTX status |
| **-32051** | MediaMTX Timeout | MediaMTX operation timeout | Retry operation |
| **-32060** | Configuration Error | Invalid configuration | Check config settings |
| **-32070** | Network Error | Network connectivity issue | Check network connection |
| **-32080** | Resource Exhausted | System resource limit reached | Reduce concurrent operations |
| **-32090** | Validation Error | Input validation failed | Fix input parameters |

## **🔄 ERROR RECOVERY PATTERNS**

### **Pattern 1: Authentication Errors**

#### **Error Response**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Authentication failed or token expired",
    "data": {
      "details": "JWT token has expired",
      "reason": "token_expired",
      "suggestion": "Re-authenticate to get a new token"
    }
  },
  "id": 1
}
```

#### **Recovery Strategy**
```javascript
async function handleAuthError(error) {
    if (error.code === -32001) {
        // Re-authenticate
        const newToken = await authenticate(username, password);
        client.setAuthToken(newToken);
        
        // Retry original request
        return retryOriginalRequest();
    }
}
```

### **Pattern 2: Camera Busy Errors**

#### **Error Response**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32011,
    "message": "Camera is busy",
    "data": {
      "details": "Camera /dev/video0 is currently in use",
      "reason": "device_busy",
      "suggestion": "Wait for current operation to complete",
      "retry_after": 5
    }
  },
  "id": 1
}
```

#### **Recovery Strategy**
```javascript
async function handleCameraBusyError(error) {
    if (error.code === -32011) {
        const retryAfter = error.data.retry_after || 5;
        
        // Wait and retry
        await sleep(retryAfter * 1000);
        return retryOriginalRequest();
    }
}
```

### **Pattern 3: Storage Errors**

#### **Error Response**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32040,
    "message": "Storage full",
    "data": {
      "details": "Insufficient storage space for recording",
      "reason": "storage_full",
      "suggestion": "Free up space or use different storage location",
      "available_space": "1.2GB",
      "required_space": "2.5GB"
    }
  },
  "id": 1
}
```

#### **Recovery Strategy**
```javascript
async function handleStorageError(error) {
    if (error.code === -32040) {
        // Check available space
        const available = error.data.available_space;
        const required = error.data.required_space;
        
        if (available < required) {
            // Clean up old files
            await cleanupOldRecordings();
            
            // Retry operation
            return retryOriginalRequest();
        }
    }
}
```

## **🛠️ ERROR HANDLING IMPLEMENTATION**

### **Client-Side Error Handling**

#### **Generic Error Handler**
```javascript
class CameraServiceClient {
    async callMethod(method, params) {
        try {
            const response = await this.sendRequest(method, params);
            
            if (response.error) {
                return await this.handleError(response.error);
            }
            
            return response.result;
        } catch (error) {
            return await this.handleNetworkError(error);
        }
    }
    
    async handleError(error) {
        switch (error.code) {
            case -32001:
                return await this.handleAuthError(error);
            case -32011:
                return await this.handleCameraBusyError(error);
            case -32040:
                return await this.handleStorageError(error);
            default:
                throw new Error(`Unhandled error: ${error.message}`);
        }
    }
}
```

#### **Retry Logic with Exponential Backoff**
```javascript
async function retryWithBackoff(operation, maxRetries = 3) {
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            return await operation();
        } catch (error) {
            if (attempt === maxRetries) {
                throw error;
            }
            
            // Exponential backoff: 1s, 2s, 4s
            const delay = Math.pow(2, attempt - 1) * 1000;
            await sleep(delay);
        }
    }
}
```

### **Server-Side Error Handling**

#### **Error Context and Logging**
```go
func (s *WebSocketServer) handleError(err error, client *ClientConnection, method string) *JsonRpcResponse {
    // Log error with context
    s.logger.WithFields(logging.Fields{
        "client_id": client.ClientID,
        "method":    method,
        "error":     err.Error(),
        "timestamp": time.Now(),
    }).Error("Method execution failed")
    
    // Determine error code and message
    errorCode, errorMessage := s.classifyError(err)
    
    return &JsonRpcResponse{
        JSONRPC: "2.0",
        Error: &JsonRpcError{
            Code:    errorCode,
            Message: errorMessage,
            Data:    s.buildErrorData(err),
        },
    }
}
```

#### **Error Classification**
```go
func (s *WebSocketServer) classifyError(err error) (int, string) {
    switch {
    case errors.Is(err, ErrAuthenticationFailed):
        return -32001, "Authentication failed or token expired"
    case errors.Is(err, ErrCameraNotFound):
        return -32010, "Camera not found"
    case errors.Is(err, ErrCameraBusy):
        return -32011, "Camera is busy"
    case errors.Is(err, ErrStorageFull):
        return -32040, "Storage full"
    default:
        return -32603, "Internal server error"
    }
}
```

## **🔍 TROUBLESHOOTING PROCEDURES**

### **Common Error Scenarios**

#### **Scenario 1: Authentication Issues**

**Symptoms:**
- All requests return `-32001 Authentication failed`
- Client can't establish connection

**Diagnosis:**
```bash
# Check JWT token validity
echo "your-jwt-token" | base64 -d

# Check server logs
journalctl -u camera-service | grep -i auth

# Test authentication endpoint
curl -X POST http://localhost:8002/ws \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"authenticate","params":{"auth_token":"test"}}'
```

**Resolution:**
1. Verify JWT token is valid and not expired
2. Check server configuration for JWT settings
3. Re-authenticate with valid credentials

#### **Scenario 2: Camera Access Issues**

**Symptoms:**
- `-32010 Camera not found` errors
- `-32011 Camera busy` errors

**Diagnosis:**
```bash
# Check available cameras
ls -la /dev/video*

# Check camera permissions
groups $USER | grep video

# Test camera access
v4l2-ctl --list-devices
```

**Resolution:**
1. Ensure cameras are connected and detected
2. Add user to video group: `sudo usermod -a -G video $USER`
3. Check for conflicting camera access

#### **Scenario 3: Storage Issues**

**Symptoms:**
- `-32040 Storage full` errors
- Recording operations fail

**Diagnosis:**
```bash
# Check disk space
df -h /opt/camera-service/recordings

# Check storage configuration
grep -r "recordings_path" /opt/camera-service/config/

# Check file permissions
ls -la /opt/camera-service/recordings/
```

**Resolution:**
1. Free up disk space
2. Adjust storage configuration
3. Check file permissions

### **Error Monitoring and Alerting**

#### **Health Check Integration**
```bash
# Check service health
curl http://localhost:8003/health

# Check detailed health
curl http://localhost:8003/health/detailed

# Check error metrics
curl http://localhost:8003/metrics
```

#### **Log Analysis**
```bash
# Monitor error logs
journalctl -u camera-service -f | grep -i error

# Analyze error patterns
journalctl -u camera-service | grep -E "error|failed" | tail -100

# Check specific error codes
journalctl -u camera-service | grep "32001\|32010\|32011"
```

## **📊 ERROR METRICS AND MONITORING**

### **Error Rate Tracking**
```go
type ErrorMetrics struct {
    TotalErrors     int64            `json:"total_errors"`
    ErrorByCode     map[int]int64    `json:"error_by_code"`
    ErrorByMethod   map[string]int64 `json:"error_by_method"`
    LastErrorTime   time.Time        `json:"last_error_time"`
}
```

### **Error Alerting**
```yaml
# Alert configuration
alerts:
  high_error_rate:
    threshold: 10  # errors per minute
    duration: 5m
    action: notify_admin
    
  authentication_failures:
    threshold: 5   # failures per minute
    duration: 2m
    action: security_alert
```

## **🎯 BEST PRACTICES**

### **For Client Developers**

1. **Always Handle Errors:** Never ignore error responses
2. **Implement Retry Logic:** Use exponential backoff for transient errors
3. **Log Errors:** Include context for debugging
4. **Graceful Degradation:** Provide fallback behavior
5. **User-Friendly Messages:** Translate technical errors to user messages

### **For Server Developers**

1. **Consistent Error Format:** Use standard error response structure
2. **Detailed Error Data:** Include helpful debugging information
3. **Proper Error Codes:** Use appropriate JSON-RPC error codes
4. **Error Logging:** Log errors with sufficient context
5. **Error Recovery:** Implement automatic recovery where possible

### **For Operations Teams**

1. **Monitor Error Rates:** Set up alerting for high error rates
2. **Analyze Error Patterns:** Identify common failure modes
3. **Capacity Planning:** Monitor resource usage trends
4. **Documentation:** Keep troubleshooting guides updated
5. **Training:** Ensure team knows error handling procedures

## **📚 REFERENCES**

- [JSON-RPC 2.0 Error Handling](https://www.jsonrpc.org/specification#error_object)
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [Error Handling Best Practices](https://docs.microsoft.com/en-us/azure/architecture/patterns/retry)
- [WebSocket Error Handling](https://tools.ietf.org/html/rfc6455#section-7.4)

---

**This error handling guide ensures robust, recoverable error handling across the MediaMTX Camera Service API, providing clear recovery strategies and troubleshooting procedures.**
