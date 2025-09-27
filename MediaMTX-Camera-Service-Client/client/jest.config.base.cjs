/**
 * Base Jest Configuration
 * 
 * Shared configuration for all test types
 * 
 * Ground Truth References:
 * - Testing Guidelines: docs/development/client-testing-guidelines.md
 * - Client Architecture: docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-CONFIG-001: Base test environment configuration
 * - REQ-CONFIG-002: Common transform and module settings
 * - REQ-CONFIG-003: Shared coverage configuration
 * 
 * Test Categories: Base
 * API Documentation Reference: mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

/** @type {import('jest').Config} */
const baseConfig = {
  // Common transform configuration
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

  // Module name mapping
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy'
  },

  // Transform ignore patterns
  transformIgnorePatterns: [
    'node_modules/(?!(ws)/)'
  ],

  // Module file extensions
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json'],

  // Common test timeout
  testTimeout: 30000,

  // Coverage configuration
  collectCoverage: true,
  coverageReporters: ['text', 'lcov', 'html'],

  // Coverage collection patterns
  collectCoverageFrom: [
    '<rootDir>/src/**/*.{ts,tsx}',
    '!<rootDir>/src/**/*.d.ts',
    '!<rootDir>/src/main.tsx',
    '!<rootDir>/src/vite-env.d.ts'
  ],

  // Coverage ignore patterns
  coveragePathIgnorePatterns: [
    '/node_modules/',
    '/tests/',
    '/coverage/',
    '/dist/',
    '/build/'
  ]
};

module.exports = baseConfig;
