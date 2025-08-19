import WebSocket from 'ws';

console.log('Testing WebSocket connection to camera service...');

const ws = new WebSocket('ws://localhost:8002');

ws.on('open', function open() {
    console.log('✅ WebSocket connection established successfully');
    
    // Test camera list request using JSON-RPC format
    const cameraListRequest = {
        jsonrpc: "2.0",
        method: "get_cameras",
        id: 1
    };
    
    ws.send(JSON.stringify(cameraListRequest));
    console.log('📡 Sent camera list request');
});

ws.on('message', function message(data) {
    console.log('📨 Received message:', data.toString());
    
    try {
        const response = JSON.parse(data.toString());
        console.log('✅ Message parsed successfully');
        
        if (response.result) {
            console.log('✅ Camera list received successfully');
            console.log('📊 Cameras found:', response.result.cameras ? response.result.cameras.length : 0);
        } else if (response.error) {
            console.log('⚠️ Server returned error:', response.error.message);
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
