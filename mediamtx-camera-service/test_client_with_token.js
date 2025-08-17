#!/usr/bin/env node
/**
 * Test script to generate valid JWT token and test JavaScript client
 */

const jwt = require('jsonwebtoken');
const { CameraClient } = require('./examples/javascript/camera_client.js');

// JWT Configuration (matches server config)
const JWT_SECRET = "dev-secret-change-me";
const USER_ID = "test_user";
const ROLE = "admin";

function generateJWTToken() {
    const payload = {
        user_id: USER_ID,
        role: ROLE,
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + (24 * 3600) // 24 hours
    };
    
    const token = jwt.sign(payload, JWT_SECRET, { algorithm: 'HS256' });
    console.log(`Generated JWT token: ${token.substring(0, 50)}...`);
    return token;
}

async function testClient() {
    console.log("🧪 Testing JavaScript Client with Valid JWT Token");
    console.log("=" .repeat(60));
    
    const token = generateJWTToken();
    
    // Create client with valid token
    const client = new CameraClient({
        host: 'localhost',
        port: 8002,
        authType: 'jwt',
        authToken: token,
        useSsl: false,
        maxRetries: 3,
        retryDelay: 1.0
    });
    
    try {
        // Connect to service
        await client.connect();
        console.log('✅ Connected to camera service');
        
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
            // Get status of first camera
            const camera = cameras[0];
            const status = await client.getCameraStatus(camera.devicePath);
            console.log(`✅ Camera status: ${status.status}`);
            console.log(`✅ Camera resolution: ${status.resolution}`);
            console.log(`✅ Camera FPS: ${status.fps}`);
        }
        
    } catch (error) {
        console.error(`❌ Error: ${error.message}`);
    } finally {
        await client.disconnect();
        console.log('✅ Disconnected');
    }
}

// Run the test
testClient().catch(console.error);
