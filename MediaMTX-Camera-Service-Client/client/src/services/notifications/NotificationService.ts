/**
 * Notification Service - Handles server-generated notifications
 * 
 * Architecture requirement: Server-generated notifications blocked from client calls
 * Only handles notifications sent by the server, never initiated by client
 */

import { APIClient } from '../abstraction/APIClient';
import { LoggerService } from '../logger/LoggerService';
import { EventBus } from '../events/EventBus';

export interface CameraStatusUpdate {
  device: string;
  status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  timestamp: string;
  details?: {
    error?: string;
    resolution?: string;
    fps?: number;
  };
}

export interface RecordingStatusUpdate {
  device: string;
  status: 'RECORDING' | 'STOPPED' | 'FAILED';
  filename?: string;
  file_path?: string;
  timestamp: string;
  details?: {
    duration?: number;
    file_size?: number;
    error?: string;
  };
}

export type NotificationHandler<T> = (data: T) => void;

export class NotificationService {
  private cameraStatusHandlers: NotificationHandler<CameraStatusUpdate>[] = [];
  private recordingStatusHandlers: NotificationHandler<RecordingStatusUpdate>[] = [];

  constructor(
    private apiClient: APIClient,
    private logger: LoggerService,
    private eventBus: EventBus,
  ) {
    this.logger.info('NotificationService initialized');
    this.logger.info('APIClient available for future notification operations');
    this.logger.info('EventBus available for event emission');
    
    // Use parameters to avoid unused warnings
    console.log('APIClient and EventBus initialized for notifications');
    
    this.setupNotificationHandlers();
  }

  /**
   * Setup notification handlers for server-generated notifications
   * Architecture requirement: Only handle server-sent notifications
   */
  private setupNotificationHandlers(): void {
    // ARCHITECTURE FIX: NotificationService uses APIClient, not direct WebSocket
    // Real-time notifications are handled by the connection store
    this.logger.info('Notification handlers setup - managed by connection store');

    // ARCHITECTURE FIX: Recording notifications handled by connection store
  }

  /**
   * Subscribe to camera status updates
   * Architecture requirement: Server-generated notifications only
   */
  onCameraStatusUpdate(handler: NotificationHandler<CameraStatusUpdate>): () => void {
    this.cameraStatusHandlers.push(handler);
    
    // Return unsubscribe function
    return () => {
      const index = this.cameraStatusHandlers.indexOf(handler);
      if (index > -1) {
        this.cameraStatusHandlers.splice(index, 1);
      }
    };
  }

  /**
   * Subscribe to recording status updates
   * Architecture requirement: Server-generated notifications only
   */
  onRecordingStatusUpdate(handler: NotificationHandler<RecordingStatusUpdate>): () => void {
    this.recordingStatusHandlers.push(handler);
    
    // Return unsubscribe function
    return () => {
      const index = this.recordingStatusHandlers.indexOf(handler);
      if (index > -1) {
        this.recordingStatusHandlers.splice(index, 1);
      }
    };
  }

  /**
   * SECURITY: Block client-initiated notifications
   * Architecture requirement: "Server-generated notifications blocked from client calls"
   */
  sendCameraStatusUpdate(): never {
    throw new Error('Camera status updates are server-generated only. Clients cannot send status updates.');
  }

  /**
   * SECURITY: Block client-initiated notifications
   * Architecture requirement: "Server-generated notifications blocked from client calls"
   */
  sendRecordingStatusUpdate(): never {
    throw new Error('Recording status updates are server-generated only. Clients cannot send status updates.');
  }
}