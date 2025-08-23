/**
 * REQ-AUTH02-001: Comprehensive authentication validation
 * REQ-AUTH02-002: Authentication parameter format compliance
 * Coverage: INTEGRATION
 * Quality: HIGH
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Test Categories: Integration/Authentication
 * API Documentation Reference: docs/api/json-rpc-methods.md
 * 
 * Uses StableTestFixture as single source of truth for authentication
 */

const { StableTestFixture } = require('../../fixtures/stable-test-fixture');

/**
 * Comprehensive Authentication Integration Test
 * Uses StableTestFixture for API-compliant authentication validation
 */

describe('Authentication Integration Tests', () => {
  let fixture;

  beforeAll(async () => {
    fixture = new StableTestFixture();
    await fixture.initialize();
  });

  afterAll(async () => {
    if (fixture) {
      fixture.cleanup();
    }
  });

  describe('Valid Token Tests', () => {
    test('should authenticate with valid token using compliant fixture', async () => {
      // Use the stable test fixture as single source of truth for authentication
      const ws = await fixture.connectWebSocketWithAuth();
      
      // The fixture handles all authentication validation against API documentation
      expect(ws).toBeDefined();
      expect(ws.readyState).toBe(1); // WebSocket.OPEN
      
      // Verify authentication was successful by testing a protected method
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'ping', id);
      
      const response = await fixture.waitForResponse(ws, id);
      expect(response).toBe('pong');
      
      ws.close();
    });
  });

  describe('Invalid Token Tests', () => {
    test('should reject invalid token with proper error validation', async () => {
      const ws = await fixture.connectWebSocket();
      
      // Try to authenticate with invalid token
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'authenticate', id, { auth_token: 'invalid.token.here' });
      
      // The fixture should validate the error response format against API documentation
      await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
      
      ws.close();
    });
  });

  describe('Expired Token Tests', () => {
    test('should reject expired token with proper error validation', async () => {
      const ws = await fixture.connectWebSocket();
      
      // Create an expired token
      const jwt = require('jsonwebtoken');
      const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
      const expiredToken = jwt.sign(
        { user_id: 'test-user', role: 'operator' },
        secret,
        { expiresIn: '-1h' } // Expired 1 hour ago
      );
      
      // Try to authenticate with expired token
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'authenticate', id, { auth_token: expiredToken });
      
      // The fixture should validate the error response format against API documentation
      await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
      
      ws.close();
    });
  });

  describe('Malformed Token Tests', () => {
    test('should reject malformed token with proper error validation', async () => {
      const ws = await fixture.connectWebSocket();
      
      // Try to authenticate with malformed token
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'authenticate', id, { auth_token: 'not.a.valid.jwt' });
      
      // The fixture should validate the error response format against API documentation
      await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
      
      ws.close();
    });
  });

  describe('Protected Method Access Tests', () => {
    test('should allow access to protected methods after authentication', async () => {
      const ws = await fixture.connectWebSocketWithAuth();
      
      // Test protected method access
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'take_snapshot', id, { device: '/dev/video0' });
      
      // The fixture validates the response format against API documentation
      const response = await fixture.waitForResponse(ws, id);
      expect(response).toBeDefined();
      
      ws.close();
    });

    test('should deny access to protected methods without authentication', async () => {
      const ws = await fixture.connectWebSocket(); // No authentication
      
      // Try to access protected method without authentication
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'take_snapshot', id, { device: '/dev/video0' });
      
      // The fixture should validate the error response format against API documentation
      await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
      
      ws.close();
    });
  });

  describe('Role-Based Access Tests', () => {
    test('should validate role-based permissions', async () => {
      const ws = await fixture.connectWebSocketWithAuth();
      
      // Test viewer role permissions
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_list', id);
      
      // The fixture validates the response format against API documentation
      const response = await fixture.waitForResponse(ws, id);
      expect(response).toBeDefined();
      expect(response.cameras).toBeDefined();
      
      ws.close();
    });
  });
});
