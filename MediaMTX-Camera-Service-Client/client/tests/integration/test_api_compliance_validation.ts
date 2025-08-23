/**
 * API Compliance Validation Tests
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Requirements Coverage:
 * - REQ-API-001: API documentation compliance validation
 * - REQ-API-002: JSON-RPC method format validation
 * - REQ-API-003: Authentication flow validation
 * - REQ-API-004: Error response format validation
 * 
 * Test Categories: Integration/API Compliance
 * API Documentation Reference: docs/api/json-rpc-methods.md
 */

import { WebSocket } from 'ws';
import { TEST_CONFIG, getWebSocketUrl, validateTestEnvironment } from '../config/test-config';

/**
 * API Compliance Test for authenticate method
 * 
 * Ground Truth: Server API documentation (json-rpc-methods.md)
 * Method: authenticate
 * Expected Request Format: { jsonrpc: "2.0", method: "authenticate", params: { auth_token: string }, id: number }
 * Expected Response Format: { jsonrpc: "2.0", result: { authenticated: boolean, role: string, permissions: string[], expires_at: string, session_id: string }, id: number }
 * Expected Error Codes: -32001 (Authentication failed)
 */
describe('API Compliance Tests - authenticate method', () => {
  let ws: WebSocket;
  let authToken: string;

  beforeAll(async () => {
    // Validate test environment setup
    if (!validateTestEnvironment()) {
      throw new Error('Test environment not properly set up. Run ./set-test-env.sh to configure authentication.');
    }
    
    // Generate valid authentication token
    const jwt = require('jsonwebtoken');
    const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
    if (!secret) {
      throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable required');
    }
    
    authToken = jwt.sign(
      { user_id: 'test_user', role: 'operator' },
      secret,
      { expiresIn: '1h' }
    );
  });

  beforeEach(async () => {
    // Establish real WebSocket connection
    ws = new WebSocket(getWebSocketUrl());
    
    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('Connection timeout')), 5000);
      
      ws.onopen = () => {
        clearTimeout(timeout);
        resolve();
      };
      
      ws.onerror = (error: any) => {
        clearTimeout(timeout);
        reject(error);
      };
    });
  });

  afterEach(async () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.close();
    }
  });

  test('REQ-API-001: authenticate validates against API documentation', async () => {
    // 1. Use documented request format from API documentation
    const request = {
      jsonrpc: "2.0",
      method: "authenticate",
      params: {
        auth_token: authToken
      },
      id: 1
    };
    
    // 2. Send request and wait for response
    const response = await sendRequest(ws, request);
    
    // 3. Validate documented response format
    expect(response).toHaveProperty('jsonrpc', '2.0');
    expect(response).toHaveProperty('id', 1);
    expect(response).toHaveProperty('result');
    
    const result = response.result;
    
    // 4. Check all documented fields are present
    const requiredFields = ["authenticated", "role", "permissions", "expires_at", "session_id"];
    requiredFields.forEach(field => {
      expect(result).toHaveProperty(field, `Missing required field '${field}' per API documentation`);
    });
    
    // 5. Validate field types and values
    expect(typeof result.authenticated).toBe('boolean');
    expect(typeof result.role).toBe('string');
    expect(Array.isArray(result.permissions)).toBe(true);
    expect(typeof result.expires_at).toBe('string');
    expect(typeof result.session_id).toBe('string');
    
    // 6. Validate authentication was successful
    expect(result.authenticated).toBe(true);
    expect(result.role).toBe('operator');
    expect(result.permissions).toContain('view');
    expect(result.permissions).toContain('control');
  });

  test('REQ-API-002: authenticate handles invalid token per API documentation', async () => {
    // 1. Use documented request format with invalid token
    const request = {
      jsonrpc: "2.0",
      method: "authenticate",
      params: {
        auth_token: "invalid_token"
      },
      id: 2
    };
    
    // 2. Send request and expect error response
    const response = await sendRequest(ws, request);
    
    // 3. Validate documented error format
    expect(response).toHaveProperty('jsonrpc', '2.0');
    expect(response).toHaveProperty('id', 2);
    expect(response).toHaveProperty('error');
    
    const error = response.error;
    
    // 4. Check documented error fields
    expect(error).toHaveProperty('code', -32001);
    expect(error).toHaveProperty('message', 'Authentication failed');
    expect(error).toHaveProperty('data');
    expect(error.data).toHaveProperty('reason');
  });
});

