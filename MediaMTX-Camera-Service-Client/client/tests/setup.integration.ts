/**
 * Integration test setup configuration
 * MANDATORY: Use this setup for all integration tests
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-SETUP-001: Integration test environment configuration
 * - REQ-SETUP-002: Real server connection setup
 * - REQ-SETUP-003: Test utilities initialization
 * 
 * Test Categories: Integration/E2E
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { loadTestEnvironment } from './utils/test-helpers';

// Load test environment variables
require('dotenv').config({ path: '.test_env' });

// Validate test environment
if (!process.env.TEST_WEBSOCKET_URL) {
  throw new Error('TEST_WEBSOCKET_URL environment variable is required for integration tests');
}

if (!process.env.TEST_JWT_SECRET) {
  throw new Error('TEST_JWT_SECRET environment variable is required for integration tests');
}

// Set test environment variables
process.env.NODE_ENV = 'test';
process.env.TEST_MOCK_MODE = 'false';

// Global test environment
let testEnvironment: any = null;

// Setup before all tests
beforeAll(async () => {
  try {
    testEnvironment = await loadTestEnvironment();
    console.log('Integration test environment loaded successfully');
  } catch (error) {
    console.error('Failed to load integration test environment:', error);
    throw error;
  }
});

// Cleanup after all tests
afterAll(async () => {
  if (testEnvironment?.apiClient) {
    await testEnvironment.apiClient.disconnect();
  }
});

// Global test utilities
declare global {
  var testEnv: any;
}

// Make test environment available globally
global.testEnv = testEnvironment;

// Custom Jest matchers for integration tests
expect.extend({
  toBeValidAPIResponse(received: any) {
    const pass = typeof received === 'object' && received !== null;
    return {
      message: () => `expected ${received} to be a valid API response`,
      pass
    };
  },
  toBeValidWebSocketMessage(received: any) {
    const pass = received && typeof received === 'object' && received.jsonrpc === '2.0';
    return {
      message: () => `expected ${received} to be a valid WebSocket message`,
      pass
    };
  },
  toBeValidAuthResult(received: any) {
    const pass = received && 
      typeof received.authenticated === 'boolean' &&
      typeof received.role === 'string' &&
      typeof received.session_id === 'string';
    return {
      message: () => `expected ${received} to be a valid auth result`,
      pass
    };
  }
});

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn()
};

// Mock localStorage for integration tests
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn()
};
global.localStorage = localStorageMock;

// Mock sessionStorage for integration tests
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn()
};
global.sessionStorage = sessionStorageMock;

// Mock fetch for HTTP requests
global.fetch = jest.fn();

// Mock URL.createObjectURL
global.URL.createObjectURL = jest.fn(() => 'mock-url');
global.URL.revokeObjectURL = jest.fn();

// Mock crypto for JWT operations
Object.defineProperty(global, 'crypto', {
  value: {
    randomUUID: jest.fn(() => 'mock-uuid'),
    getRandomValues: jest.fn((arr) => arr.map(() => Math.floor(Math.random() * 256)))
  }
});

// Cleanup after each test
afterEach(() => {
  jest.clearAllMocks();
  localStorageMock.clear();
  sessionStorageMock.clear();
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
