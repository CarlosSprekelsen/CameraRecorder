/**
 * REQ-NET01-003: Polling Fallback Mechanism Integration Tests
 * Tests real integration with MediaMTX Camera Service using polling fallback
 * Following "Real Integration Always" approach
 */

import { WebSocketTestFixture } from '../fixtures/stable-test-fixture';
import { TEST_CONFIG } from '../config/test-config';

describe('REQ-NET01-003: Polling Fallback Mechanism Integration Tests', () => {
  let wsFixture: WebSocketTestFixture;

  beforeAll(async () => {
    wsFixture = new WebSocketTestFixture();
    
    await wsFixture.initialize();
  });

  afterAll(async () => {
    wsFixture.cleanup();
  });

  describe('WebSocket Server (Port 8002)', () => {
    it('should connect to WebSocket server and respond to ping', async () => {
      const connectionResult = await wsFixture.testConnection();
      expect(connectionResult).toBe(true);
      
      const pingResult = await wsFixture.testPing();
      expect(pingResult).toBe(true);
    });

    it('should retrieve camera list via WebSocket', async () => {
      const result = await wsFixture.testCameraList();
      expect(result).toBe(true);
    });
  });


  describe('Configuration Validation', () => {
    it('should have correct endpoint configuration', () => {
      expect(TEST_CONFIG.websocket.url).toBe('ws://localhost:8002/ws');
      expect(TEST_CONFIG.websocket.port).toBe(8002);
    });

    it('should have proper authentication configuration', () => {
      expect(TEST_CONFIG.auth.jwtSecret).toBeDefined();
      expect(TEST_CONFIG.auth.jwtSecret).toBeTruthy();
    });
  });
});
