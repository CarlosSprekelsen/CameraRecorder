# Multi-Tier Snapshot Capture Architecture

**⚠️ ARCHIVED DOCUMENT**  
This document has been superseded by the consolidated **[Go Architecture Guide](../go-architecture-guide.md)**.  
The Go implementation patterns and code examples for multi-tier snapshot capture are now available in the main guide.

---

**Version:** 1.0  
**Authors:** Project Team  
**Date:** 2025-08-19  
**Status:** Approved  
**Related Epic/Story:** E1/S3 - Camera Snapshot Functionality  

**Purpose:**  
Document the multi-tier snapshot capture architecture that provides optimal user experience while maintaining power efficiency through on-demand stream activation.

---

## Overview

The MediaMTX Camera Service implements a sophisticated **multi-tier snapshot capture system** that balances speed, reliability, and power efficiency. This architecture ensures fast photo capture when possible while gracefully handling scenarios where RTSP streams need to be activated on-demand.

## Problem Statement

### User Experience Requirements
- **Immediate Response**: Photo capture should be nearly instant when possible
- **Consistent Behavior**: Predictable performance across different system states
- **Reliable Fallback**: Capture should work even when primary method fails
- **Power Efficiency**: No unnecessary resource consumption during idle periods

### Technical Challenges
- **On-Demand Activation**: RTSP streams are only active when needed (power efficiency)
- **Variable Latency**: Stream activation adds 1-3 seconds delay for first capture
- **System Reliability**: Multiple failure points (MediaMTX, FFmpeg, camera hardware)
- **Performance Tuning**: Need configurable timeouts for different environments

## Multi-Tier Architecture

### Tier 1: Immediate RTSP Capture (Fastest Path)
**Purpose**: Provide instant response when RTSP stream is already active

**Process**:
1. Quick check if RTSP stream is ready (`ready: true`, `source: {...}`)
2. If ready, capture immediately using FFmpeg from RTSP stream
3. Return snapshot with minimal latency

**Performance Characteristics**:
- **Response Time**: < 0.5 seconds (immediate)
- **Use Case**: Stream already running (during recording, active viewing)
- **Success Rate**: High when stream is active
- **Resource Usage**: Minimal (reuses existing stream)

**Configuration**:
```yaml
performance:
  snapshot_tiers:
    tier1_rtsp_ready_check_timeout: 1.0    # Quick RTSP readiness check
    immediate_response_threshold: 0.5       # Consider "immediate" if under this
```

### Tier 2: Quick Stream Activation (Balanced Path)
**Purpose**: Activate RTSP stream on-demand and capture with acceptable latency

**Process**:
1. Trigger on-demand activation via MediaMTX `runOnDemand`
2. Wait for FFmpeg process to start and stream to become ready
3. Capture from RTSP once activated
4. Return snapshot with balanced latency

**Performance Characteristics**:
- **Response Time**: 1-3 seconds (acceptable)
- **Use Case**: First snapshot after idle period, power-efficient operation
- **Success Rate**: High with proper timeout configuration
- **Resource Usage**: Moderate (starts FFmpeg process)

**Configuration**:
```yaml
performance:
  snapshot_tiers:
    tier2_activation_timeout: 3.0          # Stream activation wait time
    tier2_activation_trigger_timeout: 1.0  # Activation trigger timeout
    acceptable_response_threshold: 2.0     # Consider "acceptable" if under this
```

### Tier 3: Direct Camera Capture (Fallback Path)
**Purpose**: Bypass MediaMTX entirely for reliable capture when RTSP fails

**Process**:
1. Extract device path from stream name (e.g., "camera0" → "/dev/video0")
2. Capture directly from camera using FFmpeg with v4l2 input
3. Bypass MediaMTX and RTSP entirely
4. Return snapshot with reliable but slower performance

**Performance Characteristics**:
- **Response Time**: 2-5 seconds (slower but reliable)
- **Use Case**: MediaMTX issues, network problems, emergency capture
- **Success Rate**: Very high (direct hardware access)
- **Resource Usage**: Low (no MediaMTX dependency)

**Configuration**:
```yaml
performance:
  snapshot_tiers:
    tier3_direct_capture_timeout: 5.0      # Direct camera capture timeout
    slow_response_threshold: 5.0           # Consider "slow" if over this
```

### Tier 4: Error Handling (Last Resort)
**Purpose**: Provide detailed error information when all methods fail

**Process**:
1. All capture methods have been attempted
2. Return comprehensive error information
3. Include timing data and methods attempted
4. Provide debugging information for troubleshooting

