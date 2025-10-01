import { BaseService } from '../base/BaseService';
import { IAPIClient } from '../abstraction/IAPIClient';
import { LoggerService } from '../logger/LoggerService';
import { 
  ExternalStreamDiscoveryResult, 
  ExternalStreamAddResult, 
  ExternalStreamRemoveResult, 
  ExternalStreamsListResult,
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
export class ExternalStreamService extends BaseService {
  constructor(
    apiClient: IAPIClient,
    logger: LoggerService,
  ) {
    super(apiClient, logger);
    this.logInitialization('ExternalStreamService');
  }

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
    return this.callWithLogging('discover_external_streams', params) as Promise<ExternalStreamDiscoveryResult>;
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
    return this.callWithLogging('add_external_stream', params) as Promise<ExternalStreamAddResult>;
  }

  /**
   * Remove external stream from the system
   * Implements remove_external_stream RPC method
   */
  async removeExternalStream(streamUrl: string): Promise<ExternalStreamRemoveResult> {
    return this.callWithLogging('remove_external_stream', { stream_url: streamUrl }) as Promise<ExternalStreamRemoveResult>;
  }

  /**
   * Get list of external streams
   * Implements get_external_streams RPC method
   */
  async getExternalStreams(): Promise<ExternalStreamsListResult> {
    return this.callWithLogging('get_external_streams', {}) as Promise<ExternalStreamsListResult>;
  }

  /**
   * Set discovery interval for external streams
   * Implements set_discovery_interval RPC method
   */
  async setDiscoveryInterval(scanInterval: number): Promise<DiscoveryIntervalSetResult> {
    return this.callWithLogging('set_discovery_interval', { scan_interval: scanInterval }) as Promise<DiscoveryIntervalSetResult>;
  }
}
