/**
 * Event Subscription Service
 * Manages real-time event subscriptions with Go server
 */

import { websocketService } from './websocket';
import { RPC_METHODS, EVENT_TOPICS } from '../types/rpc';
import type { 
  EventSubscriptionParams, 
  EventSubscriptionResponse,
  EventTopic 
} from '../types/rpc';
import { logger, loggers } from './loggerService';

export interface EventSubscription {
  topics: EventTopic[];
  filters?: Record<string, unknown>;
  callback: (event: any) => void;
}

export interface SubscriptionStats {
  global_stats: {
    total_subscriptions: number;
    active_clients: number;
    topic_counts: Record<string, number>;
  };
  client_topics: string[];
  client_id: string;
}

export class EventSubscriptionService {
  private static instance: EventSubscriptionService;
  private subscriptions: Map<string, EventSubscription> = new Map();
  private subscribedTopics: Set<EventTopic> = new Set();
  private isConnected: boolean = false;

  private constructor() {
    this.setupWebSocketHandlers();
  }

  static getInstance(): EventSubscriptionService {
    if (!EventSubscriptionService.instance) {
      EventSubscriptionService.instance = new EventSubscriptionService();
    }
    return EventSubscriptionService.instance;
  }

  /**
   * Subscribe to specific event topics
   */
  public async subscribe(
    topics: EventTopic[], 
    callback: (event: any) => void,
    filters?: Record<string, unknown>
  ): Promise<EventSubscriptionResponse> {
    loggers.service.start('EventSubscriptionService', 'subscribe', { topics, filters });

    try {
      // Store subscription locally
      const subscriptionId = this.generateSubscriptionId();
      this.subscriptions.set(subscriptionId, {
        topics,
        filters,
        callback
      });

      // Add topics to subscribed set
      topics.forEach(topic => this.subscribedTopics.add(topic));

      // Subscribe with server
      const params: EventSubscriptionParams = {
        topics,
        filters
      };

      const response = await websocketService.call(
        RPC_METHODS.SUBSCRIBE_EVENTS,
        params
      ) as EventSubscriptionResponse;

      this.isConnected = true;

      loggers.service.success('EventSubscriptionService', 'subscribe', {
        subscriptionId,
        topics: response.topics
      });

      return response;
    } catch (error) {
      loggers.service.error('EventSubscriptionService', 'subscribe', error as Error);
      throw error;
    }
  }

  /**
   * Unsubscribe from specific topics or all topics
   */
  public async unsubscribe(topics?: EventTopic[]): Promise<EventSubscriptionResponse> {
    loggers.service.start('EventSubscriptionService', 'unsubscribe', { topics });

    try {
      const params = topics ? { topics } : {};
      
      const response = await websocketService.call(
        RPC_METHODS.UNSUBSCRIBE_EVENTS,
        params
      ) as EventSubscriptionResponse;

      // Remove from local subscriptions
      if (topics) {
        topics.forEach(topic => this.subscribedTopics.delete(topic));
        
        // Remove subscriptions that no longer have any topics
        for (const [id, subscription] of this.subscriptions.entries()) {
          const remainingTopics = subscription.topics.filter(t => !topics.includes(t));
          if (remainingTopics.length === 0) {
            this.subscriptions.delete(id);
          } else {
            subscription.topics = remainingTopics;
          }
        }
      } else {
        // Unsubscribe from all
        this.subscriptions.clear();
        this.subscribedTopics.clear();
      }

      loggers.service.success('EventSubscriptionService', 'unsubscribe', {
        topics: response.topics || 'all'
      });

      return response;
    } catch (error) {
      loggers.service.error('EventSubscriptionService', 'unsubscribe', error as Error);
      throw error;
    }
  }

  /**
   * Get subscription statistics
   */
  public async getSubscriptionStats(): Promise<SubscriptionStats> {
    loggers.service.start('EventSubscriptionService', 'getSubscriptionStats');

    try {
      const response = await websocketService.call(
        RPC_METHODS.GET_SUBSCRIPTION_STATS,
        {}
      ) as SubscriptionStats;

      loggers.service.success('EventSubscriptionService', 'getSubscriptionStats', {
        totalSubscriptions: response.global_stats.total_subscriptions,
        activeClients: response.global_stats.active_clients
      });

      return response;
    } catch (error) {
      loggers.service.error('EventSubscriptionService', 'getSubscriptionStats', error as Error);
      throw error;
    }
  }

  /**
   * Get currently subscribed topics
   */
  public getSubscribedTopics(): EventTopic[] {
    return Array.from(this.subscribedTopics);
  }

  /**
   * Get active subscriptions count
   */
  public getActiveSubscriptionsCount(): number {
    return this.subscriptions.size;
  }

  /**
   * Check if subscribed to a specific topic
   */
  public isSubscribedTo(topic: EventTopic): boolean {
    return this.subscribedTopics.has(topic);
  }

  /**
   * Setup WebSocket message handlers
   */
  private setupWebSocketHandlers(): void {
    websocketService.onMessage((message) => {
      if ('method' in message) {
        this.handleEventNotification(message);
      }
    });
  }

  /**
   * Handle incoming event notifications
   */
  private handleEventNotification(notification: any): void {
    const topic = this.getTopicFromMethod(notification.method);
    if (!topic || !this.isSubscribedTo(topic)) {
      return;
    }

    // Find matching subscriptions and call their callbacks
    for (const subscription of this.subscriptions.values()) {
      if (subscription.topics.includes(topic)) {
        // Apply filters if specified
        if (subscription.filters && !this.matchesFilters(notification.params, subscription.filters)) {
          continue;
        }

        try {
          subscription.callback(notification.params);
        } catch (error) {
          loggers.service.error('EventSubscriptionService', 'handleEventNotification', error as Error, {
            topic,
            method: notification.method
          });
        }
      }
    }
  }

  /**
   * Map notification method to event topic
   */
  private getTopicFromMethod(method: string): EventTopic | null {
    const methodToTopic: Record<string, EventTopic> = {
      'camera_status_update': EVENT_TOPICS.CAMERA_STATUS_CHANGE,
      'recording_status_update': EVENT_TOPICS.RECORDING_START, // Could be start or stop
      'storage_status_update': EVENT_TOPICS.SYSTEM_HEALTH,
    };

    return methodToTopic[method] || null;
  }

  /**
   * Check if event matches subscription filters
   */
  private matchesFilters(eventParams: any, filters: Record<string, unknown>): boolean {
    for (const [key, value] of Object.entries(filters)) {
      if (eventParams[key] !== value) {
        return false;
      }
    }
    return true;
  }

  /**
   * Generate unique subscription ID
   */
  private generateSubscriptionId(): string {
    return `sub_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Cleanup all subscriptions
   */
  public cleanup(): void {
    loggers.service.info('EventSubscriptionService', 'cleanup');
    
    this.subscriptions.clear();
    this.subscribedTopics.clear();
    this.isConnected = false;
  }
}

// Export singleton instance
export const eventSubscriptionService = EventSubscriptionService.getInstance();

// Export types
export type { EventSubscription, SubscriptionStats };
