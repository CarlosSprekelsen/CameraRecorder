/**
 * REQ-NET01-001: Endpoint Configuration Validation
 * Validates that tests are using correct endpoints for different operations
 */

import { WebSocketTestFixture, HealthTestFixture } from '../fixtures/stable-test-fixture';
import { TEST_CONFIG } from '../config/test-config';

describe('REQ-NET01-001: Endpoint Configuration Validation', () => {
  let wsFixture: WebSocketTestFixture;
  let healthFixture: HealthTestFixture;

  beforeAll(async () => {
    wsFixture = new WebSocketTestFixture();
    healthFixture = new HealthTestFixture();
    
    // Initialize test environment
    await wsFixture.initialize();
    await healthFixture.initialize();
  });

  afterAll(async () => {
    wsFixture.cleanup();
    healthFixture.cleanup();
  });

  describe('WebSocket Server (Port 8002)', () => {
    it('should connect to WebSocket server on port 8002', async () => {
      const result = await wsFixture.testConnection();
      expect(result).toBe(true);
    });

    it('should respond to ping on WebSocket server', async () => {
      const result = await wsFixture.testPing();
      expect(result).toBe(true);
    });

    it('should retrieve camera list via WebSocket', async () => {
      const result = await wsFixture.testCameraList();
      expect(result).toBe(true);
    });
  });

  describe('Health Server (Port 8003)', () => {
    it('should access system health endpoint on port 8003', async () => {
      const result = await healthFixture.testSystemHealth();
      expect(result).toBe(true);
    });

    it('should access camera health endpoint on port 8003', async () => {
      const result = await healthFixture.testCameraHealth();
      expect(result).toBe(true);
    });

    it('should access MediaMTX health endpoint on port 8003', async () => {
      const result = await healthFixture.testMediaMTXHealth();
      expect(result).toBe(true);
    });

    it('should access readiness endpoint on port 8003', async () => {
      const result = await healthFixture.testReadiness();
      expect(result).toBe(true);
    });
  });

  describe('Configuration Validation', () => {
    it('should have correct WebSocket URL configuration', () => {
      expect(TEST_CONFIG.websocket.url).toBe('ws://localhost:8002/ws');
      expect(TEST_CONFIG.websocket.port).toBe(8002);
    });

    it('should have correct health URL configuration', () => {
      expect(TEST_CONFIG.health.url).toBe('http://localhost:8003');
      expect(TEST_CONFIG.health.port).toBe(8003);
    });

    it('should have environment validation', () => {
      expect(TEST_CONFIG.auth.jwtSecret).toBeDefined();
    });
  });
});
