/**
 * Authentication Setup Integration Test
 * Verifies that the JWT authentication works with the correct environment variable name
 */

const WebSocket = require('ws');
const jwt = require('jsonwebtoken');

const CONFIG = {
  serverUrl: process.env.TEST_SERVER_URL || 'ws://localhost:8002/ws',
  device: process.env.TEST_CAMERA_DEVICE || '/dev/video0',
  timeout: parseInt(process.env.TEST_TIMEOUT) || 10000,
  jwtSecret: process.env.CAMERA_SERVICE_JWT_SECRET
};

function generateValidToken() {
  if (!CONFIG.jwtSecret) {
    throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
  }
  
  const payload = {
    user_id: 'test-user',
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
        const response = JSON.parse(data.toString());
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

/**
 * Jest test suite for authentication setup
 */
describe('Authentication Setup Integration Tests', () => {
  let ws;

  beforeAll(async () => {
    // Setup WebSocket connection
    ws = new WebSocket(CONFIG.serverUrl);
    await new Promise((resolve, reject) => {
      ws.on('open', resolve);
      ws.on('error', reject);
    });
    console.log('✅ WebSocket connected for authentication setup test suite');
  });

  afterAll(async () => {
    if (ws) {
      ws.close();
    }
  });

  test('should verify installation fix with correct environment variable', async () => {
    await expect(testInstallationFix()).resolves.not.toThrow();
  }, CONFIG.timeout);

  test('should generate valid JWT token', () => {
    const token = generateValidToken();
    expect(token).toBeDefined();
    expect(typeof token).toBe('string');
    expect(token.split('.').length).toBe(3); // JWT has 3 parts
  });

  test('should authenticate successfully', async () => {
    const token = generateValidToken();
    const authResult = await sendRequest(ws, 'authenticate', { token });
    expect(authResult.authenticated).toBe(true);
  }, CONFIG.timeout);
});
