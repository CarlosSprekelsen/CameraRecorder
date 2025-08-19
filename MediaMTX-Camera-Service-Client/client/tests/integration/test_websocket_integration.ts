/**
 * WebSocket Integration Tests
 * 
 * Tests real WebSocket communication with MediaMTX Camera Service
 * Following "Real Integration First" approach
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running via systemd
 * - Server accessible at ws://localhost:8002/ws
 */

import { WebSocketService } from '../../src/services/websocket';
import { RPC_METHODS, ERROR_CODES, PERFORMANCE_TARGETS, isNotification } from '../../src/types';
import type { CameraListResponse, CameraDevice, JSONRPCNotification } from '../../src/types';

describe('WebSocket Integration Tests', () => {
  let wsService: WebSocketService;
  const TEST_WEBSOCKET_URL = process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws';

  beforeAll(async () => {
    // Verify server is available before running tests
    const isServerAvailable = await checkServerAvailability();
    if (!isServerAvailable) {
      throw new Error('MediaMTX Camera Service not available. Start server before running integration tests.');
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

  describe('Connection Management', () => {
    it('should connect to real server within performance target', async () => {
      const startTime = performance.now();
      
      await wsService.connect();
      
      const connectionTime = performance.now() - startTime;
      expect(connectionTime).toBeLessThan(PERFORMANCE_TARGETS.CLIENT_WEBSOCKET_CONNECTION);
      expect(wsService.isConnected).toBe(true);
    });

    it('should handle connection resilience (disconnect/reconnect)', async () => {
      expect(wsService.isConnected).toBe(true);
      
      // Simulate network interruption
      wsService.disconnect();
      expect(wsService.isConnected).toBe(false);
      
      // Reconnect
      await wsService.connect();
      expect(wsService.isConnected).toBe(true);
    });
  });

  describe('JSON-RPC Method Validation', () => {
    it('should ping server and receive pong response', async () => {
      const startTime = performance.now();
      
      const response = await wsService.call(RPC_METHODS.PING, {});
      
      const responseTime = performance.now() - startTime;
      expect(response).toBe('pong');
      expect(responseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
    });

    it('should get camera list with correct structure', async () => {
      const startTime = performance.now();
      
      const response = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as any;
      
      const responseTime = performance.now() - startTime;
      
      // Validate response structure matches server API
      expect(response).toHaveProperty('cameras');
      expect(response).toHaveProperty('total');
      expect(response).toHaveProperty('connected');
      expect(Array.isArray(response.cameras)).toBe(true);
      expect(typeof response.total).toBe('number');
      expect(typeof response.connected).toBe('number');
      
      // Validate performance target
      expect(responseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
    });

    it('should get camera status for valid device', async () => {
      // First get camera list to find a valid device
      const cameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as CameraListResponse;
      
      if (cameraList.cameras.length === 0) {
        fail('No cameras available for status test - cannot validate core functionality');
      }

      const testDevice = cameraList.cameras[0].device;
      const startTime = performance.now();
      
      const response = await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device: testDevice }) as CameraDevice;
      
      const responseTime = performance.now() - startTime;
      
      // Validate response structure
      expect(response).toHaveProperty('device');
      expect(response).toHaveProperty('status');
      expect(response).toHaveProperty('name');
      expect(response).toHaveProperty('resolution');
      expect(response).toHaveProperty('fps');
      expect(response).toHaveProperty('streams');
      expect(response.device).toBe(testDevice);
      
      // Validate performance target
      expect(responseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
    });

    it('should handle camera not found error correctly', async () => {
      const startTime = performance.now();
      
      try {
        await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device: '/dev/video999' });
        fail('Expected error for non-existent camera');
      } catch (error) {
        const responseTime = performance.now() - startTime;
        
        // Validate error structure
        expect(error).toHaveProperty('code');
        expect((error as any).code).toBe(ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED);
        expect(error).toHaveProperty('message');
        
        // Validate performance target (errors should also be fast)
        expect(responseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      }
    });
  });

  describe('Real-time Notifications', () => {
    it('should receive camera status update notifications', async () => {
      const notificationPromise = new Promise((resolve) => {
        wsService.onMessage((message) => {
          if (isNotification(message) && message.method === 'camera_status_update') {
            resolve(message.params);
          }
        });
      });

      // Wait for notification (with timeout)
      const notification = await Promise.race([
        notificationPromise,
        new Promise((_, reject) => setTimeout(() => reject(new Error('Notification timeout')), 10000))
      ]);

      // Validate notification structure
      expect(notification).toHaveProperty('device');
      expect(notification).toHaveProperty('status');
      expect(notification).toHaveProperty('name');
      expect(notification).toHaveProperty('resolution');
      expect(notification).toHaveProperty('fps');
      expect(notification).toHaveProperty('streams');
    });

    it('should receive recording status update notifications', async () => {
      const notificationPromise = new Promise((resolve) => {
        wsService.onMessage((message) => {
          if (isNotification(message) && message.method === 'recording_status_update') {
            resolve(message.params);
          }
        });
      });

      // Wait for notification (with timeout)
      const notification = await Promise.race([
        notificationPromise,
        new Promise((_, reject) => setTimeout(() => reject(new Error('Notification timeout')), 10000))
      ]);

      // Validate notification structure
      expect(notification).toHaveProperty('device');
      expect(notification).toHaveProperty('status');
      expect(notification).toHaveProperty('filename');
      expect(notification).toHaveProperty('duration');
    });
  });

  describe('Performance Validation', () => {
    it('should meet status method performance targets', async () => {
      const measurements: number[] = [];
      
      // Measure multiple calls to get average
      for (let i = 0; i < 5; i++) {
        const startTime = performance.now();
        await wsService.call(RPC_METHODS.PING, {});
        measurements.push(performance.now() - startTime);
      }
      
      const averageTime = measurements.reduce((a, b) => a + b, 0) / measurements.length;
      expect(averageTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
    });

    it('should meet control method performance targets', async () => {
      // Test with take_snapshot (control method)
      const cameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as CameraListResponse;
      
      if (cameraList.cameras.length === 0) {
        fail('No cameras available for control method test - cannot validate core functionality');
      }

      const testDevice = cameraList.cameras[0].device;
      const startTime = performance.now();
      
      try {
        await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, { device: testDevice });
      } catch (error) {
        // Expected if camera doesn't support snapshot
        console.warn('Snapshot test failed (expected for some cameras):', (error as Error).message);
      }
      
      const responseTime = performance.now() - startTime;
      expect(responseTime).toBeLessThan(PERFORMANCE_TARGETS.CONTROL_METHODS);
    });
  });

  describe('Error Handling', () => {
    it('should handle invalid method errors', async () => {
      try {
        await wsService.call('invalid_method', {});
        fail('Expected error for invalid method');
      } catch (error) {
        expect(error).toHaveProperty('code');
        expect((error as any).code).toBe(ERROR_CODES.METHOD_NOT_FOUND);
      }
    });

    it('should handle invalid parameters', async () => {
      try {
        await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, {});
        fail('Expected error for missing device parameter');
      } catch (error) {
        expect(error).toHaveProperty('code');
        expect((error as any).code).toBe(ERROR_CODES.INVALID_PARAMS);
      }
    });

    it('should handle connection failures gracefully', async () => {
      // Disconnect and try to make a call
      wsService.disconnect();
      
      try {
        await wsService.call(RPC_METHODS.PING, {});
        fail('Expected error when disconnected');
      } catch (error) {
        expect((error as Error).message).toContain('not connected');
      }
    });
  });
});

/**
 * Check if MediaMTX Camera Service is available
 */
async function checkServerAvailability(): Promise<boolean> {
  const testWebSocketUrl = process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws';
  try {
    // Try to connect to WebSocket endpoint
    const ws = new WebSocket(testWebSocketUrl);
    
    return new Promise((resolve) => {
      const timeout = setTimeout(() => {
        ws.close();
        resolve(false);
      }, 3000);

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
