import { BaseService } from '../base/BaseService';
import { Camera, StreamsListResult, CameraListResult, CameraStatusResult, StreamUrlResult, CameraCapabilitiesResult, StreamStatusResult, StreamStartResult, StreamStopResult } from '../../types/api';
import { IDiscovery } from '../interfaces/ServiceInterfaces';
import { validateCameraDeviceId } from '../../utils/validation';
import { IAPIClient } from '../abstraction/IAPIClient';
import { LoggerService } from '../logger/LoggerService';

/**
 * Device Service - Camera discovery and stream management
 *
 * Implements the IDiscovery interface for device enumeration and stream URL management.
 * Provides methods for discovering cameras, retrieving active streams, and getting stream URLs.
 *
 * @class DeviceService
 * @implements {IDiscovery}
 *
 * @example
 * ```typescript
 * const deviceService = new DeviceService(apiClient, logger);
 * const cameras = await deviceService.getCameraList();
 * const streams = await deviceService.getStreams();
 * const url = await deviceService.getStreamUrl('camera-001');
 * ```
 *
 * @see {@link ../interfaces/ServiceInterfaces#IDiscovery} IDiscovery interface
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
export class DeviceService extends BaseService implements IDiscovery {
  constructor(
    apiClient: IAPIClient,
    logger: LoggerService,
  ) {
    super(apiClient, logger);
    this.logInitialization('DeviceService');
  }


  /**
   * Get list of all discovered cameras with their current status
   * Implements get_camera_list RPC method
   */
  async getCameraList(): Promise<Camera[]> {
    const response = await this.callWithLogging<CameraListResult>('get_camera_list', {});

    if (response.cameras && response.cameras.length > 0) {
      // Transform API cameras to store cameras with required fields
      return response.cameras.map(apiCamera => ({
        device: apiCamera.device,
        status: apiCamera.status,
        name: apiCamera.name || `Camera ${apiCamera.device}`,
        resolution: apiCamera.resolution || 'Unknown',
        fps: apiCamera.fps || 30,
        streams: apiCamera.streams || { rtsp: '', hls: '' }
      }));
    }

    return [];
  }

  /**
   * Get status for a specific camera device
   * Implements get_camera_status RPC method
   */
  async getCameraStatus(device: string): Promise<CameraStatusResult> {
    return this.callWithLogging('get_camera_status', { device }, `getCameraStatus(${device})`) as Promise<CameraStatusResult>;
  }

  /**
   * Get the stream URL for a specific camera device
   * Implements get_stream_url RPC method
   */
  async getStreamUrl(device: string): Promise<string | null> {
    if (!validateCameraDeviceId(device)) {
      throw new Error(`Invalid device ID format: ${device}. Expected format: camera[0-9]+`);
    }

    const response = await this.callWithLogging('get_stream_url', { device }, `getStreamUrl(${device})`) as StreamUrlResult;
    return response.stream_url || null;
  }

  /**
   * Get list of all active streams from MediaMTX
   * Implements get_streams RPC method
   */
  async getStreams(): Promise<StreamsListResult[]> {
    const response = await this.callWithLogging<StreamsListResult[]>('get_streams', {});

    if (Array.isArray(response) && response.length > 0) {
      return response;
    }

    return [];
  }

  /**
   * Get camera capabilities for a specific device
   * Implements get_camera_capabilities RPC method
   */
  async getCameraCapabilities(device: string): Promise<CameraCapabilitiesResult> {
    return this.callWithLogging('get_camera_capabilities', { device }, `getCameraCapabilities(${device})`) as Promise<CameraCapabilitiesResult>;
  }

  // STREAMING METHODS MOVED TO StreamingService
  // Architecture requirement: Interface segregation (Section 1.2)
  // DeviceService now only handles discovery operations per IDiscovery interface

  /**
   * Subscribe to camera status update events
   * Implements subscribe_events RPC method
   */
  async subscribeToCameraEvents(): Promise<void> {
    await this.callWithLogging('subscribe_events', {
      topics: ['camera_status_update'],
    }, 'subscribeToCameraEvents');
  }

  /**
   * Unsubscribe from camera status update events
   * Implements unsubscribe_events RPC method
   */
  async unsubscribeFromCameraEvents(): Promise<void> {
    await this.callWithLogging('unsubscribe_events', {
      topics: ['camera_status_update'],
    }, 'unsubscribeFromCameraEvents');
  }


  /**
   * Get detailed status information for a specific camera stream
   * Implements get_stream_status RPC method
   */
  async getStreamStatus(device: string): Promise<StreamStatusResult> {
    return this.callWithLogging('get_stream_status', { device }, `getStreamStatus(${device})`) as Promise<StreamStatusResult>;
  }

  /**
   * Start streaming for a specific camera device
   * Implements start_streaming RPC method
   */
  async startStreaming(device: string): Promise<StreamStartResult> {
    return this.callWithLogging('start_streaming', { device }, `startStreaming(${device})`) as Promise<StreamStartResult>;
  }

  /**
   * Stop streaming for a specific camera device
   * Implements stop_streaming RPC method
   */
  async stopStreaming(device: string): Promise<StreamStopResult> {
    return this.callWithLogging('stop_streaming', { device }, `stopStreaming(${device})`) as Promise<StreamStopResult>;
  }
}
