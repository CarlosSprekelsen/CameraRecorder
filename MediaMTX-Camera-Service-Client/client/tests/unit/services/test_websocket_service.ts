/**
 * WebSocketService unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-WS-001: WebSocket connection management
 * - REQ-WS-002: JSON-RPC message handling
 * - REQ-WS-003: Error handling and reconnection
 * - REQ-WS-004: Ping/pong heartbeat mechanism
 * - REQ-WS-005: Request timeout handling
 * 
 * Test Categories: Unit
 * API Documentation Reference: ../mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Mock WebSocket constants
const WebSocketConstants = {
  CONNECTING: 0,
  OPEN: 1,
  CLOSING: 2,
  CLOSED: 3,
};

// Mock WebSocket
const mockWebSocket = {
  readyState: WebSocketConstants.CONNECTING,
  send: jest.fn(),
  close: jest.fn(),
  onopen: null as (() => void) | null,
  onclose: null as ((event: { code: number; reason: string }) => void) | null,
  onerror: null as (() => void) | null,
  onmessage: null as ((event: { data: string }) => void) | null,
};

// Mock global WebSocket
(global as any).WebSocket = jest.fn(() => {
  // Simulate connection process
  setTimeout(() => {
    mockWebSocket.readyState = WebSocketConstants.OPEN;
    if (mockWebSocket.onopen) {
      mockWebSocket.onopen();
    }
  }, 0);
  return mockWebSocket;
});

describe('WebSocketService Unit Tests', () => {
  let webSocketService: WebSocketService;
  let mockEvents: any;

  beforeEach(() => {
    jest.clearAllMocks();
    mockEvents = {
      onConnect: jest.fn(),
      onDisconnect: jest.fn(),
      onError: jest.fn(),
      onNotification: jest.fn(),
      onResponse: jest.fn(),
    };

    webSocketService = new WebSocketService(
      {
        url: 'ws://localhost:8002/ws',
        maxReconnectAttempts: 3,
        reconnectDelay: 1000,
        maxReconnectDelay: 5000,
        pingInterval: 30000,
        pongTimeout: 5000,
      },
      mockEvents
    );
  });

  afterEach(() => {
    webSocketService.disconnect();
  });

  describe('REQ-WS-001: WebSocket connection management', () => {
    test('should connect successfully', async () => {
      const connectPromise = webSocketService.connect();

      // Simulate successful connection
      mockWebSocket.onopen?.();

      await connectPromise;

      expect(mockEvents.onConnect).toHaveBeenCalled();
      expect(webSocketService.isConnected).toBe(true);
    });

    test('should handle connection errors', async () => {
      const connectPromise = webSocketService.connect();

      // Simulate connection error
      mockWebSocket.onerror?.();

      await expect(connectPromise).rejects.toThrow('WebSocket connection failed');
      expect(mockEvents.onError).toHaveBeenCalledWith(expect.any(Error));
    });

    test('should disconnect properly', () => {
      webSocketService.connect();
      mockWebSocket.onopen?.();

      webSocketService.disconnect();

      expect(mockWebSocket.close).toHaveBeenCalledWith(1000, 'Client disconnect');
      expect(webSocketService.isConnected).toBe(false);
    });
  });

  describe('REQ-WS-002: JSON-RPC message handling', () => {
    beforeEach(async () => {
      await webSocketService.connect();
      mockWebSocket.onopen?.();
    });

    test('should send RPC requests correctly', async () => {
      const method = 'get_camera_list';
      const params = { limit: 10 };
      const expectedResult = MockDataFactory.getCameraListResult();

      // Mock successful response
      const mockResponse = {
        jsonrpc: '2.0',
        result: expectedResult,
        id: 1
      };
      
      // Simulate response after a short delay
      setTimeout(() => {
        mockWebSocket.onmessage?.({ data: JSON.stringify(mockResponse) });
      }, 10);

      const result = await webSocketService.sendRPC(method, params);

      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({
          jsonrpc: '2.0',
          method,
          params,
          id: 1,
        })
      );
      expect(result).toEqual(expectedResult);
      expect(mockEvents.onResponse).toHaveBeenCalledWith(mockResponse);
    });

    test('should send notifications correctly', () => {
      const method = 'camera_status_update';
      const params = { device: 'camera0', status: 'CONNECTED' };

      webSocketService.sendNotification(method, params);

      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({
          jsonrpc: '2.0',
          method,
          params,
        })
      );
    });

    test('should reject RPC when not connected', async () => {
      webSocketService.disconnect();

      await expect(webSocketService.sendRPC('get_camera_list')).rejects.toThrow(
        'WebSocket not connected'
      );
    });
  });

  describe('REQ-WS-003: Error handling and reconnection', () => {
    test('should handle RPC errors correctly', async () => {
      await webSocketService.connect();
      mockWebSocket.onopen?.();

      const errorResponse = MockDataFactory.getErrorResponse(-32601, 'Method Not Found');
      
      // Simulate error response
      setTimeout(() => {
        mockWebSocket.onmessage?.({ data: JSON.stringify(errorResponse) });
      }, 10);

      await expect(webSocketService.sendRPC('invalid_method')).rejects.toThrow(
        'RPC Error -32601: Method Not Found'
      );
    });

    test('should handle specific error codes', async () => {
      await webSocketService.connect();
      mockWebSocket.onopen?.();

      const authError = MockDataFactory.getErrorResponse(-32001, 'Auth Failed');
      
      setTimeout(() => {
        mockWebSocket.onmessage?.({ data: JSON.stringify(authError) });
      }, 10);

      await expect(webSocketService.sendRPC('authenticate')).rejects.toThrow(
        'Authentication failed. Please log in again.'
      );
    });
  });

  describe('REQ-WS-004: Ping/pong heartbeat mechanism', () => {
    beforeEach(() => {
      jest.useFakeTimers();
    });

    afterEach(() => {
      jest.useRealTimers();
    });

    test('should start ping interval on connection', async () => {
      await webSocketService.connect();
      mockWebSocket.onopen?.();

      // Fast-forward time to trigger ping
      jest.advanceTimersByTime(30000);

      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({
          jsonrpc: '2.0',
          method: 'ping',
          id: expect.any(Number),
        })
      );
    });
  });

  describe('REQ-WS-005: Request timeout handling', () => {
    beforeEach(() => {
      jest.useFakeTimers();
    });

    afterEach(() => {
      jest.useRealTimers();
    });

    test('should timeout requests after 30 seconds', async () => {
      await webSocketService.connect();
      mockWebSocket.onopen?.();

      const requestPromise = webSocketService.sendRPC('slow_method');

      // Fast-forward time to trigger timeout
      jest.advanceTimersByTime(30000);

      await expect(requestPromise).rejects.toThrow('RPC request timeout: slow_method');
    });
  });

  describe('Connection state management', () => {
    test('should return correct connection state', () => {
      expect(webSocketService.connectionState).toBe(WebSocketConstants.CLOSED);

      webSocketService.connect();
      expect(webSocketService.connectionState).toBe(WebSocketConstants.CONNECTING);

      mockWebSocket.onopen?.();
      expect(webSocketService.connectionState).toBe(WebSocketConstants.OPEN);
    });
  });
});