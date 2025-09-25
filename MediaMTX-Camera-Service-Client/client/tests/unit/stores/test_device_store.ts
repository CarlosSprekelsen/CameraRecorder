/**
 * DeviceStore unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-DS-001: Device state management
 * - REQ-DS-002: Device operations (list, select, update)
 * - REQ-DS-003: Error handling and state recovery
 * - REQ-DS-004: Real-time updates handling
 * - REQ-DS-005: Service injection and lifecycle
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { useDeviceStore, Camera, StreamInfo } from '../../../src/stores/device/deviceStore';
import { DeviceService } from '../../../src/services/device/DeviceService';
import { APIMocks } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Mock the DeviceService
jest.mock('../../../src/services/device/DeviceService');

describe('DeviceStore Unit Tests', () => {
  let mockDeviceService: any;
  let store: ReturnType<typeof useDeviceStore>;

  beforeEach(() => {
    // Reset store state before each test
    store = useDeviceStore.getState();
    store.reset();

    // Create mock device service
    mockDeviceService = APIMocks.createMockDeviceService() as jest.Mocked<DeviceService>;
    
    // Clear all mocks
    jest.clearAllMocks();
  });

  afterEach(() => {
    // Reset store after each test
    store.reset();
  });

  describe('REQ-DS-001: Device state management', () => {
    test('should initialize with correct initial state', () => {
      const state = useDeviceStore.getState();
      
      expect(state.cameras).toEqual([]);
      expect(state.streams).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastUpdated).toBeNull();
    });

    test('should set loading state correctly', () => {
      store.setLoading(true);
      expect(store.loading).toBe(true);
      
      store.setLoading(false);
      expect(store.loading).toBe(false);
    });

    test('should set error state correctly', () => {
      const errorMessage = 'Test error message';
      store.setError(errorMessage);
      expect(store.error).toBe(errorMessage);
      
      store.setError(null);
      expect(store.error).toBeNull();
    });

    test('should reset to initial state', () => {
      // Set some state
      store.setLoading(true);
      store.setError('Test error');
      
      // Reset
      store.reset();
      
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.cameras).toEqual([]);
      expect(store.streams).toEqual([]);
    });
  });

  describe('REQ-DS-002: Device operations', () => {
    beforeEach(() => {
      // Set up device service for tests that need it
      store.setDeviceService(mockDeviceService);
    });

    test('should get camera list successfully', async () => {
      const mockCameraList = APIMocks.getCameraListResult();
      mockDeviceService.getCameraList.mockResolvedValue(mockCameraList.cameras);

      await store.getCameraList();

      expect(mockDeviceService.getCameraList).toHaveBeenCalledTimes(1);
      expect(store.cameras).toEqual(mockCameraList.cameras);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.lastUpdated).toBeTruthy();
      expect(APIResponseValidator.validateCameraListResult({ 
        cameras: store.cameras, 
        total: store.cameras.length, 
        connected: store.cameras.filter(c => c.status === 'CONNECTED').length 
      })).toBe(true);
    });

    test('should handle camera list error', async () => {
      const errorMessage = 'Failed to get camera list';
      mockDeviceService.getCameraList.mockRejectedValue(new Error(errorMessage));

      await store.getCameraList();

      expect(store.loading).toBe(false);
      expect(store.error).toBe(errorMessage);
      expect(store.cameras).toEqual([]);
    });

    test('should get stream URL successfully', async () => {
      const device = 'camera0';
      const expectedUrl = 'rtsp://localhost:8554/camera0';
      mockDeviceService.getStreamUrl.mockResolvedValue(expectedUrl);

      const result = await store.getStreamUrl(device);

      expect(mockDeviceService.getStreamUrl).toHaveBeenCalledWith(device);
      expect(result).toBe(expectedUrl);
      expect(store.error).toBeNull();
    });

    test('should handle stream URL error', async () => {
      const device = 'camera0';
      const errorMessage = 'Failed to get stream URL';
      mockDeviceService.getStreamUrl.mockRejectedValue(new Error(errorMessage));

      const result = await store.getStreamUrl(device);

      expect(result).toBeNull();
      expect(store.error).toBe(errorMessage);
    });

    test('should get streams successfully', async () => {
      const mockStreams: StreamInfo[] = [
        {
          name: 'camera0',
          source: 'rtsp://localhost:8554/camera0',
          ready: true,
          readers: 2,
          bytes_sent: 1024000
        }
      ];
      mockDeviceService.getStreams.mockResolvedValue(mockStreams);

      await store.getStreams();

      expect(mockDeviceService.getStreams).toHaveBeenCalledTimes(1);
      expect(store.streams).toEqual(mockStreams);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.lastUpdated).toBeTruthy();
    });

    test('should handle streams error', async () => {
      const errorMessage = 'Failed to get streams';
      mockDeviceService.getStreams.mockRejectedValue(new Error(errorMessage));

      await store.getStreams();

      expect(store.loading).toBe(false);
      expect(store.error).toBe(errorMessage);
      expect(store.streams).toEqual([]);
    });
  });

  describe('REQ-DS-003: Error handling and state recovery', () => {
    test('should handle missing device service for getCameraList', async () => {
      await store.getCameraList();
      
      expect(store.error).toBe('Device service not initialized');
      expect(store.cameras).toEqual([]);
    });

    test('should handle missing device service for getStreamUrl', async () => {
      const result = await store.getStreamUrl('camera0');
      
      expect(result).toBeNull();
      expect(store.error).toBe('Device service not initialized');
    });

    test('should handle missing device service for getStreams', async () => {
      await store.getStreams();
      
      expect(store.error).toBe('Device service not initialized');
      expect(store.streams).toEqual([]);
    });

    test('should handle unknown error types', async () => {
      store.setDeviceService(mockDeviceService);
      mockDeviceService.getCameraList.mockRejectedValue('Unknown error');

      await store.getCameraList();

      expect(store.loading).toBe(false);
      expect(store.error).toBe('Failed to get camera list');
    });
  });

  describe('REQ-DS-004: Real-time updates handling', () => {
    test('should update camera status correctly', () => {
      const initialCameras: Camera[] = [
        APIMocks.getCamera('camera0'),
        APIMocks.getCamera('camera1')
      ];
      
      // Set initial cameras
      store.getState().cameras = initialCameras;
      
      // Update camera0 status
      store.updateCameraStatus('camera0', 'ERROR');
      
      const updatedCameras = store.getState().cameras;
      expect(updatedCameras[0].status).toBe('ERROR');
      expect(updatedCameras[1].status).toBe('CONNECTED'); // camera1 unchanged
    });

    test('should handle camera status update for new camera', () => {
      const newCamera: Camera = {
        device: 'camera2',
        status: 'CONNECTED',
        name: 'New Camera',
        resolution: '1920x1080',
        fps: 30,
        streams: {
          rtsp: 'rtsp://localhost:8554/camera2',
          hls: 'https://localhost/hls/camera2.m3u8'
        }
      };

      store.handleCameraStatusUpdate(newCamera);

      expect(store.cameras).toHaveLength(1);
      expect(store.cameras[0]).toEqual(newCamera);
    });

    test('should handle camera status update for existing camera', () => {
      const initialCamera: Camera = APIMocks.getCamera('camera0');
      store.getState().cameras = [initialCamera];

      const updatedCamera: Camera = {
        ...initialCamera,
        status: 'ERROR',
        name: 'Updated Camera Name'
      };

      store.handleCameraStatusUpdate(updatedCamera);

      expect(store.cameras).toHaveLength(1);
      expect(store.cameras[0]).toEqual(updatedCamera);
    });

    test('should update stream status correctly', () => {
      const initialStreams: StreamInfo[] = [
        {
          name: 'camera0',
          source: 'rtsp://localhost:8554/camera0',
          ready: true,
          readers: 2,
          bytes_sent: 1024000
        }
      ];
      
      store.getState().streams = initialStreams;
      
      store.updateStreamStatus('camera0', false, 0);
      
      const updatedStreams = store.getState().streams;
      expect(updatedStreams[0].ready).toBe(false);
      expect(updatedStreams[0].readers).toBe(0);
    });

    test('should handle stream update for new stream', () => {
      const newStream: StreamInfo = {
        name: 'camera1',
        source: 'rtsp://localhost:8554/camera1',
        ready: true,
        readers: 1,
        bytes_sent: 512000
      };

      store.handleStreamUpdate(newStream);

      expect(store.streams).toHaveLength(1);
      expect(store.streams[0]).toEqual(newStream);
    });

    test('should handle stream update for existing stream', () => {
      const initialStream: StreamInfo = {
        name: 'camera0',
        source: 'rtsp://localhost:8554/camera0',
        ready: true,
        readers: 2,
        bytes_sent: 1024000
      };
      
      store.getState().streams = [initialStream];

      const updatedStream: StreamInfo = {
        ...initialStream,
        ready: false,
        readers: 0,
        bytes_sent: 0
      };

      store.handleStreamUpdate(updatedStream);

      expect(store.streams).toHaveLength(1);
      expect(store.streams[0]).toEqual(updatedStream);
    });
  });

  describe('REQ-DS-005: Service injection and lifecycle', () => {
    test('should inject device service correctly', () => {
      store.setDeviceService(mockDeviceService);
      
      // We can't directly test the private service variable, but we can test
      // that the service is available by calling a method that requires it
      expect(() => store.getCameraList()).not.toThrow();
    });

    test('should work with multiple service injections', () => {
      const service1 = APIMocks.createMockDeviceService();
      const service2 = APIMocks.createMockDeviceService();
      
      store.setDeviceService(service1 as DeviceService);
      store.setDeviceService(service2 as DeviceService);
      
      // Should use the last injected service
      expect(() => store.getCameraList()).not.toThrow();
    });
  });

  describe('API Compliance Tests', () => {
    beforeEach(() => {
      store.setDeviceService(mockDeviceService);
    });

    test('should return cameras that match API schema', async () => {
      const mockCameras = APIMocks.getCameraListResult().cameras;
      mockDeviceService.getCameraList.mockResolvedValue(mockCameras);

      await store.getCameraList();

      store.cameras.forEach(camera => {
        expect(APIResponseValidator.validateCamera(camera)).toBe(true);
      });
    });

    test('should handle device IDs that match API pattern', () => {
      const validDevices = ['camera0', 'camera1', 'camera10'];
      
      validDevices.forEach(device => {
        expect(APIResponseValidator.validateDeviceId(device)).toBe(true);
      });
    });

    test('should handle stream URLs that match API format', async () => {
      const device = 'camera0';
      const expectedUrl = 'rtsp://localhost:8554/camera0';
      mockDeviceService.getStreamUrl.mockResolvedValue(expectedUrl);

      const result = await store.getStreamUrl(device);

      expect(APIResponseValidator.validateStreamUrl(result!)).toBe(true);
    });
  });

  describe('Edge Cases and Error Scenarios', () => {
    beforeEach(() => {
      store.setDeviceService(mockDeviceService);
    });

    test('should handle empty camera list', async () => {
      mockDeviceService.getCameraList.mockResolvedValue([]);

      await store.getCameraList();

      expect(store.cameras).toEqual([]);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should handle empty streams list', async () => {
      mockDeviceService.getStreams.mockResolvedValue([]);

      await store.getStreams();

      expect(store.streams).toEqual([]);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should handle null stream URL response', async () => {
      const device = 'camera0';
      mockDeviceService.getStreamUrl.mockResolvedValue(null);

      const result = await store.getStreamUrl(device);

      expect(result).toBeNull();
      expect(store.error).toBeNull();
    });

    test('should handle concurrent camera list requests', async () => {
      const mockCameras = APIMocks.getCameraListResult().cameras;
      mockDeviceService.getCameraList.mockResolvedValue(mockCameras);

      // Start multiple concurrent requests
      const promises = [
        store.getCameraList(),
        store.getCameraList(),
        store.getCameraList()
      ];

      await Promise.all(promises);

      // Should have been called multiple times
      expect(mockDeviceService.getCameraList).toHaveBeenCalledTimes(3);
      expect(store.cameras).toEqual(mockCameras);
    });

    test('should handle rapid status updates', () => {
      const camera: Camera = APIMocks.getCamera('camera0');
      
      // Rapid status updates
      store.handleCameraStatusUpdate({ ...camera, status: 'CONNECTED' });
      store.handleCameraStatusUpdate({ ...camera, status: 'DISCONNECTED' });
      store.handleCameraStatusUpdate({ ...camera, status: 'ERROR' });

      expect(store.cameras).toHaveLength(1);
      expect(store.cameras[0].status).toBe('ERROR');
    });
  });
});
