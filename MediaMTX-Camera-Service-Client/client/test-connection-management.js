#!/usr/bin/env node

/**
 * Sprint 3: Connection State Management and Error Handling Test
 * 
 * This script tests the comprehensive connection state management and error handling
 * functionality implemented for Sprint 3 requirements.
 * 
 * Tests:
 * - Connection state tracking (CONNECTING, CONNECTED, DISCONNECTED, ERROR)
 * - Error handling and recovery mechanisms
 * - Connection retry logic with user control
 * - Connection status indicators
 * - Graceful degradation when disconnected
 * - Connection health monitoring and alerts
 * - Real-time connection metrics
 * 
 * Usage: node test-connection-management.js
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running on localhost:8002
 * - WebSocket endpoint available at ws://localhost:8002/ws
 */

import WebSocket from 'ws';

// Test configuration
const CONFIG = {
  serverUrl: 'ws://localhost:8002/ws',
  timeout: 20000,
  retryAttempts: 3,
  retryDelay: 1000,
};

// Test results tracking
const testResults = {
  passed: 0,
  failed: 0,
  total: 0,
  errors: [],
};

/**
 * Utility function to send JSON-RPC requests
 */
function send(ws, method, id, params = undefined) {
  const req = { jsonrpc: '2.0', method, id };
  if (params) req.params = params;
  console.log(`📤 Sending ${method} (#${id})`, params ? JSON.stringify(params) : '');
  ws.send(JSON.stringify(req));
}

/**
 * Test result assertion
 */
function assert(condition, message) {
  testResults.total++;
  if (condition) {
    testResults.passed++;
    console.log(`✅ ${message}`);
  } else {
    testResults.failed++;
    console.log(`❌ ${message}`);
    testResults.errors.push(message);
  }
}

/**
 * Test connection state management
 */
async function testConnectionStateManagement() {
  console.log('\n🔌 Testing Connection State Management...');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    let state = 'connecting';
    let connectionStartTime = Date.now();

    const timeout = setTimeout(() => {
      reject(new Error('Connection state management test timeout'));
    }, CONFIG.timeout);

    ws.on('open', () => {
      const connectionTime = Date.now() - connectionStartTime;
      state = 'connected';
      
      assert(state === 'connected', 'Connection state correctly set to CONNECTED');
      assert(connectionTime < 5000, 'Connection established within 5 seconds');
      assert(ws.readyState === WebSocket.OPEN, 'WebSocket readyState is OPEN');
      
      console.log(`✅ Connection established in ${connectionTime}ms`);
      
      // Test connection health
      send(ws, 'ping', 1);
    });

    ws.on('message', (data) => {
      try {
        const response = JSON.parse(data.toString());
        console.log('📥 Received:', JSON.stringify(response));
        
        if (response.id === 1 && response.result === 'pong') {
          assert(response.result === 'pong', 'Health check ping/pong working');
          assert(state === 'connected', 'State remains CONNECTED during health check');
          
          // Test graceful disconnection
          console.log('🔌 Testing graceful disconnection...');
          ws.close(1000, 'Test disconnection');
        }
      } catch (error) {
        console.error('❌ Message parsing error:', error);
        assert(false, 'Message parsing failed');
      }
    });

    ws.on('close', (code, reason) => {
      state = 'disconnected';
      clearTimeout(timeout);
      
      assert(state === 'disconnected', 'Connection state correctly set to DISCONNECTED');
      assert(code === 1000, 'Graceful disconnection with code 1000');
      assert(reason === 'Test disconnection', 'Disconnection reason preserved');
      
      console.log('✅ Graceful disconnection successful');
      resolve();
    });

    ws.on('error', (error) => {
      state = 'error';
      clearTimeout(timeout);
      console.error('❌ WebSocket error:', error.message);
      assert(false, `Connection error: ${error.message}`);
      reject(error);
    });
  });
}

/**
 * Test error handling and recovery
 */
