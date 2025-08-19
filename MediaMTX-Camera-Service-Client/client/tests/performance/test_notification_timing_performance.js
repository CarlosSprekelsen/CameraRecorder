/**
 * Notification Timing Performance Test
 * 
 * Measures the complete loop from sending recording command to receiving notification feedback.
 * This helps understand the user feedback delay and identify intermittent notification issues.
 */

const WebSocket = require('ws');
const { performance } = require('perf_hooks');
const crypto = require('crypto');

/**
 * Generate JWT token using crypto (since we don't have jwt library)
 */
function generateJWTToken(payload, secret) {
  const header = {
    alg: 'HS256',
    typ: 'JWT'
  };
  
  const encodedHeader = Buffer.from(JSON.stringify(header)).toString('base64url');
  const encodedPayload = Buffer.from(JSON.stringify(payload)).toString('base64url');
  
  const signature = crypto
    .createHmac('sha256', secret)
    .update(`${encodedHeader}.${encodedPayload}`)
    .digest('base64url');
  
  return `${encodedHeader}.${encodedPayload}.${signature}`;
}

const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  timeout: 30000,
  testIterations: 5,
};

let ws = null;
let requestId = 0;
let notificationReceived = false;

let notificationData = null;

function send(method, params = undefined) {
  const req = { jsonrpc: '2.0', method, id: ++requestId };
  if (params) req.params = params;
  console.log(`📤 Sending ${method} (#${requestId})`, params ? JSON.stringify(params) : '');
  ws.send(JSON.stringify(req));
}

async function waitForNotification(timeout = 10000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`Notification timeout after ${timeout}ms`));
    }, timeout);

    const checkNotification = () => {
      if (notificationReceived) {
        clearTimeout(timer);
        resolve(notificationData);
      } else {
        setTimeout(checkNotification, 100);
      }
    };
    checkNotification();
  });
}

async function runNotificationTimingTest(iteration) {
  console.log(`\n🔄 Test Iteration ${iteration + 1}/${CONFIG.testIterations}`);
  
  // Reset notification state
  notificationReceived = false;
  notificationData = null;
  
  const startTime = performance.now();
  
  // Step 0: Stop any existing recording first
  console.log(`   🛑 Stopping any existing recording...`);
  send('stop_recording', { device: '/dev/video0' });
  
  // Wait for stop response
  await new Promise(resolve => setTimeout(resolve, 2000));
  
  // Step 1: Send start recording command
  const commandStartTime = performance.now();
  console.log(`   ▶️ Starting new recording...`);
  send('start_recording', { device: '/dev/video0' });
  
  // Step 2: Wait for notification
  try {
    const notification = await waitForNotification(15000);
    const notificationTime = performance.now();
    
    const commandToNotificationDelay = notificationTime - commandStartTime;
    const totalTime = notificationTime - startTime;
    
    console.log(`✅ Notification received: ${notification.method}`);
    console.log(`   Command → Notification delay: ${commandToNotificationDelay.toFixed(2)}ms`);
    console.log(`   Total test time: ${totalTime.toFixed(2)}ms`);
    console.log(`   Notification data:`, JSON.stringify(notification.params, null, 2));
    
    // Step 3: Wait a bit then send stop recording command
    await new Promise(resolve => setTimeout(resolve, 3000)); // Record for 3 seconds
    
    const stopStartTime = performance.now();
    console.log(`   ⏹️ Stopping recording...`);
    send('stop_recording', { device: '/dev/video0' });
    
    // Step 4: Wait for stop notification
    notificationReceived = false;
    notificationData = null;
    
    const stopNotification = await waitForNotification(15000);
    const stopNotificationTime = performance.now();
    
    const stopCommandToNotificationDelay = stopNotificationTime - stopStartTime;
    
    console.log(`✅ Stop notification received: ${stopNotification.method}`);
    console.log(`   Stop command → Notification delay: ${stopCommandToNotificationDelay.toFixed(2)}ms`);
    console.log(`   Stop notification data:`, JSON.stringify(stopNotification.params, null, 2));
    
    return {
      iteration: iteration + 1,
      startCommandToNotification: commandToNotificationDelay,
      stopCommandToNotification: stopCommandToNotificationDelay,
      totalTime: totalTime,
      success: true
    };
    
  } catch (error) {
    console.log(`❌ Notification timeout: ${error.message}`);
    return {
      iteration: iteration + 1,
      error: error.message,
      success: false
    };
  }
}

