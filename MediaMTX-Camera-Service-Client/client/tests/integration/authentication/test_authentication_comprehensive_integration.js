#!/usr/bin/env node

/**
 * Comprehensive Authentication Test
 * 
 * This test validates authentication functionality against the real MediaMTX Camera Service server
 * following the actual server API specification.
 * 
 * Server API Methods Tested:
 * - authenticate(token) - JWT token authentication
 * - Protected method access (take_snapshot, start_recording, stop_recording)
 * 
 * Usage: node test-authentication-comprehensive.js
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running on localhost:8002
 * - WebSocket endpoint available at ws://localhost:8002/ws
 * - Valid JWT secret configured
 */

import WebSocket from 'ws';
import jwt from 'jsonwebtoken';

// Test configuration
const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  timeout: 15000,
  device: '/dev/video0'
};

// Get JWT secret from environment (no fallback to hardcoded value)
const getJwtSecret = () => {
  const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
  if (!secret) {
    throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
  }
  return secret;
};

// Test results tracking
const testResults = {
  passed: 0,
  failed: 0,
  total: 0,
  errors: [],
  authScenarios: {
    validToken: false,
    invalidToken: false,
    expiredToken: false,
    malformedToken: false,
    protectedMethodAccess: false,
    unauthenticatedAccess: false
  }
};

/**
 * Generate a valid JWT token for authentication
 */
function generateValidToken() {
  const payload = {
    user_id: 'test_user',
    role: 'operator',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60)
  };
  
  return jwt.sign(payload, getJwtSecret(), { algorithm: 'HS256' });
}

/**
 * Generate an expired JWT token
 */
function generateExpiredToken() {
  const payload = {
    user_id: 'test_user',
    role: 'operator',
    iat: Math.floor(Date.now() / 1000) - (24 * 60 * 60), // 24 hours ago
    exp: Math.floor(Date.now() / 1000) - (60 * 60) // 1 hour ago
  };
  
  return jwt.sign(payload, getJwtSecret(), { algorithm: 'HS256' });
}

/**
 * Generate an invalid JWT token (wrong secret)
 */
function generateInvalidToken() {
  const payload = {
    user_id: 'test_user',
    role: 'operator',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60)
  };
  
  return jwt.sign(payload, 'wrong_secret', { algorithm: 'HS256' });
}

/**
 * Utility function to send JSON-RPC requests
 */
function sendRequest(ws, method, params = {}) {
  return new Promise((resolve, reject) => {
    const id = Math.floor(Math.random() * 10000);
    const request = {
      jsonrpc: '2.0',
      method: method,
      params: params,
      id: id
    };
    
    console.log(`📤 Sending ${method} (#${id})`, JSON.stringify(params));
    
    const timeout = setTimeout(() => {
      reject(new Error(`Request timeout for ${method}`));
    }, CONFIG.timeout);
    
    const messageHandler = (data) => {
      try {
        const response = JSON.parse(data);
        if (response.id === id) {
          clearTimeout(timeout);
          ws.removeListener('message', messageHandler);
          
          if (response.error) {
            console.log(`📥 Error response:`, response.error);
            reject(new Error(response.error.message || 'RPC error'));
          } else {
            console.log(`📥 Success response:`, response.result);
            resolve(response.result);
          }
        }
      } catch (error) {
        console.error('❌ Failed to parse response:', error);
        reject(error);
      }
    };
    
    ws.on('message', messageHandler);
    ws.send(JSON.stringify(request));
  });
}

/**
 * Test result assertion
 */
function assert(condition, message) {
  testResults.total++;
  if (condition) {
    testResults.passed++;
    console.log(`✅ ${message}`);
  } else {
    testResults.failed++;
    console.log(`❌ ${message}`);
    testResults.errors.push(message);
  }
}

/**
 * Test 1: Valid Token Authentication
 */
async function testValidToken(ws) {
  console.log('\n🔐 Test 1: Valid Token Authentication');
  
  try {
    const token = generateValidToken();
    console.log('\n🔑 Authenticating with valid JWT token...');
    const authResult = await sendRequest(ws, 'authenticate', { token });
    
    assert(authResult.authenticated === true, 'valid token authentication successful');
    assert(authResult.role === 'operator', 'valid token has operator role');
    assert(authResult.auth_method === 'jwt', 'valid token uses JWT auth method');
    
    testResults.authScenarios.validToken = true;
    console.log('✅ Valid token authentication test completed');
    
  } catch (error) {
    console.error('❌ Valid token authentication test failed:', error.message);
    throw error;
  }
}