**Performance Characteristics**:
- **Response Time**: < 10 seconds (with error details)
- **Use Case**: System troubleshooting, edge case handling
- **Success Rate**: N/A (error reporting)
- **Resource Usage**: Minimal (error reporting only)

**Configuration**:
```yaml
performance:
  snapshot_tiers:
    total_operation_timeout: 10.0          # Maximum total operation time
```

## Configuration Options

### Performance Tuning Parameters

All timeouts and thresholds are configurable for performance tuning:

```yaml
performance:
  snapshot_tiers:
    # Tier 1: Immediate RTSP capture
    tier1_rtsp_ready_check_timeout: 1.0    # seconds - Quick RTSP readiness check
    
    # Tier 2: Quick stream activation
    tier2_activation_timeout: 3.0          # seconds - Time to wait for stream activation
    tier2_activation_trigger_timeout: 1.0  # seconds - Timeout for triggering activation
    
    # Tier 3: Direct camera capture
    tier3_direct_capture_timeout: 5.0      # seconds - Timeout for direct camera capture
    
    # Overall operation timeout
    total_operation_timeout: 10.0          # seconds - Maximum total time for snapshot operation
    
    # User experience thresholds
    immediate_response_threshold: 0.5      # seconds - Consider "immediate" if under this
    acceptable_response_threshold: 2.0     # seconds - Consider "acceptable" if under this
    slow_response_threshold: 5.0           # seconds - Consider "slow" if over this
```

### FFmpeg Configuration

FFmpeg timeouts are also configurable:

```yaml
ffmpeg:
  snapshot:
    process_creation_timeout: 5.0    # seconds - FFmpeg process creation timeout
    execution_timeout: 8.0           # seconds - FFmpeg execution timeout
    internal_timeout: 5000000        # microseconds - FFmpeg internal timeout
    retry_attempts: 2                # Number of retry attempts
    retry_delay: 1.0                 # seconds - Delay between retries
```

## Performance Characteristics

### Response Time Analysis

| Tier | Scenario | Expected Time | User Experience | Use Case |
|------|----------|---------------|-----------------|----------|
| 1 | RTSP Ready | < 0.5s | Immediate | Active streaming/recording |
| 2 | RTSP Activation | 1-3s | Acceptable | First capture after idle |
| 3 | Direct Camera | 2-5s | Slow but reliable | MediaMTX issues |
| 4 | Error | < 10s | Failed with details | System problems |

### Success Rate Analysis

| Tier | Success Rate | Failure Modes | Recovery |
|------|-------------|---------------|----------|
| 1 | 95%+ | Stream not ready | Falls back to Tier 2 |
| 2 | 90%+ | Activation timeout | Falls back to Tier 3 |
| 3 | 98%+ | Camera hardware issues | Falls back to Tier 4 |
| 4 | N/A | All methods failed | Error reporting |

### Resource Usage Analysis

| Tier | CPU Usage | Memory Usage | Network Usage | Power Impact |
|------|-----------|--------------|---------------|--------------|
| 1 | Low | Low | Low | Minimal |
| 2 | Medium | Medium | Low | Moderate |
| 3 | Low | Low | None | Low |
| 4 | Minimal | Minimal | None | Minimal |

## Implementation Details

### Method Signature

```python
async def take_snapshot(
    self, 
    stream_name: str, 
    filename: Optional[str] = None, 
    format: str = "jpg", 
    quality: int = 85
) -> Dict[str, Any]:
```

### Return Value Structure

```python
{
    "stream_name": str,           # Stream name (e.g., "camera0")
    "filename": str,              # Generated filename
    "status": str,                # "completed", "failed", or "timeout"
    "timestamp": str,             # ISO timestamp of capture
    "file_size": int,             # File size in bytes (0 if failed)
    "file_path": str,             # Full path to snapshot file
    "capture_method": str,        # "rtsp", "direct_camera", or "failed"
    "capture_time": float,        # Total capture time in seconds
    "tier_used": int,             # Which tier was successful (1, 2, 3, or 4)
    "user_experience": str,       # "immediate", "acceptable", "slow", or "failed"
    "error": str,                 # Error message if failed
    "capture_methods_tried": list # List of methods attempted
}
```

### Error Handling

Each tier includes comprehensive error handling:

1. **Timeout Handling**: Configurable timeouts with graceful fallback
2. **Process Cleanup**: Automatic cleanup of FFmpeg processes
3. **Error Reporting**: Detailed error messages with context
4. **Logging**: Structured logging with correlation IDs

## Power Efficiency Benefits

