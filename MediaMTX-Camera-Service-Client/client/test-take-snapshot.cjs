const WebSocket = require('ws');

async function testTakeSnapshot() {
  console.log('Testing take_snapshot with format/quality options...');
  
  const ws = new WebSocket('ws://localhost:8002/ws');
  
  return new Promise((resolve, reject) => {
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Authenticate first
        console.log('\n🔐 Authenticating...');
        const authResult = await sendRequest(ws, 'authenticate', {
          token: 'test.token.123'
        });
        console.log('✅ Authentication result:', authResult);
        
        // Test 1: Basic snapshot with default format/quality
        console.log('\n📸 Test 1: Basic snapshot (default jpg, quality 85)');
        const test1 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0'
        });
        console.log('✅ Test 1 result:', test1);
        
        // Test 2: Snapshot with custom format (png)
        console.log('\n📸 Test 2: Snapshot with PNG format');
        const test2 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0',
          format: 'png'
        });
        console.log('✅ Test 2 result:', test2);
        
        // Test 3: Snapshot with custom quality
        console.log('\n📸 Test 3: Snapshot with custom quality (95)');
        const test3 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0',
          quality: 95
        });
        console.log('✅ Test 3 result:', test3);
        
        // Test 4: Snapshot with custom filename
        console.log('\n📸 Test 4: Snapshot with custom filename');
        const test4 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0',
          filename: 'test_snapshot_custom.jpg'
        });
        console.log('✅ Test 4 result:', test4);
        
        // Test 5: Snapshot with all options
        console.log('\n📸 Test 5: Snapshot with all options (png, quality 80, custom filename)');
        const test5 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0',
          format: 'png',
          quality: 80,
          filename: 'test_snapshot_all_options.png'
        });
        console.log('✅ Test 5 result:', test5);
        
        console.log('\n🎉 All take_snapshot tests completed successfully!');
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
    }, 10000);
    
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
      }
    };
    
    ws.on('message', messageHandler);
    ws.send(JSON.stringify(request));
  });
}

testTakeSnapshot().catch(console.error);
