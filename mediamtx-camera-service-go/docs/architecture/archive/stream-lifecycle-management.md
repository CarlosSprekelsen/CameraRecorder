# Stream Lifecycle Management Architecture

**⚠️ ARCHIVED DOCUMENT**  
This document has been superseded by the consolidated **[Go Architecture Guide](../go-architecture-guide.md)**.  
The Go implementation patterns and code examples for stream lifecycle management are now available in the main guide.

---

**Version**: 1.0  
**Date**: 2025-01-27  
**Status**: IMPLEMENTATION COMPLETED  

## Overview

Stream lifecycle management is critical for ensuring reliable recording operations while maintaining power efficiency. This document defines the architecture for managing MediaMTX stream activation, deactivation, and compatibility with file rotation.

## Problem Statement

### Current Issues
1. **File Rotation Incompatibility**: MediaMTX streams auto-close after 10 seconds of inactivity, breaking long recordings during file rotation
2. **Stream Lifecycle Conflicts**: Different use cases (recording, viewing, snapshots) have conflicting lifecycle requirements
3. **Power Efficiency**: Streams should not run unnecessarily when not needed
4. **Recording Reliability**: Long recordings must not be interrupted by stream lifecycle events

### Requirements
- **REQ-STREAM-001**: Streams must remain active during file rotation (30-minute intervals)
- **REQ-STREAM-002**: Different lifecycle policies for different use cases
- **REQ-STREAM-003**: Power-efficient operation with on-demand activation
- **REQ-STREAM-004**: Manual control over stream lifecycle for recording scenarios

## Architecture Design

### Stream Lifecycle Types

#### 1. Recording Streams
- **Purpose**: Long-duration video recording with file rotation
- **Lifecycle**: Manual start/stop, no auto-close
- **MediaMTX Settings**:
  ```yaml
  runOnDemandCloseAfter: 0s  # Never auto-close
  runOnDemandRestart: yes
  runOnDemandStartTimeout: 10s
  ```
- **Use Cases**: Continuous recording, timed recordings, event-triggered recording

#### 2. Viewing Streams
- **Purpose**: Live stream viewing for monitoring
- **Lifecycle**: Auto-close after inactivity
- **MediaMTX Settings**:
  ```yaml
  runOnDemandCloseAfter: 300s  # 5 minutes after last viewer
  runOnDemandRestart: yes
  runOnDemandStartTimeout: 10s
  ```
- **Use Cases**: Live monitoring, web interface viewing

#### 3. Snapshot Streams
- **Purpose**: Quick photo capture
- **Lifecycle**: Immediate activation/deactivation
- **MediaMTX Settings**:
  ```yaml
  runOnDemandCloseAfter: 60s  # 1 minute after capture
  runOnDemandRestart: no
  runOnDemandStartTimeout: 5s
  ```
- **Use Cases**: Photo capture, motion detection snapshots

### Stream Stop Conditions

#### Manual Stop (User-Initiated)
- User explicitly stops recording
- User closes viewing session
- Service shutdown

#### Automatic Stop (System-Initiated)
- Recording completion (timed recordings)
- Error recovery and restart
- Service health monitoring detects issues

#### Never Stop During
- File rotation events
- Recording session continuity
- Active viewing sessions

### File Rotation Compatibility

#### Problem
MediaMTX's default `runOnDemandCloseAfter: 10s` causes streams to stop during file rotation, breaking recording continuity.

#### Solution
- **Recording Streams**: `runOnDemandCloseAfter: 0s` (never auto-close)
- **File Rotation**: Occurs every 30 minutes without affecting stream
- **Continuity**: Recording continues seamlessly across rotation boundaries

## Implementation Strategy

### Phase 1: MediaMTX Path Configuration
- Dynamic path configuration based on use case
- Runtime modification of MediaMTX settings
- Fallback to static configuration if dynamic fails

### Phase 2: Stream Lifecycle Manager
- Centralized stream lifecycle management
- Use case-specific stream activation
- Health monitoring and recovery

### Phase 3: Integration with Recording System
- Seamless integration with file rotation
- Recording session tracking
- Error recovery and restart

## Component Design

### StreamLifecycleManager Class

