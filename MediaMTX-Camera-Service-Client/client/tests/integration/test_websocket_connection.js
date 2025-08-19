const WebSocket = require('ws');

console.log('Testing WebSocket connection to camera service...');

const ws = new WebSocket('ws://localhost:8002');

ws.on('open', function open() {
    console.log('✅ WebSocket connection established successfully');
    
    // Test camera list request
    const cameraListRequest = {
        type: 'get_cameras'
    };
    
    ws.send(JSON.stringify(cameraListRequest));
    console.log('📡 Sent camera list request');
});

ws.on('message', function message(data) {
    console.log('📨 Received message:', data.toString());
    
    try {
        const response = JSON.parse(data.toString());
        console.log('✅ Message parsed successfully:', response.type);
        
        if (response.type === 'cameras_list') {
            console.log('✅ Camera list received successfully');
            console.log('📊 Cameras found:', response.cameras ? response.cameras.length : 0);
        }
        
        // Close connection after successful test
        setTimeout(() => {
            ws.close();
            console.log('✅ WebSocket test completed successfully');
            process.exit(0);
        }, 1000);
        
    } catch (error) {
        console.error('❌ Error parsing message:', error);
        ws.close();
        process.exit(1);
    }
});

ws.on('error', function error(err) {
    console.error('❌ WebSocket error:', err.message);
    process.exit(1);
});

ws.on('close', function close() {
    console.log('🔌 WebSocket connection closed');
});

// Timeout after 10 seconds
setTimeout(() => {
    console.error('❌ WebSocket test timed out');
    process.exit(1);
}, 10000);
