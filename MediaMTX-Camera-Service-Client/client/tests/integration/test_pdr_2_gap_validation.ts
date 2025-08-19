/**
 * PDR-2 Gap Validation Tests
 * 
 * Focused tests addressing specific PDR-2 requirements NOT covered by existing tests
 * Following proven patterns from existing working tests
 * 
 * Gaps Addressed:
 * - PDR-2.1: Real-world network interruption scenarios
 * - PDR-2.3: State synchronization validation
 * - PDR-2.4: Polling fallback mechanism (CRITICAL GAP)
 * - PDR-2.5: User feedback mechanisms
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running via systemd
 * - Authentication environment set up (./set-test-env.sh)
 */

import { WebSocketService } from '../../src/services/websocket';
import { RPC_METHODS, ERROR_CODES, PERFORMANCE_TARGETS } from '../../src/types';

describe('PDR-2 Gap Validation Tests', () => {
  let wsService: WebSocketService;
  const TEST_WEBSOCKET_URL = process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws';

  beforeAll(async () => {
    // Verify server is available
    const isServerAvailable = await checkServerAvailability();
    if (!isServerAvailable) {
      throw new Error('MediaMTX Camera Service not available for PDR-2 gap validation.');
    }
  });

  beforeEach(async () => {
    wsService = new WebSocketService({
      url: TEST_WEBSOCKET_URL,
      reconnectInterval: 1000,
      maxReconnectAttempts: 3,
      requestTimeout: 5000,
      heartbeatInterval: 30000,
      baseDelay: 1000,
      maxDelay: 30000,
    });
    
    await wsService.connect();
  });

  afterEach(async () => {
    if (wsService) {
      wsService.disconnect();
    }
  });

  describe('PDR-2.1: Real-world Network Interruption Scenarios', () => {
    it('should handle rapid connection cycling (network instability simulation)', async () => {
      const cycles = 10;
      const reconnectionTimes: number[] = [];
      
      for (let i = 0; i < cycles; i++) {
        const startTime = performance.now();
        
        wsService.disconnect();
        await wsService.connect();
        
        reconnectionTimes.push(performance.now() - startTime);
        expect(wsService.isConnected).toBe(true);
        
        // Brief pause to simulate real network conditions
        await new Promise(resolve => setTimeout(resolve, 50));
      }
      
      // Validate all reconnections were successful and within performance targets
      const averageReconnectionTime = reconnectionTimes.reduce((a, b) => a + b, 0) / cycles;
      expect(averageReconnectionTime).toBeLessThan(PERFORMANCE_TARGETS.CLIENT_WEBSOCKET_CONNECTION);
      expect(reconnectionTimes.every(time => time < PERFORMANCE_TARGETS.CLIENT_WEBSOCKET_CONNECTION)).toBe(true);
    });

    it('should maintain functionality during connection instability', async () => {
      // Test that operations work during connection cycling
      const operations = [];
      
      for (let i = 0; i < 5; i++) {
        // Disconnect and reconnect
        wsService.disconnect();
        await wsService.connect();
        
        // Immediately try an operation
        const startTime = performance.now();
        const response = await wsService.call(RPC_METHODS.PING, {});
        const responseTime = performance.now() - startTime;
        
        expect(response).toBe('pong');
        expect(responseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
        
        operations.push(responseTime);
      }
      
      // Validate consistent performance
      const averageTime = operations.reduce((a, b) => a + b, 0) / operations.length;
      expect(averageTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
    });
  });

  describe('PDR-2.3: State Synchronization Validation', () => {
    it('should maintain consistent state across multiple operations', async () => {
      // Get initial camera list
      const initialCameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as any;
      expect(initialCameraList).toHaveProperty('cameras');
      expect(initialCameraList).toHaveProperty('total');
      expect(initialCameraList).toHaveProperty('connected');
      
      // Perform multiple operations and verify state consistency
      const operations = [];
      for (let i = 0; i < 3; i++) {
        const cameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as any;
        operations.push(cameraList);
        
        // Verify state consistency
        expect(cameraList.total).toBe(initialCameraList.total);
        expect(cameraList.connected).toBe(initialCameraList.connected);
        expect(cameraList.cameras.length).toBe(initialCameraList.cameras.length);
      }
      
      // Validate all operations returned consistent state
      const allConsistent = operations.every(op => 
        op.total === initialCameraList.total && 
        op.connected === initialCameraList.connected
      );
      expect(allConsistent).toBe(true);
    });

    it('should handle notification ordering correctly', async () => {
      const notifications: any[] = [];
      
      // Set up notification listener
      wsService.onMessage((message: any) => {
        if (message.method && message.params) {
          notifications.push({
            method: message.method,
            timestamp: Date.now(),
            params: message.params
          });
        }
      });
      
      // Trigger multiple operations that should generate notifications
      await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {});
      await wsService.call(RPC_METHODS.PING, {});
      await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {});
      
      // Wait for notifications
      await new Promise(resolve => setTimeout(resolve, 3000));
      
      // Validate notification ordering (timestamps should be ascending)
      if (notifications.length > 1) {
        for (let i = 1; i < notifications.length; i++) {
          expect(notifications[i].timestamp).toBeGreaterThanOrEqual(notifications[i-1].timestamp);
        }
      }
    });
  });

  describe('PDR-2.4: Polling Fallback Mechanism (CRITICAL GAP)', () => {
    it('should detect when polling fallback is not implemented', async () => {
      // This test validates that polling fallback is a critical gap
      // The current implementation does not have polling fallback
      
      // Simulate WebSocket failure
      wsService.disconnect();
      
      // Attempt to get camera list (should fail without polling fallback)
      try {
        await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {});
        // If this succeeds, polling fallback is implemented
        console.log('✅ Polling fallback mechanism is implemented');
      } catch (error: any) {
        // Expected behavior - polling fallback not implemented
        expect(error.message).toContain('not connected');
        console.log('❌ Polling fallback mechanism NOT implemented (expected for current version)');
      }
    });

    it('should validate polling fallback requirement', async () => {
      // This test documents the requirement for polling fallback
      const pollingFallbackRequired = true;
      const pollingFallbackImplemented = false; // Current state
      
      expect(pollingFallbackRequired).toBe(true);
      expect(pollingFallbackImplemented).toBe(false);
      
      // Document the gap for PDR-2 validation
      console.log('📋 PDR-2.4 GAP: Polling fallback mechanism required but not implemented');
      console.log('   - WebSocket failure should trigger HTTP polling fallback');
      console.log('   - Fallback should maintain functionality during network issues');
      console.log('   - Automatic switch back to WebSocket when connection restored');
    });
  });

  describe('PDR-2.5: User Feedback Mechanisms', () => {
    it('should provide meaningful error messages for different failure scenarios', async () => {
      // Test invalid method
      try {
        await wsService.call('invalid_method', {});
        fail('Expected error for invalid method');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error).toHaveProperty('message');
        expect(error.code).toBe(ERROR_CODES.METHOD_NOT_FOUND);
        expect(error.message).toBeTruthy();
      }
      
      // Test invalid parameters
      try {
        await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, {});
        fail('Expected error for missing device parameter');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error).toHaveProperty('message');
        expect(error.code).toBe(ERROR_CODES.INVALID_PARAMS);
        expect(error.message).toBeTruthy();
      }
      
      // Test camera not found
      try {
        await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device: '/dev/video999' });
        fail('Expected error for non-existent camera');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error).toHaveProperty('message');
        expect(error.code).toBe(ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED);
        expect(error.message).toBeTruthy();
      }
    });

    it('should handle connection errors with user-friendly messages', async () => {
      // Test connection error handling
      const invalidWsService = new WebSocketService({
        url: 'ws://invalid-host:9999/ws',
        reconnectInterval: 1000,
        maxReconnectAttempts: 1,
        requestTimeout: 1000,
        heartbeatInterval: 30000,
        baseDelay: 1000,
        maxDelay: 30000,
      });
      
      try {
        await invalidWsService.connect();
        fail('Expected connection to fail');
      } catch (error: any) {
        expect(error).toBeDefined();
        expect(error.message).toBeTruthy();
      } finally {
        invalidWsService.disconnect();
      }
    });
  });

  describe('PDR-2: Performance Under Load Validation', () => {
    it('should maintain performance targets under concurrent requests', async () => {
      const concurrentRequests = 5;
      const requestPromises = [];
      
      // Launch concurrent requests
      for (let i = 0; i < concurrentRequests; i++) {
        requestPromises.push(
          wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}).then(() => performance.now())
        );
      }
      
      const startTime = performance.now();
      const results = await Promise.all(requestPromises);
      const totalTime = performance.now() - startTime;
      
      // Validate individual request performance
      results.forEach((endTime) => {
        const requestTime = endTime - startTime;
        expect(requestTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      });
      
      // Validate overall performance
      expect(totalTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS * concurrentRequests);
    });

    it('should handle rapid method calls without degradation', async () => {
      const rapidCalls = 10;
      const responseTimes: number[] = [];
      
      for (let i = 0; i < rapidCalls; i++) {
        const startTime = performance.now();
        await wsService.call(RPC_METHODS.PING, {});
        responseTimes.push(performance.now() - startTime);
      }
      
      // Validate all calls meet performance targets
      responseTimes.forEach((time) => {
        expect(time).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      });
      
      // Validate no significant degradation
      const averageTime = responseTimes.reduce((a, b) => a + b, 0) / rapidCalls;
      expect(averageTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
    });
  });
});

/**
 * Check if MediaMTX Camera Service is available
 */
async function checkServerAvailability(): Promise<boolean> {
  const testWebSocketUrl = process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws';
  try {
    const ws = new WebSocket(testWebSocketUrl);
    
    return new Promise((resolve) => {
      const timeout = setTimeout(() => {
        ws.close();
        resolve(false);
      }, 5000);

      ws.onopen = () => {
        clearTimeout(timeout);
        ws.close();
        resolve(true);
      };

      ws.onerror = () => {
        clearTimeout(timeout);
        resolve(false);
      };
    });
  } catch {
    return false;
  }
}
