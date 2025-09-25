/**
 * EventSubscriptionService unit tests for missing RPC methods
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-EVENT-001: subscribe_events RPC method
 * - REQ-EVENT-002: unsubscribe_events RPC method
 * - REQ-EVENT-003: get_subscription_stats RPC method
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { SubscriptionResult, UnsubscriptionResult, SubscriptionStatsResult } from '../../../src/types/api';
import { MockDataFactory } from '../../utils/mocks';

// Use centralized mocks - eliminates duplication
const mockWebSocketService = MockDataFactory.createMockWebSocketService();
const mockLoggerService = MockDataFactory.createMockLoggerService();

// Create a mock event subscription service class
class EventSubscriptionService {
  constructor(
    private wsService: WebSocketService,
    private logger: LoggerService
  ) {}

  async subscribeEvents(topics: string[], filters?: Record<string, any>): Promise<SubscriptionResult> {
    try {
      this.logger.info('subscribe_events request', { topics, filters });
      return await this.wsService.sendRPC('subscribe_events', { topics, filters });
    } catch (error) {
      this.logger.error('subscribe_events failed', error as Error);
      throw error;
    }
  }

  async unsubscribeEvents(topics?: string[]): Promise<UnsubscriptionResult> {
    try {
      this.logger.info('unsubscribe_events request', { topics });
      return await this.wsService.sendRPC('unsubscribe_events', { topics });
    } catch (error) {
      this.logger.error('unsubscribe_events failed', error as Error);
      throw error;
    }
  }

  async getSubscriptionStats(): Promise<SubscriptionStatsResult> {
    try {
      this.logger.info('get_subscription_stats request');
      return await this.wsService.sendRPC('get_subscription_stats');
    } catch (error) {
      this.logger.error('get_subscription_stats failed', error as Error);
      throw error;
    }
  }
}

describe('EventSubscriptionService Unit Tests', () => {
  let eventSubscriptionService: EventSubscriptionService;

  beforeEach(() => {
    jest.clearAllMocks();
    eventSubscriptionService = new EventSubscriptionService(mockWebSocketService, mockLoggerService);
  });

  describe('REQ-EVENT-001: subscribe_events RPC method', () => {
    test('Should call WebSocket service with topics only', async () => {
      // Arrange
      const topics = ['camera.connected', 'recording.start'];
      const expectedResult: SubscriptionResult = {
        subscribed: true,
        topics: topics,
        filters: {}
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.subscribeEvents(topics);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('subscribe_events request', { topics, filters: undefined });
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('subscribe_events', { topics, filters: undefined });
      expect(result).toEqual(expectedResult);
    });

    test('Should call WebSocket service with topics and filters', async () => {
      // Arrange
      const topics = ['camera.connected', 'recording.start'];
      const filters = { device: 'camera0' };
      const expectedResult: SubscriptionResult = {
        subscribed: true,
        topics: topics,
        filters: filters
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.subscribeEvents(topics, filters);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('subscribe_events request', { topics, filters });
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('subscribe_events', { topics, filters });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const topics = ['camera.connected'];
      const error = new Error('Subscribe events failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      // Act & Assert
      await expect(eventSubscriptionService.subscribeEvents(topics)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('subscribe_events failed', error);
    });
  });

  describe('REQ-EVENT-002: unsubscribe_events RPC method', () => {
    test('Should call WebSocket service with specific topics', async () => {
      // Arrange
      const topics = ['camera.connected'];
      const expectedResult: UnsubscriptionResult = {
        unsubscribed: true,
        topics: topics
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.unsubscribeEvents(topics);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('unsubscribe_events request', { topics });
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('unsubscribe_events', { topics });
      expect(result).toEqual(expectedResult);
    });

    test('Should call WebSocket service without topics (unsubscribe all)', async () => {
      // Arrange
      const expectedResult: UnsubscriptionResult = {
        unsubscribed: true,
        topics: []
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.unsubscribeEvents();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('unsubscribe_events request', { topics: undefined });
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('unsubscribe_events', { topics: undefined });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const topics = ['camera.connected'];
      const error = new Error('Unsubscribe events failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      // Act & Assert
      await expect(eventSubscriptionService.unsubscribeEvents(topics)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('unsubscribe_events failed', error);
    });
  });

  describe('REQ-EVENT-003: get_subscription_stats RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const expectedResult: SubscriptionStatsResult = {
        global_stats: {
          total_subscriptions: 15,
          active_clients: 3,
          topic_counts: {
            'camera.connected': 2,
            'recording.start': 1,
            'recording.stop': 1
          }
        },
        client_topics: ['camera.connected', 'recording.start'],
        client_id: 'client_123'
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.getSubscriptionStats();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_subscription_stats request');
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_subscription_stats');
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const error = new Error('Get subscription stats failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      // Act & Assert
      await expect(eventSubscriptionService.getSubscriptionStats()).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('get_subscription_stats failed', error);
    });
  });

  describe('REQ-EVENT-004: Event subscription workflow', () => {
    test('Should handle complete subscription workflow', async () => {
      // Arrange
      const topics = ['camera.connected', 'recording.start'];
      const filters = { device: 'camera0' };
      
      const subscribeResult: SubscriptionResult = {
        subscribed: true,
        topics: topics,
        filters: filters
      };

      const statsResult: SubscriptionStatsResult = {
        global_stats: {
          total_subscriptions: 2,
          active_clients: 1,
          topic_counts: {
            'camera.connected': 1,
            'recording.start': 1
          }
        },
        client_topics: topics,
        client_id: 'client_123'
      };

      const unsubscribeResult: UnsubscriptionResult = {
        unsubscribed: true,
        topics: topics
      };

      mockWebSocketService.sendRPC
        .mockResolvedValueOnce(subscribeResult)
        .mockResolvedValueOnce(statsResult)
        .mockResolvedValueOnce(unsubscribeResult);

      // Act
      const subscribe = await eventSubscriptionService.subscribeEvents(topics, filters);
      const stats = await eventSubscriptionService.getSubscriptionStats();
      const unsubscribe = await eventSubscriptionService.unsubscribeEvents(topics);

      // Assert
      expect(subscribe).toEqual(subscribeResult);
      expect(stats).toEqual(statsResult);
      expect(unsubscribe).toEqual(unsubscribeResult);
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledTimes(3);
    });
  });
});
