/**
 * Event Bus - Event-Driven Communication
 * 
 * Architecture requirement: "Event-driven communication patterns" (Section 2.3)
 * Centralized event system for real-time updates and component communication
 */

import { LoggerService } from '../logger/LoggerService';

export type EventHandler<T = any> = (data: T) => void;
export type EventUnsubscribe = () => void;

export interface EventBusEvent {
  type: string;
  data: any;
  timestamp: number;
}

export class EventBus {
  private handlers: Map<string, EventHandler[]> = new Map();
  private logger: LoggerService;

  constructor(logger: LoggerService) {
    this.logger = logger;
  }

  /**
   * Subscribe to an event type
   * Architecture requirement: Event-driven communication patterns
   */
  on<T = any>(eventType: string, handler: EventHandler<T>): EventUnsubscribe {
    if (!this.handlers.has(eventType)) {
      this.handlers.set(eventType, []);
    }

    this.handlers.get(eventType)!.push(handler);
    this.logger.info(`Subscribed to event: ${eventType}`);

    // Return unsubscribe function
    return () => {
      const handlers = this.handlers.get(eventType);
      if (handlers) {
        const index = handlers.indexOf(handler);
        if (index > -1) {
          handlers.splice(index, 1);
          this.logger.info(`Unsubscribed from event: ${eventType}`);
        }
      }
    };
  }

  /**
   * Emit an event
   * Architecture requirement: Event-driven communication patterns
   */
  emit<T = any>(eventType: string, data: T extends Record<string, unknown> ? T : Record<string, unknown>): void {
    const handlers = this.handlers.get(eventType);
    if (handlers && handlers.length > 0) {
      this.logger.info(`Emitting event: ${eventType}`, data);
      
      handlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          this.logger.error(`Error in event handler for ${eventType}:`, error as Record<string, unknown>);
        }
      });
    }
  }

  /**
   * Emit event with timestamp
   * Architecture requirement: Event-driven communication patterns
   */
  emitWithTimestamp<T = any>(eventType: string, data: T): void {
    const event: EventBusEvent = {
      type: eventType,
      data,
      timestamp: Date.now()
    };

    this.emit('event_bus', event as Record<string, unknown>);
    this.emit(eventType, data as Record<string, unknown>);
  }

  /**
   * Get all registered event types
   * Architecture requirement: Event system monitoring
   */
  getRegisteredEvents(): string[] {
    return Array.from(this.handlers.keys());
  }

  /**
   * Get handler count for an event type
   * Architecture requirement: Event system monitoring
   */
  getHandlerCount(eventType: string): number {
    const handlers = this.handlers.get(eventType);
    return handlers ? handlers.length : 0;
  }

  /**
   * Clear all handlers for an event type
   * Architecture requirement: Event system cleanup
   */
  clearHandlers(eventType: string): void {
    this.handlers.delete(eventType);
    this.logger.info(`Cleared handlers for event: ${eventType}`);
  }

  /**
   * Clear all event handlers
   * Architecture requirement: Event system cleanup
   */
  clearAllHandlers(): void {
    this.handlers.clear();
    this.logger.info('Cleared all event handlers');
  }
}
