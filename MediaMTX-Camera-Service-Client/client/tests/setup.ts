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

// Simple, TypeScript-compliant Mock WebSocket for unit tests
class MockWebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;

  public readyState: number = MockWebSocket.CONNECTING;
  public readonly url: string;
  public readonly protocol: string = '';
  public readonly extensions: string = '';
  public readonly bufferedAmount: number = 0;
  public binaryType: BinaryType = 'blob';
  
  public onopen: ((event: Event) => void) | null = null;
  public onclose: ((event: CloseEvent) => void) | null = null;
  public onerror: ((event: Event) => void) | null = null;
  public onmessage: ((event: MessageEvent) => void) | null = null;
  
  public send: jest.Mock = jest.fn();
  public close: jest.Mock = jest.fn();
  public addEventListener: jest.Mock = jest.fn();
  public removeEventListener: jest.Mock = jest.fn();
  public dispatchEvent: jest.Mock = jest.fn();

  constructor(url: string, protocols?: string | string[]) {
    this.url = url;
    if (protocols) {
      (this as any).protocol = Array.isArray(protocols) ? protocols[0] : protocols;
    }
    
    // Simulate connection opening after a short delay
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN;
      if (this.onopen) {
        this.onopen(new Event('open'));
      }
    }, 10);
  }
}

// Mock global WebSocket for unit tests
global.WebSocket = MockWebSocket as unknown as typeof WebSocket;

// Mock service worker environment
Object.defineProperty(global, 'navigator', {
  value: {
    serviceWorker: {
      register: jest.fn(),
      getRegistration: jest.fn(),
      getRegistrations: jest.fn(),
    },
    userAgent: 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
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
  global.window = {
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
    document: {
      createElement: jest.fn((tagName) => ({
        tagName: tagName.toUpperCase(),
        setAttribute: jest.fn(),
        getAttribute: jest.fn(),
        appendChild: jest.fn(),
        removeChild: jest.fn(),
        querySelector: jest.fn(),
        querySelectorAll: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
        // React 19 specific properties
        _reactRootContainer: null,
        _reactInternalInstance: null,
      })),
      body: {
        appendChild: jest.fn(),
        removeChild: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
      },
      createTextNode: jest.fn((text) => ({ textContent: text })),
      getElementById: jest.fn(),
    },
  } as any;
}

if (typeof document === 'undefined') {
  global.document = {
    createElement: jest.fn((tagName) => ({
      tagName: tagName.toUpperCase(),
      setAttribute: jest.fn(),
      getAttribute: jest.fn(),
      appendChild: jest.fn(),
      removeChild: jest.fn(),
      querySelector: jest.fn(),
      querySelectorAll: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
      // React 19 specific properties
      _reactRootContainer: null,
      _reactInternalInstance: null,
    })),
    body: {
      appendChild: jest.fn(),
      removeChild: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
    },
    createTextNode: jest.fn((text) => ({ textContent: text })),
    getElementById: jest.fn(),
  } as any;
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

// Mock Event constructor with proper static properties
class MockEvent {
  static NONE = 0;
  static CAPTURING_PHASE = 1;
  static AT_TARGET = 2;
  static BUBBLING_PHASE = 3;

  public type: string;
  public preventDefault: jest.Mock;
  public stopPropagation: jest.Mock;

  constructor(type: string, options: any = {}) {
    this.type = type;
    this.preventDefault = jest.fn();
    this.stopPropagation = jest.fn();
    Object.assign(this, options);
  }
}

// Mock MessageEvent constructor
class MockMessageEvent {
  public type: string;
  public data: any;
  public origin: string;
  public lastEventId: string;
  public source: any;
  public ports: any[];

  constructor(type: string, options: any = {}) {
    this.type = type;
    this.data = options.data || '';
    this.origin = options.origin || '';
    this.lastEventId = options.lastEventId || '';
    this.source = options.source || null;
    this.ports = options.ports || [];
    Object.assign(this, options);
  }
}

// Mock CloseEvent constructor
class MockCloseEvent {
  public type: string;
  public code: number;
  public reason: string;
  public wasClean: boolean;

  constructor(type: string, options: any = {}) {
    this.type = type;
    this.code = options.code || 1000;
    this.reason = options.reason || '';
    this.wasClean = options.wasClean || true;
    Object.assign(this, options);
  }
}

// Assign mocks to global
global.Event = MockEvent as any;
global.MessageEvent = MockMessageEvent as any;
global.CloseEvent = MockCloseEvent as any; 