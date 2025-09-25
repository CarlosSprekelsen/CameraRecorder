/**
 * ExternalStreamService unit tests for missing RPC methods
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-EXT-001: discover_external_streams RPC method
 * - REQ-EXT-002: add_external_stream RPC method
 * - REQ-EXT-003: remove_external_stream RPC method
 * - REQ-EXT-004: get_external_streams RPC method
 * - REQ-EXT-005: set_discovery_interval RPC method
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { 
  ExternalStreamDiscoveryResult, 
  ExternalStreamAddResult, 
  ExternalStreamRemoveResult, 
  ExternalStreamsListResult, 
  DiscoveryIntervalSetResult 
} from '../../../src/types/api';
import { MockDataFactory } from '../../utils/mocks';

// Use centralized mocks - eliminates duplication
const mockWebSocketService = MockDataFactory.createMockWebSocketService();
const mockLoggerService = MockDataFactory.createMockLoggerService();

// Create a mock external stream service class
class ExternalStreamService {
  constructor(
    private wsService: WebSocketService,
    private logger: LoggerService
  ) {}

  async discoverExternalStreams(options: {
    skydio_enabled?: boolean;
    generic_enabled?: boolean;
    force_rescan?: boolean;
    include_offline?: boolean;
  } = {}): Promise<ExternalStreamDiscoveryResult> {
    try {
      this.logger.info('discover_external_streams request', options);
      return await this.wsService.sendRPC('discover_external_streams', options);
    } catch (error) {
      this.logger.error('discover_external_streams failed', error as Error);
      throw error;
    }
  }

  async addExternalStream(streamUrl: string, streamName: string, streamType?: string): Promise<ExternalStreamAddResult> {
    try {
      this.logger.info('add_external_stream request', { streamUrl, streamName, streamType });
      return await this.wsService.sendRPC('add_external_stream', { streamUrl, streamName, streamType });
    } catch (error) {
      this.logger.error('add_external_stream failed', error as Error);
      throw error;
    }
  }

  async removeExternalStream(streamUrl: string): Promise<ExternalStreamRemoveResult> {
    try {
      this.logger.info('remove_external_stream request', { streamUrl });
      return await this.wsService.sendRPC('remove_external_stream', { streamUrl });
    } catch (error) {
      this.logger.error('remove_external_stream failed', error as Error);
      throw error;
    }
  }

  async getExternalStreams(): Promise<ExternalStreamsListResult> {
    try {
      this.logger.info('get_external_streams request');
      return await this.wsService.sendRPC('get_external_streams');
    } catch (error) {
      this.logger.error('get_external_streams failed', error as Error);
      throw error;
    }
  }

  async setDiscoveryInterval(scanInterval: number): Promise<DiscoveryIntervalSetResult> {
    try {
      this.logger.info('set_discovery_interval request', { scanInterval });
      return await this.wsService.sendRPC('set_discovery_interval', { scanInterval });
    } catch (error) {
      this.logger.error('set_discovery_interval failed', error as Error);
      throw error;
    }
  }
}

describe('ExternalStreamService Unit Tests', () => {
  let externalStreamService: ExternalStreamService;

  beforeEach(() => {
    jest.clearAllMocks();
    externalStreamService = new ExternalStreamService(mockWebSocketService, mockLoggerService);
  });

  describe('REQ-EXT-001: discover_external_streams RPC method', () => {
    test('Should call WebSocket service with default parameters', async () => {
      // Arrange
      const expectedResult: ExternalStreamDiscoveryResult = {
        discovered_streams: [],
        skydio_streams: [],
        generic_streams: [],
        scan_timestamp: '2025-01-15T14:30:00Z',
        total_found: 0,
        discovery_options: {
          skydio_enabled: true,
          generic_enabled: false,
          force_rescan: false,
          include_offline: false
        },
        scan_duration: '2.5s',
        errors: []
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await externalStreamService.discoverExternalStreams();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('discover_external_streams request', {});
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('discover_external_streams', {});
      expect(result).toEqual(expectedResult);
    });

    test('Should call WebSocket service with custom parameters', async () => {
      // Arrange
      const options = {
        skydio_enabled: true,
        generic_enabled: true,
        force_rescan: true,
        include_offline: true
      };
      const expectedResult: ExternalStreamDiscoveryResult = {
        discovered_streams: [
          {
            url: 'rtsp://192.168.42.10:5554/subject',
            type: 'skydio_stanag4609',
            name: 'Skydio_EO_192.168.42.10_eo_/subject',
            status: 'DISCOVERED',
            discovered_at: '2025-01-15T14:30:00Z',
            last_seen: '2025-01-15T14:30:00Z',
            capabilities: {
              protocol: 'rtsp',
              format: 'stanag4609',
              source: 'skydio_uav',
              stream_type: 'eo',
              port: 5554,
              stream_path: '/subject',
              codec: 'h264',
              metadata: 'klv_mpegts'
            }
          }
        ],
        skydio_streams: [],
        generic_streams: [],
        scan_timestamp: '2025-01-15T14:30:00Z',
        total_found: 1,
        discovery_options: options,
        scan_duration: '2.5s',
        errors: []
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await externalStreamService.discoverExternalStreams(options);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('discover_external_streams request', options);
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('discover_external_streams', options);
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const error = new Error('Discovery failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      // Act & Assert
      await expect(externalStreamService.discoverExternalStreams()).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('discover_external_streams failed', error);
    });
  });

  describe('REQ-EXT-002: add_external_stream RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const streamUrl = 'rtsp://192.168.42.15:5554/subject';
      const streamName = 'Skydio_UAV_15';
      const streamType = 'skydio_stanag4609';
      const expectedResult: ExternalStreamAddResult = {
        stream_url: streamUrl,
        stream_name: streamName,
        stream_type: streamType,
        status: 'ADDED',
        timestamp: '2025-01-15T14:30:00Z'
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await externalStreamService.addExternalStream(streamUrl, streamName, streamType);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('add_external_stream request', { streamUrl, streamName, streamType });
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('add_external_stream', { streamUrl, streamName, streamType });
      expect(result).toEqual(expectedResult);
    });

    test('Should call WebSocket service with default stream type', async () => {
      // Arrange
      const streamUrl = 'rtsp://192.168.42.15:5554/subject';
      const streamName = 'Generic_UAV_15';
      const expectedResult: ExternalStreamAddResult = {
        stream_url: streamUrl,
        stream_name: streamName,
        stream_type: 'generic_rtsp',
        status: 'ADDED',
        timestamp: '2025-01-15T14:30:00Z'
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await externalStreamService.addExternalStream(streamUrl, streamName);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('add_external_stream request', { streamUrl, streamName, streamType: undefined });
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('add_external_stream', { streamUrl, streamName, streamType: undefined });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const streamUrl = 'rtsp://192.168.42.15:5554/subject';
      const streamName = 'Test_UAV';
      const error = new Error('Add external stream failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      // Act & Assert
      await expect(externalStreamService.addExternalStream(streamUrl, streamName)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('add_external_stream failed', error);
    });
  });

  describe('REQ-EXT-003: remove_external_stream RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const streamUrl = 'rtsp://192.168.42.15:5554/subject';
      const expectedResult: ExternalStreamRemoveResult = {
        stream_url: streamUrl,
        status: 'REMOVED',
        timestamp: '2025-01-15T14:30:00Z'
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await externalStreamService.removeExternalStream(streamUrl);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('remove_external_stream request', { streamUrl });
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('remove_external_stream', { streamUrl });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const streamUrl = 'rtsp://192.168.42.15:5554/subject';
      const error = new Error('Remove external stream failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      // Act & Assert
      await expect(externalStreamService.removeExternalStream(streamUrl)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('remove_external_stream failed', error);
    });
  });

  describe('REQ-EXT-004: get_external_streams RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const expectedResult: ExternalStreamsListResult = {
        external_streams: [
          {
            url: 'rtsp://192.168.42.10:5554/subject',
            type: 'skydio_stanag4609',
            name: 'Skydio_EO_192.168.42.10_eo_/subject',
            status: 'DISCOVERED',
            discovered_at: '2025-01-15T14:30:00Z',
            last_seen: '2025-01-15T14:30:00Z',
            capabilities: {
              protocol: 'rtsp',
              format: 'stanag4609',
              source: 'skydio_uav',
              stream_type: 'eo',
              port: 5554,
              stream_path: '/subject',
              codec: 'h264',
              metadata: 'klv_mpegts'
            }
          }
        ],
        skydio_streams: [],
        generic_streams: [],
        total_count: 1,
        timestamp: '2025-01-15T14:30:00Z'
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await externalStreamService.getExternalStreams();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_external_streams request');
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_external_streams');
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const error = new Error('Get external streams failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      // Act & Assert
      await expect(externalStreamService.getExternalStreams()).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('get_external_streams failed', error);
    });
  });

  describe('REQ-EXT-005: set_discovery_interval RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const scanInterval = 300;
      const expectedResult: DiscoveryIntervalSetResult = {
        scan_interval: scanInterval,
        status: 'UPDATED',
        message: 'Discovery interval updated (restart required for changes to take effect)',
        timestamp: '2025-01-15T14:30:00Z'
      };

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      // Act
      const result = await externalStreamService.setDiscoveryInterval(scanInterval);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('set_discovery_interval request', { scanInterval });
      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('set_discovery_interval', { scanInterval });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const scanInterval = 300;
      const error = new Error('Set discovery interval failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      // Act & Assert
      await expect(externalStreamService.setDiscoveryInterval(scanInterval)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('set_discovery_interval failed', error);
    });
  });
});
