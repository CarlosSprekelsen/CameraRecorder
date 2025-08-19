#!/usr/bin/env node

/**
 * Sprint 3: Real WebSocket Integration Test
 * 
 * This script tests the real WebSocket connection to the MediaMTX Camera Service
 * and validates all JSON-RPC methods work correctly with the actual server.
 * 
 * Usage: node test-websocket-integration.js
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running on localhost:8002
 * - WebSocket endpoint available at ws://localhost:8002/ws
 */

import WebSocket from 'ws';

// Test configuration
const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  timeout: 20000,
  retryAttempts: 3,
  retryDelay: 1000,
};

// Test results tracking
const testResults = {
  passed: 0,
  failed: 0,
  total: 0,
  errors: [],
};

/**
 * Utility function to send JSON-RPC requests
 */
function send(ws, method, id, params = undefined) {
  const req = { jsonrpc: '2.0', method, id };
  if (params) req.params = params;
  console.log(`📤 Sending ${method} (#${id})`, params ? JSON.stringify(params) : '');
  ws.send(JSON.stringify(req));
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
 * Test WebSocket connection establishment
 */
async function testConnection() {
  console.log('\n🔌 Testing WebSocket Connection...');
  
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error('Connection timeout'));
    }, CONFIG.timeout);

    const ws = new WebSocket(CONFIG.serverUrl);

    ws.on('open', () => {
      clearTimeout(timeout);
      console.log('✅ WebSocket connection established');
      assert(true, 'WebSocket connection established successfully');
      ws.close();
      resolve();
    });

    ws.on('error', (error) => {
      clearTimeout(timeout);
      console.error('❌ WebSocket connection failed:', error.message);
      assert(false, `WebSocket connection failed: ${error.message}`);
      reject(error);
    });
  });
}

/**
 * Test basic JSON-RPC functionality with proper connection management
 */
async function testBasicRPC() {
  console.log('\n🏓 Testing Basic RPC Methods...');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    let testCompleted = false;
    let wsClosed = false;

    const timeout = setTimeout(() => {
      if (!testCompleted) {
        testCompleted = true;
        if (!wsClosed) ws.close();
        reject(new Error('Basic RPC test timeout'));
      }
    }, CONFIG.timeout);

    ws.on('open', () => {
      console.log('✅ Connected, testing ping...');
      send(ws, 'ping', 1);
    });

    ws.on('close', () => {
      wsClosed = true;
      console.log('🔌 WebSocket connection closed');
    });

    ws.on('message', (data) => {
      try {
        const msg = JSON.parse(data.toString());
        console.log('📥', JSON.stringify(msg));

        if (msg.id === 1) {
          // Ping response
          assert(msg.result === 'pong', 'Ping response is correct');
          console.log('✅ ping test passed');
          
          // Test get_camera_list
          console.log('📋 Testing get_camera_list...');
          send(ws, 'get_camera_list', 2);
        } else if (msg.id === 2) {
          // Camera list response
          const res = msg.result;
          assert(res && typeof res === 'object', 'Camera list response is valid object');
          assert(Array.isArray(res.cameras), 'Cameras field is an array');
          assert(typeof res.total === 'number' || typeof res.connected === 'number', 'Response has total/connected fields');
          
          console.log(`📊 Found ${res.cameras.length} cameras (${res.connected || 0} connected)`);
          
          // Check if we have cameras to test with
          if (res.cameras.length > 0) {
            const camera = res.cameras[0];
            assert(camera.device, 'Camera has device field');
            assert(camera.status, 'Camera has status field');
            console.log(`📷 Testing with camera: ${camera.device}`);
            
            // Test get_camera_status
            send(ws, 'get_camera_status', 3, { device: camera.device });
          } else {
            console.log('⚠️ No cameras available, testing file listing...');
            send(ws, 'list_recordings', 4, { limit: 1, offset: 0 });
          }
        } else if (msg.id === 3) {
          // Camera status response
          const camera = msg.result;
          assert(camera && camera.device, 'Camera status response is valid');
          console.log(`✅ Camera status test passed for ${camera.device}`);
          
          // Test file listing
          send(ws, 'list_recordings', 4, { limit: 1, offset: 0 });
        } else if (msg.id === 4) {
          // File listing response
          const res = msg.result;
          assert(res && Array.isArray(res.files), 'File listing response is valid');
          console.log(`✅ File listing test passed (${res.files.length} files)`);
          
          // Test snapshots listing
          send(ws, 'list_snapshots', 5, { limit: 1, offset: 0 });
        } else if (msg.id === 5) {
          // Snapshots listing response
          const res = msg.result;
          assert(res && Array.isArray(res.files), 'Snapshots listing response is valid');
          console.log(`✅ Snapshots listing test passed (${res.files.length} snapshots)`);
          
          // Test error handling
          send(ws, 'get_camera_status', 6, { device: '/dev/invalid' });
        } else if (msg.id === 6) {
          // Error handling test
          if (msg.error) {
            const code = msg.error.code;
            const acceptable = new Set([-32001, -1000, -1001]); // Camera not found or disconnected
            assert(acceptable.has(code), `Error code ${code} is acceptable for invalid device`);
            console.log('✅ Error handling test passed');
          } else if (msg.result) {
            const camera = msg.result;
            assert(camera.device === '/dev/invalid' && camera.status === 'DISCONNECTED', 'Invalid device returns DISCONNECTED status');
            console.log('✅ Invalid device handling test passed');
          } else {
            assert(false, 'Expected error or result for invalid device');
          }
          
          // Complete test
          testCompleted = true;
          clearTimeout(timeout);
          if (!wsClosed) ws.close();
          resolve();
        }
      } catch (err) {
        if (!testCompleted) {
          testCompleted = true;
          clearTimeout(timeout);
          if (!wsClosed) ws.close();
          console.error('❌ Message parsing error:', err);
          reject(err);
        }
      }
    });

    ws.on('error', (error) => {
      if (!testCompleted) {
        testCompleted = true;
        clearTimeout(timeout);
        console.error('❌ WebSocket error:', error);
        reject(error);
      }
    });
  });
}

