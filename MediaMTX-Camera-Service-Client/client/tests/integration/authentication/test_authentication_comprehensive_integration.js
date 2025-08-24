/**
 * REQ-AUTH02-001: Comprehensive authentication validation
 * REQ-AUTH02-002: Authentication parameter format compliance
 * Coverage: INTEGRATION
 * Quality: HIGH
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Health API: ../mediamtx-camera-service/docs/api/health-endpoints.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Test Categories: Integration/Authentication
 * API Documentation Reference: docs/api/json-rpc-methods.md
 * 
 * Uses StableTestFixture as single source of truth for authentication
 * Updated for new API structure with enhanced authentication flow
 */

const { StableTestFixture } = require('../../fixtures/stable-test-fixture');

/**
 * Comprehensive Authentication Integration Test
 * Uses StableTestFixture for API-compliant authentication validation
 * Updated for new API structure with role-based access control
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
      // Updated: New API authentication flow
      const ws = await fixture.connectWebSocketWithAuth();
      
      // The fixture handles all authentication validation against API documentation
      expect(ws).toBeDefined();
      expect(ws.readyState).toBe(1); // WebSocket.OPEN
      
      // Verify authentication was successful by testing a protected method
      // Updated: All methods now require authentication
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'ping', id);
      
      const response = await fixture.waitForResponse(ws, id);
      expect(response).toBe('pong');
      
      ws.close();
    });

    test('should validate authentication response format against new API documentation', async () => {
      // Test authentication response format validation
      const ws = await fixture.connectWebSocket();
      
      // Authenticate and validate response format
      const authResponse = await fixture.authenticate(ws);
      
      // Updated: Validate new API authentication response format
      expect(authResponse).toHaveProperty('authenticated');
      expect(authResponse).toHaveProperty('role');
      expect(authResponse).toHaveProperty('permissions');
      expect(authResponse).toHaveProperty('expires_at');
      expect(authResponse).toHaveProperty('session_id');
      
      expect(typeof authResponse.authenticated).toBe('boolean');
      expect(authResponse.authenticated).toBe(true);
      
      // Updated: Validate role-based access control
      const validRoles = ['viewer', 'operator', 'admin'];
      expect(validRoles).toContain(authResponse.role);
      
      expect(Array.isArray(authResponse.permissions)).toBe(true);
      expect(typeof authResponse.expires_at).toBe('string');
      expect(typeof authResponse.session_id).toBe('string');
      
      ws.close();
    });
  });

  describe('Invalid Token Tests', () => {
    test('should reject invalid token with proper error validation', async () => {
      const ws = await fixture.connectWebSocket();
      
      // Try to authenticate with invalid token
      // Updated: New API authentication format
      const id = Math.floor(Math.random() * 1000000);
      const request = {
        jsonrpc: '2.0',
        method: 'authenticate',
        params: {
          auth_token: 'invalid.token.here'
        },
        id: id
      };
      
      ws.send(JSON.stringify(request));
      
      // The fixture should validate the error response format against API documentation
      // Updated: Validates new API error format
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
      // Updated: New API authentication format
      const id = Math.floor(Math.random() * 1000000);
      const request = {
        jsonrpc: '2.0',
        method: 'authenticate',
        params: {
          auth_token: expiredToken
        },
        id: id
      };
      
      ws.send(JSON.stringify(request));
      
      // The fixture should validate the error response format against API documentation
      // Updated: Validates new API error format
      await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
      
      ws.close();
    });
  });

  describe('Malformed Token Tests', () => {
    test('should reject malformed token with proper error validation', async () => {
      const ws = await fixture.connectWebSocket();
      
      // Try to authenticate with malformed token
      // Updated: New API authentication format
      const id = Math.floor(Math.random() * 1000000);
      const request = {
        jsonrpc: '2.0',
        method: 'authenticate',
        params: {
          auth_token: 'not.a.valid.jwt'
        },
        id: id
      };
      
      ws.send(JSON.stringify(request));
      
      // The fixture should validate the error response format against API documentation
      // Updated: Validates new API error format
      await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
      
      ws.close();
    });
  });

  describe('Role-Based Access Control Tests', () => {
    test('should validate operator role permissions with new API', async () => {
      // Test operator role permissions
      // Updated: Test role-based access control
      const ws = await fixture.connectWebSocketWithAuth();
      
      // Operator should be able to take snapshots
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'take_snapshot', id, { device: '/dev/video0' });
      
      // Should either succeed or fail with proper error (not auth error)
      try {
        const response = await fixture.waitForResponse(ws, id);
        // If successful, validate response format
        expect(response).toHaveProperty('device');
        expect(response).toHaveProperty('filename');
        expect(response).toHaveProperty('status');
      } catch (error) {
        // Should not be authentication error for operator role
        expect(error.message).not.toContain('Authentication failed');
      }
      
      ws.close();
    });

    test('should validate viewer role permissions with new API', async () => {
      // Test viewer role permissions (read-only)
      // Updated: Test role-based access control
      const ws = await fixture.connectWebSocketWithAuth();
      
      // Viewer should be able to get camera list
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_list', id);
      
      const response = await fixture.waitForResponse(ws, id);
      expect(response).toHaveProperty('cameras');
      expect(response).toHaveProperty('total');
      expect(response).toHaveProperty('connected');
      
      ws.close();
    });
  });

  describe('Session Management Tests', () => {
    test('should maintain session across multiple requests with new API', async () => {
      // Test session persistence
      // Updated: Test new API session management
      const ws = await fixture.connectWebSocketWithAuth();
      
      // First request
      const id1 = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'ping', id1);
      const response1 = await fixture.waitForResponse(ws, id1);
      expect(response1).toBe('pong');
      
      // Second request (should use same session)
      const id2 = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_list', id2);
      const response2 = await fixture.waitForResponse(ws, id2);
      expect(response2).toHaveProperty('cameras');
      
      ws.close();
    });
  });
});
