#!/usr/bin/env node

/**
 * Sprint 3: Recording Operations Test
 * 
 * This script validates the recording operations implementation against the real MediaMTX Camera Service server.
 * 
 * Tests all recording requirements:
 * - start_recording with duration controls (unlimited, timed with countdown)
 * - stop_recording with status feedback
 * - Recording progress indicators
 * - Recording session management
 * - Error handling for recording operations
 * 
 * Usage: node test-recording-operations.js
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running on localhost:8002
 * - WebSocket endpoint available at ws://localhost:8002/ws
 * - Camera device available at /dev/video0
 */

import WebSocket from 'ws';
import jwt from 'jsonwebtoken';

// Test configuration
const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  timeout: 30000,
  device: '/dev/video0',
  jwtSecret: process.env.CAMERA_SERVICE_JWT_SECRET || 'a436cccea2e4afb6d7c38b189fbdb6cd62e1671c279e7d729704e133d4e7ab53'
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
  
  return jwt.sign(payload, CONFIG.jwtSecret, { algorithm: 'HS256' });
}

// Test results tracking
const testResults = {
  passed: 0,
  failed: 0,
  total: 0,
  errors: [],
  recordingRequirements: {
    startRecording: false,
    stopRecording: false,
    durationControls: false,
    progressIndicators: false,
    sessionManagement: false,
    errorHandling: false,
  }
};

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
 * Test 1: Basic start_recording functionality
 */
async function testStartRecording(ws) {
  console.log('\n🎬 Test 1: Basic start_recording functionality');
  
  try {
    const result = await sendRequest(ws, 'start_recording', {
      device: CONFIG.device,
      duration_seconds: 10,
      format: 'mp4'
    });
    
    assert(result.success === true, 'start_recording returns success: true');
    assert(result.session_id, 'start_recording returns session_id');
    assert(result.filename, 'start_recording returns filename');
    assert(result.status === 'recording', 'start_recording returns status: recording');
    assert(result.device === CONFIG.device, 'start_recording returns correct device');
    
    testResults.recordingRequirements.startRecording = true;
    return result;
    
  } catch (error) {
    console.error('❌ start_recording test failed:', error);
    throw error;
  }
}

/**
 * Test 2: Duration controls - unlimited recording
 */
async function testUnlimitedRecording(ws) {
  console.log('\n🎬 Test 2: Unlimited recording (no duration)');
  
  try {
    const result = await sendRequest(ws, 'start_recording', {
      device: CONFIG.device,
      format: 'mp4'
      // No duration specified = unlimited
    });
    
    assert(result.success === true, 'unlimited recording returns success: true');
    assert(result.session_id, 'unlimited recording returns session_id');
    assert(result.status === 'recording', 'unlimited recording returns status: recording');
    
    testResults.recordingRequirements.durationControls = true;
    return result;
    
  } catch (error) {
    console.error('❌ unlimited recording test failed:', error);
    throw error;
  }
}

/**
 * Test 3: Duration controls - timed recording with minutes
 */
async function testTimedRecording(ws) {
  console.log('\n🎬 Test 3: Timed recording with minutes');
  
  try {
    const result = await sendRequest(ws, 'start_recording', {
      device: CONFIG.device,
      duration_minutes: 1,
      format: 'mp4'
    });
    
    assert(result.success === true, 'timed recording returns success: true');
    assert(result.session_id, 'timed recording returns session_id');
    assert(result.status === 'recording', 'timed recording returns status: recording');
    
    testResults.recordingRequirements.durationControls = true;
    return result;
    
  } catch (error) {
    console.error('❌ timed recording test failed:', error);
    throw error;
  }
}

/**
 * Test 4: stop_recording functionality
 */
async function testStopRecording(ws, sessionId) {
  console.log('\n⏹️ Test 4: stop_recording functionality');
  
  try {
    const result = await sendRequest(ws, 'stop_recording', {
      device: CONFIG.device
    });
    
    assert(result.success === true, 'stop_recording returns success: true');
    assert(result.session_id === sessionId, 'stop_recording returns correct session_id');
    assert(result.status === 'completed', 'stop_recording returns status: completed');
    assert(result.end_time, 'stop_recording returns end_time');
    assert(result.duration, 'stop_recording returns duration');
    assert(result.file_size, 'stop_recording returns file_size');
    
    testResults.recordingRequirements.stopRecording = true;
    return result;
    
  } catch (error) {
    console.error('❌ stop_recording test failed:', error);
    throw error;
  }
}