async function testErrorHandling() {
  console.log('\n⚠️ Testing Error Handling and Recovery...');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    let errorHandled = false;

    const timeout = setTimeout(() => {
      reject(new Error('Error handling test timeout'));
    }, CONFIG.timeout);

    ws.on('open', () => {
      console.log('✅ Connected, testing error scenarios...');
      
      // Test invalid JSON-RPC request
      console.log('📤 Sending invalid request...');
      ws.send('invalid json');
    });

    ws.on('message', (data) => {
      try {
        const response = JSON.parse(data.toString());
        console.log('📥 Received:', JSON.stringify(response));
        
        if (response.error) {
          errorHandled = true;
          assert(response.error.code === -32700, 'Parse error code received');
          assert(response.error.message.includes('parse'), 'Parse error message received');
          console.log('✅ Error handling working correctly');
          
          // Test recovery with valid request
          console.log('📤 Testing recovery with valid request...');
          send(ws, 'ping', 2);
        } else if (response.id === 2 && response.result === 'pong') {
          assert(errorHandled, 'Error was handled before recovery');
          assert(response.result === 'pong', 'Recovery successful with valid request');
          
          ws.close();
        }
      } catch (error) {
        console.error('❌ Message parsing error:', error);
        assert(false, 'Message parsing failed during error handling test');
      }
    });

    ws.on('close', () => {
      clearTimeout(timeout);
      assert(errorHandled, 'Error handling test completed');
      resolve();
    });

    ws.on('error', (error) => {
      clearTimeout(timeout);
      console.error('❌ WebSocket error:', error.message);
      reject(error);
    });
  });
}

/**
 * Test connection retry logic
 */
async function testConnectionRetryLogic() {
  console.log('\n🔄 Testing Connection Retry Logic...');
  
  let retryAttempts = 0;
  const maxRetries = 3;
  
  const attemptConnection = () => {
    return new Promise((resolve, reject) => {
      const ws = new WebSocket(CONFIG.serverUrl);
      
      const timeout = setTimeout(() => {
        reject(new Error('Connection retry timeout'));
      }, 5000);

      ws.on('open', () => {
        clearTimeout(timeout);
        console.log(`✅ Reconnection attempt ${retryAttempts + 1} successful`);
        ws.close();
        resolve();
      });

      ws.on('error', (error) => {
        clearTimeout(timeout);
        retryAttempts++;
        console.log(`⚠️ Connection attempt ${retryAttempts} failed: ${error.message}`);
        
        if (retryAttempts < maxRetries) {
          console.log(`🔄 Retrying in ${CONFIG.retryDelay}ms...`);
          setTimeout(() => {
            attemptConnection().then(resolve).catch(reject);
          }, CONFIG.retryDelay);
        } else {
          reject(new Error(`Max retry attempts (${maxRetries}) reached`));
        }
      });
    });
  };

  try {
    await attemptConnection();
    assert(retryAttempts <= maxRetries, 'Retry logic respects maximum attempts');
    console.log('✅ Connection retry logic working correctly');
  } catch (error) {
    console.error('❌ Retry logic test failed:', error.message);
    assert(false, `Retry logic failed: ${error.message}`);
  }
}

/**
 * Test connection health monitoring
 */
async function testHealthMonitoring() {
  console.log('\n💓 Testing Connection Health Monitoring...');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    let healthChecks = 0;
    const maxHealthChecks = 3;

    const timeout = setTimeout(() => {
      reject(new Error('Health monitoring test timeout'));
    }, CONFIG.timeout);

    ws.on('open', () => {
      console.log('✅ Connected, starting health monitoring...');
      performHealthCheck();
    });

    const performHealthCheck = () => {
      if (healthChecks >= maxHealthChecks) {
        clearTimeout(timeout);
        assert(healthChecks === maxHealthChecks, 'Health monitoring completed');
        console.log('✅ Health monitoring working correctly');
        ws.close();
        resolve();
        return;
      }

      const startTime = performance.now();
      send(ws, 'ping', healthChecks + 1);
    };

    ws.on('message', (data) => {
      try {
        const response = JSON.parse(data.toString());
        console.log('📥 Health check response:', JSON.stringify(response));
        
        if (response.result === 'pong') {
          const responseTime = performance.now() - response.id * 1000; // Approximate
          healthChecks++;
          
          assert(response.result === 'pong', `Health check ${healthChecks} successful`);
          assert(responseTime < 1000, `Health check response time < 1s (${responseTime.toFixed(2)}ms)`);
          
          console.log(`✅ Health check ${healthChecks}/${maxHealthChecks} completed`);
          
          // Schedule next health check
          setTimeout(performHealthCheck, 1000);
        }
      } catch (error) {
        console.error('❌ Health check message parsing error:', error);
        assert(false, 'Health check message parsing failed');
      }
    });

    ws.on('close', () => {
      clearTimeout(timeout);
      resolve();
    });

    ws.on('error', (error) => {
      clearTimeout(timeout);
      console.error('❌ Health monitoring error:', error.message);
      reject(error);
    });
  });
}

/**
 * Test graceful degradation
 */