/**
 * API Compliance Test for get_camera_list method
 * 
 * Ground Truth: Server API documentation (json-rpc-methods.md)
 * Method: get_camera_list
 * Expected Request Format: { jsonrpc: "2.0", method: "get_camera_list", id: number }
 * Expected Response Format: { jsonrpc: "2.0", result: { cameras: array, total: number, connected: number }, id: number }
 */
describe('API Compliance Tests - get_camera_list method', () => {
  let ws: WebSocket;
  let authToken: string;

  beforeAll(async () => {
    // Validate test environment setup
    if (!validateTestEnvironment()) {
      throw new Error('Test environment not properly set up. Run ./set-test-env.sh to configure authentication.');
    }
    
    // Generate valid authentication token
    const jwt = require('jsonwebtoken');
    const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
    if (!secret) {
      throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable required');
    }
    
    authToken = jwt.sign(
      { user_id: 'test_user', role: 'operator' },
      secret,
      { expiresIn: '1h' }
    );
  });

  beforeEach(async () => {
    // Establish real WebSocket connection
    ws = new WebSocket(getWebSocketUrl());
    
    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('Connection timeout')), 5000);
      
      ws.onopen = () => {
        clearTimeout(timeout);
        resolve();
      };
      
      ws.onerror = (error: any) => {
        clearTimeout(timeout);
        reject(error);
      };
    });

    // Authenticate first (required for get_camera_list)
    const authRequest = {
      jsonrpc: "2.0",
      method: "authenticate",
      params: {
        auth_token: authToken
      },
      id: 0
    };
    
    await sendRequest(ws, authRequest);
  });

  afterEach(async () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.close();
    }
  });

  test('REQ-API-003: get_camera_list validates against API documentation', async () => {
    // 1. Use documented request format from API documentation
    const request = {
      jsonrpc: "2.0",
      method: "get_camera_list",
      id: 3
    };
    
    // 2. Send request and wait for response
    const response = await sendRequest(ws, request);
    
    // 3. Validate documented response format
    expect(response).toHaveProperty('jsonrpc', '2.0');
    expect(response).toHaveProperty('id', 3);
    expect(response).toHaveProperty('result');
    
    const result = response.result;
    
    // 4. Check all documented fields are present
    const requiredFields = ["cameras", "total", "connected"];
    requiredFields.forEach(field => {
      expect(result).toHaveProperty(field, `Missing required field '${field}' per API documentation`);
    });
    
    // 5. Validate field types
    expect(Array.isArray(result.cameras)).toBe(true);
    expect(typeof result.total).toBe('number');
    expect(typeof result.connected).toBe('number');
    
    // 6. Validate camera object structure if cameras exist
    if (result.cameras.length > 0) {
      const camera = result.cameras[0];
      const cameraFields = ["device", "status", "name", "resolution", "fps", "streams"];
      cameraFields.forEach(field => {
        expect(camera).toHaveProperty(field, `Missing required camera field '${field}' per API documentation`);
      });
      
      // Validate streams object
      expect(camera.streams).toHaveProperty('rtsp');
      expect(camera.streams).toHaveProperty('webrtc');
      expect(camera.streams).toHaveProperty('hls');
    }
  });
});

/**
 * Helper function to send JSON-RPC requests and wait for responses
 */
function sendRequest(ws: WebSocket, request: any): Promise<any> {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error(`Request timeout for ${request.method}`));
    }, 10000);
    
    const messageHandler = (data: WebSocket.Data) => {
      try {
        const response = JSON.parse(data.toString());
        if (response.id === request.id) {
          clearTimeout(timeout);
          ws.removeListener('message', messageHandler);
          resolve(response);
        }
      } catch (error) {
        console.error('Failed to parse response:', error);
        reject(error);
      }
    };
    
    ws.on('message', messageHandler);
    ws.send(JSON.stringify(request));
  });
}