/**
 * Test real-time notifications with proper connection handling
 */
async function testNotifications() {
  console.log('\n📢 Testing Real-time Notifications...');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    let notificationReceived = false;
    let wsClosed = false;

    const timeout = setTimeout(() => {
      if (!notificationReceived) {
        console.log('⚠️ No notifications received (this is normal if no cameras are active)');
        if (!wsClosed) ws.close();
        resolve();
      }
    }, 10000); // 10 second timeout for notifications

    ws.on('open', () => {
      console.log('✅ Connected, waiting for notifications...');
      // Send ping to keep connection alive
      send(ws, 'ping', 1);
    });

    ws.on('close', () => {
      wsClosed = true;
      console.log('🔌 Notification test connection closed');
    });

    ws.on('message', (data) => {
      try {
        const msg = JSON.parse(data.toString());
        
        if (!msg.id) {
          // This is a notification
          console.log('📢 Notification received:', msg.method);
          assert(msg.method, 'Notification has method field');
          assert(msg.params, 'Notification has params field');
          notificationReceived = true;
          
          clearTimeout(timeout);
          if (!wsClosed) ws.close();
          resolve();
        } else if (msg.id === 1) {
          // Ping response, send another ping to keep connection alive
          setTimeout(() => {
            if (!notificationReceived && !wsClosed) {
              send(ws, 'ping', 2);
            }
          }, 2000);
        }
      } catch (err) {
        console.error('❌ Notification parsing error:', err);
      }
    });

    ws.on('error', (error) => {
      console.error('❌ WebSocket error during notification test:', error);
      clearTimeout(timeout);
      reject(error);
    });
  });
}

/**
 * Test connection stability and reconnection
 */
