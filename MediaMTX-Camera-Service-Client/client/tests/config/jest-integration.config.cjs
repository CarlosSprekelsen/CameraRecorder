/**
 * Integration test configuration
 * MANDATORY: Use this configuration for all integration tests
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-CONFIG-001: Integration test environment configuration
 * - REQ-CONFIG-002: Real server connection
 * - REQ-CONFIG-003: Coverage thresholds
 * 
 * Test Categories: Integration/E2E
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

/** @type {import('jest').Config} */
module.exports = {
  testEnvironment: 'node',
  setupFilesAfterEnv: ['<rootDir>/../setup.integration.ts'],
  testMatch: [
    '<rootDir>/../integration/**/test_*.{js,ts,tsx}',
    '<rootDir>/../e2e/**/test_*.{js,ts,tsx}'
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
    '^@/(.*)$': '<rootDir>/../../src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy'
  },
  testTimeout: 30000,
  transformIgnorePatterns: [
    'node_modules/(?!(ws)/)'
  ],
  collectCoverage: true,
  coverageDirectory: 'coverage/integration',
  coverageReporters: ['text', 'lcov', 'html'],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70
    }
  },
  collectCoverageFrom: [
    '<rootDir>/../../src/**/*.{ts,tsx}',
    '!<rootDir>/../../src/**/*.d.ts',
    '!<rootDir>/../../src/main.tsx',
    '!<rootDir>/../../src/vite-env.d.ts'
  ],
  coveragePathIgnorePatterns: [
    '/node_modules/',
    '/tests/',
    '/coverage/',
    '/dist/'
  ],
  // Integration test specific settings
  maxWorkers: 1, // Run integration tests sequentially
  forceExit: true, // Force exit after tests complete
  detectOpenHandles: true // Detect open handles
};