```python
class StreamLifecycleManager:
    """
    Manages stream lifecycle for different use cases.
    """
    
    async def start_recording_stream(self, device_path: str) -> bool:
        """Start stream optimized for recording with file rotation."""
        
    async def start_viewing_stream(self, device_path: str) -> bool:
        """Start stream optimized for live viewing."""
        
    async def start_snapshot_stream(self, device_path: str) -> bool:
        """Start stream optimized for quick snapshot capture."""
        
    async def stop_stream(self, device_path: str, reason: str) -> bool:
        """Stop stream with proper cleanup and logging."""
        
    async def monitor_stream_health(self, device_path: str) -> bool:
        """Monitor stream health during long operations."""
        
    async def configure_mediamtx_path(self, device_path: str, use_case: str) -> bool:
        """Configure MediaMTX path with appropriate settings."""
```

### MediaMTX Path Configuration

#### Recording Configuration
```yaml
paths:
  camera0:
    runOnDemand: "ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp rtsp://localhost:8554/camera0"
    runOnDemandRestart: yes
    runOnDemandStartTimeout: 10s
    runOnDemandCloseAfter: 0s  # Never auto-close for recording
    runOnUnDemand: ""
```

#### Viewing Configuration
```yaml
paths:
  camera0:
    runOnDemand: "ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp rtsp://localhost:8554/camera0"
    runOnDemandRestart: yes
    runOnDemandStartTimeout: 10s
    runOnDemandCloseAfter: 300s  # 5 minutes after last viewer
    runOnUnDemand: ""
```

## Configuration Management

### Development Configuration
- Configuration files in `config/` directory
- Environment-specific settings
- Runtime configuration updates

### Deployment Configuration
- Production settings in `/opt/mediamtx/config/`
- Service-managed configuration updates
- Fallback to static configuration

## Error Handling and Recovery

### Stream Failure Scenarios
1. **FFmpeg Process Crash**: Restart with exponential backoff
2. **MediaMTX Unreachable**: Circuit breaker pattern
3. **Device Access Issues**: Fallback to alternative methods
4. **Configuration Errors**: Use default settings

### Recovery Strategies
1. **Automatic Restart**: For transient failures
2. **Manual Intervention**: For persistent issues
3. **Fallback Methods**: Direct camera access if MediaMTX fails
4. **Health Monitoring**: Proactive detection of issues

## Performance Considerations

### Response Times
- **Recording Start**: < 3 seconds (with stream activation)
- **Viewing Start**: < 2 seconds (stream already active)
- **Snapshot Capture**: < 1 second (direct camera access)

### Resource Usage
- **Memory**: Minimal overhead for lifecycle management
- **CPU**: FFmpeg process only when streams active
- **Network**: RTSP traffic only during active sessions

## Monitoring and Observability

### Metrics
- Stream activation/deactivation times
- Recording session durations
- File rotation success rates
- Error rates and recovery times

### Logging
- Stream lifecycle events
- Configuration changes
- Error conditions and recovery
- Performance metrics

## Future Enhancements

### Planned Features
1. **Adaptive Timeouts**: Dynamic timeout adjustment based on usage patterns
2. **Stream Pooling**: Reuse active streams for multiple operations
3. **Predictive Activation**: Start streams before they're needed
4. **Advanced Health Monitoring**: ML-based failure prediction

### API Extensions
1. **Stream Status API**: Real-time stream health information
2. **Lifecycle Control API**: Manual stream lifecycle management
3. **Configuration API**: Runtime configuration updates
4. **Metrics API**: Performance and health metrics

## Implementation Notes

### Development Guidelines
1. **Test All Use Cases**: Recording, viewing, and snapshot scenarios
2. **Validate File Rotation**: Ensure no interruption during rotation
3. **Monitor Performance**: Track response times and resource usage
4. **Document Changes**: Update this document as implementation progresses

### Testing Strategy
1. **Unit Tests**: Individual component testing
2. **Integration Tests**: MediaMTX integration testing
3. **End-to-End Tests**: Complete workflow testing
4. **Performance Tests**: Response time and resource usage testing

### Deployment Considerations
1. **Configuration Migration**: Handle existing MediaMTX configurations
2. **Backward Compatibility**: Support existing API clients
3. **Rollback Strategy**: Ability to revert to previous implementation
4. **Monitoring Setup**: Ensure proper observability in production

## Conclusion

This architecture provides a comprehensive solution for stream lifecycle management that addresses the current issues while supporting future enhancements. The implementation will be phased to ensure stability and proper testing at each stage.

**Implementation Status**: 
- ✅ Phase 1: MediaMTX Path Configuration - COMPLETED
- ✅ Phase 2: Stream Lifecycle Manager - COMPLETED  
- ✅ Phase 3: Integration with Recording System - COMPLETED
- ✅ Unit Tests - COMPLETED

**Next Steps**: Deploy and test in development environment to validate file rotation compatibility.
