# On-Demand Stream Activation

**⚠️ ARCHIVED DOCUMENT**  
This document has been superseded by the consolidated **[Go Architecture Guide](../go-architecture-guide.md)**.  
The Go implementation patterns and code examples for on-demand stream activation are now available in the main guide.

---

## Overview

The MediaMTX Camera Service implements **on-demand stream activation** to optimize power efficiency and resource usage. This document explains how stream activation works and how it differs from traditional auto-start approaches.

## Architecture Decision

**Decision**: Use on-demand stream activation with MediaMTX `runOnDemand` configuration instead of always-running FFmpeg processes.

**Rationale**: 
- **Power Efficiency**: No unnecessary FFmpeg processes running when not needed
- **Resource Optimization**: Reduced CPU and memory usage during idle periods
- **Scalability**: Better support for multiple cameras without resource exhaustion
- **Reliability**: Automatic restart of failed FFmpeg processes on demand

## How It Works

### 1. Camera Detection Phase
When a camera is detected:
- MediaMTX path is created with `runOnDemand` configuration
- **No FFmpeg process is started immediately**
- Path is configured but inactive (`ready: false`, `source: null`)

### 2. On-Demand Activation Phase
When a client requests camera operations:
- First access triggers FFmpeg process start via `runOnDemand`
- FFmpeg captures from camera and publishes to MediaMTX
- Stream becomes active (`ready: true`, `source: {...}`)
- Subsequent requests use the active stream

### 3. Configuration Example
```json
{
  "name": "camera0",
  "source": null,  // Initially no FFmpeg process
  "ready": false,  // Stream not active
  "runOnDemand": "ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:8554/camera0",
  "runOnDemandRestart": true
}
```

## Configuration Settings

### auto_start_streams Parameter

The `auto_start_streams` configuration parameter controls **path creation**, not **process activation**:

```yaml
camera:
  auto_start_streams: true  # Creates MediaMTX paths on camera detection
```

**What it does:**
- ✅ Creates MediaMTX paths when cameras are detected
- ✅ Configures `runOnDemand` commands for each camera
- ❌ Does NOT start FFmpeg processes immediately

**What it doesn't do:**
- ❌ Start FFmpeg processes at startup
- ❌ Keep streams always active
- ❌ Consume resources when not needed

## API Behavior

### Recording Operations
```python
# This will trigger on-demand activation
result = await server._method_start_recording({
    "device": "/dev/video0",
    "duration": 30
})
```

**Process:**
1. Validate stream readiness via MediaMTX API
2. If stream not ready, trigger `runOnDemand` activation
3. Wait for FFmpeg process to start and stream to become ready
4. Proceed with recording operation

### Snapshot Operations
```python
# This will trigger on-demand activation
result = await server._method_take_snapshot({
    "device": "/dev/video0",
    "format": "jpg"
})
```

**Process:**
1. Check if stream is active
2. If not active, trigger `runOnDemand` activation
3. Capture snapshot from active stream
4. Return snapshot data

## Error Handling

### Stream Not Ready Errors
When operations are attempted on inactive streams:

```python
# Expected behavior
try:
    result = await server._method_start_recording(params)
except MediaMTXError as e:
    if "Stream camera0 is not active and ready" in str(e):
        # This is expected for on-demand activation
        # The operation should trigger stream activation
```

### Power Efficiency Validation
Tests verify that no unnecessary processes are running:

```python
# Verify power efficiency
paths = await mediamtx_api.get_paths()
for path in paths:
    assert path['source'] is None, "No FFmpeg process should be running initially"
    assert not path['ready'], "Stream should not be ready initially"
```

## Testing Strategy

### Power Efficiency Tests
- Verify no FFmpeg processes running at startup
- Confirm streams are inactive initially
- Validate resource usage is minimal

### On-Demand Activation Tests
- Test that operations trigger stream activation
- Verify FFmpeg processes start when needed
- Confirm streams become ready after activation

### Error Handling Tests
- Test proper error messages for inactive streams
- Verify graceful handling of activation failures
- Confirm recovery from FFmpeg process failures

## Benefits

### 1. Power Efficiency
- **Idle State**: No FFmpeg processes running
- **Active State**: Processes only when needed
- **Resource Usage**: Minimal CPU/memory during idle periods

### 2. Scalability
- **Multiple Cameras**: Each camera activates independently
- **Resource Management**: No resource exhaustion with many cameras
- **Dynamic Scaling**: Processes start/stop based on demand

### 3. Reliability
- **Automatic Recovery**: Failed processes restart on next request
- **Isolation**: One camera failure doesn't affect others
- **Health Monitoring**: Process health tracked per camera

### 4. User Experience
- **Fast Response**: Streams activate quickly when needed
- **Transparent**: Users don't need to manage stream activation
- **Reliable**: Consistent behavior across different scenarios

## Migration from Auto-Start

If migrating from an auto-start approach:

### Before (Auto-Start)
```yaml
# All FFmpeg processes start immediately
camera:
  auto_start_streams: true  # Starts all processes at startup
```

### After (On-Demand)
```yaml
# Paths created, processes start on-demand
camera:
  auto_start_streams: true  # Creates paths, processes start when needed
```

### Code Changes Required
1. **Error Handling**: Handle "stream not ready" errors gracefully
2. **Activation Logic**: Wait for stream activation before operations
3. **Testing**: Update tests to expect on-demand behavior
4. **Documentation**: Clarify activation timing expectations

## Troubleshooting

### Common Issues

#### 1. "Stream not ready" errors
**Cause**: Operation attempted before stream activation
**Solution**: Implement proper activation waiting logic

#### 2. Slow first operation
**Cause**: FFmpeg process startup time
**Solution**: Acceptable for on-demand activation, consider pre-warming for critical use cases

#### 3. Process startup failures
**Cause**: Camera unavailable or FFmpeg configuration issues
**Solution**: Check camera availability and FFmpeg command configuration

### Debugging Commands
```bash
# Check MediaMTX paths
curl http://localhost:9997/v3/paths/list | jq

# Check specific path configuration
curl http://localhost:9997/v3/config/paths/get/camera0 | jq

# Monitor FFmpeg processes
ps aux | grep ffmpeg
```

## Conclusion

On-demand stream activation provides significant benefits in power efficiency, scalability, and reliability. The system maintains the same API interface while optimizing resource usage through intelligent process management.

**Key Takeaway**: `auto_start_streams: true` creates paths but doesn't start processes - they activate on-demand when operations are requested.
