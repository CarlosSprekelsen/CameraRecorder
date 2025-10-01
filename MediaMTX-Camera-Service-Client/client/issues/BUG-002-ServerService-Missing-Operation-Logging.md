# BUG-002: ServerService Missing Operation Logging

## Summary
ServerService methods `subscribeEvents()`, `unsubscribeEvents()`, and `getSubscriptionStats()` do not log their operations, making debugging and monitoring impossible.

## Type
Defect - Missing Functionality

## Priority
High

## Severity
Major

## Affected Component
- **Service**: `ServerService`
- **File**: `src/services/server/ServerService.ts`
- **Methods**: `subscribeEvents()`, `unsubscribeEvents()`, `getSubscriptionStats()`

## Environment
- **Version**: Current development branch
- **Architecture**: Client-side service layer

## Steps to Reproduce
1. Initialize `ServerService` with APIClient and LoggerService
2. Call `subscribeEvents(['camera_connected'])`
3. Inspect logger output
4. Observe no log entries for the operation

## Expected Behavior
All service methods should log their operations following the standard RPC logging pattern:
- **Request log**: `'{method_name} request', params`
- **Error log**: `'{method_name} failed', error`

Example:
```typescript
async subscribeEvents(topics: string[], filters?: Record<string, unknown>): Promise<SubscriptionResult> {
  try {
    this.logger.info('subscribe_events request', { topics, filters });
    const result = await this.apiClient.call('subscribe_events', { topics, filters });
    return result;
  } catch (error) {
    this.logger.error('subscribe_events failed', error);
    throw error;
  }
}
```

## Actual Behavior
Methods have no logging whatsoever. Operations execute silently:

```typescript
// Current implementation - NO LOGGING
async subscribeEvents(
  topics: string[],
  filters?: Record<string, unknown>,
): Promise<SubscriptionResult> {
  return this.apiClient.call('subscribe_events', { topics, filters });
}
```

## Root Cause Analysis

### Code Location
File: `src/services/server/ServerService.ts`

### Affected Methods

**1. subscribeEvents() - Lines 95-100**
```typescript
async subscribeEvents(
  topics: string[],
  filters?: Record<string, unknown>,
): Promise<SubscriptionResult> {
  return this.apiClient.call('subscribe_events', { topics, filters });
}
```
**Issue**: No logging at all

**2. unsubscribeEvents() - Lines 102-104**
```typescript
async unsubscribeEvents(topics?: string[]): Promise<UnsubscriptionResult> {
  return this.apiClient.call('unsubscribe_events', { topics });
}
```
**Issue**: No logging at all

**3. getSubscriptionStats() - Lines 106-108**
```typescript
async getSubscriptionStats(): Promise<SubscriptionStatsResult> {
  return this.apiClient.call('get_subscription_stats');
}
```
**Issue**: No logging at all

### Comparison with Correct Methods

**Other methods in ServerService that DO have logging:**
```typescript
// getMetrics() - HAS logging (lines 86-93)
async getMetrics(): Promise<MetricsResult> {
  try {
    this.logger.info('get_metrics request');
    const result = await this.apiClient.call<MetricsResult>('get_metrics');
    return result;
  } catch (error) {
    this.logger.error('Error getting metrics:', error as Error);
    throw error;
  }
}
```

Note: Even `getMetrics()` doesn't follow the standard pattern (`'Error getting metrics:'` instead of `'get_metrics failed'`), but at least it has logging.

### Why This Occurred
Methods were implemented as simple pass-through wrappers without considering logging requirements. The implementation prioritized brevity over observability.

### Impact Assessment
- **Debugging**: Impossible to trace event subscription operations in logs
- **Monitoring**: Cannot track subscription activity or failures
- **Testing**: Tests fail because they expect logging to occur (9 test failures)
- **Production Support**: No visibility into subscription management operations
- **Audit Trail**: No record of which events were subscribed/unsubscribed

## Test Evidence

### Failing Tests
File: `tests/unit/services/event_subscription_service.test.ts`

Failing assertions (9 total):

**Test 1: Should call WebSocket service with topics only**
```typescript
// Line 54
expect(mockLoggerService.info).toHaveBeenCalledWith('subscribe_events request', { topics, filters: undefined });
// Actual: Logger was never called (only "ServerService initialized")
```

**Test 2: Should call WebSocket service with topics and filters**
```typescript
// Line 75
expect(mockLoggerService.info).toHaveBeenCalledWith('subscribe_events request', { topics, filters });
// Actual: Logger was never called
```

**Test 3: Should handle errors correctly**
```typescript
// Line 88
expect(mockLoggerService.error).toHaveBeenCalledWith('subscribe_events failed', error);
// Actual: Logger was never called (no error logging)
```

**Test 4: Should call WebSocket service with specific topics (unsubscribe)**
```typescript
// Line 107
expect(mockLoggerService.info).toHaveBeenCalledWith('unsubscribe_events request', { topics });
// Actual: Logger was never called
```

**Test 5: Should call WebSocket service without topics (unsubscribe all)**
```typescript
// Line 125
expect(mockLoggerService.info).toHaveBeenCalledWith('unsubscribe_events request', { topics: undefined });
// Actual: Logger was never called
```

**Test 6: Should handle errors correctly (unsubscribe)**
```typescript
// Line 138
expect(mockLoggerService.error).toHaveBeenCalledWith('unsubscribe_events failed', error);
// Actual: Logger was never called
```

**Test 7: Should call WebSocket service with correct parameters (get stats)**
```typescript
// Line 165
expect(mockLoggerService.info).toHaveBeenCalledWith('get_subscription_stats request');
// Actual: Logger was never called
```

