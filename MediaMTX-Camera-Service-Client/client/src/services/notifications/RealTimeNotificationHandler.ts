/**
 * RealTimeNotificationHandler - Architecture Compliance
 * 
 * Architecture requirement: "RealTimeNotificationHandler service" (Section 5.2)
 * Implements real-time notification routing to store handlers
 */

import { useDeviceStore } from '../../stores/device/deviceStore';
import { useRecordingStore } from '../../stores/recording/recordingStore';
import { useServerStore } from '../../stores/server/serverStore';
import { Camera, RecordingInfo } from '../../types/api';

export class RealTimeNotificationHandler {
  /**
   * Handle camera status update notifications
   * Architecture requirement: Route notifications to appropriate store handlers
   */
  handleCameraStatusUpdate(camera: Camera): void {
    console.log('RealTimeNotificationHandler: Processing camera_status_update', camera);
    
    // Route to device store handler
    useDeviceStore.getState().handleCameraStatusUpdate(camera);
  }

  /**
   * Handle recording status update notifications
   * Architecture requirement: Route notifications to appropriate store handlers
   */
  handleRecordingStatusUpdate(recording: RecordingInfo): void {
    console.log('RealTimeNotificationHandler: Processing recording_status_update', recording);
    
    // Route to recording store handler
    useRecordingStore.getState().handleRecordingStatusUpdate(recording);
  }

  /**
   * Handle system health update notifications
   * Architecture requirement: Route notifications to appropriate store handlers
   */
  handleSystemHealthUpdate(status: any): void {
    console.log('RealTimeNotificationHandler: Processing system_health_update', status);
    
    // Route to server store handler
    useServerStore.getState().handleSystemStatusUpdate(status);
  }
}