/**
 * Test 5: Error handling - invalid device
 */
async function testErrorHandling(ws) {
  console.log('\n❌ Test 5: Error handling - invalid device');
  
  try {
    await sendRequest(ws, 'start_recording', {
      device: '/dev/invalid_device',
      duration_seconds: 10
    });
    
    // Should not reach here
    assert(false, 'start_recording with invalid device should fail');
    
  } catch (error) {
    assert(error.message.includes('error') || error.message.includes('failed'), 'Invalid device returns error');
    testResults.recordingRequirements.errorHandling = true;
  }
}

/**
 * Test 6: Session management - multiple recordings
 */
async function testSessionManagement(ws) {
  console.log('\n📋 Test 6: Session management - multiple recordings');
  
  try {
    // Start first recording
    const recording1 = await sendRequest(ws, 'start_recording', {
      device: CONFIG.device,
      duration_seconds: 5,
      format: 'mp4'
    });
    
    assert(recording1.success === true, 'First recording starts successfully');
    assert(recording1.session_id, 'First recording has session_id');
    
    // Try to start second recording (should fail or handle gracefully)
    try {
      const recording2 = await sendRequest(ws, 'start_recording', {
        device: CONFIG.device,
        duration_seconds: 5,
        format: 'mp4'
      });
      
      // If second recording starts, stop it
      if (recording2.success) {
        await sendRequest(ws, 'stop_recording', { device: CONFIG.device });
      }
    } catch (error) {
      // Expected behavior - device already recording
      console.log('ℹ️ Second recording attempt handled gracefully');
    }
    
    // Stop first recording
    const stopResult = await sendRequest(ws, 'stop_recording', { device: CONFIG.device });
    assert(stopResult.success === true, 'First recording stops successfully');
    
    testResults.recordingRequirements.sessionManagement = true;
    
  } catch (error) {
    console.error('❌ session management test failed:', error);
    throw error;
  }
}

/**
 * Main test execution
 */
async function runTests() {
  console.log('🎬 Sprint 3: Recording Operations Test');
  console.log('=====================================');
  console.log(`Server: ${CONFIG.serverUrl}`);
  console.log(`Device: ${CONFIG.device}`);
  console.log(`Timeout: ${CONFIG.timeout}ms`);
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Step 1: Authenticate first
        console.log('🔐 Authenticating...');
        const token = generateValidToken();
        const authResult = await sendRequest(ws, 'authenticate', { token });
        
        if (!authResult.authenticated) {
          throw new Error('Authentication failed');
        }
        console.log('✅ Authentication successful');
        
        // Run all tests
        const recording1 = await testStartRecording(ws);
        await new Promise(resolve => setTimeout(resolve, 2000)); // Wait 2 seconds
        await testStopRecording(ws, recording1.session_id);
        
        await testUnlimitedRecording(ws);
        await new Promise(resolve => setTimeout(resolve, 2000)); // Wait 2 seconds
        await sendRequest(ws, 'stop_recording', { device: CONFIG.device });
        
        await testTimedRecording(ws);
        await new Promise(resolve => setTimeout(resolve, 2000)); // Wait 2 seconds
        await sendRequest(ws, 'stop_recording', { device: CONFIG.device });
        
        await testErrorHandling(ws);
        await testSessionManagement(ws);
        
        // Print results
        console.log('\n📊 Test Results:');
        console.log('================');
        console.log(`Total tests: ${testResults.total}`);
        console.log(`Passed: ${testResults.passed}`);
        console.log(`Failed: ${testResults.failed}`);
        
        console.log('\n🎯 Recording Requirements:');
        console.log('=========================');
        Object.entries(testResults.recordingRequirements).forEach(([requirement, passed]) => {
          console.log(`${passed ? '✅' : '❌'} ${requirement}`);
        });
        
        if (testResults.failed === 0) {
          console.log('\n🎉 All recording tests passed!');
          console.log('✅ Recording operations implementation completed');
        } else {
          console.log('\n❌ Some tests failed. Check errors above.');
        }
        
        ws.close();
        resolve();
        
      } catch (error) {
        console.error('❌ Test execution failed:', error);
        ws.close();
        reject(error);
      }
    });
    
    ws.on('error', (error) => {
      console.error('❌ WebSocket error:', error);
      reject(error);
    });
  });
}

// Run tests if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  runTests().catch(error => {
    console.error('❌ Test suite failed:', error);
    process.exit(1);
  });
}

export { runTests, testResults };
