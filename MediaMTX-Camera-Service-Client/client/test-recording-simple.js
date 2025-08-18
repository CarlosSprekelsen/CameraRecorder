#!/usr/bin/env node

/**
 * Simple Recording Test
 * Tests the basic recording functionality against the real server
 */

import WebSocket from 'ws';

const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  device: '/dev/video0',
  timeout: 10000,
  authToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdF91c2VyIiwicm9sZSI6Im9wZXJhdG9yIiwiaWF0IjoxNzU1NTM3MzQ4LCJleHAiOjE3NTU2MjM3NDh9.28vtGjOIBFbkNw7mhRKuflTHtGQ-oub_UzVXi2I4in0'
};

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

async function testRecording() {
  console.log('🎬 Testing Recording Operations');
  console.log('==============================');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Authenticate first
        console.log('\n🔐 Authenticating...');
        const authResult = await sendRequest(ws, 'authenticate', {
          token: CONFIG.authToken
        });
        console.log('✅ Authentication successful');
        
        // Test 1: Start recording
        console.log('\n🎬 Test 1: Start recording (10 seconds)');
        const startResult = await sendRequest(ws, 'start_recording', {
          device: CONFIG.device,
          duration_seconds: 10,
          format: 'mp4'
        });
        
        console.log('✅ Recording started successfully');
        console.log('Session ID:', startResult.session_id);
        console.log('Filename:', startResult.filename);
        console.log('Status:', startResult.status);
        
        // Wait 3 seconds
        console.log('\n⏳ Waiting 3 seconds...');
        await new Promise(resolve => setTimeout(resolve, 3000));
        
        // Test 2: Stop recording
        console.log('\n⏹️ Test 2: Stop recording');
        const stopResult = await sendRequest(ws, 'stop_recording', {
          device: CONFIG.device
        });
        
        console.log('✅ Recording stopped successfully');
        console.log('Session ID:', stopResult.session_id);
        console.log('Filename:', stopResult.filename);
        console.log('Status:', stopResult.status);
        console.log('Duration:', stopResult.duration, 'seconds');
        console.log('File size:', stopResult.file_size, 'bytes');
        
        console.log('\n🎉 All recording tests passed!');
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
testRecording().catch(error => {
  console.error('❌ Test suite failed:', error);
  process.exit(1);
});