async function main() {
  console.log('🔍 Notification Timing Analysis');
  console.log('================================');
  console.log(`Server: ${CONFIG.serverUrl}`);
  console.log(`Iterations: ${CONFIG.testIterations}`);
  console.log(`Timeout: ${CONFIG.timeout}ms`);
  
  return new Promise((resolve, reject) => {
    ws = new WebSocket(CONFIG.serverUrl);
    
    ws.on('open', async () => {
      console.log('✅ Connected to WebSocket server');
      
      try {
        // Authenticate first - use environment variable only
        const jwtSecret = process.env.CAMERA_SERVICE_JWT_SECRET;
        if (!jwtSecret) {
          throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
        }
        
        // Generate a proper JWT token for testing
        const payload = {
          user_id: 'test_user',
          role: 'operator',
          iat: Math.floor(Date.now() / 1000),
          exp: Math.floor(Date.now() / 1000) + 3600
        };
        
        // Generate token using crypto (since we don't have jwt library)
        const token = generateJWTToken(payload, jwtSecret);
        
        send('authenticate', { token });
        
        const results = [];
        
        for (let i = 0; i < CONFIG.testIterations; i++) {
          const result = await runNotificationTimingTest(i);
          results.push(result);
          
          // Wait between tests
          if (i < CONFIG.testIterations - 1) {
            await new Promise(resolve => setTimeout(resolve, 2000));
          }
        }
        
        // Analyze results
        console.log('\n📊 Notification Timing Analysis Results');
        console.log('=====================================');
        
        const successfulTests = results.filter(r => r.success);
        const failedTests = results.filter(r => !r.success);
        
        console.log(`Total tests: ${results.length}`);
        console.log(`Successful: ${successfulTests.length}`);
        console.log(`Failed: ${failedTests.length}`);
        console.log(`Success rate: ${((successfulTests.length / results.length) * 100).toFixed(1)}%`);
        
        if (successfulTests.length > 0) {
          const startDelays = successfulTests.map(r => r.startCommandToNotification);
          const stopDelays = successfulTests.map(r => r.stopCommandToNotification);
          
          console.log('\n⏱️ Start Recording Notification Delays:');
          console.log(`   Min: ${Math.min(...startDelays).toFixed(2)}ms`);
          console.log(`   Max: ${Math.max(...startDelays).toFixed(2)}ms`);
          console.log(`   Avg: ${(startDelays.reduce((a, b) => a + b, 0) / startDelays.length).toFixed(2)}ms`);
          
          console.log('\n⏱️ Stop Recording Notification Delays:');
          console.log(`   Min: ${Math.min(...stopDelays).toFixed(2)}ms`);
          console.log(`   Max: ${Math.max(...stopDelays).toFixed(2)}ms`);
          console.log(`   Avg: ${(stopDelays.reduce((a, b) => a + b, 0) / stopDelays.length).toFixed(2)}ms`);
        }
        
        if (failedTests.length > 0) {
          console.log('\n❌ Failed Tests:');
          failedTests.forEach(test => {
            console.log(`   Iteration ${test.iteration}: ${test.error}`);
          });
        }
        
        ws.close();
        resolve(results);
        
      } catch (error) {
        console.error('❌ Test execution failed:', error);
        ws.close();
        reject(error);
      }
    });
    
    ws.on('message', (data) => {
      try {
        const response = JSON.parse(data.toString());
        
        // Handle JSON-RPC response
        if (response.id !== undefined) {
          console.log(`📥 Response #${response.id}:`, JSON.stringify(response));
        }
        // Handle JSON-RPC notification
        else if (response.method !== undefined && response.id === undefined) {
          console.log(`📡 Notification: ${response.method}`);
          notificationReceived = true;
          notificationData = response;
        }
        
      } catch (error) {
        console.error('❌ Message parsing error:', error);
      }
    });
    
    ws.on('error', (error) => {
      console.error('❌ WebSocket error:', error.message);
      reject(error);
    });
    
    ws.on('close', () => {
      console.log('🔌 WebSocket connection closed');
    });
  });
}

/**
 * Jest test suite for notification timing performance
 */
describe('Notification Timing Performance Tests', () => {
  let ws;

  beforeAll(async () => {
    // Setup WebSocket connection
    ws = new WebSocket(CONFIG.serverUrl);
    await new Promise((resolve, reject) => {
      ws.on('open', resolve);
      ws.on('error', reject);
    });
    console.log('✅ WebSocket connected for performance test suite');
  });

  afterAll(async () => {
    if (ws) {
      ws.close();
    }
  });

  test('should complete notification timing analysis', async () => {
    const results = await main();
    expect(Array.isArray(results)).toBe(true);
    expect(results.length).toBeGreaterThan(0);
  }, 60000);

  test('should meet performance targets from guidelines', async () => {
    const results = await main();
    const successfulTests = results.filter(r => r.success);
    
    if (successfulTests.length > 0) {
      const startDelays = successfulTests.map(r => r.startCommandToNotification);
      const avgStartDelay = startDelays.reduce((a, b) => a + b, 0) / startDelays.length;
      
      // Performance target: <100ms (p95 under load)
      expect(avgStartDelay).toBeLessThan(100);
    }
  }, 60000);
});
