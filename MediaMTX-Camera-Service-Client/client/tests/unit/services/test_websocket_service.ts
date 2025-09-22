/**
 * WebSocket Service Unit Tests
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Requirements Coverage:
 * - REQ-CONN01-001: WebSocket connection must be stable and reliable
 * - REQ-CONN01-002: JSON-RPC 2.0 protocol compliance must be enforced
 * - REQ-CONN01-003: Connection recovery must handle network interruptions
 * - REQ-CONN01-004: Message queuing must work during disconnections
 * 
 * Coverage: UNIT
 * Quality: HIGH
 */

import { websocketService } from '../../../src/services/websocket';
import { RPC_METHODS, ERROR_CODES } from '../../../src/types/rpc';

// Mock WebSocket
const mockWebSocket = {
  readyState: WebSocket.CONNECTING,
  send: jest.fn(),
  close: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  CONNECTING: WebSocket.CONNECTING,
  OPEN: WebSocket.OPEN,
  CLOSING: WebSocket.CLOSING,
  CLOSED: WebSocket.CLOSED,
};

// Mock global WebSocket
(global as any).WebSocket = jest.fn(() => mockWebSocket);

// Mock logger service
jest.mock('../../../src/services/loggerService', () => ({
  logger: {
    info: jest.fn(),
    error: jest.fn(),
    warn: jest.fn(),
    debug: jest.fn(),
  },
  loggers: {
    websocket: {
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
    },
  },
}));

