/**
 * Device Actions - Unidirectional Data Flow
 * 
 * Architecture requirement: "Centralized state store with unidirectional data flow" (ADR-002)
 * Actions dispatch events to stores, stores update state, components react to state changes
 */

import { DeviceStore } from '../device/deviceStore';
import { APIClient } from '../../services/abstraction/APIClient';
import { LoggerService } from '../../services/logger/LoggerService';

export class DeviceActions {
  constructor(
    private deviceStore: DeviceStore,
    private apiClient: APIClient,
    private logger: LoggerService
  ) {}

  /**
   * Load camera list - follows unidirectional data flow
   * Event → Action → Reducer → State → View
   */
  async loadCameraList(): Promise<void> {
    try {
      this.logger.info('Loading camera list');
      
      // Dispatch loading state
      this.deviceStore.setLoading(true);
      
      // Execute API call through abstraction layer
      const response = await this.apiClient.call('get_camera_list', {});
      
      // Dispatch success state
      this.deviceStore.setCameras(response.cameras || []);
      this.deviceStore.setLoading(false);
      
      this.logger.info('Camera list loaded successfully');
    } catch (error) {
      this.logger.error('Failed to load camera list', error as Record<string, unknown>);
      
      // Dispatch error state
      this.deviceStore.setError(error as Error);
      this.deviceStore.setLoading(false);
    }
  }

  /**
   * Select camera - follows unidirectional data flow
   */
  selectCamera(deviceId: string): void {
    this.logger.info(`Selecting camera: ${deviceId}`);
    this.deviceStore.setSelectedCamera(deviceId);
  }

  /**
   * Update camera status - follows unidirectional data flow
   */
  updateCameraStatus(deviceId: string, status: string): void {
    this.logger.info(`Updating camera status: ${deviceId} -> ${status}`);
    this.deviceStore.updateCameraStatus(deviceId, status);
  }

  /**
   * Clear error state - follows unidirectional data flow
   */
  clearError(): void {
    this.deviceStore.clearError();
  }
}
