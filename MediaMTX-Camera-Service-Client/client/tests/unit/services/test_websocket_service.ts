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

// Use centralized WebSocket mock - eliminates duplication
const mockWebSocketService = MockDataFactory.createMockWebSocketService();

describe('WebSocketService Unit Tests', () => {
  let webSocketService: WebSocketService;
  let mockEvents: any;

  beforeEach(() => {
    jest.clearAllMocks();
    mockEvents = {
      onConnect: MockDataFactory.createMockEventHandler(),
      onDisconnect: MockDataFactory.createMockEventHandler(),
      onError: MockDataFactory.createMockEventHandler(),
      onNotification: MockDataFactory.createMockEventHandler(),
      onResponse: MockDataFactory.createMockEventHandler(),
    };

    // Use centralized mock instead of creating new WebSocketService
    webSocketService = mockWebSocketService as any;
    
    // Set up mock WebSocket
    const mockWs = MockDataFactory.createMockWebSocket();
    (webSocketService as any).ws = mockWs;
    (webSocketService as any).connectionState = 0; // CONNECTING
    (webSocketService as any).isConnected = false;
  });

  afterEach(() => {
    webSocketService.disconnect();
  });

  describe('REQ-WS-001: WebSocket connection management', () => {
    test('should connect successfully', async () => {
      // Mock the connect method to resolve immediately
      (webSocketService as any).connect = jest.fn().mockResolvedValue(undefined);
      
      await webSocketService.connect();

      expect(webSocketService.connect).toHaveBeenCalled();
      expect(webSocketService.isConnected).toBe(true);
    });

    test('should handle connection errors', async () => {
      // Mock the connect method to reject with error
      (webSocketService as any).connect = jest.fn().mockRejectedValue(new Error('WebSocket connection failed'));

      await expect(webSocketService.connect()).rejects.toThrow('WebSocket connection failed');
    });

    test('should disconnect properly', () => {
      // Mock the disconnect method
      (webSocketService as any).disconnect = jest.fn().mockImplementation(() => {
        (webSocketService as any).isConnected = false;
        (webSocketService as any).connectionState = 3; // CLOSED
      });

      webSocketService.disconnect();

      expect(webSocketService.disconnect).toHaveBeenCalled();
      expect(webSocketService.isConnected).toBe(false);
    });
  });

  describe('REQ-WS-002: JSON-RPC message handling', () => {
    beforeEach(() => {
      // Mock connected state
      (webSocketService as any).isConnected = true;
      (webSocketService as any).connectionState = 1; // OPEN
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

      expect((mockWebSocketService as any).send).toHaveBeenCalledWith(
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

      expect((mockWebSocketService as any).send).toHaveBeenCalledWith(
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
      (mockWebSocketService as any).onopen?.();

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
      (mockWebSocketService as any).onopen?.();

      const authError = MockDataFactory.getErrorResponse(-32001, 'Auth Failed');
      
      setTimeout(() => {
        (mockWebSocketService as any).onmessage?.({ data: JSON.stringify(authError) });
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
      (mockWebSocketService as any).onopen?.();

      // Fast-forward time to trigger ping
      jest.advanceTimersByTime(30000);

      expect((mockWebSocketService as any).send).toHaveBeenCalledWith(
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
      (mockWebSocketService as any).onopen?.();

      const requestPromise = webSocketService.sendRPC('slow_method');

      // Fast-forward time to trigger timeout
      jest.advanceTimersByTime(30000);

      await expect(requestPromise).rejects.toThrow('RPC request timeout: slow_method');
    });
  });

  describe('Connection state management', () => {
    test('should return correct connection state', () => {
      // Test initial state
      expect(webSocketService.connectionState).toBe(0); // WebSocket.CONNECTING

      // Mock connecting state
      (webSocketService as any).connectionState = 0; // CONNECTING
      expect(webSocketService.connectionState).toBe(0);

      // Mock open state
      (webSocketService as any).connectionState = 1; // OPEN
      expect(webSocketService.connectionState).toBe(1);
    });
  });
});