describe('WebSocket Service', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockWebSocket.readyState = WebSocket.CONNECTING;
  });

  afterEach(() => {
    websocketService.disconnect();
  });

  describe('REQ-CONN01-001: WebSocket Connection Stability', () => {
    it('should connect to WebSocket server', async () => {
      const connectPromise = websocketService.connect('ws://localhost:8002/ws');
      
      // Simulate connection success
      mockWebSocket.readyState = WebSocket.OPEN;
      const openEvent = new Event('open');
      mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'open')?.[1](openEvent);

      await connectPromise;

      expect(websocketService.isConnected()).toBe(true);
      expect(websocketService.getConnectionState()).toBe('connected');
    });

    it('should handle connection failures', async () => {
      const connectPromise = websocketService.connect('ws://invalid-host:8002/ws');
      
      // Simulate connection failure
      const errorEvent = new Event('error');
      mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'error')?.[1](errorEvent);

      await expect(connectPromise).rejects.toThrow();
      expect(websocketService.isConnected()).toBe(false);
    });

    it('should maintain connection state accurately', () => {
      expect(websocketService.isConnected()).toBe(false);
      expect(websocketService.getConnectionState()).toBe('disconnected');

      // Simulate connection
      mockWebSocket.readyState = WebSocket.OPEN;
      websocketService.connect('ws://localhost:8002/ws');

      expect(websocketService.isConnected()).toBe(true);
      expect(websocketService.getConnectionState()).toBe('connected');
    });
  });

  describe('REQ-CONN01-002: JSON-RPC 2.0 Protocol Compliance', () => {
    beforeEach(async () => {
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');
    });

    it('should send properly formatted JSON-RPC requests', async () => {
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: 'pong',
        id: 1,
      };

      // Mock successful response
      const responseHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'message')?.[1];
      
      const callPromise = websocketService.call(RPC_METHODS.PING, {});
      
      // Simulate response
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify(mockResponse),
      });
      responseHandler?.(messageEvent);

      const result = await callPromise;

      expect(mockWebSocket.send).toHaveBeenCalledWith(
        expect.stringContaining('"jsonrpc":"2.0"')
      );
      expect(mockWebSocket.send).toHaveBeenCalledWith(
        expect.stringContaining('"method":"ping"')
      );
      expect(result).toBe('pong');
    });

    it('should handle JSON-RPC errors correctly', async () => {
      const mockError = {
        jsonrpc: '2.0' as const,
        error: {
          code: ERROR_CODES.METHOD_NOT_FOUND,
          message: 'Method not found',
        },
        id: 1,
      };

      const responseHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'message')?.[1];
      
      const callPromise = websocketService.call('invalid_method', {});
      
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify(mockError),
      });
      responseHandler?.(messageEvent);

      await expect(callPromise).rejects.toMatchObject({
        code: ERROR_CODES.METHOD_NOT_FOUND,
        message: 'Method not found',
      });
    });

    it('should validate JSON-RPC response format', async () => {
      const invalidResponse = {
        // Missing jsonrpc field
        result: 'pong',
        id: 1,
      };

      const responseHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'message')?.[1];
      
      const callPromise = websocketService.call(RPC_METHODS.PING, {});
      
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify(invalidResponse),
      });
      responseHandler?.(messageEvent);

      await expect(callPromise).rejects.toThrow('Invalid JSON-RPC response');
    });
  });

  describe('REQ-CONN01-003: Connection Recovery', () => {
    it('should automatically reconnect on connection loss', async () => {
      // Initial connection
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      expect(websocketService.isConnected()).toBe(true);

      // Simulate connection loss
      const closeEvent = new CloseEvent('close', { code: 1006 });
      const closeHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'close')?.[1];
      closeHandler?.(closeEvent);

      // Should attempt reconnection
      expect(websocketService.getConnectionState()).toBe('reconnecting');
    });

    it('should handle reconnection failures gracefully', async () => {
      // Initial connection
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      // Simulate connection loss
      const closeEvent = new CloseEvent('close', { code: 1006 });
      const closeHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'close')?.[1];
      closeHandler?.(closeEvent);

      // Simulate reconnection failure
      const errorEvent = new Event('error');
      const errorHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'error')?.[1];
      errorHandler?.(errorEvent);

      expect(websocketService.getConnectionState()).toBe('disconnected');
    });

    it('should respect reconnection limits', async () => {
      // Configure low reconnection limit
      websocketService.configure({ maxReconnectAttempts: 2 });

      // Initial connection
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      // Simulate multiple connection losses
      const closeEvent = new CloseEvent('close', { code: 1006 });
      const closeHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'close')?.[1];

      // First reconnection attempt
      closeHandler?.(closeEvent);
      expect(websocketService.getConnectionState()).toBe('reconnecting');

      // Second reconnection attempt
      closeHandler?.(closeEvent);
      expect(websocketService.getConnectionState()).toBe('reconnecting');

      // Third attempt should stop reconnecting
      closeHandler?.(closeEvent);
      expect(websocketService.getConnectionState()).toBe('disconnected');
    });
  });

  describe('REQ-CONN01-004: Message Queuing', () => {
    it('should queue messages during disconnection', async () => {
      // Start disconnected
      expect(websocketService.isConnected()).toBe(false);

      // Queue a message
      const callPromise = websocketService.call(RPC_METHODS.PING, {});
      
      // Connect and send queued message
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: 'pong',
        id: 1,
      };

      const responseHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'message')?.[1];
      
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify(mockResponse),
      });
      responseHandler?.(messageEvent);

      const result = await callPromise;
      expect(result).toBe('pong');
    });

    it('should flush queued messages on reconnection', async () => {
      // Initial connection
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      // Disconnect
      websocketService.disconnect();

      // Queue multiple messages
      const call1 = websocketService.call(RPC_METHODS.PING, {});
      const call2 = websocketService.call(RPC_METHODS.GET_CAMERAS, {});

      // Reconnect
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      // Both calls should be sent
      expect(mockWebSocket.send).toHaveBeenCalledTimes(2);
    });

    it('should handle queue overflow', async () => {
      // Configure small queue size
      websocketService.configure({ maxQueueSize: 2 });

      // Start disconnected
      expect(websocketService.isConnected()).toBe(false);

      // Queue more messages than queue size
      const call1 = websocketService.call(RPC_METHODS.PING, {});
      const call2 = websocketService.call(RPC_METHODS.GET_CAMERAS, {});
      const call3 = websocketService.call(RPC_METHODS.GET_STATUS, {});

      // Third call should be rejected due to queue overflow
      await expect(call3).rejects.toThrow('Message queue overflow');
    });
  });

  describe('Event Handling', () => {
    it('should emit connection events', async () => {
      const onConnect = jest.fn();
      const onDisconnect = jest.fn();

      websocketService.on('connect', onConnect);
      websocketService.on('disconnect', onDisconnect);

      // Connect
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      expect(onConnect).toHaveBeenCalled();

      // Disconnect
      websocketService.disconnect();
      expect(onDisconnect).toHaveBeenCalled();
    });

    it('should emit error events', async () => {
      const onError = jest.fn();
      websocketService.on('error', onError);

      // Simulate error
      const errorEvent = new Event('error');
      const errorHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'error')?.[1];
      errorHandler?.(errorEvent);

      expect(onError).toHaveBeenCalled();
    });

    it('should emit message events', async () => {
      const onMessage = jest.fn();
      websocketService.on('message', onMessage);

      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify({ test: 'data' }),
      });
      const messageHandler = mockWebSocket.addEventListener.mock.calls
        .find(call => call[0] === 'message')?.[1];
      messageHandler?.(messageEvent);

      expect(onMessage).toHaveBeenCalled();
    });
  });

  describe('Configuration', () => {
    it('should apply configuration settings', () => {
      const config = {
        maxReconnectAttempts: 5,
        reconnectInterval: 2000,
        maxQueueSize: 100,
        pingInterval: 30000,
      };

      websocketService.configure(config);

      expect(websocketService.getConfig()).toMatchObject(config);
    });

    it('should validate configuration values', () => {
      const invalidConfig = {
        maxReconnectAttempts: -1,
        reconnectInterval: 0,
        maxQueueSize: -1,
      };

      expect(() => {
        websocketService.configure(invalidConfig);
      }).toThrow('Invalid configuration');
    });
  });

  describe('Cleanup', () => {
    it('should cleanup resources on disconnect', async () => {
      mockWebSocket.readyState = WebSocket.OPEN;
      await websocketService.connect('ws://localhost:8002/ws');

      websocketService.disconnect();

      expect(mockWebSocket.close).toHaveBeenCalled();
      expect(websocketService.isConnected()).toBe(false);
    });

    it('should clear event listeners on cleanup', () => {
      const onConnect = jest.fn();
      websocketService.on('connect', onConnect);

      websocketService.cleanup();

      // Event listeners should be cleared
      expect(mockWebSocket.removeEventListener).toHaveBeenCalled();
    });
  });
});
