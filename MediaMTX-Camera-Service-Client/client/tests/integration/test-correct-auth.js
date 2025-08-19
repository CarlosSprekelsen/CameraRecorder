import WebSocket from 'ws';
import { generateValidToken, validateTestEnvironment } from './auth-utils.js';

console.log('Testing camera operations with correct authentication...');

// Validate test environment first
if (!validateTestEnvironment()) {
    process.exit(1);
}

// Generate valid token using current server secret
const validToken = generateValidToken('test_user', 'operator');
console.log('🔐 Generated valid JWT token using current server secret');

const ws = new WebSocket('ws://localhost:8002');

let testResults = {
    connection: false,
    authentication: false,
    cameraList: false,
    snapshot: false,
    recording: false,
    fileDownload: false
};

ws.on('open', function open() {
    console.log('✅ WebSocket connection established successfully');
    testResults.connection = true;
    
    // Test 1: Get camera list (no auth required)
    const cameraListRequest = {
        jsonrpc: "2.0",
        method: "get_cameras",
        id: 1
    };
    
    ws.send(JSON.stringify(cameraListRequest));
    console.log('📡 Test 1: Sent camera list request');
});

ws.on('message', function message(data) {
    console.log('📨 Received message:', data.toString());
    
    try {
        const response = JSON.parse(data.toString());
        
        if (response.result && response.id === 1) {
            // Camera list response
            console.log('✅ Test 1: Camera list received successfully');
            console.log('📊 Cameras found:', response.result.cameras.length);
            testResults.cameraList = true;
            
            if (response.result.cameras.length > 0) {
                const camera = response.result.cameras[0];
                console.log('📷 Testing with camera:', camera.name);
                
                // Test 2: Take snapshot with auth token in params
                const snapshotRequest = {
                    jsonrpc: "2.0",
                    method: "take_snapshot",
                    params: {
                        device: camera.device,
                        filename: "test_snapshot.jpg",
                        auth_token: validToken
                    },
                    id: 2
                };
                
                ws.send(JSON.stringify(snapshotRequest));
                console.log('📡 Test 2: Sent snapshot request with auth token');
            }
        } else if (response.result && response.id === 2) {
            // Snapshot response
            console.log('✅ Test 2: Snapshot taken successfully');
            console.log('📸 Snapshot file:', response.result.filename);
            testResults.snapshot = true;
            testResults.authentication = true;
            
            // Test 3: Start recording with auth token
            const recordingRequest = {
                jsonrpc: "2.0",
                method: "start_recording",
                params: {
                    device: "/dev/video0",
                    duration: 5,
                    auth_token: validToken
                },
                id: 3
            };
            
            ws.send(JSON.stringify(recordingRequest));
            console.log('📡 Test 3: Sent recording start request');
            
        } else if (response.result && response.id === 3) {
            // Recording start response
            console.log('✅ Test 3: Recording started successfully');
            console.log('🎥 Recording file:', response.result.filename);
            testResults.recording = true;
            
            // Test 4: Get file list for download
            const fileListRequest = {
                jsonrpc: "2.0",
                method: "list_snapshots",
                params: {
                    auth_token: validToken
                },
                id: 4
            };
            
            ws.send(JSON.stringify(fileListRequest));
            console.log('📡 Test 4: Sent snapshots list request');
            
        } else if (response.result && response.id === 4) {
            // File list response
            console.log('✅ Test 4: Snapshots list received successfully');
            console.log('📁 Snapshots available:', response.result.snapshots ? response.result.snapshots.length : 0);
            testResults.fileDownload = true;
            
            // Complete all tests
            setTimeout(() => {
                ws.close();
                console.log('\n🎯 TEST RESULTS SUMMARY:');
                console.log('Connection:', testResults.connection ? '✅ PASS' : '❌ FAIL');
                console.log('Authentication:', testResults.authentication ? '✅ PASS' : '❌ FAIL');
                console.log('Camera List:', testResults.cameraList ? '✅ PASS' : '❌ FAIL');
                console.log('Snapshot:', testResults.snapshot ? '✅ PASS' : '❌ FAIL');
                console.log('Recording:', testResults.recording ? '✅ PASS' : '❌ FAIL');
                console.log('File Download:', testResults.fileDownload ? '✅ PASS' : '❌ FAIL');
                
                const allPassed = Object.values(testResults).every(result => result);
                console.log('\n🎉 OVERALL RESULT:', allPassed ? '✅ ALL TESTS PASSED' : '❌ SOME TESTS FAILED');
                
                process.exit(allPassed ? 0 : 1);
            }, 1000);
        } else if (response.error) {
            console.log('⚠️ Server returned error:', response.error.message);
            if (response.id === 2) {
                console.log('❌ Snapshot failed - authentication issue');
                // Try without auth to see if it works
                const camera = { device: "/dev/video0" };
                const snapshotRequest = {
                    jsonrpc: "2.0",
                    method: "take_snapshot",
                    params: {
                        device: camera.device,
                        filename: "test_snapshot_noauth.jpg"
                    },
                    id: 5
                };
                
                ws.send(JSON.stringify(snapshotRequest));
                console.log('📡 Test 5: Sent snapshot request without auth');
            }
        }
        
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

// Timeout after 30 seconds
setTimeout(() => {
    console.error('❌ Camera operations test timed out');
    process.exit(1);
}, 30000);
