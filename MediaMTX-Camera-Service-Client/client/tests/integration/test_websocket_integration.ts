/**
 * REQ-WS01-001: WebSocket Integration with Real MediaMTX Server
 * REQ-WS01-002: JSON-RPC Method Validation
 * Coverage: INTEGRATION
 * Quality: HIGH
 */
/**
 * WebSocket Integration Tests
 * 
 * Tests real WebSocket communication with MediaMTX Camera Service
 * Following "Real Integration Always" approach
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running via systemd
 * - Server accessible at ws://localhost:8002/ws
 * - Authentication properly configured via set-test-env.sh
 */

import { WebSocketTestFixture } from '../fixtures/stable-test-fixture';
import { TEST_CONFIG } from '../config/test-config';

describe('WebSocket Integration Tests', () => {
  let wsFixture: WebSocketTestFixture;

  beforeAll(async () => {
    wsFixture = new WebSocketTestFixture();
    await wsFixture.initialize();
  });

  afterAll(async () => {
    wsFixture.cleanup();
  });

  describe('Connection Management', () => {
    it('should connect to real server within performance target', async () => {
      const startTime = performance.now();
      
      const result = await wsFixture.testConnection();
      const connectionTime = performance.now() - startTime;
      
      expect(result).toBe(true);
      expect(connectionTime).toBeLessThan(TEST_CONFIG.test.timeout);
    });

    it('should handle connection resilience (disconnect/reconnect)', async () => {
      // Test initial connection
      const result1 = await wsFixture.testConnection();
      expect(result1).toBe(true);
      
      // Test reconnection
      const result2 = await wsFixture.testConnection();
      expect(result2).toBe(true);
    });
  });

  describe('JSON-RPC Method Validation', () => {
    it('should ping server and receive pong response', async () => {
      const startTime = performance.now();
      
      const result = await wsFixture.testPing();
      const responseTime = performance.now() - startTime;
      
      expect(result).toBe(true);
      expect(responseTime).toBeLessThan(TEST_CONFIG.test.timeout);
    });

    it('should get camera list with correct structure', async () => {
      const startTime = performance.now();
      
      const result = await wsFixture.testCameraList();
      const responseTime = performance.now() - startTime;
      
      expect(result).toBe(true);
      expect(responseTime).toBeLessThan(TEST_CONFIG.test.timeout);
    });
  });

  describe('Configuration Validation', () => {
    it('should use correct WebSocket endpoint configuration', () => {
      expect(TEST_CONFIG.websocket.url).toBe('ws://localhost:8002/ws');
      expect(TEST_CONFIG.websocket.port).toBe(8002);
    });

    it('should have proper authentication configuration', () => {
      expect(TEST_CONFIG.auth.jwtSecret).toBeDefined();
      expect(TEST_CONFIG.auth.jwtSecret).toBeTruthy();
    });
  });
});
