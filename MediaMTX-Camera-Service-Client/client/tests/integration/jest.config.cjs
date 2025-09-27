/**
 * Jest Configuration for Integration Tests
 * 
 * Configuration for integration tests that require real server
 */

module.exports = {
  // Test environment
  testEnvironment: 'node',
  
  // Test file patterns
  testMatch: [
    '<rootDir>/test_basic_connectivity.ts'
  ],
  
  // Setup files
  setupFilesAfterEnv: ['<rootDir>/setup.ts'],
  
  // Coverage configuration
  collectCoverage: true,
  coverageDirectory: '<rootDir>/coverage/integration',
  coverageReporters: ['text', 'lcov', 'html'],
  
  // Coverage thresholds (lower for integration tests)
  coverageThreshold: {
    global: {
      statements: 60,
      branches: 50,
      functions: 60,
      lines: 60
    }
  },
  
  // Test timeout (longer for integration tests)
  testTimeout: 30000,
  
  // Verbose output
  verbose: true,
  
  // Transform configuration
  transform: {
    '^.+\\.ts$': 'ts-jest'
  },
  
  // Module file extensions
  moduleFileExtensions: ['ts', 'js', 'json'],
  
  // TypeScript configuration
  globals: {
    'ts-jest': {
      tsconfig: '<rootDir>/tsconfig.json'
    }
  },
  
  // Test environment variables
  // setupFiles: ['<rootDir>/tests/integration/env.ts'],
  
  // Global setup and teardown
  // globalSetup: '<rootDir>/tests/integration/globalSetup.ts',
  // globalTeardown: '<rootDir>/tests/integration/globalTeardown.ts'
};
