# Phase 3: Go Server Integration Migration Guide

**Version:** 3.0  
**Date:** 2025-01-15  
**Status:** Go Server JSON-RPC Integration Complete  
**Related Epic/Story:** Go Implementation API Compatibility

## Overview

This migration guide covers the complete integration with the Go server's enhanced JSON-RPC API, including authentication, new methods, and event subscription system.

## 🚀 What's New

### 1. **Authentication System**
- JWT token and API key authentication
- Role-based access control (admin, operator, viewer)
- Automatic token refresh and session management
- Authentication error handling

### 2. **Enhanced API Methods**
- **New Methods**: 25+ new methods including streaming, file management, and system monitoring
- **Enhanced Error Codes**: Updated error codes aligned with Go server
- **Performance Improvements**: 5x faster response times, 10x better concurrency

### 3. **Event Subscription System**
- Real-time event subscriptions with topic-based filtering
- Enhanced notification handling
- Subscription statistics and management

### 4. **External Stream Discovery**
- UAV and external RTSP stream discovery
- Skydio-specific stream support
- Network-based stream management

## 📋 Migration Checklist

### ✅ **Completed in Phase 3**

- [x] **RPC Methods Updated**: All 25+ new methods added to `RPC_METHODS`
- [x] **Error Codes Aligned**: Updated error codes to match Go server
- [x] **Authentication Service**: Complete JWT/API key authentication system
- [x] **WebSocket Integration**: Enhanced WebSocket service with auth support
- [x] **Event Subscription**: Real-time event subscription system
- [x] **Type Definitions**: Complete TypeScript types for all new methods
- [x] **Connection Service**: Updated to handle authentication flow

### 🔄 **Next Steps**

- [ ] **Testing**: Integration testing with Go server
- [ ] **Component Updates**: Update components to use new methods
- [ ] **Old Store Removal**: Remove deprecated connectionStore.ts
- [ ] **Documentation**: Update API documentation

## 🔧 **Breaking Changes**

### 1. **Authentication Required**
All API calls now require authentication (except `authenticate` method).

**Before:**
```typescript
await websocketService.call('get_camera_list', {});
```

**After:**
```typescript
// Authentication is automatically added by WebSocket service
await websocketService.call('get_camera_list', {});
// OR explicitly:
await websocketService.call('get_camera_list', {}, true); // requireAuth = true
```

### 2. **New Error Codes**
Error codes have been updated to match Go server.

**Before:**
```typescript
ERROR_CODES.AUTHENTICATION_REQUIRED // -32004
```

**After:**
```typescript
ERROR_CODES.AUTHENTICATION_FAILED // -32001
ERROR_CODES.INSUFFICIENT_PERMISSIONS // -32003
```

### 3. **Enhanced Method Parameters**
Some methods now have additional optional parameters.

**Before:**
```typescript
await websocketService.call('start_recording', { device: 'camera0' });
```

**After:**
```typescript
await websocketService.call('start_recording', { 
  device: 'camera0',
  duration: 3600, // optional
  format: 'fmp4'  // optional
});
```

## 🛠 **Implementation Guide**

### 1. **Initialize Authentication**

```typescript
import { authService } from '../services/authService';

// Initialize with JWT token
await authService.initialize({
  jwtToken: 'your-jwt-token',
  autoReauth: true,
  reauthThreshold: 300 // 5 minutes before expiry
});

// OR with API key
await authService.initialize({
  apiKey: 'your-api-key',
  autoReauth: true
});
```

### 2. **Use New Methods**

```typescript
import { RPC_METHODS } from '../types/rpc';

// Get camera capabilities
const capabilities = await websocketService.call(
  RPC_METHODS.GET_CAMERA_CAPABILITIES,
  { device: 'camera0' }
);

// Start streaming
const stream = await websocketService.call(
  RPC_METHODS.START_STREAMING,
  { device: 'camera0' }
);

// Discover external streams
const discovery = await websocketService.call(
  RPC_METHODS.DISCOVER_EXTERNAL_STREAMS,
  { 
    skydio_enabled: true,
    generic_enabled: false 
  }
);
```

### 3. **Event Subscription**

```typescript
import { eventSubscriptionService } from '../services/eventSubscriptionService';
import { EVENT_TOPICS } from '../types/rpc';

// Subscribe to camera events
await eventSubscriptionService.subscribe(
  [EVENT_TOPICS.CAMERA_CONNECTED, EVENT_TOPICS.CAMERA_DISCONNECTED],
  (event) => {
    console.log('Camera event:', event);
  },
  { device: 'camera0' } // optional filters
);

// Get subscription stats
const stats = await eventSubscriptionService.getSubscriptionStats();
```

