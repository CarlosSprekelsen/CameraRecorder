/**
 * Connection Service Unit Tests
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Requirements Coverage:
 * - REQ-CONN02-001: Connection state management must be accurate
 * - REQ-CONN02-002: Connection health monitoring must be reliable
 * - REQ-CONN02-003: Connection recovery must be automatic and robust
 * - REQ-CONN02-004: Connection metrics must be tracked accurately
 * 
 * Coverage: UNIT
 * Quality: HIGH
 */

import { ConnectionService } from '../../../src/services/connectionService';
import { websocketService } from '../../../src/services/websocket';

// Mock dependencies
jest.mock('../../../src/services/websocket', () => ({
  websocketService: {
    connect: jest.fn(),
    disconnect: jest.fn(),
    isConnected: jest.fn(),
    getConnectionState: jest.fn(),
    on: jest.fn(),
    off: jest.fn(),
  },
}));

jest.mock('../../../src/services/loggerService', () => ({
  logger: {
    info: jest.fn(),
    error: jest.fn(),
    warn: jest.fn(),
    debug: jest.fn(),
  },
  loggers: {
    connection: {
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
    },
  },
}));

describe('Connection Service', () => {
  let connectionService: ConnectionService;
  const mockWebSocketService = websocketService as jest.Mocked<typeof websocketService>;

  beforeEach(() => {
    connectionService = new ConnectionService();
    jest.clearAllMocks();
  });

  afterEach(() => {
    connectionService.cleanup();
  });

  describe('REQ-CONN02-001: Connection State Management', () => {
    it('should initialize with disconnected state', () => {
      expect(connectionService.isConnected()).toBe(false);
      expect(connectionService.getConnectionState()).toBe('disconnected');
      expect(connectionService.getConnectionUrl()).toBeNull();
    });

    it('should connect to WebSocket server', async () => {
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);
      mockWebSocketService.getConnectionState.mockReturnValue('connected');

      await connectionService.connect('ws://localhost:8002/ws');

      expect(mockWebSocketService.connect).toHaveBeenCalledWith('ws://localhost:8002/ws');
      expect(connectionService.isConnected()).toBe(true);
      expect(connectionService.getConnectionState()).toBe('connected');
      expect(connectionService.getConnectionUrl()).toBe('ws://localhost:8002/ws');
    });

    it('should disconnect from WebSocket server', async () => {
      // First connect
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);
      await connectionService.connect('ws://localhost:8002/ws');

      // Then disconnect
      mockWebSocketService.disconnect.mockImplementation(() => {});
      mockWebSocketService.isConnected.mockReturnValue(false);
      mockWebSocketService.getConnectionState.mockReturnValue('disconnected');

      connectionService.disconnect();

      expect(mockWebSocketService.disconnect).toHaveBeenCalled();
      expect(connectionService.isConnected()).toBe(false);
      expect(connectionService.getConnectionState()).toBe('disconnected');
    });

    it('should track connection attempts', async () => {
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);

      await connectionService.connect('ws://localhost:8002/ws');

      expect(connectionService.getConnectionAttempts()).toBe(1);
      expect(connectionService.getLastConnectionAttempt()).toBeDefined();
    });

    it('should track connection duration', async () => {
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);

      await connectionService.connect('ws://localhost:8002/ws');

      const duration = connectionService.getConnectionDuration();
      expect(duration).toBeGreaterThanOrEqual(0);
    });
  });

  describe('REQ-CONN02-002: Connection Health Monitoring', () => {
    beforeEach(async () => {
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);
      await connectionService.connect('ws://localhost:8002/ws');
    });

    it('should monitor connection health', () => {
      const health = connectionService.getConnectionHealth();
      
      expect(health).toHaveProperty('isHealthy');
      expect(health).toHaveProperty('lastHeartbeat');
      expect(health).toHaveProperty('latency');
      expect(health).toHaveProperty('packetLoss');
    });

    it('should detect unhealthy connections', () => {
      // Simulate unhealthy connection
      connectionService.updateHeartbeat(new Date(Date.now() - 60000)); // 1 minute ago

      const health = connectionService.getConnectionHealth();
      expect(health.isHealthy).toBe(false);
    });

    it('should track connection latency', () => {
      const startTime = Date.now();
      connectionService.updateLatency(50); // 50ms latency

      const health = connectionService.getConnectionHealth();
      expect(health.latency).toBe(50);
    });

    it('should track packet loss', () => {
      connectionService.updatePacketLoss(0.05); // 5% packet loss

      const health = connectionService.getConnectionHealth();
      expect(health.packetLoss).toBe(0.05);
    });

    it('should emit health status changes', () => {
      const onHealthChange = jest.fn();
      connectionService.on('healthChange', onHealthChange);

      connectionService.updateHeartbeat(new Date());
      connectionService.updateLatency(100);

      expect(onHealthChange).toHaveBeenCalled();
    });
  });

  describe('REQ-CONN02-003: Connection Recovery', () => {
    it('should automatically reconnect on connection loss', async () => {
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);

      await connectionService.connect('ws://localhost:8002/ws');

      // Simulate connection loss
      mockWebSocketService.isConnected.mockReturnValue(false);
      mockWebSocketService.getConnectionState.mockReturnValue('disconnected');

      // Trigger reconnection
      connectionService.handleConnectionLoss();

      expect(connectionService.getConnectionState()).toBe('reconnecting');
    });

    it('should respect reconnection limits', async () => {
      connectionService.configure({ maxReconnectAttempts: 2 });

      mockWebSocketService.connect.mockRejectedValue(new Error('Connection failed'));

      // Attempt multiple reconnections
      await connectionService.connect('ws://localhost:8002/ws');
      connectionService.handleConnectionLoss();
      connectionService.handleConnectionLoss();
      connectionService.handleConnectionLoss();

      expect(connectionService.getConnectionState()).toBe('disconnected');
      expect(connectionService.getReconnectionAttempts()).toBe(2);
    });

    it('should implement exponential backoff', async () => {
      connectionService.configure({ 
        reconnectInterval: 1000,
        maxReconnectInterval: 10000,
        reconnectMultiplier: 2,
      });

      mockWebSocketService.connect.mockRejectedValue(new Error('Connection failed'));

      await connectionService.connect('ws://localhost:8002/ws');
      
      const firstAttempt = connectionService.getNextReconnectTime();
      connectionService.handleConnectionLoss();
      
      const secondAttempt = connectionService.getNextReconnectTime();
      
      expect(secondAttempt - firstAttempt).toBeGreaterThan(1000);
    });

    it('should reset reconnection attempts on successful connection', async () => {
      connectionService.configure({ maxReconnectAttempts: 3 });

      // First connection fails
      mockWebSocketService.connect.mockRejectedValue(new Error('Connection failed'));
      await connectionService.connect('ws://localhost:8002/ws');
      connectionService.handleConnectionLoss();

      expect(connectionService.getReconnectionAttempts()).toBe(1);

      // Second connection succeeds
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);
      await connectionService.connect('ws://localhost:8002/ws');

      expect(connectionService.getReconnectionAttempts()).toBe(0);
    });
  });

  describe('REQ-CONN02-004: Connection Metrics', () => {
    beforeEach(async () => {
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);
      await connectionService.connect('ws://localhost:8002/ws');
    });

    it('should track connection metrics', () => {
      const metrics = connectionService.getConnectionMetrics();

      expect(metrics).toHaveProperty('totalConnections');
      expect(metrics).toHaveProperty('totalDisconnections');
      expect(metrics).toHaveProperty('averageConnectionDuration');
      expect(metrics).toHaveProperty('reconnectionCount');
      expect(metrics).toHaveProperty('lastConnectionTime');
    });

    it('should update metrics on connection events', async () => {
      const initialMetrics = connectionService.getConnectionMetrics();
      const initialConnections = initialMetrics.totalConnections;

      // Disconnect and reconnect
      connectionService.disconnect();
      await connectionService.connect('ws://localhost:8002/ws');

      const updatedMetrics = connectionService.getConnectionMetrics();
      expect(updatedMetrics.totalConnections).toBe(initialConnections + 1);
    });

    it('should track average connection duration', async () => {
      // Connect for a short duration
      await new Promise(resolve => setTimeout(resolve, 100));
      connectionService.disconnect();

      const metrics = connectionService.getConnectionMetrics();
      expect(metrics.averageConnectionDuration).toBeGreaterThan(0);
    });

    it('should track reconnection count', async () => {
      const initialMetrics = connectionService.getConnectionMetrics();
      const initialReconnections = initialMetrics.reconnectionCount;

      // Simulate reconnection
      connectionService.handleConnectionLoss();

      const updatedMetrics = connectionService.getConnectionMetrics();
      expect(updatedMetrics.reconnectionCount).toBe(initialReconnections + 1);
    });

    it('should reset metrics when requested', () => {
      connectionService.resetMetrics();

      const metrics = connectionService.getConnectionMetrics();
      expect(metrics.totalConnections).toBe(0);
      expect(metrics.totalDisconnections).toBe(0);
      expect(metrics.reconnectionCount).toBe(0);
    });
  });

  describe('Event Handling', () => {
    it('should emit connection events', async () => {
      const onConnect = jest.fn();
      const onDisconnect = jest.fn();
      const onReconnecting = jest.fn();

      connectionService.on('connect', onConnect);
      connectionService.on('disconnect', onDisconnect);
      connectionService.on('reconnecting', onReconnecting);

      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);

      await connectionService.connect('ws://localhost:8002/ws');
      expect(onConnect).toHaveBeenCalled();

      connectionService.disconnect();
      expect(onDisconnect).toHaveBeenCalled();

      connectionService.handleConnectionLoss();
      expect(onReconnecting).toHaveBeenCalled();
    });

    it('should remove event listeners', () => {
      const onConnect = jest.fn();
      connectionService.on('connect', onConnect);
      connectionService.off('connect', onConnect);

      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);

      connectionService.connect('ws://localhost:8002/ws');
      expect(onConnect).not.toHaveBeenCalled();
    });
  });

  describe('Configuration', () => {
    it('should apply configuration settings', () => {
      const config = {
        maxReconnectAttempts: 5,
        reconnectInterval: 2000,
        healthCheckInterval: 30000,
        connectionTimeout: 10000,
      };

      connectionService.configure(config);

      expect(connectionService.getConfig()).toMatchObject(config);
    });

    it('should validate configuration values', () => {
      const invalidConfig = {
        maxReconnectAttempts: -1,
        reconnectInterval: 0,
        healthCheckInterval: -1,
      };

      expect(() => {
        connectionService.configure(invalidConfig);
      }).toThrow('Invalid configuration');
    });
  });

  describe('Cleanup', () => {
    it('should cleanup resources', async () => {
      mockWebSocketService.connect.mockResolvedValue(undefined);
      mockWebSocketService.isConnected.mockReturnValue(true);
      await connectionService.connect('ws://localhost:8002/ws');

      connectionService.cleanup();

      expect(mockWebSocketService.disconnect).toHaveBeenCalled();
      expect(mockWebSocketService.off).toHaveBeenCalled();
    });

    it('should clear all timers on cleanup', () => {
      const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout');
      const clearIntervalSpy = jest.spyOn(global, 'clearInterval');

      connectionService.cleanup();

      expect(clearTimeoutSpy).toHaveBeenCalled();
      expect(clearIntervalSpy).toHaveBeenCalled();

      clearTimeoutSpy.mockRestore();
      clearIntervalSpy.mockRestore();
    });
  });
});
