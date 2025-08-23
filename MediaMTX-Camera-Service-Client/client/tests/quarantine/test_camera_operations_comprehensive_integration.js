/**
 * REQ-CAM03-001: Comprehensive camera operations validation
 * REQ-CAM03-002: Camera operations response format compliance
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
 * Comprehensive Camera Operations Integration Test
 * Uses StableTestFixture for API-compliant camera operations validation
 */

describe('Camera Operations Integration Tests', () => {
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

  describe('Basic API Functionality', () => {
    test('REQ-CAM03-001: should handle ping using compliant fixture', async () => {
      const ws = await fixture.connectWebSocketWithAuth();
      
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'ping', id);
      
      // The fixture validates the response format against API documentation
      const response = await fixture.waitForResponse(ws, id);
      expect(response).toBe('pong');
      
      ws.close();
    });

    test('REQ-CAM03-002: should get camera list using compliant fixture', async () => {
      const ws = await fixture.connectWebSocketWithAuth();
      
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_list', id);
      
      // The fixture validates the response format against API documentation
      const response = await fixture.waitForResponse(ws, id);
      expect(response).toBeDefined();
      expect(response.cameras).toBeDefined();
      expect(Array.isArray(response.cameras)).toBe(true);
      expect(typeof response.total).toBe('number');
      expect(typeof response.connected).toBe('number');
      
      ws.close();
    });

    test('REQ-CAM03-003: should get camera status using compliant fixture', async () => {
      const ws = await fixture.connectWebSocketWithAuth();
      
      // First get camera list to find a valid device
      const listId = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_list', listId);
      const cameraList = await fixture.waitForResponse(ws, listId);
      
      if (cameraList.cameras.length > 0) {
        const testDevice = cameraList.cameras[0].device;
        
        const statusId = Math.floor(Math.random() * 1000000);
        fixture.sendRequest(ws, 'get_camera_status', statusId, { device: testDevice });
        
        // The fixture validates the response format against API documentation
        const response = await fixture.waitForResponse(ws, statusId);
        expect(response).toBeDefined();
        expect(response.device).toBe(testDevice);
        expect(response.status).toBeDefined();
        expect(response.name).toBeDefined();
        expect(response.resolution).toBeDefined();
        expect(response.fps).toBeDefined();
        expect(response.streams).toBeDefined();
      }
      
      ws.close();
    });
  });

  describe('Camera Control Operations', () => {
    test('REQ-CAM03-004: should take snapshot using compliant fixture', async () => {
      const ws = await fixture.connectWebSocketWithAuth();
      
      // First get camera list to find a valid device
      const listId = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_list', listId);
      const cameraList = await fixture.waitForResponse(ws, listId);
      
      if (cameraList.cameras.length > 0) {
        const testDevice = cameraList.cameras[0].device;
        
        const snapshotId = Math.floor(Math.random() * 1000000);
        fixture.sendRequest(ws, 'take_snapshot', snapshotId, { device: testDevice });
        
        // The fixture validates the response format against API documentation
        const response = await fixture.waitForResponse(ws, snapshotId);
        expect(response).toBeDefined();
        expect(response.status).toBeDefined();
        expect(response.filename).toBeDefined();
        expect(response.file_size).toBeDefined();
        expect(response.format).toBeDefined();
        expect(response.quality).toBeDefined();
      }
      
      ws.close();
    });

    test('REQ-CAM03-005: should start recording using compliant fixture', async () => {
      const ws = await fixture.connectWebSocketWithAuth();
      
      // First get camera list to find a valid device
      const listId = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_list', listId);
      const cameraList = await fixture.waitForResponse(ws, listId);
      
      if (cameraList.cameras.length > 0) {
        const testDevice = cameraList.cameras[0].device;
        
        const startId = Math.floor(Math.random() * 1000000);
        fixture.sendRequest(ws, 'start_recording', startId, { device: testDevice });
        
        // The fixture validates the response format against API documentation
        const response = await fixture.waitForResponse(ws, startId);
        expect(response).toBeDefined();
        expect(response.session_id).toBeDefined();
        expect(response.status).toBeDefined();
        expect(response.start_time).toBeDefined();
        
        // Stop recording
        const stopId = Math.floor(Math.random() * 1000000);
        fixture.sendRequest(ws, 'stop_recording', stopId, { device: testDevice });
        
        const stopResponse = await fixture.waitForResponse(ws, stopId);
        expect(stopResponse).toBeDefined();
        expect(stopResponse.session_id).toBeDefined();
        expect(stopResponse.status).toBeDefined();
        expect(stopResponse.start_time).toBeDefined();
        expect(stopResponse.end_time).toBeDefined();
        expect(stopResponse.duration).toBeDefined();
        expect(stopResponse.file_size).toBeDefined();
      }
      
      ws.close();
    });
  });

  describe('Error Handling', () => {
    test('REQ-CAM03-006: should handle invalid device errors properly', async () => {
      const ws = await fixture.connectWebSocketWithAuth();
      
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_status', id, { device: '/dev/invalid' });
      
      // The fixture should validate the error response format against API documentation
      await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
      
      ws.close();
    });

    test('REQ-CAM03-007: should handle unauthorized access properly', async () => {
      const ws = await fixture.connectWebSocket(); // No authentication
      
      const id = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'take_snapshot', id, { device: '/dev/video0' });
      
      // The fixture should validate the error response format against API documentation
      await expect(fixture.waitForResponse(ws, id)).rejects.toThrow();
      
      ws.close();
    });
  });
});

/**
 * Legacy test function for backward compatibility
 * Now uses the stable test fixture instead of custom implementation
 */
async function testComprehensiveCameraOperations() {
  console.log('🔍 Testing Comprehensive Camera Operations with Compliant Fixture');
  console.log('==================================================================');
  
  const fixture = new StableTestFixture();
  
  try {
    await fixture.initialize();
    const ws = await fixture.connectWebSocketWithAuth();
    
    console.log('✅ Authentication successful using compliant fixture');
    
    // Test basic API functionality
    const pingId = Math.floor(Math.random() * 1000000);
    fixture.sendRequest(ws, 'ping', pingId);
    const pingResult = await fixture.waitForResponse(ws, pingId);
    console.log('✅ Ping test passed:', pingResult);
    
    // Test camera list
    const listId = Math.floor(Math.random() * 1000000);
    fixture.sendRequest(ws, 'get_camera_list', listId);
    const cameraList = await fixture.waitForResponse(ws, listId);
    console.log('✅ Camera list test passed:', cameraList.cameras.length, 'cameras found');
    
    // Test camera status if cameras available
    if (cameraList.cameras.length > 0) {
      const testDevice = cameraList.cameras[0].device;
      const statusId = Math.floor(Math.random() * 1000000);
      fixture.sendRequest(ws, 'get_camera_status', statusId, { device: testDevice });
      const statusResult = await fixture.waitForResponse(ws, statusId);
      console.log('✅ Camera status test passed for device:', testDevice);
    }
    
    console.log('\n✅ All camera operations tests passed using stable fixture');
    console.log('✅ API compliance validation working correctly');
    
    ws.close();
    return 'Comprehensive camera operations test passed using stable fixture';
  } catch (error) {
    console.error('❌ Camera operations test failed:', error.message);
    throw error;
  }
}

// Export for backward compatibility
module.exports = { testComprehensiveCameraOperations };