/**
 * Test 2: Invalid Token Authentication
 */
async function testInvalidToken(ws) {
  console.log('\n❌ Test 2: Invalid Token Authentication');
  
  try {
    const token = generateInvalidToken();
    console.log('\n🔑 Attempting authentication with invalid JWT token...');
    
    try {
      await sendRequest(ws, 'authenticate', { token });
      assert(false, 'should have rejected invalid token');
    } catch (error) {
      assert(error.message.includes('Invalid authentication token'), 'invalid token properly rejected');
    }
    
    testResults.authScenarios.invalidToken = true;
    console.log('✅ Invalid token authentication test completed');
    
  } catch (error) {
    console.error('❌ Invalid token authentication test failed:', error.message);
    throw error;
  }
}

/**
 * Test 3: Expired Token Authentication
 */
async function testExpiredToken(ws) {
  console.log('\n⏰ Test 3: Expired Token Authentication');
  
  try {
    const token = generateExpiredToken();
    console.log('\n🔑 Attempting authentication with expired JWT token...');
    
    try {
      await sendRequest(ws, 'authenticate', { token });
      assert(false, 'should have rejected expired token');
    } catch (error) {
      assert(error.message.includes('Invalid authentication token') || error.message.includes('expired'), 'expired token properly rejected');
    }
    
    testResults.authScenarios.expiredToken = true;
    console.log('✅ Expired token authentication test completed');
    
  } catch (error) {
    console.error('❌ Expired token authentication test failed:', error.message);
    throw error;
  }
}

/**
 * Test 4: Malformed Token Authentication
 */
async function testMalformedToken(ws) {
  console.log('\n🔧 Test 4: Malformed Token Authentication');
  
  try {
    const malformedToken = 'not.a.valid.jwt.token';
    console.log('\n🔑 Attempting authentication with malformed JWT token...');
    
    try {
      await sendRequest(ws, 'authenticate', { token: malformedToken });
      assert(false, 'should have rejected malformed token');
    } catch (error) {
      assert(error.message.includes('Invalid authentication token'), 'malformed token properly rejected');
    }
    
    testResults.authScenarios.malformedToken = true;
    console.log('✅ Malformed token authentication test completed');
    
  } catch (error) {
    console.error('❌ Malformed token authentication test failed:', error.message);
    throw error;
  }
}

/**
 * Test 5: Protected Method Access (Authenticated)
 */
async function testProtectedMethodAccess(ws) {
  console.log('\n🔒 Test 5: Protected Method Access (Authenticated)');
  
  try {
    // First authenticate
    const token = generateValidToken();
    const authResult = await sendRequest(ws, 'authenticate', { token });
    assert(authResult.authenticated === true, 'authentication successful for protected method test');
    
    // Test protected method access
    console.log('\n📸 Testing protected method: take_snapshot');
    try {
      await sendRequest(ws, 'take_snapshot', { device: CONFIG.device });
      console.log('✅ Protected method accessible after authentication');
    } catch (error) {
      // Snapshot might fail due to hardware, but authentication should work
      if (error.message.includes('Authentication required')) {
        assert(false, 'protected method still requires authentication after auth');
      } else {
        console.log('⚠️ Protected method failed for non-auth reason (expected):', error.message);
      }
    }
    
    testResults.authScenarios.protectedMethodAccess = true;
    console.log('✅ Protected method access test completed');
    
  } catch (error) {
    console.error('❌ Protected method access test failed:', error.message);
    throw error;
  }
}

/**
 * Test 6: Unauthenticated Access to Protected Methods
 */
