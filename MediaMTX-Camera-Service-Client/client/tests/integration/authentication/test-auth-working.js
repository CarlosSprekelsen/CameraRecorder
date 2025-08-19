#!/usr/bin/env node

/**
 * Working Authentication Test
 * Tests authentication with the correct JWT secret from the server
 */

import WebSocket from 'ws';
import jwt from 'jsonwebtoken';

const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  device: '/dev/video0',
  timeout: 10000,
  // This is the actual JWT secret from the server environment
  jwtSecret: process.env.CAMERA_SERVICE_JWT_SECRET || 'a436cccea2e4afb6d7c38b189fbdb6cd62e1671c279e7d729704e133d4e7ab53'
};

function generateValidToken() {
  const payload = {
    user_id: 'test_user',
    role: 'operator',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24 hours
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

async function testAuthentication() {
  console.log('🔐 Testing Authentication');
  console.log('========================');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Generate a valid token
        const token = generateValidToken();
        console.log('\n🔑 Generated valid JWT token');
        console.log('Token:', token);
        
        // Test 1: Authenticate with valid token
        console.log('\n🔐 Test 1: Authenticate with valid token');
        const authResult = await sendRequest(ws, 'authenticate', {
          token: token
        });
        
        console.log('✅ Authentication successful');
        console.log('Authenticated:', authResult.authenticated);
        console.log('Role:', authResult.role);
        console.log('Auth method:', authResult.auth_method);
        
        // Test 2: Try protected method (take_snapshot)
        console.log('\n📸 Test 2: Try protected method (take_snapshot)');
        const snapshotResult = await sendRequest(ws, 'take_snapshot', {
          device: CONFIG.device
        });
        
        console.log('✅ Snapshot successful');
        console.log('Filename:', snapshotResult.filename);
        console.log('Status:', snapshotResult.status);
        
        // Test 3: Try recording operations
        console.log('\n🎬 Test 3: Start recording');
        const startResult = await sendRequest(ws, 'start_recording', {
          device: CONFIG.device,
          duration_seconds: 5,
          format: 'mp4'
        });
        
        console.log('✅ Recording started');
        console.log('Session ID:', startResult.session_id);
        console.log('Filename:', startResult.filename);
        console.log('Status:', startResult.status);
        
        // Wait 2 seconds
        console.log('\n⏳ Waiting 2 seconds...');
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        // Test 4: Stop recording
        console.log('\n⏹️ Test 4: Stop recording');
        const stopResult = await sendRequest(ws, 'stop_recording', {
          device: CONFIG.device
        });
        
        console.log('✅ Recording stopped');
        console.log('Session ID:', stopResult.session_id);
        console.log('Filename:', stopResult.filename);
        console.log('Status:', stopResult.status);
        console.log('Duration:', stopResult.duration, 'seconds');
        console.log('File size:', stopResult.file_size, 'bytes');
        
        console.log('\n🎉 All authentication and recording tests passed!');
        console.log('✅ Recording operations implementation completed');
        
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
testAuthentication().catch(error => {
  console.error('❌ Test suite failed:', error);
  process.exit(1);
});
