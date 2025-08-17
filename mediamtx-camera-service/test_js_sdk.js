#!/usr/bin/env node
/**
 * Test script to validate JavaScript SDK functionality.
 */

const jwt = require('jsonwebtoken');
const { CameraClient } = require('./sdk/javascript/dist/index.js');

async function testJsSdk() {
    console.log("🧪 Testing JavaScript SDK Functionality");
    console.log("=" .repeat(50));
    
    // Generate JWT token
    const JWT_SECRET = "dev-secret-change-me";
    const USER_ID = "test_user";
    const ROLE = "admin";
    
    const payload = {
        user_id: USER_ID,
        role: ROLE,
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + (24 * 3600)
    };
    
    const token = jwt.sign(payload, JWT_SECRET, { algorithm: 'HS256' });
    console.log(`Generated JWT token: ${token.substring(0, 50)}...`);
    
    // Create client
    const client = new CameraClient({
        host: 'localhost',
        port: 8002,
        authType: 'jwt',
        authToken: token
    });
    
    try {
        // Connect
        await client.connect();
        console.log("✅ Connected to camera service");
        
        // Test ping
        const pong = await client.ping();
        console.log(`✅ Ping response: ${pong}`);
        
        // Get camera list
        const cameras = await client.getCameraList();
        console.log(`✅ Found ${cameras.length} cameras:`);
        for (const camera of cameras) {
            console.log(`  - ${camera.name} (${camera.devicePath}) - ${camera.status}`);
        }
        
        if (cameras.length > 0) {
            // Test get camera status (this should work with SDK)
            const camera = cameras[0];
            const status = await client.getCameraStatus(camera.devicePath);
            console.log(`✅ Camera status: ${status.status}`);
            
            // Test snapshot
            const snapshot = await client.takeSnapshot(camera.devicePath);
            console.log(`✅ Snapshot taken: ${snapshot.filename}`);
        }
        
    } catch (error) {
        console.error(`❌ JavaScript SDK test error: ${error.message}`);
    } finally {
        await client.disconnect();
        console.log("✅ Disconnected");
    }
}

testJsSdk().catch(console.error);
