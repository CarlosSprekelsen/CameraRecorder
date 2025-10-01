# BUG-022: Server Subscription Stats API Compliance Issue

## Summary
Server `get_subscription_stats` method returns incorrect field names that do not match the documented JSON-RPC API specification, causing integration tests to fail.

## Impact
- **API compliance violation** - Server not following documented JSON-RPC specification
- **Integration tests failing** - Tests correctly validate against documentation
- **Client compatibility broken** - Applications following docs will fail

## Root Cause Analysis

### Primary Issue: Server Field Name Mismatch

**Documentation (Ground Truth)**: `docs/api/json_rpc_methods.md` lines 2101-2118
**Server Implementation**: Returns wrong field names

### Expected vs Actual Response Structure

**✅ DOCUMENTED API (Correct)**:
```json
{
  "global_stats": {
    "total_subscriptions": 15,
    "active_clients": 3,
    "topic_counts": {
      "camera.connected": 2,
      "recording.start": 1
    }
  },
  "client_topics": ["camera.connected", "recording.start"],
  "client_id": "client_123"
}
```

**❌ SERVER RESPONSE (Incorrect)**:
```json
{
  "global_stats": {
    "active_subscriptions": 1,        // ❌ Should be "total_subscriptions"
    "topic_distribution": {...},      // ❌ Should be "topic_counts"
    "total_clients": 1,               // ❌ Should be "active_clients"
    "total_topics": 1                 // ❌ Not in specification
  }
}
```

### Test Evidence
```javascript
// Test correctly follows documentation
const stats = await apiClient.call('get_subscription_stats');
expect(stats.global_stats.total_subscriptions).toBeGreaterThanOrEqual(0); // ❌ Server returns undefined
expect(stats.client_topics).toContain('camera.connected'); // ✅ Works
expect(stats.client_topics).toContain('camera.disconnected'); // ✅ Works  
expect(typeof stats.client_id).toBe('string'); // ✅ Works
```

## Classification: SERVER IMPLEMENTATION BUG

**Ground Truth**: JSON-RPC API Documentation is authoritative
**Issue**: Server implementation does not match documented specification
**Resolution**: Fix server, not test

### Field Mapping Required:
- `active_subscriptions` → `total_subscriptions`
- `topic_distribution` → `topic_counts` 
- `total_clients` → `active_clients`
- `total_topics` → Remove (not in specification)

## Recommended Actions

1. **Fix server implementation** in `get_subscription_stats` method
2. **Update response structure** to match documentation exactly
3. **Keep test unchanged** - Test correctly validates API compliance
4. **Verify against documentation** - Ensure 100% compliance with JSON-RPC spec

## Files Affected

- **Server**: `get_subscription_stats` method implementation
- **Documentation**: Already correct (`docs/api/json_rpc_methods.md`)
- **Tests**: Already correct, following documentation

## Priority: HIGH

This is a critical API compliance issue that breaks the contract between server and client applications following the documented specification.

## Evidence

### Debug Output
```
Connected, authenticating...
Auth response: {"authenticated": true, "role": "admin", ...}
Authenticated, subscribing to events...
Subscribe response: {"subscribed": true, "topics": ["camera.connected"]}
Getting subscription stats...
Subscription stats response: {
  "global_stats": {
    "active_subscriptions": 1,        // ❌ Wrong field name
    "topic_distribution": {...},      // ❌ Wrong field name
    "total_clients": 1,               // ❌ Wrong field name
    "total_topics": 1                 // ❌ Not in spec
  }
}
```

### Test Failure
```
expect(stats.global_stats.total_subscriptions).toBeGreaterThanOrEqual(0);
Matcher error: received value must be a number or bigint
Received has value: undefined
```

## Resolution

The server must be updated to return the correct field names as documented in the JSON-RPC specification. The test is correct and should not be modified.