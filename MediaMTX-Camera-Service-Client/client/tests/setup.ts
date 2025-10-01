/**
 * Unit test setup configuration
 * MANDATORY: Use this setup for all unit tests
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-SETUP-001: Unit test environment configuration
 * - REQ-SETUP-002: Mock setup
 * - REQ-SETUP-003: Test utilities initialization
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import dotenv from 'dotenv';
import path from 'path';
import '@testing-library/jest-dom';
import { configure } from '@testing-library/react';

// Load test environment tokens from server-generated .test_env
dotenv.config({ path: path.join(__dirname, '../.test_env') });

// Configure testing library
configure({
  testIdAttribute: 'data-testid'
});

// Mock WebSocket for unit tests
global.WebSocket = class MockWebSocket {
  private listeners: { [key: string]: Function[] } = {};
  public readyState = WebSocket.CONNECTING;
  public url: string;
  private timeouts: NodeJS.Timeout[] = [];

  constructor(url: string) {
    this.url = url;
    // Simulate connection
    const timeout = setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      this.listeners['open']?.forEach(listener => listener());
    }, 10);
    this.timeouts.push(timeout);
  }

  on(event: string, listener: Function): void {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(listener);
  }

  send(data: string): void {
    // Mock response based on request
    const request = JSON.parse(data);
    const response = this.getMockResponse(request);
    
    const timeout = setTimeout(() => {
      this.listeners['message']?.forEach(listener => {
        listener({ data: JSON.stringify(response) });
      });
    }, 10);
    this.timeouts.push(timeout);
  }

  close(): void {
    this.readyState = WebSocket.CLOSED;
    this.listeners['close']?.forEach(listener => listener());
    // Clear all timeouts
    this.timeouts.forEach(timeout => clearTimeout(timeout));
    this.timeouts = [];
  }

  private getMockResponse(request: any): any {
    switch (request.method) {
      case 'ping':
        return { jsonrpc: '2.0', result: 'pong', id: request.id };
      case 'authenticate':
        return {
          jsonrpc: '2.0',
          result: {
            authenticated: true,
            role: 'admin',
            permissions: ['read', 'write', 'delete', 'admin'],
            expires_at: new Date(Date.now() + 3600000).toISOString(),
            session_id: 'test-session-id'
          },
          id: request.id
        };
      case 'get_camera_list':
        return {
          jsonrpc: '2.0',
          result: {
            cameras: [
              {
                device: 'camera0',
                status: 'CONNECTED',
                name: 'Test Camera',
                resolution: '1920x1080',
                fps: 30,
                streams: {
                  rtsp: 'rtsp://localhost:8554/camera0',
                  hls: 'https://localhost/hls/camera0.m3u8'
                }
              }
            ],
            total: 1,
            connected: 1
          },
          id: request.id
        };
      default:
        return {
          jsonrpc: '2.0',
          error: {
            code: -32601,
            message: 'Method Not Found'
          },
          id: request.id
        };
    }
  }
} as any;

// Mock fetch for HTTP requests
global.fetch = jest.fn();

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn()
};

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn()
};
global.localStorage = localStorageMock;

// Mock sessionStorage
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn()
};
global.sessionStorage = sessionStorageMock;

// Mock IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  observe() {}
  disconnect() {}
  unobserve() {}
};

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  constructor() {}
  observe() {}
  disconnect() {}
  unobserve() {}
};

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock URL.createObjectURL
global.URL.createObjectURL = jest.fn(() => 'mock-url');
global.URL.revokeObjectURL = jest.fn();

// Mock crypto for JWT operations
Object.defineProperty(global, 'crypto', {
  value: {
    randomUUID: jest.fn(() => 'mock-uuid'),
    getRandomValues: jest.fn((arr) => arr.map(() => Math.floor(Math.random() * 256)))
  }
});

// Setup test environment variables
process.env.NODE_ENV = 'test';
process.env.TEST_MOCK_MODE = 'true';
process.env.TEST_WEBSOCKET_URL = 'ws://localhost:8002/ws';
process.env.TEST_JWT_SECRET = 'test-secret-key';
process.env.TEST_TIMEOUT = '30000';

// Global test utilities
declare global {
  namespace jest {
    interface Matchers<R> {
      toBeValidISOString(): R;
      toBeValidDeviceId(): R;
      toBeValidStreamUrl(): R;
    }
  }
}

// Custom Jest matchers
expect.extend({
  toBeValidISOString(received: string) {
    const pass = typeof received === 'string' && !isNaN(Date.parse(received));
    return {
      message: () => `expected ${received} to be a valid ISO string`,
      pass
    };
  },
  toBeValidDeviceId(received: string) {
    const pass = /^camera[0-9]+$/.test(received);
    return {
      message: () => `expected ${received} to be a valid device ID`,
      pass
    };
  },
  toBeValidStreamUrl(received: string) {
    const pass = typeof received === 'string' && (received.startsWith('rtsp://') || received.startsWith('https://'));
    return {
      message: () => `expected ${received} to be a valid stream URL`,
      pass
    };
  }
});

// Cleanup after each test
afterEach(() => {
  jest.clearAllMocks();
  localStorageMock.clear();
  sessionStorageMock.clear();
});
