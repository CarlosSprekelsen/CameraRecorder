/**
 * Integration Test Environment Variables
 * 
 * Environment configuration for integration tests
 */

// Set test environment variables
process.env.NODE_ENV = 'test';
process.env.INTEGRATION_TEST = 'true';
process.env.SERVER_URL = 'ws://localhost:8002/ws';
process.env.TEST_TIMEOUT = '30000';
process.env.PERFORMANCE_TEST = 'true';
process.env.SECURITY_TEST = 'true';

// Mock console for cleaner test output
const originalConsole = console;
global.console = {
  ...originalConsole,
  log: (...args: any[]) => {
    if (process.env.VERBOSE_TESTS === 'true') {
      originalConsole.log(...args);
    }
  },
  error: (...args: any[]) => {
    originalConsole.error(...args);
  },
  warn: (...args: any[]) => {
    originalConsole.warn(...args);
  }
};

// Performance monitoring
global.performance = {
  now: () => Date.now(),
  mark: (name: string) => {
    // Mock performance mark
  },
  measure: (name: string, start: string, end: string) => {
    // Mock performance measure
  }
} as any;
