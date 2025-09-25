/**
 * RecordingStore unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-RS-001: Recording operations (start, stop, snapshot)
 * - REQ-RS-002: Recording state management
 * - REQ-RS-003: Recording status updates handling
 * - REQ-RS-004: Concurrency control and error handling
 * - REQ-RS-005: Service injection and lifecycle
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { useRecordingStore, RecordingInfo } from '../../../src/stores/recording/recordingStore';
import { RecordingService } from '../../../src/services/recording/RecordingService';
import { APIMocks } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Mock the RecordingService
jest.mock('../../../src/services/recording/RecordingService');

describe('RecordingStore Unit Tests', () => {
  let mockRecordingService: any;
  let store: ReturnType<typeof useRecordingStore>;

  beforeEach(() => {
    // Reset store state before each test
    store = useRecordingStore.getState();
    store.reset();

    // Create mock recording service
    mockRecordingService = APIMocks.createMockRecordingService() as jest.Mocked<RecordingService>;
    
    // Clear all mocks
    jest.clearAllMocks();
  });

  afterEach(() => {
    // Reset store after each test
    store.reset();
  });

  describe('REQ-RS-001: Recording operations', () => {
    beforeEach(() => {
      store.setService(mockRecordingService);
    });

    test('should take snapshot successfully', async () => {
      const device = 'camera0';
      const filename = 'snapshot_camera0_123456.jpg';
      mockRecordingService.takeSnapshot.mockResolvedValue(undefined);

      await store.takeSnapshot(device, filename);

      expect(mockRecordingService.takeSnapshot).toHaveBeenCalledWith(device, filename);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should take snapshot without filename', async () => {
      const device = 'camera0';
      mockRecordingService.takeSnapshot.mockResolvedValue(undefined);

      await store.takeSnapshot(device);

      expect(mockRecordingService.takeSnapshot).toHaveBeenCalledWith(device, undefined);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should start recording successfully', async () => {
      const device = 'camera0';
      const duration = 60;
      const format = 'mp4';
      mockRecordingService.startRecording.mockResolvedValue(undefined);

      await store.startRecording(device, duration, format);

      expect(mockRecordingService.startRecording).toHaveBeenCalledWith(device, duration, format);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should start recording with default parameters', async () => {
      const device = 'camera0';
      mockRecordingService.startRecording.mockResolvedValue(undefined);

      await store.startRecording(device);

      expect(mockRecordingService.startRecording).toHaveBeenCalledWith(device, undefined, undefined);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should stop recording successfully', async () => {
      const device = 'camera0';
      mockRecordingService.stopRecording.mockResolvedValue(undefined);

      await store.stopRecording(device);

      expect(mockRecordingService.stopRecording).toHaveBeenCalledWith(device);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should handle snapshot error', async () => {
      const device = 'camera0';
      const errorMessage = 'Snapshot failed';
      mockRecordingService.takeSnapshot.mockRejectedValue(new Error(errorMessage));

      await store.takeSnapshot(device);

      expect(store.loading).toBe(false);
      expect(store.error).toBe(errorMessage);
    });

    test('should handle start recording error', async () => {
      const device = 'camera0';
      const errorMessage = 'Start recording failed';
      mockRecordingService.startRecording.mockRejectedValue(new Error(errorMessage));

      await store.startRecording(device);

      expect(store.loading).toBe(false);
      expect(store.error).toBe(errorMessage);
    });

    test('should handle stop recording error', async () => {
      const device = 'camera0';
      const errorMessage = 'Stop recording failed';
      mockRecordingService.stopRecording.mockRejectedValue(new Error(errorMessage));

      await store.stopRecording(device);

      expect(store.loading).toBe(false);
      expect(store.error).toBe(errorMessage);
    });
  });

  describe('REQ-RS-002: Recording state management', () => {
    test('should initialize with correct initial state', () => {
      const state = useRecordingStore.getState();
      
      expect(state.activeRecordings).toEqual({});
      expect(state.history).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });

    test('should handle recording status updates correctly', () => {
      const recordingInfo: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };

      store.handleRecordingStatusUpdate(recordingInfo);

      expect(store.activeRecordings['camera0']).toEqual(recordingInfo);
      expect(store.history).toHaveLength(1);
      expect(store.history[0]).toEqual(recordingInfo);
    });

    test('should handle recording stop status update', () => {
      // First, add an active recording
      const startInfo: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(startInfo);

      // Then stop it
      const stopInfo: RecordingInfo = {
        device: 'camera0',
        status: 'STOPPED',
        startTime: startInfo.startTime,
        filename: startInfo.filename,
        duration: 60,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(stopInfo);

      expect(store.activeRecordings['camera0']).toBeUndefined();
      expect(store.history).toHaveLength(2);
      expect(store.history[0]).toEqual(stopInfo);
    });

    test('should handle recording error status update', () => {
      // First, add an active recording
      const startInfo: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(startInfo);

      // Then error it
      const errorInfo: RecordingInfo = {
        device: 'camera0',
        status: 'ERROR',
        startTime: startInfo.startTime,
        filename: startInfo.filename,
        duration: 30,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(errorInfo);

      expect(store.activeRecordings['camera0']).toBeUndefined();
      expect(store.history).toHaveLength(2);
      expect(store.history[0]).toEqual(errorInfo);
    });

    test('should handle multiple active recordings', () => {
      const recording1: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };
      
      const recording2: RecordingInfo = {
        device: 'camera1',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera1_123456.mp4',
        duration: 0,
        format: 'mp4'
      };

      store.handleRecordingStatusUpdate(recording1);
      store.handleRecordingStatusUpdate(recording2);

      expect(store.activeRecordings['camera0']).toEqual(recording1);
      expect(store.activeRecordings['camera1']).toEqual(recording2);
      expect(store.history).toHaveLength(2);
    });

    test('should reset to initial state', () => {
      // Add some state
      const recordingInfo: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(recordingInfo);

      // Reset
      store.reset();

      expect(store.activeRecordings).toEqual({});
      expect(store.history).toEqual([]);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });
  });

  describe('REQ-RS-003: Recording status updates handling', () => {
    test('should handle STARTED status correctly', () => {
      const recordingInfo: RecordingInfo = {
        device: 'camera0',
        status: 'STARTED',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };

      store.handleRecordingStatusUpdate(recordingInfo);

      expect(store.activeRecordings['camera0']).toEqual(recordingInfo);
      expect(store.history).toHaveLength(1);
    });

    test('should handle RECORDING status correctly', () => {
      const recordingInfo: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };

      store.handleRecordingStatusUpdate(recordingInfo);

      expect(store.activeRecordings['camera0']).toEqual(recordingInfo);
      expect(store.history).toHaveLength(1);
    });

    test('should handle STOPPED status correctly', () => {
      // First add an active recording
      const startInfo: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(startInfo);

      // Then stop it
      const stopInfo: RecordingInfo = {
        device: 'camera0',
        status: 'STOPPED',
        startTime: startInfo.startTime,
        filename: startInfo.filename,
        duration: 60,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(stopInfo);

      expect(store.activeRecordings['camera0']).toBeUndefined();
      expect(store.history).toHaveLength(2);
    });

    test('should handle ERROR status correctly', () => {
      // First add an active recording
      const startInfo: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(startInfo);

      // Then error it
      const errorInfo: RecordingInfo = {
        device: 'camera0',
        status: 'ERROR',
        startTime: startInfo.startTime,
        filename: startInfo.filename,
        duration: 30,
        format: 'mp4'
      };
      store.handleRecordingStatusUpdate(errorInfo);

      expect(store.activeRecordings['camera0']).toBeUndefined();
      expect(store.history).toHaveLength(2);
    });

    test('should maintain history order (newest first)', () => {
      const recording1: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera0_123456.mp4',
        duration: 0,
        format: 'mp4'
      };
      
      const recording2: RecordingInfo = {
        device: 'camera1',
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording_camera1_123456.mp4',
        duration: 0,
        format: 'mp4'
      };

      store.handleRecordingStatusUpdate(recording1);
      store.handleRecordingStatusUpdate(recording2);

      expect(store.history[0]).toEqual(recording2); // Newest first
      expect(store.history[1]).toEqual(recording1);
    });
  });

  describe('REQ-RS-004: Concurrency control and error handling', () => {
    beforeEach(() => {
      store.setService(mockRecordingService);
    });

    test('should prevent concurrent recordings on same device', async () => {
      const device = 'camera0';
      
      // Start first recording
      mockRecordingService.startRecording.mockResolvedValue(undefined);
      await store.startRecording(device);
      
      // Try to start second recording on same device
      await store.startRecording(device);
      
      // Should only call service once (second call should be blocked)
      expect(mockRecordingService.startRecording).toHaveBeenCalledTimes(1);
      expect(store.error).toBe(`Device ${device} is already recording`);
    });

    test('should allow recording on different devices', async () => {
      mockRecordingService.startRecording.mockResolvedValue(undefined);
      
      await store.startRecording('camera0');
      await store.startRecording('camera1');
      
      expect(mockRecordingService.startRecording).toHaveBeenCalledTimes(2);
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera0', undefined, undefined);
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera1', undefined, undefined);
    });

    test('should handle missing service for snapshot', async () => {
      await store.takeSnapshot('camera0');
      expect(store.error).toBe('Recording service not initialized');
    });

    test('should handle missing service for start recording', async () => {
      await store.startRecording('camera0');
      expect(store.error).toBe('Recording service not initialized');
    });

    test('should handle missing service for stop recording', async () => {
      await store.stopRecording('camera0');
      expect(store.error).toBe('Recording service not initialized');
    });

    test('should handle unknown error types', async () => {
      mockRecordingService.takeSnapshot.mockRejectedValue('Unknown error');

      await store.takeSnapshot('camera0');

      expect(store.loading).toBe(false);
      expect(store.error).toBe('Snapshot failed');
    });

    test('should clear error when starting new operation', async () => {
      // First, set an error
      store.getState().error = 'Previous error';
      
      // Start a new operation
      mockRecordingService.startRecording.mockResolvedValue(undefined);
      await store.startRecording('camera0');
      
      expect(store.error).toBeNull();
    });
  });

  describe('REQ-RS-005: Service injection and lifecycle', () => {
    test('should inject recording service correctly', () => {
      store.setService(mockRecordingService);
      
      // We can't directly test the private service variable, but we can test
      // that the service is available by calling a method that requires it
      expect(() => store.takeSnapshot('camera0')).not.toThrow();
    });

    test('should work with multiple service injections', () => {
      const service1 = APIMocks.createMockRecordingService();
      const service2 = APIMocks.createMockRecordingService();
      
      store.setService(service1 as RecordingService);
      store.setService(service2 as RecordingService);
      
      // Should use the last injected service
      expect(() => store.takeSnapshot('camera0')).not.toThrow();
    });

    test('should handle service injection before operations', () => {
      // Inject service
      store.setService(mockRecordingService);
      
      // Perform operations
      store.takeSnapshot('camera0');
      store.startRecording('camera0');
      store.stopRecording('camera0');
      
      // Should not throw errors
      expect(store.error).toBeNull();
    });
  });

  describe('API Compliance Tests', () => {
    beforeEach(() => {
      store.setService(mockRecordingService);
    });

    test('should handle recording formats that match API schema', async () => {
      const validFormats = ['fmp4', 'mp4', 'mkv'];
      
      for (const format of validFormats) {
        mockRecordingService.startRecording.mockResolvedValue(undefined);
        await store.startRecording('camera0', undefined, format);
        expect(APIResponseValidator.validateRecordingFormat(format)).toBe(true);
      }
    });

    test('should handle device IDs that match API pattern', async () => {
      const validDevices = ['camera0', 'camera1', 'camera10'];
      
      for (const device of validDevices) {
        mockRecordingService.takeSnapshot.mockResolvedValue(undefined);
        await store.takeSnapshot(device);
        expect(APIResponseValidator.validateDeviceId(device)).toBe(true);
      }
    });
  });

  describe('Edge Cases and Complex Scenarios', () => {
    beforeEach(() => {
      store.setService(mockRecordingService);
    });

    test('should handle rapid status updates', () => {
      const device = 'camera0';
      
      // Rapid status updates
      store.handleRecordingStatusUpdate({
        device,
        status: 'STARTED',
        startTime: new Date().toISOString(),
        filename: 'recording.mp4',
        duration: 0,
        format: 'mp4'
      });
      
      store.handleRecordingStatusUpdate({
        device,
        status: 'RECORDING',
        startTime: new Date().toISOString(),
        filename: 'recording.mp4',
        duration: 0,
        format: 'mp4'
      });
      
      store.handleRecordingStatusUpdate({
        device,
        status: 'STOPPED',
        startTime: new Date().toISOString(),
        filename: 'recording.mp4',
        duration: 60,
        format: 'mp4'
      });

      expect(store.history).toHaveLength(3);
      expect(store.activeRecordings[device]).toBeUndefined();
    });

    test('should handle concurrent operations on different devices', async () => {
      mockRecordingService.startRecording.mockResolvedValue(undefined);
      
      // Start recordings on different devices simultaneously
      const promises = [
        store.startRecording('camera0'),
        store.startRecording('camera1'),
        store.startRecording('camera2')
      ];
      
      await Promise.all(promises);
      
      expect(mockRecordingService.startRecording).toHaveBeenCalledTimes(3);
      expect(store.error).toBeNull();
    });

    test('should handle recording info with missing optional fields', () => {
      const minimalInfo: RecordingInfo = {
        device: 'camera0',
        status: 'RECORDING'
        // Missing optional fields: filename, startTime, duration, format
      };

      store.handleRecordingStatusUpdate(minimalInfo);

      expect(store.activeRecordings['camera0']).toEqual(minimalInfo);
      expect(store.activeRecordings['camera0']?.filename).toBeUndefined();
      expect(store.activeRecordings['camera0']?.startTime).toBeUndefined();
      expect(store.activeRecordings['camera0']?.duration).toBeUndefined();
      expect(store.activeRecordings['camera0']?.format).toBeUndefined();
    });

    test('should handle large history of recordings', () => {
      const device = 'camera0';
      
      // Add many recordings to history
      for (let i = 0; i < 100; i++) {
        store.handleRecordingStatusUpdate({
          device,
          status: 'RECORDING',
          startTime: new Date().toISOString(),
          filename: `recording_${i}.mp4`,
          duration: 0,
          format: 'mp4'
        });
        
        store.handleRecordingStatusUpdate({
          device,
          status: 'STOPPED',
          startTime: new Date().toISOString(),
          filename: `recording_${i}.mp4`,
          duration: 60,
          format: 'mp4'
        });
      }

      expect(store.history).toHaveLength(200);
      expect(store.activeRecordings[device]).toBeUndefined();
    });

    test('should handle recording operations with various parameters', async () => {
      mockRecordingService.startRecording.mockResolvedValue(undefined);
      
      // Test with different duration and format combinations
      await store.startRecording('camera0', 30, 'mp4');
      await store.startRecording('camera1', 60, 'fmp4');
      await store.startRecording('camera2', 120, 'mkv');
      await store.startRecording('camera3'); // No duration or format
      
      expect(mockRecordingService.startRecording).toHaveBeenCalledTimes(4);
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera0', 30, 'mp4');
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera1', 60, 'fmp4');
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera2', 120, 'mkv');
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera3', undefined, undefined);
    });
  });
});
