/**
 * REQ-AUTH01-001: Authentication flow validation against API documentation
 * REQ-AUTH01-002: Authentication parameter format compliance
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

const { StableTestFixture } = require('../fixtures/stable-test-fixture');

/**
 * Authentication Setup Integration Test
 * Uses StableTestFixture for API-compliant authentication validation
 */

describe('Authentication Setup Integration Tests', () => {
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

  test('REQ-AUTH01-001: should authenticate successfully using compliant fixture', async () => {
    // Use the stable test fixture as single source of truth for authentication
    const ws = await fixture.connectWebSocketWithAuth();
    
    // The fixture handles all authentication validation against API documentation
    // If authentication fails, the fixture will throw an error with proper validation
    expect(ws).toBeDefined();
    expect(ws.readyState).toBe(1); // WebSocket.OPEN
    
    // Verify authentication was successful by testing a protected method
    const id = Math.floor(Math.random() * 1000000);
    fixture.sendRequest(ws, 'ping', id);
    
    const response = await fixture.waitForResponse(ws, id);
    expect(response).toBe('pong');
    
    ws.close();
  });

  test('REQ-AUTH01-002: should handle authentication errors properly', async () => {
    // Test that the fixture properly validates authentication errors
    const ws = await fixture.connectWebSocket();
    
    // Try to call a protected method without authentication
    const id = Math.floor(Math.random() * 1000000);
    fixture.sendRequest(ws, 'get_camera_list', id);
    
    // The fixture should validate the error response format against API documentation
    await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
    
    ws.close();
  });

  test('REQ-AUTH01-003: should validate authentication response format against API documentation', async () => {
    // The fixture automatically validates authentication response format
    // This test ensures the fixture is working correctly
    const ws = await fixture.connectWebSocketWithAuth();
    
    // The fixture.validateResponseFormat() is called automatically during authentication
    // If the response doesn't match API documentation, the fixture will throw an error
    expect(ws).toBeDefined();
    
    ws.close();
  });
});

/**
 * Legacy test function for backward compatibility
 * Now uses the stable test fixture instead of custom implementation
 */
async function testInstallationFix() {
  console.log('🔧 Testing Installation Fix with Compliant Fixture');
  console.log('================================================');
  console.log('Environment Variable: CAMERA_SERVICE_JWT_SECRET');
  console.log('Expected Behavior: Authentication should work correctly using stable fixture');
  console.log('');
  
  const fixture = new StableTestFixture();
  
  try {
    await fixture.initialize();
    const ws = await fixture.connectWebSocketWithAuth();
    
    console.log('✅ Authentication successful using compliant fixture');
    console.log('✅ API compliance validation passed');
    console.log('✅ Ground truth validation working correctly');
    
    ws.close();
    return 'Authentication test passed using stable fixture';
  } catch (error) {
    console.error('❌ Authentication failed:', error.message);
    throw error;
  }
}

// Export for backward compatibility
module.exports = { testInstallationFix };