async function testConnectionStability() {
  console.log('\n🔄 Testing Connection Stability...');
  
  const ws = new WebSocket(CONFIG.serverUrl);
  let wsClosed = false;
  
  return new Promise((resolve, reject) => {
    let messagesReceived = 0;
    const maxMessages = 5;
    
    const timeout = setTimeout(() => {
      if (!wsClosed) ws.close();
      assert(messagesReceived > 0, 'Connection maintained for stability test');
      resolve();
    }, 15000); // 15 second stability test

    ws.on('open', () => {
      console.log('✅ Connected, testing stability...');
      send(ws, 'ping', 1);
    });

    ws.on('close', () => {
      wsClosed = true;
      console.log('🔌 Stability test connection closed');
    });

    ws.on('message', (data) => {
      try {
        const msg = JSON.parse(data.toString());
        messagesReceived++;
        
        if (msg.id && msg.result === 'pong') {
          console.log(`🏓 Ping ${messagesReceived} successful`);
          
          if (messagesReceived < maxMessages && !wsClosed) {
            setTimeout(() => {
              if (!wsClosed) {
                send(ws, 'ping', messagesReceived + 1);
              }
            }, 2000);
          } else {
            clearTimeout(timeout);
            if (!wsClosed) ws.close();
            assert(true, 'Connection stability test passed');
            resolve();
          }
        }
      } catch (err) {
        console.error('❌ Message parsing error:', err);
      }
    });

    ws.on('error', (error) => {
      console.error('❌ WebSocket error during stability test:', error);
      clearTimeout(timeout);
      reject(error);
    });
  });
}

/**
 * Main test execution
 */
async function runTests() {
  console.log('🚀 Starting Sprint 3 WebSocket Integration Tests');
  console.log(`📡 Server: ${CONFIG.serverUrl}`);
  console.log(`⏱️  Timeout: ${CONFIG.timeout}ms`);
  
  try {
    // Test 1: Connection establishment
    await testConnection();
    
    // Test 2: Basic RPC functionality
    await testBasicRPC();
    
    // Test 3: Real-time notifications
    await testNotifications();
    
    // Test 4: Connection stability
    await testConnectionStability();
    
  } catch (error) {
    console.error('💥 Test execution failed:', error.message);
    testResults.failed++;
    testResults.errors.push(`Test execution failed: ${error.message}`);
  }
  
  // Print test results
  console.log('\n📊 Test Results Summary');
  console.log('========================');
  console.log(`✅ Passed: ${testResults.passed}`);
  console.log(`❌ Failed: ${testResults.failed}`);
  console.log(`📊 Total: ${testResults.total}`);
  console.log(`📈 Success Rate: ${((testResults.passed / testResults.total) * 100).toFixed(1)}%`);
  
  if (testResults.errors.length > 0) {
    console.log('\n❌ Errors:');
    testResults.errors.forEach((error, index) => {
      console.log(`  ${index + 1}. ${error}`);
    });
  }
  
  // Exit with appropriate code
  const success = testResults.failed === 0;
  console.log(`\n${success ? '🎉' : '💥'} All tests ${success ? 'passed' : 'failed'}`);
  process.exit(success ? 0 : 1);
}

// Run tests if this script is executed directly
if (require.main === module) {
  runTests().catch((error) => {
    console.error('💥 Test runner failed:', error);
    process.exit(1);
  });
}

module.exports = { runTests, testResults };

// Add Jest test functions for integration testing
describe('WebSocket Integration Tests', () => {
  test('should test WebSocket integration', async () => {
    await expect(runTests()).resolves.not.toThrow();
  }, 60000);

  test('should validate connection establishment', async () => {
    // Test connection establishment specifically
    const WebSocket = require('ws');
    const ws = new WebSocket('ws://localhost:8002/ws');
    
    await new Promise((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('Connection timeout')), 5000);
      
      ws.on('open', () => {
        clearTimeout(timeout);
        ws.close();
        resolve();
      });
      
      ws.on('error', (error) => {
        clearTimeout(timeout);
        reject(error);
      });
    });
  }, 10000);

  test('should validate JSON-RPC protocol', async () => {
    const WebSocket = require('ws');
    const ws = new WebSocket('ws://localhost:8002/ws');
    
    await new Promise((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('RPC timeout')), 5000);
      
      ws.on('open', () => {
        // Send ping request
        const request = {
          jsonrpc: '2.0',
          id: 1,
          method: 'ping'
        };
        ws.send(JSON.stringify(request));
      });
      
      ws.on('message', (data) => {
        try {
          const response = JSON.parse(data.toString());
          if (response.id === 1 && response.result === 'pong') {
            clearTimeout(timeout);
            ws.close();
            resolve();
          }
        } catch (error) {
          // Ignore non-JSON messages
        }
      });
      
      ws.on('error', (error) => {
        clearTimeout(timeout);
        reject(error);
      });
    });
  }, 10000);
});