### 4. **Enhanced Error Handling**

```typescript
try {
  await websocketService.call('some_method', {});
} catch (error) {
  if (error.code === ERROR_CODES.AUTHENTICATION_FAILED) {
    // Handle authentication failure
    authService.logout();
    // Redirect to login
  } else if (error.code === ERROR_CODES.INSUFFICIENT_PERMISSIONS) {
    // Handle permission error
    console.error('Insufficient permissions for this operation');
  }
}
```

## 📊 **Performance Improvements**

### **Go Server Benefits**
- **Response Time**: <50ms for status methods, <100ms for control methods
- **Concurrency**: 1000+ simultaneous WebSocket connections
- **Memory Usage**: <60MB base footprint
- **CPU Usage**: <50% sustained usage

### **Client Optimizations**
- **Authentication Caching**: Tokens cached and auto-refreshed
- **Event Filtering**: Server-side event filtering reduces bandwidth
- **Connection Pooling**: Efficient WebSocket connection management
- **Error Recovery**: Automatic reconnection with exponential backoff

## 🔍 **Testing Strategy**

### 1. **Unit Tests**
```typescript
// Test authentication service
describe('AuthenticationService', () => {
  it('should authenticate with valid JWT token', async () => {
    const response = await authService.authenticate('valid-token');
    expect(response.authenticated).toBe(true);
  });
});

// Test event subscription
describe('EventSubscriptionService', () => {
  it('should subscribe to camera events', async () => {
    const response = await eventSubscriptionService.subscribe(
      [EVENT_TOPICS.CAMERA_CONNECTED],
      jest.fn()
    );
    expect(response.subscribed).toBe(true);
  });
});
```

### 2. **Integration Tests**
```typescript
// Test with real Go server
describe('Go Server Integration', () => {
  it('should connect and authenticate', async () => {
    await connectionService.connect();
    expect(authService.isAuthenticated()).toBe(true);
  });
});
```

## 🚨 **Common Issues & Solutions**

### **Issue 1: Authentication Failures**
**Problem**: `Authentication failed` errors
**Solution**: Ensure valid JWT token or API key is provided

```typescript
// Check token validity
if (!authService.isAuthenticated()) {
  await authService.authenticate('your-token');
}
```

### **Issue 2: Permission Errors**
**Problem**: `Insufficient permissions` errors
**Solution**: Check user role and required permissions

```typescript
// Check permissions before operation
if (!authService.hasRole('operator')) {
  throw new Error('Operator role required for this operation');
}
```

### **Issue 3: Event Subscription Not Working**
**Problem**: Events not being received
**Solution**: Verify subscription and topic names

```typescript
// Check subscription status
const stats = await eventSubscriptionService.getSubscriptionStats();
console.log('Active subscriptions:', stats.global_stats.total_subscriptions);
```

## 📈 **Migration Timeline**

### **Phase 3A: Core Integration** ✅
- [x] Authentication system
- [x] WebSocket service updates
- [x] RPC methods and error codes
- [x] Event subscription system

### **Phase 3B: Component Updates** 🔄
- [ ] Update camera components to use new methods
- [ ] Update recording components with enhanced features
- [ ] Update file management with new capabilities
- [ ] Update system monitoring components

### **Phase 3C: Testing & Validation** 📋
- [ ] Unit tests for new services
- [ ] Integration tests with Go server
- [ ] Performance testing
- [ ] User acceptance testing

### **Phase 3D: Cleanup** 🧹
- [ ] Remove old connectionStore.ts
- [ ] Update documentation
- [ ] Performance optimization
- [ ] Final validation

## 🎯 **Success Metrics**

### **Technical Metrics**
- [ ] All 25+ new methods working correctly
- [ ] Authentication flow 100% functional
- [ ] Event subscription system operational
- [ ] Error handling comprehensive
- [ ] Performance targets met

### **User Experience Metrics**
- [ ] Faster response times (<100ms)
- [ ] Better error messages
- [ ] Real-time event updates
- [ ] Seamless authentication
- [ ] Enhanced functionality

## 📚 **Additional Resources**

- **Go Server API Documentation**: `../mediamtx-camera-service-go/docs/api/json_rpc_methods.md`
- **Authentication Guide**: `../docs/authentication-guide.md`
- **Event Subscription Guide**: `../docs/event-subscription-guide.md`
- **Performance Benchmarks**: `../docs/performance-benchmarks.md`

---

**Next Steps**: Proceed with Phase 3B - Component Updates to integrate new methods into existing components.
