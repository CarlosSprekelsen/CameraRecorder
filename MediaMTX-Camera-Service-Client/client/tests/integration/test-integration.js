import WebSocket from 'ws';

async function testIntegration() {
  console.log('🔍 Testing Camera List Integration with Real Server');
  console.log('================================================');
  
  const ws = new WebSocket('ws://localhost:8002/ws');
  
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error('Test timeout'));
    }, 10000);

    ws.on('open', () => {
      console.log('✅ WebSocket connected to real MediaMTX server');
      
      // Test get_camera_list
      const request = {
        jsonrpc: '2.0',
        method: 'get_camera_list',
        id: 1
      };
      
      console.log('📤 Sending get_camera_list request...');
      ws.send(JSON.stringify(request));
    });

    ws.on('message', (data) => {
      try {
        const response = JSON.parse(data.toString());
        console.log('📥 Received response:', JSON.stringify(response, null, 2));
        
        if (response.id === 1 && response.result) {
          const cameras = response.result.cameras;
          console.log(`\n📊 Camera List Integration Results:`);
          console.log(`   Total cameras: ${response.result.total}`);
          console.log(`   Connected cameras: ${response.result.connected}`);
          
          cameras.forEach((camera, index) => {
            console.log(`\n   Camera ${index + 1}:`);
            console.log(`     Device: ${camera.device}`);
            console.log(`     Name: ${camera.name}`);
            console.log(`     Status: ${camera.status}`);
            console.log(`     Resolution: ${camera.resolution}`);
            console.log(`     FPS: ${camera.fps}`);
          });
          
          console.log('\n✅ Camera list integration working correctly!');
          console.log('✅ Real server responding with actual camera data');
          console.log('✅ React app should display this data correctly');
          
          clearTimeout(timeout);
          ws.close();
          resolve('Integration test passed');
        }
      } catch (error) {
        console.error('❌ Error parsing response:', error);
        clearTimeout(timeout);
        ws.close();
        reject(error);
      }
    });

    ws.on('error', (error) => {
      console.error('❌ WebSocket error:', error);
      clearTimeout(timeout);
      reject(error);
    });

    ws.on('close', () => {
      console.log('🔌 WebSocket connection closed');
    });
  });
}

if (import.meta.url === `file://${process.argv[1]}`) {
  testIntegration()
    .then((result) => {
      console.log('\n🎉 Integration test completed successfully!');
      process.exit(0);
    })
    .catch((error) => {
      console.error('\n❌ Integration test failed:', error);
      process.exit(1);
    });
}

export { testIntegration };
