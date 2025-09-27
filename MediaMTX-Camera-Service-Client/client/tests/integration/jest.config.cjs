/**
 * Jest Configuration for Integration Tests
 * 
 * Configuration for integration tests that require real server
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-CONFIG-001: Integration test environment configuration
 * - REQ-CONFIG-002: Real server communication
 * - REQ-CONFIG-003: API compliance validation
 * 
 * Test Categories: Integration
 * API Documentation Reference: mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

const baseConfig = require('../../jest.config.base.cjs');

module.exports = {
  ...baseConfig,
  
  // Test environment
  testEnvironment: 'node',
  
  // Standardized test file patterns - use *.test.ts convention
  testMatch: [
    '<rootDir>/**/*.test.{js,ts,tsx}'
  ],
  
  // Setup files
  setupFilesAfterEnv: ['<rootDir>/setup.ts'],
  
  // Integration-specific coverage settings
  coverageDirectory: '<rootDir>/coverage/integration',
  coverageThreshold: {
    global: {
      statements: 60,
      branches: 50,
      functions: 60,
      lines: 60
    }
  },
  
  // Verbose output for integration tests
  verbose: true
};
