import { APIClient } from '../abstraction/APIClient';
import { LoggerService } from '../logger/LoggerService';
import { 
  ExternalStreamDiscoveryResult, 
  ExternalStreamAddResult, 
  ExternalStreamRemoveResult, 
  DiscoveryIntervalSetResult
} from '../../types/api';

/**
 * External Stream Service - External stream management
 *
 * Implements external stream discovery, management, and configuration.
 * Provides methods for discovering external streams, adding/removing them,
 * and managing discovery intervals.
 *
 * @class ExternalStreamService
 *
 * @example
 * ```typescript
 * const externalStreamService = new ExternalStreamService(wsService, logger);
 *
 * // Discover external streams
 * const streams = await externalStreamService.discoverExternalStreams();
 *
 * // Add external stream
 * const result = await externalStreamService.addExternalStream({
 *   name: 'external-cam',
 *   url: 'rtsp://192.168.1.100/stream'
 * });
 *
 * // Remove external stream
 * await externalStreamService.removeExternalStream('external-cam');
 * ```
 *
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
export class ExternalStreamService {
  constructor(
    private apiClient: APIClient,
    private logger: LoggerService,
  ) {}

  /**
   * Discover external streams based on criteria
   * Implements discover_external_streams RPC method
   */
  async discoverExternalStreams(params: {
    skydio_enabled?: boolean;
    generic_enabled?: boolean;
    force_rescan?: boolean;
    include_offline?: boolean;
  } = {}): Promise<ExternalStreamDiscoveryResult> {
    try {
      this.logger.info('Discovering external streams', params);
      const response = await this.apiClient.call('discover_external_streams', params) as ExternalStreamDiscoveryResult;
      this.logger.info(`Discovered ${response.streams?.length || 0} external streams`);
      return response;
    } catch (error) {
      this.logger.error('Failed to discover external streams', error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Add external stream to the system
   * Implements add_external_stream RPC method
   */
  async addExternalStream(params: {
    stream_url: string;
    stream_name: string;
    stream_type?: string;
  }): Promise<ExternalStreamAddResult> {
    try {
      this.logger.info('Adding external stream', params);
      const response = await this.apiClient.call('add_external_stream', params) as ExternalStreamAddResult;
      this.logger.info(`External stream added: ${response.stream_name}`);
      return response;
    } catch (error) {
      this.logger.error('Failed to add external stream', error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Remove external stream from the system
   * Implements remove_external_stream RPC method
   */
  async removeExternalStream(streamUrl: string): Promise<ExternalStreamRemoveResult> {
    try {
      this.logger.info(`Removing external stream: ${streamUrl}`);
      const response = await this.apiClient.call('remove_external_stream', { stream_url: streamUrl }) as ExternalStreamRemoveResult;
      this.logger.info(`External stream removed: ${streamUrl}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to remove external stream: ${streamUrl}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get list of external streams
   * Implements get_external_streams RPC method
   */
  async getExternalStreams(): Promise<ExternalStreamListResult> {
    try {
      this.logger.info('Getting external streams');
      const response = await this.apiClient.call('get_external_streams', {}) as ExternalStreamListResult;
      this.logger.info(`Retrieved ${response.streams?.length || 0} external streams`);
      return response;
    } catch (error) {
      this.logger.error('Failed to get external streams', error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Set discovery interval for external streams
   * Implements set_discovery_interval RPC method
   */
  async setDiscoveryInterval(scanInterval: number): Promise<DiscoveryIntervalSetResult> {
    try {
      this.logger.info(`Setting discovery interval: ${scanInterval} seconds`);
      const response = await this.apiClient.call('set_discovery_interval', { scan_interval: scanInterval }) as DiscoveryIntervalSetResult;
      this.logger.info(`Discovery interval set to ${response.scan_interval} seconds`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to set discovery interval: ${scanInterval}`, error as Record<string, unknown>);
      throw error;
    }
  }
}
