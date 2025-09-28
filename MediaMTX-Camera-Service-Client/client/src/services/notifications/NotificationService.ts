/**
 * Notification Service - Handles server-generated notifications
 * 
 * Architecture requirement: Server-generated notifications blocked from client calls
 * Only handles notifications sent by the server, never initiated by client
 */

import { WebSocketService } from '../websocket/WebSocketService';
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
    private wsService: WebSocketService,
    private logger: LoggerService,
    private eventBus: EventBus,
  ) {
    this.setupNotificationHandlers();
  }

  /**
   * Setup notification handlers for server-generated notifications
   * Architecture requirement: Only handle server-sent notifications
   */
  private setupNotificationHandlers(): void {
    this.wsService.onNotification('camera_status_update', (data: CameraStatusUpdate) => {
      this.logger.info('Received camera status update from server', data);
      
      // Emit event through event bus for real-time updates
      this.eventBus.emitWithTimestamp('camera_status_update', data);
      
      this.cameraStatusHandlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          this.logger.error('Error in camera status handler', error as Record<string, unknown>);
        }
      });
    });

    this.wsService.onNotification('recording_status_update', (data: RecordingStatusUpdate) => {
      this.logger.info('Received recording status update from server', data);
      
      // Emit event through event bus for real-time updates
      this.eventBus.emitWithTimestamp('recording_status_update', data);
      
      this.recordingStatusHandlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          this.logger.error('Error in recording status handler', error as Record<string, unknown>);
        }
      });
    });
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