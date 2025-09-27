import { WebSocketService } from '../websocket/WebSocketService';
import { LoggerService } from '../logger/LoggerService';
import { 
  ExternalStreamDiscoveryResult, 
  ExternalStreamAddResult, 
  ExternalStreamRemoveResult, 
  ExternalStreamListResult,
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
    private wsService: WebSocketService,
    private logger: LoggerService,
  ) {}

  /**
   * Discover external streams based on criteria
   * Implements discover_external_streams RPC method
   */
  async discoverExternalStreams(criteria?: Record<string, unknown>): Promise<ExternalStreamDiscoveryResult> {
    try {
      this.logger.info('Discovering external streams', { criteria });
      const response = await this.wsService.sendRPC('discover_external_streams', { criteria }) as ExternalStreamDiscoveryResult;
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
  async addExternalStream(streamConfig: Record<string, unknown>): Promise<ExternalStreamAddResult> {
    try {
      this.logger.info('Adding external stream', { streamConfig });
      const response = await this.wsService.sendRPC('add_external_stream', { streamConfig }) as ExternalStreamAddResult;
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
  async removeExternalStream(streamId: string): Promise<ExternalStreamRemoveResult> {
    try {
      this.logger.info(`Removing external stream: ${streamId}`);
      const response = await this.wsService.sendRPC('remove_external_stream', { stream_id: streamId }) as ExternalStreamRemoveResult;
      this.logger.info(`External stream removed: ${streamId}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to remove external stream: ${streamId}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get list of external streams
   * Implements get_external_streams RPC method
   */
  async getExternalStreams(filter?: Record<string, unknown>): Promise<ExternalStreamListResult> {
    try {
      this.logger.info('Getting external streams', { filter });
      const response = await this.wsService.sendRPC('get_external_streams', { filter }) as ExternalStreamListResult;
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
  async setDiscoveryInterval(interval: number): Promise<DiscoveryIntervalSetResult> {
    try {
      this.logger.info(`Setting discovery interval: ${interval} seconds`);
      const response = await this.wsService.sendRPC('set_discovery_interval', { interval }) as DiscoveryIntervalSetResult;
      this.logger.info(`Discovery interval set to ${response.scan_interval} seconds`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to set discovery interval: ${interval}`, error as Record<string, unknown>);
      throw error;
    }
  }
}
