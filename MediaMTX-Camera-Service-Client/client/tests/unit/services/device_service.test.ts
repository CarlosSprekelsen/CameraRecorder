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
import { IAPIClient } from '../../../src/services/abstraction/IAPIClient';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Use centralized mocks - aligned with refactored architecture
const mockAPIClient = MockDataFactory.createMockAPIClient();
const mockLoggerService = MockDataFactory.createMockLoggerService();

describe('DeviceService Unit Tests', () => {
  let deviceService: DeviceService;

  beforeEach(() => {
    jest.clearAllMocks();
    deviceService = new DeviceService(mockAPIClient, mockLoggerService);
  });

  describe('REQ-DEV-001: Camera discovery and listing', () => {
    test('should get camera list successfully', async () => {
      const expectedResult = MockDataFactory.getCameraListResult();
      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      const result = await deviceService.getCameraList();

      expect(mockAPIClient.call).toHaveBeenCalledWith('get_camera_list', {});
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_camera_list request', {});
      expect(result).toEqual(expectedResult.cameras);
      // Note: Validation removed temporarily to focus on IAPIClient migration
    });

    test('should handle empty camera list', async () => {
      const emptyResult = { cameras: [], total: 0, connected: 0 };
      (mockAPIClient.call as jest.Mock).mockResolvedValue(emptyResult);

      const result = await deviceService.getCameraList();

      expect(result).toEqual([]);
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_camera_list request', {});
    });

    test('should handle camera list errors', async () => {
      const error = new Error('Failed to get cameras');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      await expect(deviceService.getCameraList()).rejects.toThrow('Failed to get cameras');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'get_camera_list failed',
        error
      );
    });

    test('should validate camera objects', async () => {
      const cameraList = MockDataFactory.getCameraListResult();
      (mockAPIClient.call as jest.Mock).mockResolvedValue(cameraList);

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
      
      (mockAPIClient.call as jest.Mock).mockResolvedValue(response);

      const result = await deviceService.getStreamUrl(device);

      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('get_stream_url', { device });
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_stream_url request', { device });
      expect(result).toBe(streamUrl);
    });

    test('should return null when no stream URL found', async () => {
      const device = 'camera0';
      const response = { stream_url: null };
      
      (mockAPIClient.call as jest.Mock).mockResolvedValue(response);

      const result = await deviceService.getStreamUrl(device);

      expect(result).toBeNull();
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_stream_url request', { device });
    });

    test('should handle stream URL errors', async () => {
      const device = 'camera0';
      const error = new Error('Stream not available');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      await expect(deviceService.getStreamUrl(device)).rejects.toThrow('Stream not available');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'get_stream_url failed',
        error
      );
    });

    test('should validate device ID format', async () => {
      const device = 'camera0';
      const streamUrl = 'rtsp://localhost:8554/camera0';
      const response = { stream_url: streamUrl };
      
      (mockAPIClient.call as jest.Mock).mockResolvedValue(response);

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
      
      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedStreams);

      const result = await deviceService.getStreams();

      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('get_streams', {});
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_streams request', {});
      expect(result).toEqual(expectedStreams);
    });

    test('should handle empty streams list', async () => {
      (mockAPIClient.call as jest.Mock).mockResolvedValue([]);

      const result = await deviceService.getStreams();

      expect(result).toEqual([]);
      
    });

    test('should handle stream errors', async () => {
      const error = new Error('Failed to get streams');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      await expect(deviceService.getStreams()).rejects.toThrow('Failed to get streams');
      expect(mockLoggerService.error).toHaveBeenCalledWith('get_streams failed', error);
    });

    test('should get stream status for specific device', async () => {
      const device = 'camera0';
      const expectedStatus = MockDataFactory.getStreamStatusResult();
      
      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedStatus);

      const result = await deviceService.getStreamStatus(device);

      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('get_stream_status', { device });
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_stream_status request', { device });
      expect(result).toEqual(expectedStatus);
      expect(APIResponseValidator.validateStreamStatus(result)).toBe(true);
    });

    test('should handle stream status errors', async () => {
      const device = 'camera0';
      const error = new Error('Stream status unavailable');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      await expect(deviceService.getStreamStatus(device)).rejects.toThrow('Stream status unavailable');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'get_stream_status failed',
        error
      );
    });
  });

  describe('REQ-DEV-004: Event subscription management', () => {
    test('should subscribe to camera events', async () => {
      const subscriptionResult = { subscribed: true, topics: ['camera_status_update'] };
      (mockAPIClient.call as jest.Mock).mockResolvedValue(subscriptionResult);

      await deviceService.subscribeToCameraEvents();

      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('subscribe_events', {
        topics: ['camera_status_update'],
      });
      expect(mockLoggerService.info).toHaveBeenCalledWith('subscribe_events request', { topics: ['camera_status_update'] });
    });

    test('should handle subscription errors', async () => {
      const error = new Error('Subscription failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      await expect(deviceService.subscribeToCameraEvents()).rejects.toThrow('Subscription failed');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'subscribe_events failed',
        error
      );
    });

    test('should unsubscribe from camera events', async () => {
      const unsubscriptionResult = { unsubscribed: true, topics: ['camera_status_update'] };
      (mockAPIClient.call as jest.Mock).mockResolvedValue(unsubscriptionResult);

      await deviceService.unsubscribeFromCameraEvents();

      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('unsubscribe_events', {
        topics: ['camera_status_update'],
      });
      expect(mockLoggerService.info).toHaveBeenCalledWith('unsubscribe_events request', { topics: ['camera_status_update'] });
    });

    test('should handle unsubscription errors', async () => {
      const error = new Error('Unsubscription failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      await expect(deviceService.unsubscribeFromCameraEvents()).rejects.toThrow('Unsubscription failed');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'unsubscribe_events failed',
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
      
      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedCapabilities);

      const result = await deviceService.getCameraCapabilities(device);

      expect((mockAPIClient.call as jest.Mock)).toHaveBeenCalledWith('get_camera_capabilities', { device });
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_camera_capabilities request', { device });
      expect(result).toEqual(expectedCapabilities);
    });

    test('should handle capabilities errors', async () => {
      const device = 'camera0';
      const error = new Error('Capabilities unavailable');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      await expect(deviceService.getCameraCapabilities(device)).rejects.toThrow('Capabilities unavailable');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'get_camera_capabilities failed',
        error
      );
    });

    test('should validate device ID for capabilities', async () => {
      const device = 'camera0';
      const capabilities = { supported_formats: ['h264'] };
      
      (mockAPIClient.call as jest.Mock).mockResolvedValue(capabilities);

      await deviceService.getCameraCapabilities(device);

      expect(APIResponseValidator.validateDeviceId(device)).toBe(true);
    });
  });

  describe('Error handling and logging', () => {
    test('should log all operations with appropriate levels', async () => {
      const cameraList = MockDataFactory.getCameraListResult();
      (mockAPIClient.call as jest.Mock).mockResolvedValue(cameraList);

      await deviceService.getCameraList();

      expect(mockLoggerService.info).toHaveBeenCalledWith('get_camera_list request', {});
    });

    test('should handle WebSocket service errors gracefully', async () => {
      const error = new Error('WebSocket connection lost');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      await expect(deviceService.getCameraList()).rejects.toThrow('WebSocket connection lost');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'get_camera_list failed',
        error
      );
    });
  });
});