### On-Demand Activation
- **Idle State**: No FFmpeg processes running when not needed
- **Active State**: Processes start only when snapshot requested
- **Automatic Cleanup**: Failed processes are cleaned up automatically

### Resource Optimization
- **CPU Usage**: Minimal during idle periods
- **Memory Usage**: Reduced footprint when streams inactive
- **Battery Life**: Extended battery life for mobile/embedded deployments

### Scalability
- **Multiple Cameras**: Each camera activates independently
- **Resource Management**: No resource exhaustion with many cameras
- **Dynamic Scaling**: Processes start/stop based on demand

## Testing Strategy

### Performance Testing
- **Response Time**: Measure actual vs expected response times
- **Success Rate**: Validate success rates for each tier
- **Resource Usage**: Monitor CPU, memory, and power consumption
- **Timeout Validation**: Test timeout configurations

### User Experience Testing
- **Immediate Response**: Verify < 0.5s response when RTSP ready
- **Acceptable Response**: Verify 1-3s response for activation
- **Fallback Reliability**: Test direct camera capture reliability
- **Error Handling**: Validate error reporting and recovery

### Edge Case Testing
- **Network Issues**: Test behavior with MediaMTX unavailable
- **Camera Hardware**: Test with camera disconnection/reconnection
- **Resource Exhaustion**: Test with limited system resources
- **Concurrent Operations**: Test multiple simultaneous snapshots

## Configuration Recommendations

### Development Environment
```yaml
performance:
  snapshot_tiers:
    tier1_rtsp_ready_check_timeout: 0.5    # Faster for development
    tier2_activation_timeout: 2.0          # Shorter for testing
    tier3_direct_capture_timeout: 3.0      # Faster fallback
    immediate_response_threshold: 0.3       # Stricter for development
```

### Production Environment
```yaml
performance:
  snapshot_tiers:
    tier1_rtsp_ready_check_timeout: 1.0    # Standard production
    tier2_activation_timeout: 3.0          # Balanced for production
    tier3_direct_capture_timeout: 5.0      # Reliable fallback
    immediate_response_threshold: 0.5       # Standard for production
```

### High-Performance Environment
```yaml
performance:
  snapshot_tiers:
    tier1_rtsp_ready_check_timeout: 0.2    # Very fast check
    tier2_activation_timeout: 1.5          # Fast activation
    tier3_direct_capture_timeout: 2.0      # Fast fallback
    immediate_response_threshold: 0.2       # Very strict
```

## Troubleshooting

### Common Issues

#### Slow Response Times
- **Check Tier 1**: Verify RTSP stream readiness
- **Check Tier 2**: Monitor stream activation time
- **Check Tier 3**: Validate direct camera access
- **Adjust Timeouts**: Tune configuration for environment

#### High Failure Rates
- **Check MediaMTX**: Verify MediaMTX service health
- **Check Camera**: Validate camera hardware and permissions
- **Check FFmpeg**: Ensure FFmpeg is properly installed
- **Check Logs**: Review detailed error messages

#### Resource Exhaustion
- **Monitor Processes**: Check for orphaned FFmpeg processes
- **Review Configuration**: Adjust timeout and retry settings
- **Check Concurrent Operations**: Limit simultaneous snapshots
- **Monitor System Resources**: Check CPU, memory, and disk usage

### Debugging Commands

```bash
# Check MediaMTX paths
curl http://localhost:9997/v3/paths/list | jq

# Check specific path status
curl http://localhost:9997/v3/config/paths/get/camera0 | jq

# Monitor FFmpeg processes
ps aux | grep ffmpeg

# Check camera device permissions
ls -la /dev/video*

# Test direct camera access
ffmpeg -f v4l2 -i /dev/video0 -vframes 1 -y test.jpg
```

## Conclusion

The multi-tier snapshot capture architecture provides:

1. **Optimal User Experience**: Fast response when possible, graceful degradation when needed
2. **Power Efficiency**: On-demand activation with minimal resource usage
3. **Reliability**: Multiple fallback mechanisms ensure capture success
4. **Configurability**: All timeouts and thresholds are tunable for different environments
5. **Observability**: Comprehensive logging and error reporting for troubleshooting

This architecture successfully balances the competing requirements of speed, reliability, and power efficiency while providing a consistent and predictable user experience for photo capture operations.

---

**Related Documentation:**
- [On-Demand Stream Activation](../architecture/on-demand-stream-activation.md)
- [Performance Tuning Guide](../development/performance-tuning.md)
- [Configuration Reference](../configuration/reference.md)
