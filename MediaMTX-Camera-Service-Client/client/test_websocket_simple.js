/**
 * Simple WebSocket Connection Test
 * 
 * This test validates basic WebSocket connectivity to the MediaMTX Camera Service
 * Used for PDR-1 validation to identify connection issues
 */

const WebSocket = require('ws');

const TEST_WEBSOCKET_URL = process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws';

async function testWebSocketConnection() {
  console.log('🔌 Testing WebSocket connection to:', TEST_WEBSOCKET_URL);
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(TEST_WEBSOCKET_URL);
    
    const timeout = setTimeout(() => {
      console.log('⏰ Connection timeout after 10 seconds');
      ws.close();
      reject(new Error('Connection timeout'));
    }, 10000);

    ws.on('open', () => {
      console.log('✅ WebSocket connected successfully');
      clearTimeout(timeout);
      ws.close();
      resolve(true);
    });

    ws.on('error', (error) => {
      console.log('❌ WebSocket connection error:', error.message);
      clearTimeout(timeout);
      reject(error);
    });

    ws.on('close', (code, reason) => {
      console.log('🔌 WebSocket closed:', code, reason?.toString());
    });
  });
}

async function testJSONRPCMethods() {
  console.log('📡 Testing JSON-RPC methods...');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(TEST_WEBSOCKET_URL);
    
    const timeout = setTimeout(() => {
      console.log('⏰ JSON-RPC test timeout');
      ws.close();
      reject(new Error('JSON-RPC test timeout'));
    }, 15000);

    ws.on('open', async () => {
      console.log('✅ WebSocket connected for JSON-RPC test');
      
      try {
        // Test ping method
        const pingRequest = {
          jsonrpc: '2.0',
          id: 1,
          method: 'ping',
          params: {}
        };
        
        console.log('📤 Sending ping request...');
        ws.send(JSON.stringify(pingRequest));
        
        // Wait for response
        setTimeout(() => {
          console.log('⏰ No response received for ping');
          clearTimeout(timeout);
          ws.close();
          reject(new Error('No response to ping'));
        }, 5000);
        
      } catch (error) {
        console.log('❌ Error in JSON-RPC test:', error.message);
        clearTimeout(timeout);
        ws.close();
        reject(error);
      }
    });

    ws.on('message', (data) => {
      console.log('📨 Received message:', data.toString());
      clearTimeout(timeout);
      ws.close();
      resolve(true);
    });

    ws.on('error', (error) => {
      console.log('❌ WebSocket error in JSON-RPC test:', error.message);
      clearTimeout(timeout);
      reject(error);
    });
  });
}

async function runTests() {
  console.log('🚀 Starting PDR-1 WebSocket validation tests...\n');
  
  try {
    // Test 1: Basic connection
    console.log('=== Test 1: Basic WebSocket Connection ===');
    await testWebSocketConnection();
    console.log('✅ Basic connection test passed\n');
    
    // Test 2: JSON-RPC methods
    console.log('=== Test 2: JSON-RPC Method Testing ===');
    await testJSONRPCMethods();
    console.log('✅ JSON-RPC test passed\n');
    
    console.log('🎉 All PDR-1 WebSocket validation tests passed!');
    
  } catch (error) {
    console.log('❌ PDR-1 WebSocket validation failed:', error.message);
    process.exit(1);
  }
}

// Run tests if this file is executed directly
if (require.main === module) {
  runTests();
}

module.exports = { testWebSocketConnection, testJSONRPCMethods };
