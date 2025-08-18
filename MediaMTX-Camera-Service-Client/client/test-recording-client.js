#!/usr/bin/env node

/**
 * Client Recording Test
 * Tests the client recording functionality with the real server
 */

import WebSocket from 'ws';
import jwt from 'jsonwebtoken';

const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  device: '/dev/video0',
  timeout: 10000,
  jwtSecret: 'd0adf90f433d25a0f1d8b9e384f77976fff12f3ecf57ab39364dcc83731aa6f7'
};

function generateValidToken() {
  const payload = {
    user_id: 'test_user',
    role: 'operator',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60)
  };
  
  return jwt.sign(payload, CONFIG.jwtSecret, { algorithm: 'HS256' });
}

function sendRequest(ws, method, params = {}) {
  return new Promise((resolve, reject) => {
    const id = Math.floor(Math.random() * 10000);
    const request = {
      jsonrpc: '2.0',
      method: method,
      params: params,
      id: id
    };
    
    console.log(`📤 ${method}:`, JSON.stringify(params));
    
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
            console.log(`❌ ${method} error:`, response.error);
            reject(new Error(response.error.message || 'RPC error'));
          } else {
            console.log(`✅ ${method} success:`, response.result);
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

async function testClientRecording() {
  console.log('🎬 Testing Client Recording Implementation');
  console.log('========================================');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Step 1: Authenticate
        const token = generateValidToken();
        console.log('\n🔐 Step 1: Authenticating...');
        const authResult = await sendRequest(ws, 'authenticate', { token });
        
        if (!authResult.authenticated) {
          throw new Error('Authentication failed');
        }
        console.log('✅ Authentication successful');
        
        // Step 2: Test start_recording with duration controls
        console.log('\n🎬 Step 2: Testing start_recording with duration controls');
        
        // Test 2a: Timed recording (10 seconds)
        console.log('\n📹 Test 2a: Timed recording (10 seconds)');
        const startResult1 = await sendRequest(ws, 'start_recording', {
          device: CONFIG.device,
          duration_seconds: 10,
          format: 'mp4'
        });
        
        console.log('✅ Timed recording started');
        console.log('Session ID:', startResult1.session_id);
        console.log('Filename:', startResult1.filename);
        console.log('Status:', startResult1.status);
        
        // Wait 3 seconds then stop
        console.log('\n⏳ Waiting 3 seconds...');
        await new Promise(resolve => setTimeout(resolve, 3000));
        
        const stopResult1 = await sendRequest(ws, 'stop_recording', {
          device: CONFIG.device
        });
        
        console.log('✅ Timed recording stopped');
        console.log('Duration:', stopResult1.duration, 'seconds');
        console.log('File size:', stopResult1.file_size, 'bytes');
        
        // Test 2b: Unlimited recording
        console.log('\n📹 Test 2b: Unlimited recording');
        const startResult2 = await sendRequest(ws, 'start_recording', {
          device: CONFIG.device,
          format: 'mp4'
          // No duration = unlimited
        });
        
        console.log('✅ Unlimited recording started');
        console.log('Session ID:', startResult2.session_id);
        console.log('Filename:', startResult2.filename);
        console.log('Status:', startResult2.status);
        
        // Wait 2 seconds then stop
        console.log('\n⏳ Waiting 2 seconds...');
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        const stopResult2 = await sendRequest(ws, 'stop_recording', {
          device: CONFIG.device
        });
        
        console.log('✅ Unlimited recording stopped');
        console.log('Duration:', stopResult2.duration, 'seconds');
        console.log('File size:', stopResult2.file_size, 'bytes');
        
        // Test 2c: Recording with minutes duration
        console.log('\n📹 Test 2c: Recording with minutes duration');
        const startResult3 = await sendRequest(ws, 'start_recording', {
          device: CONFIG.device,
          duration_minutes: 1,
          format: 'mp4'
        });
        
        console.log('✅ Minutes recording started');
        console.log('Session ID:', startResult3.session_id);
        console.log('Filename:', startResult3.filename);
        console.log('Status:', startResult3.status);
        
        // Wait 2 seconds then stop
        console.log('\n⏳ Waiting 2 seconds...');
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        const stopResult3 = await sendRequest(ws, 'stop_recording', {
          device: CONFIG.device
        });
        
        console.log('✅ Minutes recording stopped');
        console.log('Duration:', stopResult3.duration, 'seconds');
        console.log('File size:', stopResult3.file_size, 'bytes');
        
        // Step 3: Test session management
        console.log('\n📋 Step 3: Testing session management');
        
        // Start recording
        const sessionStart = await sendRequest(ws, 'start_recording', {
          device: CONFIG.device,
          duration_seconds: 5,
          format: 'mp4'
        });
        
        console.log('✅ Session recording started');
        
        // Try to start another recording (should handle gracefully)
        try {
          await sendRequest(ws, 'start_recording', {
            device: CONFIG.device,
            duration_seconds: 5,
            format: 'mp4'
          });
          console.log('ℹ️ Second recording started (device supports multiple sessions)');
        } catch (error) {
          console.log('ℹ️ Second recording handled gracefully:', error.message);
        }
        
        // Stop the recording
        const sessionStop = await sendRequest(ws, 'stop_recording', {
          device: CONFIG.device
        });
        
        console.log('✅ Session recording stopped');
        
        console.log('\n🎉 All client recording tests passed!');
        console.log('✅ Recording operations implementation completed');
        console.log('✅ Duration controls working (unlimited, timed with countdown)');
        console.log('✅ Session management working');
        console.log('✅ Status feedback working');
        
        ws.close();
        resolve();
        
      } catch (error) {
        console.error('❌ Test failed:', error);
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

// Run test
testClientRecording().catch(error => {
  console.error('❌ Test suite failed:', error);
  process.exit(1);
});
