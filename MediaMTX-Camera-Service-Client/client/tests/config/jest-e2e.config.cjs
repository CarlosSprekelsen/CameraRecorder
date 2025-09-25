/**
 * E2E test configuration
 * MANDATORY: Use this configuration for all E2E tests
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-CONFIG-001: E2E test environment configuration
 * - REQ-CONFIG-002: Real hardware interaction
 * - REQ-CONFIG-003: Performance validation
 * 
 * Test Categories: E2E/Performance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

/** @type {import('jest').Config} */
module.exports = {
  testEnvironment: 'node',
  setupFilesAfterEnv: ['<rootDir>/tests/setup.integration.ts'],
  testMatch: [
    '<rootDir>/tests/e2e/**/test_*.{js,ts,tsx}'
  ],
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx',
        skipLibCheck: true,
        esModuleInterop: true,
        allowSyntheticDefaultImports: true,
        typeRoots: ['<rootDir>/tests/types', 'node_modules/@types']
      }
    }],
    '^.+\\.js$': 'babel-jest'
  },
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy'
  },
  testTimeout: 60000, // Longer timeout for E2E tests
  transformIgnorePatterns: [
    'node_modules/(?!(ws)/)'
  ],
  collectCoverage: false, // No coverage for E2E tests
  maxWorkers: 1, // Run E2E tests sequentially
  forceExit: true, // Force exit after tests complete
  detectOpenHandles: true, // Detect open handles
  // E2E specific settings
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
