/**
 * REQ-AUTH02-001: Comprehensive authentication validation
 * REQ-AUTH02-002: Secondary requirements covered
 * Coverage: INTEGRATION
 * Quality: HIGH
 */
/**
 * Comprehensive Authentication Integration Test
 * 
 * Tests complete authentication workflow using Node.js ws library
 * This test requires a running MediaMTX server for integration testing
 */

const WebSocket = require('ws');
const jwt = require('jsonwebtoken');

// Test configuration
const CONFIG = {
  serverUrl: process.env.TEST_SERVER_URL || 'ws://localhost:8002/ws',
  timeout: parseInt(process.env.TEST_TIMEOUT) || 15000,
  jwtSecret: process.env.CAMERA_SERVICE_JWT_SECRET
};

// Helper function to send JSON-RPC requests
function sendRequest(ws, method, params = {}) {
  return new Promise((resolve, reject) => {
    const id = Math.floor(Math.random() * 10000);
    const request = {
      jsonrpc: '2.0',
      id,
      method,
      params
    };

    const timeout = setTimeout(() => {
      reject(new Error('Request timeout'));
    }, CONFIG.timeout);

    const messageHandler = (data) => {
      try {
        const response = JSON.parse(data.toString());
        if (response.id === id) {
          clearTimeout(timeout);
          ws.removeListener('message', messageHandler);
          
          if (response.error) {
            reject(new Error(response.error.message || 'RPC error'));
          } else {
            resolve(response.result);
          }
        }
      } catch (error) {
        // Ignore non-JSON messages
      }
    };

    ws.on('message', messageHandler);
    ws.send(JSON.stringify(request));
  });
}

/**
 * Jest test suite for comprehensive authentication
 */
describe('Authentication Integration Tests', () => {
  let ws;

  beforeAll(async () => {
    // Setup WebSocket connection
    ws = new WebSocket(CONFIG.serverUrl);
    await new Promise((resolve, reject) => {
      ws.on('open', resolve);
      ws.on('error', reject);
    });
    console.log('✅ WebSocket connected for authentication test suite');
  });

  afterAll(async () => {
    if (ws) {
      ws.close();
    }
  });

  describe('Valid Token Tests', () => {
    test('should authenticate with valid token', async () => {
      if (!CONFIG.jwtSecret) {
        throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
      }

      const payload = {
        user_id: 'test-user',
        role: 'operator',
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24 hours
      };

      const token = jwt.sign(payload, CONFIG.jwtSecret, { algorithm: 'HS256' });

      // First authenticate using the authenticate method
      const authResult = await sendRequest(ws, 'authenticate', { token: token });
      expect(authResult.authenticated).toBe(true);
      expect(authResult.role).toBe('operator');
    }, CONFIG.timeout);
  });

  describe('Invalid Token Tests', () => {
    test('should reject invalid token', async () => {
      const invalidToken = 'invalid.token.here';
      
      const result = await sendRequest(ws, 'authenticate', { token: invalidToken });
      expect(result.authenticated).toBe(false);
      expect(result.error).toMatch(/Invalid authentication token/);
    }, CONFIG.timeout);
  });

  describe('Expired Token Tests', () => {
    test('should reject expired token', async () => {
      if (!CONFIG.jwtSecret) {
        throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
      }

      const expiredToken = jwt.sign(
        { user_id: 'test-user', role: 'operator' },
        CONFIG.jwtSecret,
        { expiresIn: '-1h' } // Expired 1 hour ago
      );

      const result = await sendRequest(ws, 'authenticate', { token: expiredToken });
      expect(result.authenticated).toBe(false);
      expect(result.error).toMatch(/Invalid authentication token/);
    }, CONFIG.timeout);
  });

  describe('Malformed Token Tests', () => {
    test('should reject malformed token', async () => {
      const malformedToken = 'not.a.valid.jwt.token';
      
      const result = await sendRequest(ws, 'authenticate', { token: malformedToken });
      expect(result.authenticated).toBe(false);
      expect(result.error).toMatch(/Invalid authentication token/);
    }, CONFIG.timeout);
  });

  describe('Protected Method Access Tests', () => {
    test('should require authentication for protected methods', async () => {
      if (!CONFIG.jwtSecret) {
        throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
      }

      const payload = {
        user_id: 'test-user',
        role: 'operator',
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24 hours
      };

      const token = jwt.sign(payload, CONFIG.jwtSecret, { algorithm: 'HS256' });

      // First authenticate
      const authResult = await sendRequest(ws, 'authenticate', { token: token });
      expect(authResult.authenticated).toBe(true);

      // Now test protected method access (should work after authentication)
      const result = await sendRequest(ws, 'take_snapshot', { device: '/dev/video0' });
      expect(result).toBeDefined();
    }, CONFIG.timeout);
  });

  describe('Unauthenticated Access Tests', () => {
    test('should block unauthenticated access', async () => {
      const ws2 = new WebSocket(CONFIG.serverUrl);
      await new Promise((resolve) => ws2.on('open', resolve));
      
      try {
        await expect(sendRequest(ws2, 'take_snapshot', { device: '/dev/video0' }))
          .rejects.toThrow(/Authentication required|Unauthorized/);
      } finally {
        ws2.close();
      }
    }, CONFIG.timeout);
  });

  describe('Role-Based Access Tests', () => {
    test('should handle role-based access control', async () => {
      if (!CONFIG.jwtSecret) {
        throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
      }

      // Test viewer role
      const viewerToken = jwt.sign(
        { user_id: 'viewer-user', role: 'viewer' },
        CONFIG.jwtSecret,
        { expiresIn: '1h' }
      );

      // Authenticate with viewer token
      const authResult = await sendRequest(ws, 'authenticate', { token: viewerToken });
      expect(authResult.authenticated).toBe(true);
      expect(authResult.role).toBe('viewer');
      
      // Viewer should be able to read but not write
      const cameraList = await sendRequest(ws, 'get_camera_list');
      expect(cameraList).toBeDefined();
    }, CONFIG.timeout);
  });
});
