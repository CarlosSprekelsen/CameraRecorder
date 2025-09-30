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

import { ServerService } from '../../../src/services/server/ServerService';
import { IAPIClient } from '../../../src/services/abstraction/IAPIClient';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { SubscriptionResult, UnsubscriptionResult, SubscriptionStatsResult } from '../../../src/types/api';
import { MockDataFactory } from '../../utils/mocks';

// Use centralized mocks - aligned with refactored architecture
const mockAPIClient = MockDataFactory.createMockAPIClient();
const mockLoggerService = MockDataFactory.createMockLoggerService();

// Use real EventSubscriptionService with APIClient
// Note: EventSubscriptionService is part of ServerService in the real implementation

describe('EventSubscriptionService Unit Tests', () => {
  let eventSubscriptionService: ServerService;

  beforeEach(() => {
    jest.clearAllMocks();
    eventSubscriptionService = new ServerService(mockAPIClient, mockLoggerService);
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

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.subscribeEvents(topics);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('subscribe_events request', { topics, filters: undefined });
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('subscribe_events', { topics, filters: undefined });
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

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.subscribeEvents(topics, filters);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('subscribe_events request', { topics, filters });
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('subscribe_events', { topics, filters });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const topics = ['camera.connected'];
      const error = new Error('Subscribe events failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

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

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.unsubscribeEvents(topics);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('unsubscribe_events request', { topics });
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('unsubscribe_events', { topics });
      expect(result).toEqual(expectedResult);
    });

    test('Should call WebSocket service without topics (unsubscribe all)', async () => {
      // Arrange
      const expectedResult: UnsubscriptionResult = {
        unsubscribed: true,
        topics: []
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.unsubscribeEvents();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('unsubscribe_events request', { topics: undefined });
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('unsubscribe_events', { topics: undefined });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const topics = ['camera.connected'];
      const error = new Error('Unsubscribe events failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

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

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await eventSubscriptionService.getSubscriptionStats();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_subscription_stats request');
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('get_subscription_stats');
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const error = new Error('Get subscription stats failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

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

      (mockAPIClient.call as jest.Mock)
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
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledTimes(3);
    });
  });
});
