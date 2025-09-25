import { WebSocketService } from '../websocket/WebSocketService';
import { LoggerService } from '../logger/LoggerService';
import { Camera, StreamInfo } from '../../stores/device/deviceStore';
import { IDiscovery } from '../interfaces/ServiceInterfaces';
import { CameraListResult, StreamUrlResult, CameraCapabilitiesResult, StreamStatusResult } from '../../types/api';

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
 * const deviceService = new DeviceService(wsService, logger);
 * const cameras = await deviceService.getCameraList();
 * const streams = await deviceService.getStreams();
 * const url = await deviceService.getStreamUrl('camera-001');
 * ```
 *
 * @see {@link ../interfaces/ServiceInterfaces#IDiscovery} IDiscovery interface
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
export class DeviceService implements IDiscovery {
  constructor(
    private wsService: WebSocketService,
    private logger: LoggerService,
  ) {}

  /**
   * Get list of all discovered cameras with their current status
   * Implements get_camera_list RPC method
   */
  async getCameraList(): Promise<Camera[]> {
    try {
      this.logger.info('Getting camera list');

      const response = await this.wsService.sendRPC('get_camera_list') as CameraListResult;

      if (response.cameras && response.cameras.length > 0) {
        this.logger.info(`Retrieved ${response.cameras.length} cameras`);
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

      this.logger.warn('No cameras found in response');
      return [];
    } catch (error) {
      this.logger.error('Failed to get camera list', error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get the stream URL for a specific camera device
   * Implements get_stream_url RPC method
   */
  async getStreamUrl(device: string): Promise<string | null> {
    try {
      this.logger.info(`Getting stream URL for device: ${device}`);

      const response = await this.wsService.sendRPC('get_stream_url', { device }) as StreamUrlResult;

      if (response.stream_url) {
        this.logger.info(`Retrieved stream URL for ${device}`);
        return response.stream_url;
      }

      this.logger.warn(`No stream URL found for device: ${device}`);
      return null;
    } catch (error) {
      this.logger.error(`Failed to get stream URL for device: ${device}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get list of all active streams from MediaMTX
   * Implements get_streams RPC method
   */
  async getStreams(): Promise<StreamInfo[]> {
    try {
      this.logger.info('Getting active streams');

      const response = await this.wsService.sendRPC('get_streams') as StreamInfo[];

      if (Array.isArray(response) && response.length > 0) {
        this.logger.info(`Retrieved ${response.length} active streams`);
        return response;
      }

      this.logger.warn('No streams found in response');
      return [];
    } catch (error) {
      this.logger.error('Failed to get streams', error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Subscribe to camera status update events
   * Implements subscribe_events RPC method
   */
  async subscribeToCameraEvents(): Promise<void> {
    try {
      this.logger.info('Subscribing to camera status updates');

      await this.wsService.sendRPC('subscribe_events', {
        topics: ['camera_status_update'],
      });

      this.logger.info('Successfully subscribed to camera events');
    } catch (error) {
      this.logger.error('Failed to subscribe to camera events', error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Unsubscribe from camera status update events
   * Implements unsubscribe_events RPC method
   */
  async unsubscribeFromCameraEvents(): Promise<void> {
    try {
      this.logger.info('Unsubscribing from camera status updates');

      await this.wsService.sendRPC('unsubscribe_events', {
        topics: ['camera_status_update'],
      });

      this.logger.info('Successfully unsubscribed from camera events');
    } catch (error) {
      this.logger.error('Failed to unsubscribe from camera events', error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get detailed capabilities and supported formats for a specific camera device
   * Implements get_camera_capabilities RPC method
   */
  async getCameraCapabilities(device: string): Promise<CameraCapabilitiesResult> {
    try {
      this.logger.info(`Getting capabilities for device: ${device}`);

      const response = await this.wsService.sendRPC('get_camera_capabilities', { device }) as CameraCapabilitiesResult;

      this.logger.info(`Retrieved capabilities for ${device}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to get capabilities for device: ${device}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get detailed status information for a specific camera stream
   * Implements get_stream_status RPC method
   */
  async getStreamStatus(device: string): Promise<StreamStatusResult> {
    try {
      this.logger.info(`Getting stream status for device: ${device}`);

      const response = await this.wsService.sendRPC('get_stream_status', { device }) as StreamStatusResult;

      this.logger.info(`Retrieved stream status for ${device}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to get stream status for device: ${device}`, error as Record<string, unknown>);
      throw error;
    }
  }
}
