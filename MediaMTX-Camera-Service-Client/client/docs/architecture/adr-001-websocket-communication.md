# ADR-001: WebSocket Communication Protocol

**Status:** Accepted  
**Date:** January 2025  
**Deciders:** Development Team  
**Technical Story:** [Sprint-1] WebSocket Foundation

## Context

The MediaMTX Camera Service Client requires real-time bidirectional communication with the server for:
- Camera status updates
- Recording progress notifications  
- System health monitoring
- Command acknowledgments

## Decision

We will use **WebSocket with JSON-RPC 2.0** as the primary communication protocol.

## Rationale

### WebSocket Benefits
- **Real-time Communication:** Persistent connection enables instant updates
- **Bidirectional:** Server can push notifications without client polling
- **Low Latency:** Direct TCP connection without HTTP overhead
- **Browser Support:** Native WebSocket API in all modern browsers

### JSON-RPC 2.0 Benefits
- **Structured Messages:** Clear request/response format
- **Error Handling:** Standardized error codes and messages
- **Batch Operations:** Multiple requests in single message
- **Extensibility:** Easy to add new methods without breaking changes

## Alternatives Considered

### 1. HTTP REST with Polling
**Rejected because:**
- High latency for real-time updates
- Server resource intensive with frequent polling
- No server-initiated notifications

### 2. Server-Sent Events (SSE)
**Rejected because:**
- Unidirectional (server to client only)
- No request/response pattern for commands
- Limited browser support for bidirectional communication

### 3. WebRTC Data Channels
**Rejected because:**
- Overkill for simple command/status communication
- Complex setup and maintenance
- Not suitable for client-server architecture

## Implementation Details

### Connection Management
```typescript
class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  
  async connect(url: string): Promise<void> {
    // Implementation with exponential backoff
  }
}
```

### JSON-RPC 2.0 Format
```typescript
// Request
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "take_snapshot",
  "params": { "device": "camera-001", "filename": "snap.jpg" }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": { "success": true, "filename": "snap.jpg" }
}
```

### Error Handling
```typescript
// Error Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32601,
    "message": "Method not found"
  }
}
```

## Consequences

### Positive
- **Real-time Updates:** Instant camera status changes
- **Efficient Communication:** No unnecessary polling
- **Standardized Protocol:** JSON-RPC 2.0 is well-documented
- **Scalable:** Handles multiple concurrent connections

### Negative
- **Connection Management:** Requires reconnection logic
- **State Synchronization:** Need to handle connection drops
- **Browser Limitations:** Some corporate firewalls block WebSocket

### Risks
- **Connection Drops:** Network issues can interrupt communication
- **Memory Leaks:** Improper cleanup of event listeners
- **Security:** WebSocket connections need proper authentication

## Mitigation Strategies

### Connection Reliability
- Automatic reconnection with exponential backoff
- Heartbeat/ping mechanism to detect dead connections
- Connection state management in store

### Error Recovery
- Request timeout handling
- Retry logic for failed operations
- Graceful degradation when disconnected

### Security
- JWT token authentication over WebSocket
- Secure WebSocket (WSS) in production
- Input validation for all RPC parameters

## Monitoring

### Metrics to Track
- Connection uptime percentage
- Message latency (p95, p99)
- Reconnection frequency
- Error rates by method

### Alerts
- Connection failures > 5% in 5 minutes
- Message latency > 500ms p95
- Authentication failures

## Related ADRs
- [ADR-002: State Management](#) - How state is managed during connection drops
- [ADR-003: Authentication](#) - How authentication works over WebSocket

## References
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
- [MediaMTX Documentation](https://github.com/mediamtx/mediamtx)

---

**Last Updated:** January 2025  
**Review Date:** April 2025  
**Next Review:** July 2025
