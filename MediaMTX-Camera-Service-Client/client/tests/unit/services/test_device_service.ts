/**
 * DeviceService unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-DEV-001: Camera discovery and listing
 * - REQ-DEV-002: Stream URL management
 * - REQ-DEV-003: Stream status monitoring
 * - REQ-DEV-004: Event subscription management
 * - REQ-DEV-005: Camera capabilities retrieval
 * 
 * Test Categories: Unit
 * API Documentation Reference: ../mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

import { DeviceService } from '../../../src/services/device/DeviceService';
import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Mock dependencies
const mockWebSocketService = {
  sendRPC: jest.fn(),
} as jest.Mocked<WebSocketService>;

const mockLoggerService = {
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
} as jest.Mocked<LoggerService>;

describe('DeviceService Unit Tests', () => {
  let deviceService: DeviceService;

  beforeEach(() => {
    jest.clearAllMocks();
    deviceService = new DeviceService(mockWebSocketService, mockLoggerService);
  });

  describe('REQ-DEV-001: Camera discovery and listing', () => {
    test('should get camera list successfully', async () => {
      const expectedResult = MockDataFactory.getCameraListResult();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await deviceService.getCameraList();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_camera_list');
      expect(mockLoggerService.info).toHaveBeenCalledWith('Getting camera list');
      expect(mockLoggerService.info).toHaveBeenCalledWith(
        `Retrieved ${expectedResult.cameras.length} cameras`
      );
      expect(result).toEqual(expectedResult.cameras);
      expect(APIResponseValidator.validateCameraListResult(expectedResult)).toBe(true);
    });

    test('should handle empty camera list', async () => {
      const emptyResult = { cameras: [], total: 0, connected: 0 };
      mockWebSocketService.sendRPC.mockResolvedValue(emptyResult);

      const result = await deviceService.getCameraList();

      expect(result).toEqual([]);
      expect(mockLoggerService.warn).toHaveBeenCalledWith('No cameras found in response');
    });

    test('should handle camera list errors', async () => {
      const error = new Error('Failed to get cameras');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(deviceService.getCameraList()).rejects.toThrow('Failed to get cameras');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'Failed to get camera list',
        error
      );
    });

    test('should validate camera objects', async () => {
      const cameraList = MockDataFactory.getCameraListResult();
      mockWebSocketService.sendRPC.mockResolvedValue(cameraList);

      const result = await deviceService.getCameraList();

      result.forEach(camera => {
        expect(APIResponseValidator.validateCamera(camera)).toBe(true);
      });
    });
  });

  describe('REQ-DEV-002: Stream URL management', () => {
    test('should get stream URL for device', async () => {
      const device = 'camera0';
      const streamUrl = 'rtsp://localhost:8554/camera0';
      const response = { stream_url: streamUrl };
      
      mockWebSocketService.sendRPC.mockResolvedValue(response);

      const result = await deviceService.getStreamUrl(device);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_stream_url', { device });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Getting stream URL for device: ${device}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Retrieved stream URL for ${device}`);
      expect(result).toBe(streamUrl);
    });

    test('should return null when no stream URL found', async () => {
      const device = 'camera0';
      const response = { stream_url: null };
      
      mockWebSocketService.sendRPC.mockResolvedValue(response);

      const result = await deviceService.getStreamUrl(device);

      expect(result).toBeNull();
      expect(mockLoggerService.warn).toHaveBeenCalledWith(`No stream URL found for device: ${device}`);
    });

    test('should handle stream URL errors', async () => {
      const device = 'camera0';
      const error = new Error('Stream not available');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(deviceService.getStreamUrl(device)).rejects.toThrow('Stream not available');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        `Failed to get stream URL for device: ${device}`,
        error
      );
    });

    test('should validate device ID format', async () => {
      const device = 'camera0';
      const streamUrl = 'rtsp://localhost:8554/camera0';
      const response = { stream_url: streamUrl };
      
      mockWebSocketService.sendRPC.mockResolvedValue(response);

      await deviceService.getStreamUrl(device);

      expect(APIResponseValidator.validateDeviceId(device)).toBe(true);
    });
  });

  describe('REQ-DEV-003: Stream status monitoring', () => {
    test('should get active streams', async () => {
      const expectedStreams = [
        {
          device: 'camera0',
          status: 'ACTIVE',
          url: 'rtsp://localhost:8554/camera0'
        },
        {
          device: 'camera1',
          status: 'ACTIVE',
          url: 'rtsp://localhost:8554/camera1'
        }
      ];
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedStreams);

      const result = await deviceService.getStreams();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_streams');
      expect(mockLoggerService.info).toHaveBeenCalledWith('Getting active streams');
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Retrieved ${expectedStreams.length} active streams`);
      expect(result).toEqual(expectedStreams);
    });

    test('should handle empty streams list', async () => {
      mockWebSocketService.sendRPC.mockResolvedValue([]);

      const result = await deviceService.getStreams();

      expect(result).toEqual([]);
      expect(mockLoggerService.warn).toHaveBeenCalledWith('No streams found in response');
    });

    test('should handle stream errors', async () => {
      const error = new Error('Failed to get streams');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(deviceService.getStreams()).rejects.toThrow('Failed to get streams');
      expect(mockLoggerService.error).toHaveBeenCalledWith('Failed to get streams', error);
    });

    test('should get stream status for specific device', async () => {
      const device = 'camera0';
      const expectedStatus = MockDataFactory.getStreamStatus(device);
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedStatus);

      const result = await deviceService.getStreamStatus(device);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_stream_status', { device });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Getting stream status for device: ${device}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Retrieved stream status for ${device}`);
      expect(result).toEqual(expectedStatus);
      expect(APIResponseValidator.validateStreamStatus(result)).toBe(true);
    });

    test('should handle stream status errors', async () => {
      const device = 'camera0';
      const error = new Error('Stream status unavailable');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(deviceService.getStreamStatus(device)).rejects.toThrow('Stream status unavailable');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        `Failed to get stream status for device: ${device}`,
        error
      );
    });
  });

  describe('REQ-DEV-004: Event subscription management', () => {
    test('should subscribe to camera events', async () => {
      const subscriptionResult = { subscribed: true, topics: ['camera_status_update'] };
      mockWebSocketService.sendRPC.mockResolvedValue(subscriptionResult);

      await deviceService.subscribeToCameraEvents();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('subscribe_events', {
        topics: ['camera_status_update'],
      });
      expect(mockLoggerService.info).toHaveBeenCalledWith('Subscribing to camera status updates');
      expect(mockLoggerService.info).toHaveBeenCalledWith('Successfully subscribed to camera events');
    });

    test('should handle subscription errors', async () => {
      const error = new Error('Subscription failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(deviceService.subscribeToCameraEvents()).rejects.toThrow('Subscription failed');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'Failed to subscribe to camera events',
        error
      );
    });

    test('should unsubscribe from camera events', async () => {
      const unsubscriptionResult = { unsubscribed: true, topics: ['camera_status_update'] };
      mockWebSocketService.sendRPC.mockResolvedValue(unsubscriptionResult);

      await deviceService.unsubscribeFromCameraEvents();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('unsubscribe_events', {
        topics: ['camera_status_update'],
      });
      expect(mockLoggerService.info).toHaveBeenCalledWith('Unsubscribing from camera status updates');
      expect(mockLoggerService.info).toHaveBeenCalledWith('Successfully unsubscribed from camera events');
    });

    test('should handle unsubscription errors', async () => {
      const error = new Error('Unsubscription failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(deviceService.unsubscribeFromCameraEvents()).rejects.toThrow('Unsubscription failed');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'Failed to unsubscribe from camera events',
        error
      );
    });
  });

  describe('REQ-DEV-005: Camera capabilities retrieval', () => {
    test('should get camera capabilities', async () => {
      const device = 'camera0';
      const expectedCapabilities = {
        supported_formats: ['h264', 'h265'],
        max_resolution: '4K',
        supported_fps: [15, 30, 60],
        features: ['night_vision', 'motion_detection']
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedCapabilities);

      const result = await deviceService.getCameraCapabilities(device);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_camera_capabilities', { device });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Getting capabilities for device: ${device}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Retrieved capabilities for ${device}`);
      expect(result).toEqual(expectedCapabilities);
    });

    test('should handle capabilities errors', async () => {
      const device = 'camera0';
      const error = new Error('Capabilities unavailable');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(deviceService.getCameraCapabilities(device)).rejects.toThrow('Capabilities unavailable');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        `Failed to get capabilities for device: ${device}`,
        error
      );
    });

    test('should validate device ID for capabilities', async () => {
      const device = 'camera0';
      const capabilities = { supported_formats: ['h264'] };
      
      mockWebSocketService.sendRPC.mockResolvedValue(capabilities);

      await deviceService.getCameraCapabilities(device);

      expect(APIResponseValidator.validateDeviceId(device)).toBe(true);
    });
  });

  describe('Error handling and logging', () => {
    test('should log all operations with appropriate levels', async () => {
      const cameraList = MockDataFactory.getCameraListResult();
      mockWebSocketService.sendRPC.mockResolvedValue(cameraList);

      await deviceService.getCameraList();

      expect(mockLoggerService.info).toHaveBeenCalledWith('Getting camera list');
      expect(mockLoggerService.info).toHaveBeenCalledWith(
        `Retrieved ${cameraList.cameras.length} cameras`
      );
    });

    test('should handle WebSocket service errors gracefully', async () => {
      const error = new Error('WebSocket connection lost');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(deviceService.getCameraList()).rejects.toThrow('WebSocket connection lost');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'Failed to get camera list',
        error
      );
    });
  });
});
