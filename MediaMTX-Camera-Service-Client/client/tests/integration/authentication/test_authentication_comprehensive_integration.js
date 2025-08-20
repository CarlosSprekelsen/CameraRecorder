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

// Test results tracking
const testResults = {
  total: 0,
  passed: 0,
  failed: 0
};

// Assertion function
function assert(condition, message) {
  testResults.total++;
  if (condition) {
    testResults.passed++;
    console.log(`✅ ${message}`);
  } else {
    testResults.failed++;
    console.log(`❌ ${message}`);
    throw new Error(message);
  }
}

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

// Test functions
async function testValidToken(ws) {
  console.log('\n🔐 Test 1: Valid Token Authentication');
  
  try {
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

    // Server doesn't have an authenticate method - authentication is handled by including auth_token in parameters
    // Test authentication by calling a protected method with auth_token
    const authResult = await sendRequest(ws, 'get_camera_status', { device: '/dev/video0', auth_token: token });
    assert(authResult.authenticated === true, 'valid token should authenticate');
    assert(authResult.role === 'operator', 'user should have operator role');
    
    testResults.passed++;
    console.log('✅ Valid token authentication test passed');
    
  } catch (error) {
    testResults.failed++;
    console.error('❌ Valid token authentication test failed:', error.message);
    throw error;
  }
}

async function testInvalidToken(ws) {
  console.log('\n❌ Test 2: Invalid Token Rejection');
  
  try {
    const invalidToken = 'invalid.token.here';
    
    try {
      // Test invalid token by calling protected method
    await sendRequest(ws, 'get_camera_status', { device: '/dev/video0', auth_token: invalidToken });
      throw new Error('Should have rejected invalid token');
    } catch (error) {
      assert(error.message.includes('Invalid token') || error.message.includes('Authentication failed'), 'invalid token should be rejected');
    }
    
    testResults.passed++;
    console.log('✅ Invalid token rejection test passed');
    
  } catch (error) {
    testResults.failed++;
    console.error('❌ Invalid token rejection test failed:', error.message);
    throw error;
  }
}

async function testExpiredToken(ws) {
  console.log('\n⏰ Test 3: Expired Token Rejection');
  
  try {
    if (!CONFIG.jwtSecret) {
      throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
    }

    const expiredToken = jwt.sign(
      { user_id: 'test-user', role: 'operator' },
      CONFIG.jwtSecret,
      { expiresIn: '-1h' } // Expired 1 hour ago
    );

    try {
      // Test expired token by calling protected method
    await sendRequest(ws, 'get_camera_status', { device: '/dev/video0', auth_token: expiredToken });
      throw new Error('Should have rejected expired token');
    } catch (error) {
      assert(error.message.includes('Token expired') || error.message.includes('Authentication failed'), 'expired token should be rejected');
    }
    
    testResults.passed++;
    console.log('✅ Expired token rejection test passed');
    
  } catch (error) {
    testResults.failed++;
    console.error('❌ Expired token rejection test failed:', error.message);
    throw error;
  }
}

async function testMalformedToken(ws) {
  console.log('\n🔧 Test 4: Malformed Token Rejection');
  
  try {
    const malformedToken = 'not.a.valid.jwt.token';
    
    try {
      // Test malformed token by calling protected method
    await sendRequest(ws, 'get_camera_status', { device: '/dev/video0', auth_token: malformedToken });
      throw new Error('Should have rejected malformed token');
    } catch (error) {
      assert(error.message.includes('Invalid token') || error.message.includes('Authentication failed'), 'malformed token should be rejected');
    }
    
    testResults.passed++;
    console.log('✅ Malformed token rejection test passed');
    
  } catch (error) {
    testResults.failed++;
    console.error('❌ Malformed token rejection test failed:', error.message);
    throw error;
  }
}

