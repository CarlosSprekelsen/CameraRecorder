#!/usr/bin/env node

/**
 * Comprehensive Camera Operations Test
 * 
 * This test validates all camera operations against the real MediaMTX Camera Service server
 * following the actual server API specification.
 * 
 * Server API Methods Tested:
 * - take_snapshot(device, filename?) - Only device and filename parameters
 * - start_recording(device, duration?, format?) - Only device, duration, format parameters
 * - stop_recording(device) - Only device parameter
 * 
 * Usage: node test-camera-operations-comprehensive.js
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running on localhost:8002
 * - WebSocket endpoint available at ws://localhost:8002/ws
 * - Camera device available at /dev/video0
 * - Valid JWT authentication
 */

import WebSocket from 'ws';
import jwt from 'jsonwebtoken';

// Test configuration
const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  timeout: 30000,
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
  apiMethods: {
    takeSnapshot: false,
    startRecording: false,
    stopRecording: false,
    getCameraStatus: false,
    getCameraList: false,
    listRecordings: false,
    listSnapshots: false
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
 * Test 1: Basic API Methods (No Auth Required)
 */
async function testBasicAPI(ws) {
  console.log('\n🔍 Test 1: Basic API Methods (No Auth Required)');
  
  try {
    // Test ping
    console.log('\n🏓 Testing ping...');
    const pingResult = await sendRequest(ws, 'ping');
    assert(pingResult === 'pong', 'ping returns pong');
    
    // Test get_camera_list
    console.log('\n📋 Testing get_camera_list...');
    const cameraList = await sendRequest(ws, 'get_camera_list');
    assert(cameraList && Array.isArray(cameraList.cameras), 'get_camera_list returns valid camera list');
    assert(typeof cameraList.total === 'number', 'get_camera_list has total count');
    assert(typeof cameraList.connected === 'number', 'get_camera_list has connected count');
    testResults.apiMethods.getCameraList = true;
    
    // Test get_camera_status
    console.log('\n📊 Testing get_camera_status...');
    const cameraStatus = await sendRequest(ws, 'get_camera_status', { device: CONFIG.device });
    assert(cameraStatus && cameraStatus.device === CONFIG.device, 'get_camera_status returns valid status');
    assert(cameraStatus.status, 'camera status has status field');
    assert(cameraStatus.name, 'camera status has name field');
    testResults.apiMethods.getCameraStatus = true;
    
    // Test list_recordings (no auth required)
    console.log('\n🎬 Testing list_recordings...');
    const recordings = await sendRequest(ws, 'list_recordings', { limit: 5, offset: 0 });
    assert(recordings && Array.isArray(recordings.files), 'list_recordings returns valid file list');
    testResults.apiMethods.listRecordings = true;
    
    // Test list_snapshots (no auth required)
    console.log('\n📸 Testing list_snapshots...');
    const snapshots = await sendRequest(ws, 'list_snapshots', { limit: 5, offset: 0 });
    assert(snapshots && Array.isArray(snapshots.files), 'list_snapshots returns valid file list');
    testResults.apiMethods.listSnapshots = true;
    
  } catch (error) {
    console.error('❌ Basic API test failed:', error.message);
    throw error;
  }
}

/**
 * Test 2: Authentication
 */
async function testAuthentication(ws) {
  console.log('\n🔐 Test 2: Authentication');
  
  try {
    const token = generateValidToken();
    console.log('\n🔑 Authenticating with JWT token...');
    const authResult = await sendRequest(ws, 'authenticate', { token });
    
    assert(authResult.authenticated === true, 'authentication successful');
    assert(authResult.role === 'operator', 'user has operator role');
    console.log('✅ Authentication successful');
    
  } catch (error) {
    console.error('❌ Authentication test failed:', error.message);
    throw error;
  }
}

/**
 * Test 3: Take Snapshot (Auth Required)
 */
async function testTakeSnapshot(ws) {
  console.log('\n📸 Test 3: Take Snapshot (Auth Required)');
  
  try {
    // Test 3a: Basic snapshot with default filename
    console.log('\n📸 Test 3a: Basic snapshot (default filename)');
    const snapshot1 = await sendRequest(ws, 'take_snapshot', {
      device: CONFIG.device
    });
    assert(snapshot1 && snapshot1.device === CONFIG.device, 'snapshot has correct device');
    assert(snapshot1.filename, 'snapshot has filename');
    assert(snapshot1.status === 'completed', 'snapshot status is completed');
    assert(snapshot1.timestamp, 'snapshot has timestamp');
    assert(typeof snapshot1.file_size === 'number', 'snapshot has file size');
    assert(snapshot1.file_path, 'snapshot has file path');
    
    // Test 3b: Snapshot with custom filename
    console.log('\n📸 Test 3b: Snapshot with custom filename');
    const customFilename = `test_snapshot_${Date.now()}.jpg`;
    const snapshot2 = await sendRequest(ws, 'take_snapshot', {
      device: CONFIG.device,
      filename: customFilename
    });
    assert(snapshot2.filename === customFilename, 'snapshot uses custom filename');
    assert(snapshot2.status === 'completed', 'custom filename snapshot completed');
    
    testResults.apiMethods.takeSnapshot = true;
    console.log('✅ Take snapshot tests completed');
    
  } catch (error) {
    console.error('❌ Take snapshot test failed:', error.message);
    throw error;
  }
}

/**
 * Test 4: Recording Operations (Auth Required)
 */
async function testRecordingOperations(ws) {
  console.log('\n🎬 Test 4: Recording Operations (Auth Required)');
  
  try {
    // Test 4a: Start recording with duration
    console.log('\n🎬 Test 4a: Start recording (10 seconds)');
    const startResult = await sendRequest(ws, 'start_recording', {
      device: CONFIG.device,
      duration: 10,
      format: 'mp4'
    });
    assert(startResult && startResult.device === CONFIG.device, 'start recording has correct device');
    assert(startResult.session_id, 'start recording has session ID');
    assert(startResult.filename, 'start recording has filename');
    assert(startResult.status === 'STARTED', 'recording status is STARTED');
    assert(startResult.start_time, 'recording has start time');
    assert(startResult.duration === 10, 'recording has correct duration');
    assert(startResult.format === 'mp4', 'recording has correct format');
    
    testResults.apiMethods.startRecording = true;
    
    // Wait a moment for recording to start
    console.log('⏳ Waiting 2 seconds for recording to establish...');
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    // Test 4b: Stop recording
    console.log('\n⏹️ Test 4b: Stop recording');
    const stopResult = await sendRequest(ws, 'stop_recording', {
      device: CONFIG.device
    });
    assert(stopResult && stopResult.device === CONFIG.device, 'stop recording has correct device');
    assert(stopResult.session_id, 'stop recording has session ID');
    assert(stopResult.filename, 'stop recording has filename');
    assert(stopResult.status === 'STOPPED', 'recording status is STOPPED');
    assert(stopResult.start_time, 'stop recording has start time');
    assert(stopResult.end_time, 'stop recording has end time');
    assert(typeof stopResult.duration === 'number', 'stop recording has duration');
    assert(typeof stopResult.file_size === 'number', 'stop recording has file size');
    
    testResults.apiMethods.stopRecording = true;
    console.log('✅ Recording operations tests completed');
    
  } catch (error) {
    console.error('❌ Recording operations test failed:', error.message);
    throw error;
  }
}

/**
 * Test 5: Error Handling
 */
async function testErrorHandling(ws) {
  console.log('\n⚠️ Test 5: Error Handling');
  
  try {
    // Test 5a: Invalid device
    console.log('\n⚠️ Test 5a: Invalid device');
    try {
      await sendRequest(ws, 'get_camera_status', { device: '/dev/invalid' });
      assert(false, 'should have thrown error for invalid device');
    } catch (error) {
      assert(error.message.includes('Camera not found') || error.message.includes('DISCONNECTED'), 'invalid device handled correctly');
    }
    
    // Test 5b: Unauthenticated protected method
    console.log('\n⚠️ Test 5b: Unauthenticated protected method');
    const ws2 = new WebSocket(CONFIG.serverUrl);
    await new Promise((resolve) => ws2.on('open', resolve));
    
    try {
      await sendRequest(ws2, 'take_snapshot', { device: CONFIG.device });
      assert(false, 'should have thrown authentication error');
    } catch (error) {
      assert(error.message.includes('Authentication required'), 'unauthenticated access blocked correctly');
    }
    
    ws2.close();
    console.log('✅ Error handling tests completed');
    
  } catch (error) {
    console.error('❌ Error handling test failed:', error.message);
    throw error;
  }
}

/**
 * Main test execution
 */
async function runTests() {
  console.log('🎯 Comprehensive Camera Operations Test');
  console.log('=====================================');
  console.log(`Server: ${CONFIG.serverUrl}`);
  console.log(`Device: ${CONFIG.device}`);
  console.log(`Timeout: ${CONFIG.timeout}ms`);
  console.log('');
  
  const ws = new WebSocket(CONFIG.serverUrl);
  
  return new Promise((resolve, reject) => {
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Run all test suites
        await testBasicAPI(ws);
        await testAuthentication(ws);
        await testTakeSnapshot(ws);
        await testRecordingOperations(ws);
        await testErrorHandling(ws);
        
        // Summary
        console.log('\n📊 Test Summary');
        console.log('==============');
        console.log(`Total Tests: ${testResults.total}`);
        console.log(`Passed: ${testResults.passed}`);
        console.log(`Failed: ${testResults.failed}`);
        console.log(`Success Rate: ${Math.round((testResults.passed / testResults.total) * 100)}%`);
        
        console.log('\n🔧 API Method Coverage');
        console.log('====================');
        Object.entries(testResults.apiMethods).forEach(([method, tested]) => {
          console.log(`${tested ? '✅' : '❌'} ${method}`);
        });
        
        if (testResults.failed === 0) {
          console.log('\n🎉 All tests passed! Camera operations are working correctly.');
        } else {
          console.log('\n❌ Some tests failed. Please check the errors above.');
        }
        
        ws.close();
        resolve();
        
      } catch (error) {
        console.error('❌ Test suite failed:', error.message);
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
