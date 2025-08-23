/**
 * REQ-CAM02-001: Camera list retrieval and management
 * REQ-CAM02-002: Camera list response format validation against API documentation
 * Coverage: INTEGRATION
 * Quality: HIGH
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Test Categories: Integration/Camera Operations
 * API Documentation Reference: docs/api/json-rpc-methods.md
 * 
 * Uses StableTestFixture as single source of truth for authentication and validation
 */

const { StableTestFixture } = require('../fixtures/stable-test-fixture');

/**
 * Camera List Integration Test
 * Uses StableTestFixture for API-compliant camera list validation
 */

describe('Camera List Integration Tests', () => {
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

  test('REQ-CAM02-001: should retrieve camera list using compliant fixture', async () => {
    // Use the stable test fixture as single source of truth for authentication
    const ws = await fixture.connectWebSocketWithAuth();
    
    // Send get_camera_list request using the fixture
    const id = Math.floor(Math.random() * 1000000);
    fixture.sendRequest(ws, 'get_camera_list', id);
    
    // The fixture automatically validates the response format against API documentation
    const response = await fixture.waitForResponse(ws, id);
    
    // The fixture has already validated the response format, but we can add additional checks
    expect(response).toBeDefined();
    expect(response.cameras).toBeDefined();
    expect(Array.isArray(response.cameras)).toBe(true);
    expect(typeof response.total).toBe('number');
    expect(typeof response.connected).toBe('number');
    
    // Log camera information for debugging
    console.log(`📊 Camera List Results:`);
    console.log(`   Total cameras: ${response.total}`);
    console.log(`   Connected cameras: ${response.connected}`);
    
    if (response.cameras.length > 0) {
      response.cameras.forEach((camera, index) => {
        console.log(`\n   Camera ${index + 1}:`);
        console.log(`     Device: ${camera.device}`);
        console.log(`     Name: ${camera.name}`);
        console.log(`     Status: ${camera.status}`);
        console.log(`     Resolution: ${camera.resolution}`);
        console.log(`     FPS: ${camera.fps}`);
      });
    }
    
    ws.close();
  });

  test('REQ-CAM02-002: should validate camera list response format against API documentation', async () => {
    // This test ensures the fixture properly validates the camera list response format
    const ws = await fixture.connectWebSocketWithAuth();
    
    const id = Math.floor(Math.random() * 1000000);
    fixture.sendRequest(ws, 'get_camera_list', id);
    
    // The fixture.validateResponseFormat() is called automatically
    // If the response doesn't match API documentation, the fixture will throw an error
    const response = await fixture.waitForResponse(ws, id);
    
    // Additional validation that the fixture should have already done
    expect(response).toHaveProperty('cameras');
    expect(response).toHaveProperty('total');
    expect(response).toHaveProperty('connected');
    
    ws.close();
  });

  test('REQ-CAM02-003: should handle unauthorized access properly', async () => {
    // Test that unauthorized access is properly handled
    const ws = await fixture.connectWebSocket(); // No authentication
    
    const id = Math.floor(Math.random() * 1000000);
    fixture.sendRequest(ws, 'get_camera_list', id);
    
    // The fixture should validate the error response format against API documentation
    await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
    
    ws.close();
  });
});

/**
 * Legacy test function for backward compatibility
 * Now uses the stable test fixture instead of custom implementation
 */
async function testIntegration() {
  console.log('🔍 Testing Camera List Integration with Compliant Fixture');
  console.log('========================================================');
  
  const fixture = new StableTestFixture();
  
  try {
    await fixture.initialize();
    const ws = await fixture.connectWebSocketWithAuth();
    
    // Send get_camera_list request using the fixture
    const id = Math.floor(Math.random() * 1000000);
    fixture.sendRequest(ws, 'get_camera_list', id);
    
    // The fixture handles all validation against API documentation
    const response = await fixture.waitForResponse(ws, id);
    
    console.log(`\n📊 Camera List Integration Results:`);
    console.log(`   Total cameras: ${response.total}`);
    console.log(`   Connected cameras: ${response.connected}`);
    
    if (response.cameras.length > 0) {
      response.cameras.forEach((camera, index) => {
        console.log(`\n   Camera ${index + 1}:`);
        console.log(`     Device: ${camera.device}`);
        console.log(`     Name: ${camera.name}`);
        console.log(`     Status: ${camera.status}`);
        console.log(`     Resolution: ${camera.resolution}`);
        console.log(`     FPS: ${camera.fps}`);
      });
    }
    
    console.log('\n✅ Camera list integration working correctly!');
    console.log('✅ API compliance validation passed');
    console.log('✅ Ground truth validation working correctly');
    
    ws.close();
    return 'Integration test passed using stable fixture';
  } catch (error) {
    console.error('❌ Integration test failed:', error.message);
    throw error;
  }
}

// Export for backward compatibility
module.exports = { testIntegration };
