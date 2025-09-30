/**
 * RealTimeNotificationHandler - Architecture Compliance
 * 
 * Architecture requirement: "RealTimeNotificationHandler service" (Section 5.2)
 * Implements real-time notification routing to store handlers
 */

import { useDeviceStore } from '../../stores/device/deviceStore';
import { useRecordingStore } from '../../stores/recording/recordingStore';
import { useServerStore } from '../../stores/server/serverStore';
import { Camera, SystemStatus, SystemReadinessStatus } from '../../types/api';
import { RecordingSessionInfo } from '../../stores/recording/recordingStore';

export class RealTimeNotificationHandler {
  /**
   * Validate camera status update notification
   * Leverage existing Camera interface for type safety
   */
  private validateCameraStatusUpdate(params: any): Camera {
    if (!params || typeof params.device !== 'string') {
      throw new Error('Invalid camera status update: missing device');
    }
    if (!['CONNECTED', 'DISCONNECTED', 'ERROR'].includes(params.status)) {
      throw new Error('Invalid camera status update: invalid status');
    }
    return params as Camera;
  }

  /**
   * Validate recording status update notification
   * Leverage existing RecordingSessionInfo interface for type safety
   */
  private validateRecordingStatusUpdate(params: any): RecordingSessionInfo {
    if (!params || typeof params.device !== 'string') {
      throw new Error('Invalid recording status update: missing device');
    }
    if (!['RECORDING', 'STOPPED', 'ERROR', 'STARTING'].includes(params.status)) {
      throw new Error('Invalid recording status update: invalid status');
    }
    return params as RecordingSessionInfo;
  }

  /**
   * Validate system health update notification
   * Leverage existing SystemStatus interface for type safety
   */
  private validateSystemHealthUpdate(params: any): SystemStatus | SystemReadinessStatus {
    if (!params) {
      throw new Error('Invalid system health update: missing parameters');
    }
    // Accept both SystemStatus and SystemReadinessStatus
    return params as SystemStatus | SystemReadinessStatus;
  }

  /**
   * Handle camera status update notifications
   * Architecture requirement: Route notifications to appropriate store handlers
   */
  handleCameraStatusUpdate(camera: any): void {
    console.log('RealTimeNotificationHandler: Processing camera_status_update', camera);
    
    try {
      // Validate notification parameters using existing Camera interface
      const validatedCamera = this.validateCameraStatusUpdate(camera);
      
      // Route to device store handler
      useDeviceStore.getState().handleCameraStatusUpdate(validatedCamera);
    } catch (error) {
      console.error('Error processing camera status update:', error);
      // Leverage existing store error handling pattern
      useDeviceStore.getState().setError(
        error instanceof Error ? error.message : 'Failed to process camera status update'
      );
    }
  }

  /**
   * Handle recording status update notifications
   * Architecture requirement: Route notifications to appropriate store handlers
   */
  handleRecordingStatusUpdate(recording: any): void {
    console.log('RealTimeNotificationHandler: Processing recording_status_update', recording);
    
    try {
      // Validate notification parameters using existing RecordingSessionInfo interface
      const validatedRecording = this.validateRecordingStatusUpdate(recording);
      
      // Route to recording store handler
      useRecordingStore.getState().handleRecordingStatusUpdate(validatedRecording);
    } catch (error) {
      console.error('Error processing recording status update:', error);
      // Leverage existing store error handling pattern
      useRecordingStore.getState().setError(
        error instanceof Error ? error.message : 'Failed to process recording status update'
      );
    }
  }

  /**
   * Handle system health update notifications
   * Architecture requirement: Route notifications to appropriate store handlers
   */
  handleSystemHealthUpdate(status: any): void {
    console.log('RealTimeNotificationHandler: Processing system_health_update', status);
    
    try {
      // Validate notification parameters using existing SystemStatus/SystemReadinessStatus interfaces
      const validatedStatus = this.validateSystemHealthUpdate(status);
      
      // Route to server store handler
      useServerStore.getState().handleSystemStatusUpdate(validatedStatus);
    } catch (error) {
      console.error('Error processing system health update:', error);
      // Leverage existing store error handling pattern
      useServerStore.getState().setError(
        error instanceof Error ? error.message : 'Failed to process system health update'
      );
    }
  }
}