**Test 8: Should handle errors correctly (get stats)**
```typescript
// Line 177
expect(mockLoggerService.error).toHaveBeenCalledWith('get_subscription_stats failed', error);
// Actual: Logger was never called
```

### Test Output
```
FAIL tests/unit/services/event_subscription_service.test.ts
  ● EventSubscriptionService Unit Tests › REQ-EVENT-001: subscribe_events RPC method › Should call WebSocket service with topics only
    
    expect(jest.fn()).toHaveBeenCalledWith(...expected)
    
    Expected: "subscribe_events request", {"filters": undefined, "topics": ["camera.connected", "recording.start"]}
    Received: "ServerService initialized"
    
    Number of calls: 1
```

## Comparison with Correct Implementation

### AuthService (Correct Implementation)
File: `src/services/auth/AuthService.ts`
```typescript
async authenticate(token: string): Promise<AuthenticateResult> {
  if (!this.apiClient.isConnected()) {
    throw new Error('WebSocket not connected');
  }

  const params: AuthenticateParams = { auth_token: token };
  const result = await this.apiClient.call<AuthenticateResult>('authenticate', params as unknown as Record<string, unknown>);

  return result;
}
```
Note: AuthService also lacks proper logging. This is a broader issue.

### DeviceService (Better Implementation)
File: `src/services/device/DeviceService.ts`
```typescript
async getCameraList(): Promise<Camera[]> {
  this.logger.info('Getting camera list');
  const result = await this.apiClient.call<CameraListResult>('get_camera_list', {});
  this.logger.info(`Retrieved ${result.cameras.length} cameras`);
  return result.cameras;
}
```
Note: DeviceService has logging but doesn't follow standard pattern.

## Related Documentation
- **Testing Guidelines**: `docs/development/client-testing-guidelines.md`
- **Architecture**: `docs/architecture/client-architechture.md`
- **API Documentation**: `docs/api/mediamtx_camera_service_openrpc.json`

## Security Implications
Event subscriptions are critical for real-time monitoring. Without logging:
- Cannot audit which events are being monitored
- Cannot detect unauthorized subscription attempts
- Cannot troubleshoot subscription failures
- Cannot track subscription lifecycle

## Acceptance Criteria
1. `subscribeEvents()` logs request and errors
2. `unsubscribeEvents()` logs request and errors
3. `getSubscriptionStats()` logs request and errors
4. All logs follow standard pattern: `'{method_name} request'` and `'{method_name} failed'`
5. All 9 failing tests in `event_subscription_service.test.ts` pass

## Proposed Fix

### Changes Required
Update all three methods in `src/services/server/ServerService.ts`:

**1. subscribeEvents() - Lines 95-100**
```typescript
// BEFORE
async subscribeEvents(
  topics: string[],
  filters?: Record<string, unknown>,
): Promise<SubscriptionResult> {
  return this.apiClient.call('subscribe_events', { topics, filters });
}

// AFTER
async subscribeEvents(
  topics: string[],
  filters?: Record<string, unknown>,
): Promise<SubscriptionResult> {
  try {
    this.logger.info('subscribe_events request', { topics, filters });
    const result = await this.apiClient.call('subscribe_events', { topics, filters });
    return result;
  } catch (error) {
    this.logger.error('subscribe_events failed', error);
    throw error;
  }
}
```

**2. unsubscribeEvents() - Lines 102-104**
```typescript
// BEFORE
async unsubscribeEvents(topics?: string[]): Promise<UnsubscriptionResult> {
  return this.apiClient.call('unsubscribe_events', { topics });
}

// AFTER
async unsubscribeEvents(topics?: string[]): Promise<UnsubscriptionResult> {
  try {
    this.logger.info('unsubscribe_events request', { topics });
    const result = await this.apiClient.call('unsubscribe_events', { topics });
    return result;
  } catch (error) {
    this.logger.error('unsubscribe_events failed', error);
    throw error;
  }
}
```

**3. getSubscriptionStats() - Lines 106-108**
```typescript
// BEFORE
async getSubscriptionStats(): Promise<SubscriptionStatsResult> {
  return this.apiClient.call('get_subscription_stats');
}

// AFTER
async getSubscriptionStats(): Promise<SubscriptionStatsResult> {
  try {
    this.logger.info('get_subscription_stats request');
    const result = await this.apiClient.call('get_subscription_stats');
    return result;
  } catch (error) {
    this.logger.error('get_subscription_stats failed', error);
    throw error;
  }
}
```

## Verification Steps
1. Apply proposed changes to `src/services/server/ServerService.ts`
2. Run unit tests: `npm run test:unit -- --testPathPattern=event_subscription_service.test.ts`
3. Verify all 9 previously failing tests now pass
4. Verify log output appears during integration testing
5. Verify error handling still works correctly
6. Verify no regressions in subscription functionality

## Additional Considerations

### Performance Impact
Adding logging has minimal performance impact:
- Log statements are typically async/non-blocking
- Production log levels can filter out debug messages
- Benefits of observability far outweigh minimal overhead

### Alternative Approach
Could use AOP (Aspect-Oriented Programming) or decorators to add logging automatically to all service methods, but this would require broader architectural changes.

## Related Issues
- **BUG-001**: ExternalStreamService uses non-standard logging pattern
- Consider audit of all services to ensure logging consistency

## Attachments
- Test failure output: See test run from 2025-10-01
- Service implementation: `src/services/server/ServerService.ts`
- Test expectations: `tests/unit/services/event_subscription_service.test.ts`

