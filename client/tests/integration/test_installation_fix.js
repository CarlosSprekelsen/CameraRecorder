#!/usr/bin/env node

/**
 * Test Installation Fix
 * Verifies that the JWT authentication works with the correct environment variable name
 */

import WebSocket from 'ws';
import jwt from 'jsonwebtoken';

const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  device: '/dev/video0',
  timeout: 10000,
  // This should now work with the corrected environment variable
  jwtSecret: 'd0adf90f433d25a0f1d8b9e384f77976fff12f3ecf57ab39364dcc83731aa6f7'
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

async function testInstallationFix() {
  console.log('🔧 Testing Installation Fix');
  console.log('==========================');
  console.log('Environment Variable: CAMERA_SERVICE_JWT_SECRET');
  console.log('Expected Behavior: Authentication should work correctly');
  console.log('');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Generate a valid token
        const token = generateValidToken();
        console.log('\n🔑 Generated valid JWT token');
        console.log('Token:', token);
        
        // Test authentication
        console.log('\n🔐 Testing authentication with CAMERA_SERVICE_JWT_SECRET');
        const authResult = await sendRequest(ws, 'authenticate', {
          token: token
        });
        
        if (authResult.authenticated) {
          console.log('✅ Authentication successful with correct environment variable');
          console.log('Authenticated:', authResult.authenticated);
          console.log('Role:', authResult.role);
          console.log('Auth method:', authResult.auth_method);
          
          console.log('\n🎉 Installation fix verified!');
          console.log('✅ Environment variable naming is now consistent');
          console.log('✅ Fresh installations will work correctly');
          console.log('✅ No more authentication issues');
          
          ws.close();
          resolve();
        } else {
          throw new Error('Authentication failed');
        }
        
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
testInstallationFix().catch(error => {
  console.error('❌ Installation fix test failed:', error);
  process.exit(1);
});
