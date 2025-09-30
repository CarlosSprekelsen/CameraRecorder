/**
 * StreamingService unit tests for missing RPC methods
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-STREAM-001: start_streaming RPC method
 * - REQ-STREAM-002: stop_streaming RPC method
 * - REQ-STREAM-003: get_stream_url RPC method
 * - REQ-STREAM-004: get_stream_status RPC method
 * - REQ-STREAM-005: get_streams RPC method
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { StreamingService } from '../../../src/services/streaming/StreamingService';
import { IAPIClient } from '../../../src/services/abstraction/IAPIClient';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { StreamStartResult, StreamStopResult, StreamUrlResult, StreamStatusResult, StreamsListResult } from '../../../src/types/api';
import { MockDataFactory } from '../../utils/mocks';

// Use centralized mocks - aligned with refactored architecture
const mockAPIClient = MockDataFactory.createMockAPIClient();
const mockLoggerService = MockDataFactory.createMockLoggerService();

// Use real StreamingService with APIClient

describe('StreamingService Unit Tests', () => {
  let streamingService: StreamingService;

  beforeEach(() => {
    jest.clearAllMocks();
    streamingService = new StreamingService(mockAPIClient, mockLoggerService);
  });

  describe('REQ-STREAM-001: start_streaming RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const device = 'camera0';
      const expectedResult: StreamStartResult = {
        device,
        stream_name: 'camera_video0_viewing',
        stream_url: 'rtsp://localhost:8554/camera_video0_viewing',
        status: 'STARTED',
        start_time: '2025-01-15T14:30:00Z',
        auto_close_after: '300s'
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await streamingService.startStreaming(device);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('start_streaming request', { device });
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('start_streaming', { device });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const device = 'camera0';
      const error = new Error('Streaming failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(streamingService.startStreaming(device)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('start_streaming failed', error);
    });
  });

  describe('REQ-STREAM-002: stop_streaming RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const device = 'camera0';
      const expectedResult: StreamStopResult = {
        device,
        stream_name: 'camera_video0_viewing',
        status: 'STOPPED',
        start_time: '2025-01-15T14:30:00Z',
        end_time: '2025-01-15T14:35:00Z',
        duration: 300,
        stream_continues: false
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await streamingService.stopStreaming(device);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('stop_streaming request', { device });
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('stop_streaming', { device });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const device = 'camera0';
      const error = new Error('Stop streaming failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(streamingService.stopStreaming(device)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('stop_streaming failed', error);
    });
  });

  describe('REQ-STREAM-003: get_stream_url RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const device = 'camera0';
      const expectedResult: StreamUrlResult = {
        device,
        stream_name: 'camera_video0_viewing',
        stream_url: 'rtsp://localhost:8554/camera_video0_viewing',
        available: true,
        active_consumers: 2,
        stream_status: 'READY'
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await streamingService.getStreamUrl(device);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_stream_url request', { device });
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('get_stream_url', { device });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const device = 'camera0';
      const error = new Error('Get stream URL failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(streamingService.getStreamUrl(device)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('get_stream_url failed', error);
    });
  });

  describe('REQ-STREAM-004: get_stream_status RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const device = 'camera0';
      const expectedResult: StreamStatusResult = {
        device,
        stream_name: 'camera_video0_viewing',
        status: 'ACTIVE',
        ready: true,
        ffmpeg_process: {
          running: true,
          pid: 12345,
          uptime: 300
        },
        mediamtx_path: {
          exists: true,
          ready: true,
          readers: 2
        },
        metrics: {
          bytes_sent: 12345678,
          frames_sent: 9000,
          bitrate: 600000,
          fps: 30
        },
        start_time: '2025-01-15T14:30:00Z'
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await streamingService.getStreamStatus(device);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_stream_status request', { device });
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('get_stream_status', { device });
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const device = 'camera0';
      const error = new Error('Get stream status failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(streamingService.getStreamStatus(device)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('get_stream_status failed', error);
    });
  });

  describe('REQ-STREAM-005: get_streams RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const expectedResult: StreamsListResult = [
        {
          name: 'camera0',
          source: 'ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -profile:v baseline -level 3.0 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:8554/camera0',
          ready: true,
          readers: 2,
          bytes_sent: 12345678
        },
        {
          name: 'camera1',
          source: 'ffmpeg -f v4l2 -i /dev/video1 -c:v libx264 -profile:v baseline -level 3.0 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:8554/camera1',
          ready: false,
          readers: 0,
          bytes_sent: 0
        }
      ];

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await streamingService.getStreams();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_streams request');
      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('get_streams');
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const error = new Error('Get streams failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(streamingService.getStreams()).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('get_streams failed', error);
    });
  });
});
