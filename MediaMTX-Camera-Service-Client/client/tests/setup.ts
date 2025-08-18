/**
 * Jest setup file for client tests
 * 
 * Configures test environment for:
 * - WebSocket mocking
 * - Service worker compatibility
 * - Timer mocking
 */

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