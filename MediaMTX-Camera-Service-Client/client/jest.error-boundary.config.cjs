/** @type {import('jest').Config} */
module.exports = {
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/tests/setup.ts'],
  testMatch: [
    '<rootDir>/tests/unit/components/ErrorBoundaries/**/*.test.{js,ts,tsx}'
  ],
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx'
      }
    }],
    '^.+\\.js$': 'babel-jest'
  },
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy'
  },
  testTimeout: 30000,
  
  // Mock problematic modules
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    '^../../../../src/services/loggerService$': '<rootDir>/tests/unit/components/ErrorBoundaries/__mocks__/loggerService.js'
  },
  
  // Transform ignore patterns
  transformIgnorePatterns: [
    'node_modules/(?!(ws)/)'
  ],
  
  // Coverage configuration
  collectCoverageFrom: [
    'src/components/ErrorBoundaries/**/*.{ts,tsx}',
    '!src/components/ErrorBoundaries/**/*.d.ts',
    '!src/components/ErrorBoundaries/**/*.test.{ts,tsx}',
    '!src/components/ErrorBoundaries/**/*.spec.{ts,tsx}'
  ],
  
  coverageThreshold: {
    global: {
      branches: 90,
      functions: 90,
      lines: 90,
      statements: 90
    },
    'src/components/ErrorBoundaries/': {
      branches: 95,
      functions: 95,
      lines: 95,
      statements: 95
    }
  }
};
