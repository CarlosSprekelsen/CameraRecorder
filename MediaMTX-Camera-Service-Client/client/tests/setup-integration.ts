/**
 * Integration Test Setup
 * 
 * Configures testing environment for real server integration
 * Following "Real Integration First" approach
 */

import { shouldUseMockServer } from './fixtures/mock-server';

// Global test configuration
beforeAll(async () => {
  console.log('Setting up integration test environment...');
  
  // Check if we should use mock server
  if (shouldUseMockServer()) {
    console.log('Using mock server for integration tests (USE_MOCK_SERVER=true)');
    return;
  }
  
  // Validate real server availability
  const isServerAvailable = await validateServerAvailability();
  if (!isServerAvailable) {
    throw new Error(`
      MediaMTX Camera Service not available for integration tests.
      
      To run integration tests with real server:
      1. Start MediaMTX Camera Service: sudo systemctl start mediamtx-camera-service
      2. Verify service is running: sudo systemctl status mediamtx-camera-service
      3. Check WebSocket endpoint: ws://localhost:8002/ws
      
      To run with mock server:
      USE_MOCK_SERVER=true npm test -- --testPathPattern=integration
    `);
  }
  
  console.log('Real server available for integration tests');
});

// Global test teardown
afterAll(async () => {
  console.log('Cleaning up integration test environment...');
  
  // Clean up any test data or connections
  await cleanupTestEnvironment();
});

// Test timeout configuration for integration tests
jest.setTimeout(30000); // 30 seconds for integration tests

/**
 * Validate server availability
 */
async function validateServerAvailability(): Promise<boolean> {
  const TEST_WEBSOCKET_URL = process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws';
  const TEST_API_URL = process.env.TEST_API_URL || 'http://localhost:8003/health/system';
  
  try {
    // Check WebSocket endpoint
    const wsAvailable = await checkWebSocketAvailability(TEST_WEBSOCKET_URL);
    if (!wsAvailable) {
      console.warn('WebSocket endpoint not available:', TEST_WEBSOCKET_URL);
      return false;
    }
    
    // Check API endpoint
    const apiAvailable = await checkApiAvailability(TEST_API_URL);
    if (!apiAvailable) {
      console.warn('API endpoint not available:', TEST_API_URL);
      return false;
    }
    
    return true;
  } catch (error) {
    console.error('Error validating server availability:', error);
    return false;
  }
}

/**
 * Check WebSocket availability
 */
async function checkWebSocketAvailability(url: string): Promise<boolean> {
  return new Promise((resolve) => {
    try {
      const ws = new WebSocket(url);
      
      const timeout = setTimeout(() => {
        ws.close();
        resolve(false);
      }, 5000);

      ws.onopen = () => {
        clearTimeout(timeout);
        ws.close();
        resolve(true);
      };

      ws.onerror = () => {
        clearTimeout(timeout);
        resolve(false);
      };
    } catch {
      resolve(false);
    }
  });
}

/**
 * Check API availability
 */
async function checkApiAvailability(url: string): Promise<boolean> {
  try {
    const response = await fetch(url);
    return response.ok;
  } catch {
    return false;
  }
}

/**
 * Cleanup test environment
 */
async function cleanupTestEnvironment(): Promise<void> {
  // Clean up any test files or data
  // This could include removing test recordings, snapshots, etc.
  
  try {
    // Example cleanup operations
    // await cleanupTestFiles();
    // await resetTestState();
  } catch (error) {
    console.warn('Error during test environment cleanup:', error);
  }
}

// Global error handling for integration tests
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

process.on('uncaughtException', (error) => {
  console.error('Uncaught Exception:', error);
});

// Performance monitoring for integration tests
const originalConsoleLog = console.log;
console.log = (...args) => {
  if (args[0]?.includes('Performance Test')) {
    // Log performance metrics to file for analysis
    const fs = require('fs');
    const performanceLog = args.join(' ') + '\n';
    fs.appendFileSync('test-results/performance.log', performanceLog);
  }
  originalConsoleLog(...args);
};
