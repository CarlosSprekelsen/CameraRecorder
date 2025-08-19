/**
 * Jest Configuration for MediaMTX Camera Service Client
 * 
 * Supports unified testing strategy with real server integration
 * Following "Real Integration First" approach
 * 
 * CRITICAL: All paths relative to client/ directory only
 */

export default {
  // Test environment
  testEnvironment: 'jsdom',
  
  // Test environment options
  testEnvironmentOptions: {
    url: 'http://localhost:3000',
    customExportConditions: ['node', 'node-addons']
  },
  
  // React 18 compatibility
  setupFilesAfterEnv: [
    '<rootDir>/tests/setup.ts'
  ],
  
  // Module name mapping for React 18 compatibility
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@tests/(.*)$': '<rootDir>/tests/$1',
    '^@fixtures/(.*)$': '<rootDir>/tests/fixtures/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    '\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$': '<rootDir>/tests/fixtures/fileMock.js',
    '^react-dom/client$': 'react-dom/client',
    '^react-dom$': 'react-dom'
  },
  
  // Test file patterns - ALL relative to client/
  testMatch: [
    '<rootDir>/tests/**/*.test.{ts,tsx}',
    '<rootDir>/tests/**/*.spec.{ts,tsx}',
    '<rootDir>/tests/**/test_*_unit.js',
    '<rootDir>/tests/**/test_*_integration.js',
    '<rootDir>/tests/**/test_*_e2e.js',
    '<rootDir>/tests/**/test_*_performance.js',
    '<rootDir>/tests/**/test_*_validation.ts',
    '<rootDir>/tests/**/test_*.{ts,tsx}',
    '<rootDir>/tests/**/test_*.js',
    '<rootDir>/src/**/*.test.{ts,tsx}',
    '<rootDir>/src/**/*.spec.{ts,tsx}'
  ],
  
  // Test file exclusions
  testPathIgnorePatterns: [
    '/node_modules/',
    '/dist/',
    '/build/'
  ],
  
  // Coverage configuration
  collectCoverage: true,
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.test.{ts,tsx}',
    '!src/**/*.spec.{ts,tsx}',
    '!src/index.tsx',
    '!src/vite-env.d.ts',
    '!src/main.tsx'
  ],
  
  // Coverage thresholds (from testing guidelines)
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  },
  
  // Coverage reporters
  coverageReporters: [
    'text',
    'lcov',
    'html'
  ],
  
  // Transform configuration
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: '<rootDir>/tsconfig.test.json',
      useESM: false
    }],
    '^.+\\.(js|jsx)$': 'babel-jest'
  },
  
  // Transform ignore patterns for ES modules
  transformIgnorePatterns: [
    'node_modules/(?!(ws|buffer)/)'
  ],
  
  // Module file extensions
  moduleFileExtensions: [
    'ts',
    'tsx',
    'js',
    'jsx',
    'json'
  ],
  
  // Test timeout configuration
  testTimeout: 30000, // 30 seconds for integration tests
  
  // Performance monitoring
  verbose: true,
  
  // Clear mocks between tests
  clearMocks: true,
  
  // Restore mocks between tests
  restoreMocks: true,
  
  // Reset modules between tests
  resetModules: true,
  
  // Module resolution - ensure we use client's node_modules ONLY
  moduleDirectories: ['<rootDir>/node_modules', 'src'],
  
  // Test environment variables
  testEnvironmentOptions: {
    url: 'http://localhost:3000'
  }
};
