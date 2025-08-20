/**
 * Jest setup file for integration tests
 * 
 * Configures test environment for:
 * - Real WebSocket connections using Node.js ws library
 * - Real server integration
 * - No mocking of WebSocket
 */

// Import jest-dom matchers
import '@testing-library/jest-dom';

// For integration tests, we want to use the real WebSocket library
// The Node.js environment should provide the real WebSocket from 'ws' library

// Mock console methods to reduce noise in tests but keep errors
global.console = {
  ...console,
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  // Keep error logging for debugging
  error: console.error,
};

// Ensure environment variables are loaded
if (!process.env.CAMERA_SERVICE_JWT_SECRET) {
  console.warn('⚠️ CAMERA_SERVICE_JWT_SECRET not set. Authentication tests may fail.');
}

if (!process.env.TEST_SERVER_URL) {
  console.warn('⚠️ TEST_SERVER_URL not set. Integration tests may fail.');
}
