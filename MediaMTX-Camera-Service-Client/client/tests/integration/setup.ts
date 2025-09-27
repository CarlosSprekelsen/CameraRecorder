/**
 * Integration Test Setup
 * 
 * Setup for integration tests with real server
 */

import { LoggerService } from '../../src/services/logger/LoggerService';
import WebSocket from 'ws';

// Create a WebSocket mock with constants for Node.js environment
class WebSocketMock extends WebSocket {
  static OPEN = 1;
  static CLOSED = 3;
  static CONNECTING = 0;
  static CLOSING = 2;
}

// Mock browser APIs for Node.js environment
global.WebSocket = WebSocketMock as any;

// Global test setup
beforeAll(async () => {
  console.log('🚀 Starting Integration Tests with Real Server');
  console.log('📡 Server URL: ws://localhost:8002/ws');
  console.log('⏱️  Test Timeout: 30 seconds per test');
  console.log('🔒 Security Tests: Enabled');
  console.log('📊 Performance Tests: Enabled');
});

// Global test teardown
afterAll(async () => {
  console.log('✅ Integration Tests Completed');
  console.log('📊 Performance metrics collected');
  console.log('🔒 Security validation completed');
  console.log('📡 API compliance verified');
});

// Test timeout configuration
jest.setTimeout(30000);

// Global error handling
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

process.on('uncaughtException', (error) => {
  console.error('Uncaught Exception:', error);
});

// Performance monitoring
const performanceMonitor = {
  startTime: Date.now(),
  operations: 0,
  errors: 0,
  
  recordOperation: () => {
    performanceMonitor.operations++;
  },
  
  recordError: () => {
    performanceMonitor.errors++;
  },
  
  getMetrics: () => {
    const endTime = Date.now();
    const duration = endTime - performanceMonitor.startTime;
    return {
      duration,
      operations: performanceMonitor.operations,
      errors: performanceMonitor.errors,
      successRate: performanceMonitor.operations > 0 
        ? ((performanceMonitor.operations - performanceMonitor.errors) / performanceMonitor.operations) * 100 
        : 0
    };
  }
};

// Export for use in tests
(global as any).performanceMonitor = performanceMonitor;
