/**
 * E2E test configuration
 * MANDATORY: Use this configuration for all E2E tests
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-CONFIG-001: E2E test environment configuration
 * - REQ-CONFIG-002: Real hardware interaction
 * - REQ-CONFIG-003: Performance validation
 * 
 * Test Categories: E2E/Performance
 * API Documentation Reference: mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

const baseConfig = require('../../jest.config.base.cjs');

/** @type {import('jest').Config} */
module.exports = {
  ...baseConfig,
  
  rootDir: '../../',
  testEnvironment: 'node',
  setupFilesAfterEnv: ['<rootDir>/tests/setup.integration.ts'],
  
  // Standardized test file patterns - use *.test.ts convention
  testMatch: [
    '<rootDir>/tests/e2e/**/*.test.{js,ts,tsx}'
  ],
  
  // E2E-specific settings
  testTimeout: 60000, // Longer timeout for E2E tests
  collectCoverage: false, // No coverage for E2E tests
  maxWorkers: 1, // Run E2E tests sequentially
  forceExit: true, // Force exit after tests complete
  detectOpenHandles: true, // Detect open handles
  verbose: true, // Verbose output for E2E tests
  bail: false, // Don't bail on first failure
  
  // Performance monitoring
  reporters: [
    'default',
    ['jest-html-reporters', {
      publicPath: './coverage/e2e',
      filename: 'e2e-report.html'
    }]
  ]
};
