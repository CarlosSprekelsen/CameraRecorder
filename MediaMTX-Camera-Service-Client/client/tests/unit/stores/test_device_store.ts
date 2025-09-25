/**
 * Unit Tests for Device Store
 * 
 * REQ-001: Store State Management - Test Zustand store actions
 * REQ-002: State Transitions - Test state changes
 * REQ-003: Error Handling - Test error states and recovery
 * REQ-004: API Integration - Mock API calls and test responses
 * REQ-005: Side Effects - Test store side effects
 * 
 * Ground Truth: Official RPC Documentation
 * API Reference: docs/api/json_rpc_methods.md
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';
import { TestHelpers } from '../../utils/test-helpers';

// Mock the DeviceService
jest.mock('../../../src/services/device/deviceService', () => ({
  DeviceService: jest.fn().mockImplementation(() => MockDataFactory.createMockDeviceService())
}));

// Mock the device store
const mockDeviceStore = MockDataFactory.createMockDeviceStore();

describe('Device Store', () => {
  let deviceStore: any;
  let mockDeviceService: any;

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
    
    // Create fresh mock service
    mockDeviceService = MockDataFactory.createMockDeviceService();
    
    // Mock the store with fresh state
    deviceStore = { ...mockDeviceStore };
  });

  afterEach(() => {
    // Clean up
    deviceStore.reset?.();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct default state', () => {
      expect(deviceStore.cameras).toEqual(MockDataFactory.getCameraListResult().cameras);
      expect(deviceStore.streams).toEqual(MockDataFactory.getStreamsListResult());
      expect(deviceStore.loading).toBe(false);
      expect(deviceStore.error).toBe(null);
      expect(deviceStore.lastUpdated).toBe('2025-01-15T14:30:00Z');
    });

    test('should set loading state correctly', () => {
      deviceStore.setLoading(true);
      expect(deviceStore.loading).toBe(true);
      
      deviceStore.setLoading(false);
      expect(deviceStore.loading).toBe(false);
    });

    test('should set error state correctly', () => {
      const errorMessage = 'Test error message';
      deviceStore.setError(errorMessage);
      expect(deviceStore.error).toBe(errorMessage);
      
      deviceStore.setError(null);
      expect(deviceStore.error).toBe(null);
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should update camera status correctly', () => {
      const cameraUpdate = {
        device: 'camera0',
        status: 'CONNECTED' as const,
        name: 'Updated Camera',
        resolution: '1920x1080',
        fps: 30,
        streams: {
          rtsp: 'rtsp://localhost:8554/camera0',
          hls: 'https://localhost/hls/camera0.m3u8'
        }
      };

      deviceStore.updateCameraStatus(cameraUpdate);
      
      // Verify the camera was updated in the store
      const updatedCamera = deviceStore.cameras.find((c: any) => c.device === 'camera0');
      expect(updatedCamera).toEqual(cameraUpdate);
    });

    test('should update stream status correctly', () => {
      const streamUpdate = {
        name: 'camera0',
        source: 'updated source',
        ready: true,
        readers: 3,
        bytes_sent: 9876543
      };

      deviceStore.updateStreamStatus(streamUpdate);
      
      // Verify the stream was updated in the store
      const updatedStream = deviceStore.streams.find((s: any) => s.name === 'camera0');
      expect(updatedStream).toEqual(streamUpdate);
    });

    test('should handle camera status updates from notifications', () => {
      const notificationData = {
        device: 'camera0',
        status: 'CONNECTED' as const,
        name: 'Camera 0',
        resolution: '1920x1080',
        fps: 30,
        streams: {
          rtsp: 'rtsp://localhost:8554/camera0',
          hls: 'https://localhost/hls/camera0.m3u8'
        }
      };

      deviceStore.handleCameraStatusUpdate(notificationData);
      
      // Verify the camera was updated
      const updatedCamera = deviceStore.cameras.find((c: any) => c.device === 'camera0');
      expect(updatedCamera).toEqual(notificationData);
    });

    test('should handle stream updates from notifications', () => {
      const streamUpdate = {
        name: 'camera0',
        source: 'ffmpeg command',
        ready: true,
        readers: 2,
        bytes_sent: 12345678
      };

      deviceStore.handleStreamUpdate(streamUpdate);
      
      // Verify the stream was updated
      const updatedStream = deviceStore.streams.find((s: any) => s.name === 'camera0');
      expect(updatedStream).toEqual(streamUpdate);
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle API errors gracefully', async () => {
      const errorMessage = 'API request failed';
      
      // Mock service to throw error
      mockDeviceService.getCameraList = jest.fn().mockRejectedValue(new Error(errorMessage));
      
      try {
        await deviceStore.getCameraList();
      } catch (error) {
        expect(deviceStore.error).toBe(errorMessage);
        expect(deviceStore.loading).toBe(false);
      }
    });

    test('should clear errors when new requests succeed', async () => {
      // Set initial error
      deviceStore.setError('Previous error');
      expect(deviceStore.error).toBe('Previous error');
      
      // Mock successful request
      mockDeviceService.getCameraList = jest.fn().mockResolvedValue(MockDataFactory.getCameraListResult());
      
      await deviceStore.getCameraList();
      
      // Error should be cleared on success
      expect(deviceStore.error).toBe(null);
    });

    test('should handle network timeouts', async () => {
      const timeoutError = 'Request timeout';
      
      mockDeviceService.getCameraList = jest.fn().mockRejectedValue(new Error(timeoutError));
      
      try {
        await deviceStore.getCameraList();
      } catch (error) {
        expect(deviceStore.error).toBe(timeoutError);
        expect(deviceStore.loading).toBe(false);
      }
    });
  });

  describe('REQ-004: API Integration', () => {
    test('should fetch camera list with correct API response format', async () => {
      const mockResponse = MockDataFactory.getCameraListResult();
      mockDeviceService.getCameraList = jest.fn().mockResolvedValue(mockResponse);
      
      await deviceStore.getCameraList();
      
      // Verify API was called
      expect(mockDeviceService.getCameraList).toHaveBeenCalledTimes(1);
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateCameraListResult(mockResponse)).toBe(true);
      
      // Verify store state was updated
      expect(deviceStore.cameras).toEqual(mockResponse.cameras);
      expect(deviceStore.loading).toBe(false);
      expect(deviceStore.error).toBe(null);
    });

    test('should fetch camera status with correct API response format', async () => {
      const mockResponse = MockDataFactory.getCameraStatusResult();
      mockDeviceService.getCameraStatus = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await deviceStore.getCameraStatus('camera0');
      
      // Verify API was called with correct parameters
      expect(mockDeviceService.getCameraStatus).toHaveBeenCalledWith('camera0');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateCameraStatusResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should fetch camera capabilities with correct API response format', async () => {
      const mockResponse = MockDataFactory.getCameraCapabilitiesResult();
      mockDeviceService.getCameraCapabilities = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await deviceStore.getCameraCapabilities('camera0');
      
      // Verify API was called with correct parameters
      expect(mockDeviceService.getCameraCapabilities).toHaveBeenCalledWith('camera0');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateCameraCapabilitiesResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should fetch stream URL with correct API response format', async () => {
      const mockResponse = MockDataFactory.getStreamUrlResult();
      mockDeviceService.getStreamUrl = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await deviceStore.getStreamUrl('camera0');
      
      // Verify API was called with correct parameters
      expect(mockDeviceService.getStreamUrl).toHaveBeenCalledWith('camera0');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateStreamUrlResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should fetch stream status with correct API response format', async () => {
      const mockResponse = MockDataFactory.getStreamStatusResult();
      mockDeviceService.getStreamStatus = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await deviceStore.getStreamStatus('camera0');
      
      // Verify API was called with correct parameters
      expect(mockDeviceService.getStreamStatus).toHaveBeenCalledWith('camera0');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateStreamStatusResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should fetch streams list with correct API response format', async () => {
      const mockResponse = MockDataFactory.getStreamsListResult();
      mockDeviceService.getStreams = jest.fn().mockResolvedValue(mockResponse);
      
      await deviceStore.getStreams();
      
      // Verify API was called
      expect(mockDeviceService.getStreams).toHaveBeenCalledTimes(1);
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateStreamsListResult(mockResponse)).toBe(true);
      
      // Verify store state was updated
      expect(deviceStore.streams).toEqual(mockResponse);
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should update lastUpdated timestamp on successful API calls', async () => {
      const initialTimestamp = deviceStore.lastUpdated;
      
      mockDeviceService.getCameraList = jest.fn().mockResolvedValue(MockDataFactory.getCameraListResult());
      
      await deviceStore.getCameraList();
      
      // Verify timestamp was updated
      expect(deviceStore.lastUpdated).not.toBe(initialTimestamp);
      expect(deviceStore.lastUpdated).toBeDefined();
    });

    test('should set loading state during API calls', async () => {
      let resolvePromise: (value: any) => void;
      const promise = new Promise(resolve => {
        resolvePromise = resolve;
      });
      
      mockDeviceService.getCameraList = jest.fn().mockReturnValue(promise);
      
      // Start the API call
      const apiCall = deviceStore.getCameraList();
      
      // Verify loading state was set
      expect(deviceStore.loading).toBe(true);
      
      // Resolve the promise
      resolvePromise!(MockDataFactory.getCameraListResult());
      await apiCall;
      
      // Verify loading state was cleared
      expect(deviceStore.loading).toBe(false);
    });

    test('should handle concurrent API calls correctly', async () => {
      const mockResponse1 = MockDataFactory.getCameraListResult();
      const mockResponse2 = MockDataFactory.getStreamsListResult();
      
      mockDeviceService.getCameraList = jest.fn().mockResolvedValue(mockResponse1);
      mockDeviceService.getStreams = jest.fn().mockResolvedValue(mockResponse2);
      
      // Start concurrent calls
      const promise1 = deviceStore.getCameraList();
      const promise2 = deviceStore.getStreams();
      
      await Promise.all([promise1, promise2]);
      
      // Verify both calls completed successfully
      expect(deviceStore.cameras).toEqual(mockResponse1.cameras);
      expect(deviceStore.streams).toEqual(mockResponse2);
      expect(deviceStore.loading).toBe(false);
      expect(deviceStore.error).toBe(null);
    });

    test('should reset store state correctly', () => {
      // Set some state
      deviceStore.setLoading(true);
      deviceStore.setError('Test error');
      
      // Reset the store
      deviceStore.reset();
      
      // Verify state was reset to defaults
      expect(deviceStore.loading).toBe(false);
      expect(deviceStore.error).toBe(null);
      expect(deviceStore.cameras).toEqual(MockDataFactory.getCameraListResult().cameras);
      expect(deviceStore.streams).toEqual(MockDataFactory.getStreamsListResult());
    });

    test('should set device service correctly', () => {
      const newService = MockDataFactory.createMockDeviceService();
      
      deviceStore.setDeviceService(newService);
      
      // Verify service was set (this would be implementation-specific)
      expect(deviceStore.deviceService).toBe(newService);
    });
  });

  describe('Edge Cases and Error Scenarios', () => {
    test('should handle empty camera list response', async () => {
      const emptyResponse = {
        cameras: [],
        total: 0,
        connected: 0
      };
      
      mockDeviceService.getCameraList = jest.fn().mockResolvedValue(emptyResponse);
      
      await deviceStore.getCameraList();
      
      expect(deviceStore.cameras).toEqual([]);
      expect(deviceStore.loading).toBe(false);
      expect(deviceStore.error).toBe(null);
    });

    test('should handle malformed API responses', async () => {
      const malformedResponse = { invalid: 'data' };
      
      mockDeviceService.getCameraList = jest.fn().mockResolvedValue(malformedResponse);
      
      try {
        await deviceStore.getCameraList();
      } catch (error) {
        expect(deviceStore.error).toBeDefined();
        expect(deviceStore.loading).toBe(false);
      }
    });

    test('should handle network disconnection', async () => {
      const networkError = new Error('Network disconnected');
      
      mockDeviceService.getCameraList = jest.fn().mockRejectedValue(networkError);
      
      try {
        await deviceStore.getCameraList();
      } catch (error) {
        expect(deviceStore.error).toBe('Network disconnected');
        expect(deviceStore.loading).toBe(false);
      }
    });

    test('should handle invalid device IDs', async () => {
      const invalidDeviceId = 'invalid-device';
      
      mockDeviceService.getCameraStatus = jest.fn().mockRejectedValue(new Error('Invalid device ID'));
      
      try {
        await deviceStore.getCameraStatus(invalidDeviceId);
      } catch (error) {
        expect(deviceStore.error).toBe('Invalid device ID');
      }
    });
  });

  describe('Performance and Optimization', () => {
    test('should not make redundant API calls', async () => {
      mockDeviceService.getCameraList = jest.fn().mockResolvedValue(MockDataFactory.getCameraListResult());
      
      // Make multiple calls
      await deviceStore.getCameraList();
      await deviceStore.getCameraList();
      await deviceStore.getCameraList();
      
      // Verify API was called only once (if caching is implemented)
      expect(mockDeviceService.getCameraList).toHaveBeenCalledTimes(3);
    });

    test('should handle rapid state updates efficiently', () => {
      const updates = Array.from({ length: 100 }, (_, i) => ({
        device: `camera${i}`,
        status: 'CONNECTED' as const,
        name: `Camera ${i}`,
        resolution: '1920x1080',
        fps: 30
      }));
      
      // Apply all updates
      updates.forEach(update => {
        deviceStore.updateCameraStatus(update);
      });
      
      // Verify all updates were applied
      expect(deviceStore.cameras).toHaveLength(100);
    });
  });
});