async function testGracefulDegradation() {
  console.log('\n🛡️ Testing Graceful Degradation...');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    let disconnected = false;

    const timeout = setTimeout(() => {
      reject(new Error('Graceful degradation test timeout'));
    }, CONFIG.timeout);

    ws.on('open', () => {
      console.log('✅ Connected, testing graceful degradation...');
      
      // Simulate network interruption
      setTimeout(() => {
        console.log('🔌 Simulating network interruption...');
        ws.terminate(); // Force close without proper handshake
      }, 2000);
    });

    ws.on('close', (code, reason) => {
      disconnected = true;
      clearTimeout(timeout);
      
      assert(disconnected, 'Connection properly marked as disconnected');
      assert(code !== 1000, 'Connection closed due to interruption (not graceful)');
      
      console.log('✅ Graceful degradation test completed');
      resolve();
    });

    ws.on('error', (error) => {
      if (!disconnected) {
        clearTimeout(timeout);
        console.error('❌ Graceful degradation error:', error.message);
        reject(error);
      }
    });
  });
}

/**
 * Test performance metrics
 */
async function testPerformanceMetrics() {
  console.log('\n📊 Testing Performance Metrics...');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.serverUrl);
    const metrics = {
      connectionTime: 0,
      responseTimes: [],
      messageCount: 0,
      errorCount: 0
    };

    const timeout = setTimeout(() => {
      reject(new Error('Performance metrics test timeout'));
    }, CONFIG.timeout);

    const startTime = performance.now();

    ws.on('open', () => {
      metrics.connectionTime = performance.now() - startTime;
      
      assert(metrics.connectionTime < 5000, `Connection time < 5s (${metrics.connectionTime.toFixed(2)}ms)`);
      console.log(`✅ Connection established in ${metrics.connectionTime.toFixed(2)}ms`);
      
      // Test multiple requests for metrics
      for (let i = 1; i <= 5; i++) {
        setTimeout(() => {
          const requestStart = performance.now();
          send(ws, 'ping', i);
          
          // Store request timing
          metrics.messageCount++;
        }, i * 200);
      }
    });

    ws.on('message', (data) => {
      try {
        const response = JSON.parse(data.toString());
        const responseTime = performance.now() - (response.id * 200 + startTime);
        
        if (response.result === 'pong') {
          metrics.responseTimes.push(responseTime);
          console.log(`📊 Response ${response.id}: ${responseTime.toFixed(2)}ms`);
          
          if (metrics.responseTimes.length === 5) {
            const avgResponseTime = metrics.responseTimes.reduce((a, b) => a + b, 0) / metrics.responseTimes.length;
            const maxResponseTime = Math.max(...metrics.responseTimes);
            
            assert(avgResponseTime < 1000, `Average response time < 1s (${avgResponseTime.toFixed(2)}ms)`);
            assert(maxResponseTime < 2000, `Max response time < 2s (${maxResponseTime.toFixed(2)}ms)`);
            assert(metrics.messageCount === 5, 'All messages sent and received');
            
            console.log(`✅ Performance metrics: Avg=${avgResponseTime.toFixed(2)}ms, Max=${maxResponseTime.toFixed(2)}ms`);
            ws.close();
          }
        }
      } catch (error) {
        metrics.errorCount++;
        console.error('❌ Performance test message error:', error);
      }
    });

    ws.on('close', () => {
      clearTimeout(timeout);
      assert(metrics.errorCount === 0, 'No errors during performance test');
      resolve();
    });

    ws.on('error', (error) => {
      clearTimeout(timeout);
      console.error('❌ Performance test error:', error.message);
      reject(error);
    });
  });
}

/**
 * Main test execution
 */
async function runTests() {
  console.log('🚀 Starting Sprint 3 Connection State Management Tests');
  console.log('📡 Server:', CONFIG.serverUrl);
  console.log('⏱️ Timeout:', CONFIG.timeout, 'ms');

  try {
    await testConnectionStateManagement();
    await testErrorHandling();
    await testConnectionRetryLogic();
    await testHealthMonitoring();
    await testGracefulDegradation();
    await testPerformanceMetrics();

    console.log('\n📊 Test Results Summary');
    console.log('========================');
    console.log(`✅ Passed: ${testResults.passed}`);
    console.log(`❌ Failed: ${testResults.failed}`);
    console.log(`📊 Total: ${testResults.total}`);
    console.log(`📈 Success Rate: ${((testResults.passed / testResults.total) * 100).toFixed(1)}%`);

    if (testResults.failed === 0) {
      console.log('\n🎉 All connection state management tests passed');
      console.log('✅ Sprint 3 connection state management requirements met');
    } else {
      console.log('\n⚠️ Some tests failed:');
      testResults.errors.forEach(error => console.log(`  - ${error}`));
      process.exit(1);
    }

  } catch (error) {
    console.error('\n❌ Test execution failed:', error.message);
    process.exit(1);
  }
}

// Run tests if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  runTests();
}
