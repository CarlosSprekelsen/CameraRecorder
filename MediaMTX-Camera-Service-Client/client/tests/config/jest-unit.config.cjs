/**
 * Unit test configuration
 * MANDATORY: Use this configuration for all unit tests
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-CONFIG-001: Unit test environment configuration
 * - REQ-CONFIG-002: Mock setup
 * - REQ-CONFIG-003: Coverage thresholds
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

const baseConfig = require('../../jest.config.base.cjs');

/** @type {import('jest').Config} */
module.exports = {
  ...baseConfig,
  
  rootDir: '../../',
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/tests/setup.ts'],
  
  // Standardized test file patterns - use *.test.ts convention
  testMatch: [
    '<rootDir>/tests/unit/**/*.test.{js,ts,tsx}',
    '<rootDir>/src/**/*.test.{js,ts,tsx}'
  ],
  
  // Unit-specific coverage settings
  coverageDirectory: 'coverage/unit',
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  }
};
