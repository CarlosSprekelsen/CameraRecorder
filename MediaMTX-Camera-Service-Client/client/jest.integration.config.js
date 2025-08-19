/**
 * Jest Configuration for Integration Tests
 * 
 * Uses Node.js environment for server integration tests
 * Separate from main Jest config which uses jsdom for React components
 */

export default {
  // Test environment - Node.js for server integration
  testEnvironment: 'node',
  
  // Test environment options
  testEnvironmentOptions: {
    url: 'http://localhost:3000'
  },
  
  // Module name mapping
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@tests/(.*)$': '<rootDir>/tests/$1',
    '^@fixtures/(.*)$': '<rootDir>/tests/fixtures/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    '\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$': '<rootDir>/tests/fixtures/fileMock.js'
  },
  
  // Test file patterns - integration tests only
  testMatch: [
    '<rootDir>/tests/integration/**/*.js',
    '<rootDir>/tests/integration/**/*.ts'
  ],
  
  // Test file exclusions
  testPathIgnorePatterns: [
    '/node_modules/',
    '/dist/',
    '/build/'
  ],
  
  // Coverage configuration
  collectCoverage: false, // Disable coverage for integration tests
  
  // Setup files
  setupFilesAfterEnv: [
    '<rootDir>/tests/setup-integration.ts'
  ],
  
  // Transform configuration
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: '<rootDir>/tsconfig.test.json'
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
  
  // Module resolution
  moduleDirectories: ['node_modules', 'src']
};