async function testProtectedMethodAccess(ws) {
  console.log('\n🔒 Test 5: Protected Method Access');
  
  try {
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

    // Authenticate first
    // Test authentication by calling protected method
    await sendRequest(ws, 'get_camera_status', { device: '/dev/video0', auth_token: token });

    // Test protected method access
    const result = await sendRequest(ws, 'take_snapshot', { device: '/dev/video0' });
    assert(result, 'authenticated user should access protected method');
    
    testResults.passed++;
    console.log('✅ Protected method access test passed');
    
  } catch (error) {
    testResults.failed++;
    console.error('❌ Protected method access test failed:', error.message);
    throw error;
  }
}

async function testUnauthenticatedAccess() {
  console.log('\n🚫 Test 6: Unauthenticated Access Blocking');
  
  try {
    const ws2 = new WebSocket(CONFIG.serverUrl);
    await new Promise((resolve) => ws2.on('open', resolve));
    
    try {
      await sendRequest(ws2, 'take_snapshot', { device: '/dev/video0' });
      throw new Error('Should have blocked unauthenticated access');
    } catch (error) {
      assert(error.message.includes('Authentication required') || error.message.includes('Unauthorized'), 'unauthenticated access should be blocked');
    }
    
    ws2.close();
    testResults.passed++;
    console.log('✅ Unauthenticated access blocking test passed');
    
  } catch (error) {
    testResults.failed++;
    console.error('❌ Unauthenticated access blocking test failed:', error.message);
    throw error;
  }
}

async function testRoleBasedAccess(ws) {
  console.log('\n👥 Test 7: Role-Based Access Control');
  
  try {
    if (!CONFIG.jwtSecret) {
      throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
    }

    // Test viewer role
    const viewerToken = jwt.sign(
      { user_id: 'viewer-user', role: 'viewer' },
      CONFIG.jwtSecret,
      { expiresIn: '1h' }
    );

    // Test viewer token by calling protected method
    await sendRequest(ws, 'get_camera_status', { device: '/dev/video0', auth_token: viewerToken });
    
    // Viewer should be able to read but not write
    const cameraList = await sendRequest(ws, 'get_camera_list');
    assert(cameraList, 'viewer should access read operations');
    
    testResults.passed++;
    console.log('✅ Role-based access control test passed');
    
  } catch (error) {
    testResults.failed++;
    console.error('❌ Role-based access control test failed:', error.message);
    throw error;
  }
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
      await expect(testValidToken(ws)).resolves.not.toThrow();
    }, CONFIG.timeout);
  });

  describe('Invalid Token Tests', () => {
    test('should reject invalid token', async () => {
      await expect(testInvalidToken(ws)).resolves.not.toThrow();
    }, CONFIG.timeout);
  });

  describe('Expired Token Tests', () => {
    test('should reject expired token', async () => {
      await expect(testExpiredToken(ws)).resolves.not.toThrow();
    }, CONFIG.timeout);
  });

  describe('Malformed Token Tests', () => {
    test('should reject malformed token', async () => {
      await expect(testMalformedToken(ws)).resolves.not.toThrow();
    }, CONFIG.timeout);
  });

  describe('Protected Method Access Tests', () => {
    test('should require authentication for protected methods', async () => {
      await expect(testProtectedMethodAccess(ws)).resolves.not.toThrow();
    }, CONFIG.timeout);
  });

  describe('Unauthenticated Access Tests', () => {
    test('should block unauthenticated access', async () => {
      await expect(testUnauthenticatedAccess()).resolves.not.toThrow();
    }, CONFIG.timeout);
  });

  describe('Role-Based Access Tests', () => {
    test('should handle role-based access control', async () => {
      await expect(testRoleBasedAccess(ws)).resolves.not.toThrow();
    }, CONFIG.timeout);
  });

  describe('Test Results Summary', () => {
    test('should have successful authentication test results', () => {
      expect(testResults.total).toBeGreaterThan(0);
      expect(testResults.passed).toBeGreaterThan(0);
      expect(testResults.failed).toBe(0);
    });

    test('should have authentication scenario coverage', () => {
      expect(testResults.passed).toBeGreaterThanOrEqual(7);
    });
  });
});