async function testUnauthenticatedAccess() {
  console.log('\n🚫 Test 6: Unauthenticated Access to Protected Methods');
  
  try {
    const ws2 = new WebSocket(CONFIG.serverUrl);
    await new Promise((resolve) => ws2.on('open', resolve));
    
    // Test unauthenticated access to protected methods
    console.log('\n📸 Testing unauthenticated access to take_snapshot');
    try {
      await sendRequest(ws2, 'take_snapshot', { device: CONFIG.device });
      assert(false, 'should have blocked unauthenticated access');
    } catch (error) {
      assert(error.message.includes('Authentication required'), 'unauthenticated access properly blocked');
    }
    
    console.log('\n🎬 Testing unauthenticated access to start_recording');
    try {
      await sendRequest(ws2, 'start_recording', { device: CONFIG.device, duration: 10 });
      assert(false, 'should have blocked unauthenticated access');
    } catch (error) {
      assert(error.message.includes('Authentication required'), 'unauthenticated access properly blocked');
    }
    
    ws2.close();
    testResults.authScenarios.unauthenticatedAccess = true;
    console.log('✅ Unauthenticated access test completed');
    
  } catch (error) {
    console.error('❌ Unauthenticated access test failed:', error.message);
    throw error;
  }
}

/**
 * Test 7: Role-Based Access Control
 */
async function testRoleBasedAccess(ws) {
  console.log('\n👥 Test 7: Role-Based Access Control');
  
  try {
    // Test with operator role
    const operatorToken = generateValidToken(); // Default is operator
    const operatorAuth = await sendRequest(ws, 'authenticate', { token: operatorToken });
    assert(operatorAuth.role === 'operator', 'operator role correctly assigned');
    
    // Test with viewer role (if supported)
    const viewerPayload = {
      user_id: 'test_viewer',
      role: 'viewer',
      iat: Math.floor(Date.now() / 1000),
      exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60)
    };
    const viewerToken = jwt.sign(viewerPayload, CONFIG.jwtSecret, { algorithm: 'HS256' });
    
    try {
      const viewerAuth = await sendRequest(ws, 'authenticate', { token: viewerToken });
      assert(viewerAuth.role === 'viewer', 'viewer role correctly assigned');
      console.log('✅ Role-based access control test completed');
    } catch (error) {
      console.log('⚠️ Viewer role test failed (may not be supported):', error.message);
    }
    
  } catch (error) {
    console.error('❌ Role-based access control test failed:', error.message);
    throw error;
  }
}

/**
 * Main test execution
 */
async function runTests() {
  console.log('🔐 Comprehensive Authentication Test');
  console.log('==================================');
  console.log(`Server: ${CONFIG.serverUrl}`);
  console.log(`Timeout: ${CONFIG.timeout}ms`);
  console.log('');
  
  const ws = new WebSocket(CONFIG.serverUrl);
  
  return new Promise((resolve, reject) => {
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Run all test suites
        await testValidToken(ws);
        await testInvalidToken(ws);
        await testExpiredToken(ws);
        await testMalformedToken(ws);
        await testProtectedMethodAccess(ws);
        await testUnauthenticatedAccess();
        await testRoleBasedAccess(ws);
        
        // Summary
        console.log('\n📊 Test Summary');
        console.log('==============');
        console.log(`Total Tests: ${testResults.total}`);
        console.log(`Passed: ${testResults.passed}`);
        console.log(`Failed: ${testResults.failed}`);
        console.log(`Success Rate: ${Math.round((testResults.passed / testResults.total) * 100)}%`);
        
        console.log('\n🔐 Authentication Scenario Coverage');
        console.log('==================================');
        Object.entries(testResults.authScenarios).forEach(([scenario, tested]) => {
          console.log(`${tested ? '✅' : '❌'} ${scenario}`);
        });
        
        if (testResults.failed === 0) {
          console.log('\n🎉 All authentication tests passed! Security is properly implemented.');
        } else {
          console.log('\n❌ Some authentication tests failed. Please check the errors above.');
        }
        
        ws.close();
        resolve();
        
      } catch (error) {
        console.error('❌ Authentication test suite failed:', error.message);
        ws.close();
        reject(error);
      }
    });
    
    ws.on('error', (error) => {
      console.error('❌ WebSocket connection failed:', error.message);
      reject(error);
    });
  });
}

// Run the tests
if (import.meta.url === `file://${process.argv[1]}`) {
  runTests().catch(error => {
    console.error('❌ Test execution failed:', error);
    process.exit(1);
  });
}
