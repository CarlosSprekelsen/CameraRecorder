/**
 * DeviceStore Coverage Tests - targeting uncovered lines
 * 
 * Focus: Lines 104-105,112-113,119-120,133,164-165,172-173
 * Coverage Target: Increase DeviceStore from 77.08% to 80%+
 */

import { useDeviceStore } from '../../../src/stores/device/deviceStore';
import { DeviceService } from '../../../src/services/device/DeviceService';
import { MockDataFactory } from '../../utils/mocks';

// Mock DeviceService
const mockDeviceService = MockDataFactory.createMockDeviceService();

describe('DeviceStore Coverage Tests', () => {
  beforeEach(() => {
    // Reset the store to initial state
    useDeviceStore.getState().reset();
    
    // Reset all mocks
    jest.clearAllMocks();
    
    // Set up default mock implementations
    const cameraListResult = MockDataFactory.getCameraListResult();
    mockDeviceService.getCameraList.mockResolvedValue(cameraListResult.cameras);
    mockDeviceService.getStreamUrl.mockResolvedValue('rtsp://localhost:8554/camera0');
    mockDeviceService.getStreams.mockResolvedValue(MockDataFactory.getStreamsListResult());
  });

  afterEach(() => {
    useDeviceStore.getState().reset();
  });

  describe('Coverage: DeviceStore uncovered lines', () => {
    test('should handle getCameraList with service error - line 104-105', async () => {
      const { getCameraList, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service
      setDeviceService(mockDeviceService as any);
      
      // Mock service to throw error
      mockDeviceService.getCameraList.mockRejectedValue(new Error('Service error'));
      
      // Start the action
      await getCameraList();
      
      // Check error state
      const state = useDeviceStore.getState();
      expect(state.error).toBe('Service error');
      expect(state.loading).toBe(false);
    });

    test('should handle getStreamUrl with service error - line 112-113', async () => {
      const { getStreamUrl, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service
      setDeviceService(mockDeviceService as any);
      
      // Mock service to throw error
      mockDeviceService.getStreamUrl.mockRejectedValue(new Error('Stream URL error'));
      
      // Start the action
      const result = await getStreamUrl('camera0');
      
      // Check error state and return value
      expect(result).toBeNull();
      const state = useDeviceStore.getState();
      expect(state.error).toBe('Stream URL error');
    });

    test('should handle getStreams with service error - line 119-120', async () => {
      const { getStreams, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service
      setDeviceService(mockDeviceService as any);
      
      // Mock service to throw error
      mockDeviceService.getStreams.mockRejectedValue(new Error('Streams error'));
      
      // Start the action
      await getStreams();
      
      // Check error state
      const state = useDeviceStore.getState();
      expect(state.error).toBe('Streams error');
      expect(state.loading).toBe(false);
    });

    test('should handle setLoading and setError methods', () => {
      const { setLoading, setError } = useDeviceStore.getState();
      
      // Test setLoading
      setLoading(true);
      expect(useDeviceStore.getState().loading).toBe(true);
      
      setLoading(false);
      expect(useDeviceStore.getState().loading).toBe(false);
      
      // Test setError
      setError('Test error');
      expect(useDeviceStore.getState().error).toBe('Test error');
      
      setError(null);
      expect(useDeviceStore.getState().error).toBeNull();
    });

    test('should handle handleCameraStatusUpdate with existing camera - line 164-165', () => {
      const { handleCameraStatusUpdate, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service
      setDeviceService(mockDeviceService as any);
      
      // Set initial cameras
      useDeviceStore.setState({
        cameras: [
          {
            device: 'camera0',
            status: 'offline',
            name: 'Camera 0',
            resolution: '1920x1080',
            fps: 30,
            streams: { rtsp: '', hls: '' }
          }
        ]
      });
      
      // Update existing camera
      const updatedCamera = {
        device: 'camera0',
        status: 'online',
        name: 'Camera 0 Updated',
        resolution: '1920x1080',
        fps: 30,
        streams: { rtsp: 'rtsp://localhost:8554/camera0', hls: '' }
      };
      
      handleCameraStatusUpdate(updatedCamera);
      
      // Check that camera was updated
      const state = useDeviceStore.getState();
      expect(state.cameras[0].status).toBe('online');
      expect(state.cameras[0].name).toBe('Camera 0 Updated');
    });

    test('should handle handleStreamUpdate with existing stream - line 172-173', () => {
      const { handleStreamUpdate, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service
      setDeviceService(mockDeviceService as any);
      
      // Set initial streams
      useDeviceStore.setState({
        streams: [
          {
            name: 'camera0',
            source: 'ffmpeg -f v4l2 -i /dev/video0',
            ready: false,
            readers: 0,
            bytes_sent: 0
          }
        ]
      });
      
      // Update existing stream
      const updatedStream = {
        name: 'camera0',
        source: 'ffmpeg -f v4l2 -i /dev/video0',
        ready: true,
        readers: 2,
        bytes_sent: 12345678
      };
      
      handleStreamUpdate(updatedStream);
      
      // Check that stream was updated
      const state = useDeviceStore.getState();
      expect(state.streams[0].ready).toBe(true);
      expect(state.streams[0].readers).toBe(2);
    });
  });
});
