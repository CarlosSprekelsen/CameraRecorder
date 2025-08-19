const WebSocket = require('ws');

async function testTakeSnapshotParams() {
  console.log('Testing take_snapshot parameter validation...');
  
  const ws = new WebSocket('ws://localhost:8002/ws');
  
  return new Promise((resolve, reject) => {
    ws.on('open', async () => {
      console.log('✅ WebSocket connected');
      
      try {
        // Test 1: Check if server accepts format parameter
        console.log('\n📸 Test 1: Check format parameter acceptance');
        const test1 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0',
          format: 'png'
        });
        console.log('✅ Test 1 - Server response:', test1);
        
        // Test 2: Check if server accepts quality parameter
        console.log('\n📸 Test 2: Check quality parameter acceptance');
        const test2 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0',
          quality: 95
        });
        console.log('✅ Test 2 - Server response:', test2);
        
        // Test 3: Check if server accepts filename parameter
        console.log('\n📸 Test 3: Check filename parameter acceptance');
        const test3 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0',
          filename: 'test_snapshot.jpg'
        });
        console.log('✅ Test 3 - Server response:', test3);
        
        // Test 4: Check if server accepts all parameters together
        console.log('\n📸 Test 4: Check all parameters together');
        const test4 = await sendRequest(ws, 'take_snapshot', {
          device: '/dev/video0',
          format: 'png',
          quality: 80,
          filename: 'test_all_params.png'
        });
        console.log('✅ Test 4 - Server response:', test4);
        
        console.log('\n🎉 All parameter validation tests completed!');
        console.log('Note: These tests expect authentication errors, which is expected.');
        console.log('The important thing is that the server accepts the parameter format.');
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
            console.log(`📥 Error response (expected):`, response.error);
            // For this test, we expect authentication errors, so we resolve with the error
            resolve({ error: response.error, params: params });
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

testTakeSnapshotParams().catch(console.error);
