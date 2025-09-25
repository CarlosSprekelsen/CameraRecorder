/**
 * Unit test configuration
 * MANDATORY: Use this configuration for all unit tests
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-CONFIG-001: Unit test environment configuration
 * - REQ-CONFIG-002: Mock setup
 * - REQ-CONFIG-003: Coverage thresholds
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

/** @type {import('jest').Config} */
module.exports = {
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/../setup.ts'],
  testMatch: [
    '<rootDir>/../unit/**/*.test.{js,ts,tsx}',
    '<rootDir>/../unit/**/test_*.{js,ts,tsx}',
    '<rootDir>/../../src/**/*.test.{js,ts,tsx}',
    '<rootDir>/../../src/**/test_*.{js,ts,tsx}'
  ],
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx',
        skipLibCheck: true,
        esModuleInterop: true,
        allowSyntheticDefaultImports: true
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
  coverageDirectory: 'coverage/unit',
  coverageReporters: ['text', 'lcov', 'html'],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
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
  ]
};
