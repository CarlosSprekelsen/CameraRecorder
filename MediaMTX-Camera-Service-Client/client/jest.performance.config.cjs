/** @type {import('jest').Config} */
module.exports = {
  testEnvironment: 'node',
  setupFilesAfterEnv: ['<rootDir>/tests/setup.integration.ts'],
  testMatch: [
    '<rootDir>/tests/performance/**/test_*.{js,ts,tsx}'
  ],
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx'
      }
    }],
    '^.+\\.js$': 'babel-jest'
  },
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy'
  },
  testTimeout: 60000, // Longer timeout for performance tests
  
  transformIgnorePatterns: [
    'node_modules/(?!(ws)/)'
  ],
  
  // Performance test specific settings
  verbose: true,
  collectCoverage: false, // Performance tests don't need coverage
  maxWorkers: 1, // Run performance tests sequentially to avoid interference
};
