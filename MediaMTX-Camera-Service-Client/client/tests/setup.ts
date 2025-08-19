/**
 * Jest setup file for client tests
 * 
 * Configures test environment for:
 * - WebSocket mocking
 * - Service worker compatibility
 * - Timer mocking
 * - DOM testing utilities
 */

// Import jest-dom matchers
import '@testing-library/jest-dom';

// Mock WebSocket for tests
class MockWebSocket {
  public readyState: number = WebSocket.CONNECTING;
  public url: string;
  public onopen: (() => void) | null = null;
  public onclose: ((event: { wasClean: boolean }) => void) | null = null;
  public onerror: ((event: Event) => void) | null = null;
  public onmessage: ((event: { data: string }) => void) | null = null;
  public send: jest.Mock = jest.fn();
  public close: jest.Mock = jest.fn();

  constructor(url: string) {
    this.url = url;
  }
}

// Mock global WebSocket
global.WebSocket = MockWebSocket as unknown as typeof WebSocket;

// Mock service worker environment
Object.defineProperty(global, 'navigator', {
  value: {
    serviceWorker: {
      register: jest.fn(),
      getRegistration: jest.fn(),
      getRegistrations: jest.fn(),
    },
  },
  writable: true,
});

// Mock window.location if not already defined
if (!global.location) {
  Object.defineProperty(global, 'location', {
    value: {
      href: 'http://localhost:3000',
      origin: 'http://localhost:3000',
      protocol: 'http:',
      host: 'localhost:3000',
      hostname: 'localhost',
      port: '3000',
      pathname: '/',
      search: '',
      hash: '',
    },
    writable: true,
  });
}

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
};

// Ensure React DOM is properly set up for tests
if (typeof window === 'undefined') {
  global.window = {} as any;
}

if (typeof document === 'undefined') {
  global.document = {} as any;
}

// Mock React 18 features that might cause issues
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

// Mock matchMedia for Material-UI compatibility
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
}); 