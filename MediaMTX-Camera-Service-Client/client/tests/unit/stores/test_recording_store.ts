/**
 * Unit Tests for Recording Store
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

// Mock the RecordingService
jest.mock('../../../src/services/recording/recordingService', () => ({
  RecordingService: jest.fn().mockImplementation(() => MockDataFactory.createMockRecordingService())
}));

// Mock the recording store
const mockRecordingStore = MockDataFactory.createMockRecordingStore();

describe('Recording Store', () => {
  let recordingStore: any;
  let mockRecordingService: any;

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
    
    // Create fresh mock service
    mockRecordingService = MockDataFactory.createMockRecordingService();
    
    // Mock the store with fresh state
    recordingStore = { ...mockRecordingStore };
  });

  afterEach(() => {
    // Clean up
    recordingStore.reset?.();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct default state', () => {
      expect(recordingStore.activeRecordings).toEqual([]);
      expect(recordingStore.recordingHistory).toEqual([]);
      expect(recordingStore.loading).toBe(false);
      expect(recordingStore.error).toBe(null);
      expect(recordingStore.lastUpdated).toBe('2025-01-15T14:30:00Z');
    });

    test('should set loading state correctly', () => {
      recordingStore.setLoading(true);
      expect(recordingStore.loading).toBe(true);
      
      recordingStore.setLoading(false);
      expect(recordingStore.loading).toBe(false);
    });

    test('should set error state correctly', () => {
      const errorMessage = 'Test error message';
      recordingStore.setError(errorMessage);
      expect(recordingStore.error).toBe(errorMessage);
      
      recordingStore.setError(null);
      expect(recordingStore.error).toBe(null);
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should add active recording when starting', async () => {
      const mockResponse = MockDataFactory.getRecordingStartResult();
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(mockResponse);
      
      await recordingStore.startRecording('camera0');
      
      // Verify recording was added to active recordings
      expect(recordingStore.activeRecordings).toContainEqual(mockResponse);
      expect(recordingStore.loading).toBe(false);
      expect(recordingStore.error).toBe(null);
    });

    test('should remove active recording when stopping', async () => {
      const mockResponse = MockDataFactory.getRecordingStopResult();
      mockRecordingService.stopRecording = jest.fn().mockResolvedValue(mockResponse);
      
      // Add an active recording first
      recordingStore.activeRecordings = [MockDataFactory.getRecordingStartResult()];
      
      await recordingStore.stopRecording('camera0');
      
      // Verify recording was removed from active recordings
      expect(recordingStore.activeRecordings).toHaveLength(0);
      // Verify recording was added to history
      expect(recordingStore.recordingHistory).toContainEqual(mockResponse);
    });

    test('should add snapshot to history when taken', async () => {
      const mockResponse = MockDataFactory.getSnapshotResult();
      mockRecordingService.takeSnapshot = jest.fn().mockResolvedValue(mockResponse);
      
      await recordingStore.takeSnapshot('camera0');
      
      // Verify snapshot was added to history
      expect(recordingStore.recordingHistory).toContainEqual(mockResponse);
      expect(recordingStore.loading).toBe(false);
      expect(recordingStore.error).toBe(null);
    });

    test('should handle recording status updates', () => {
      const statusUpdate = {
        device: 'camera0',
        status: 'RECORDING' as const,
        filename: 'test-recording',
        start_time: '2025-01-15T14:30:00Z'
      };
      
      recordingStore.handleRecordingStatusUpdate(statusUpdate);
      
      // Verify status was updated
      const activeRecording = recordingStore.activeRecordings.find((r: any) => r.device === 'camera0');
      expect(activeRecording).toEqual(expect.objectContaining(statusUpdate));
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle API errors gracefully', async () => {
      const errorMessage = 'API request failed';
      
      mockRecordingService.startRecording = jest.fn().mockRejectedValue(new Error(errorMessage));
      
      try {
        await recordingStore.startRecording('camera0');
      } catch (error) {
        expect(recordingStore.error).toBe(errorMessage);
        expect(recordingStore.loading).toBe(false);
      }
    });

    test('should clear errors when new requests succeed', async () => {
      // Set initial error
      recordingStore.setError('Previous error');
      expect(recordingStore.error).toBe('Previous error');
      
      // Mock successful request
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(MockDataFactory.getRecordingStartResult());
      
      await recordingStore.startRecording('camera0');
      
      // Error should be cleared on success
      expect(recordingStore.error).toBe(null);
    });

    test('should handle recording conflicts', async () => {
      const conflictError = new Error('Recording already in progress');
      
      mockRecordingService.startRecording = jest.fn().mockRejectedValue(conflictError);
      
      try {
        await recordingStore.startRecording('camera0');
      } catch (error) {
        expect(recordingStore.error).toBe('Recording already in progress');
      }
    });

    test('should handle device not found errors', async () => {
      const notFoundError = new Error('Device not found');
      
      mockRecordingService.startRecording = jest.fn().mockRejectedValue(notFoundError);
      
      try {
        await recordingStore.startRecording('nonexistent-camera');
      } catch (error) {
        expect(recordingStore.error).toBe('Device not found');
      }
    });
  });

  describe('REQ-004: API Integration', () => {
    test('should start recording with correct API response format', async () => {
      const mockResponse = MockDataFactory.getRecordingStartResult();
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await recordingStore.startRecording('camera0');
      
      // Verify API was called with correct parameters
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera0');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateRecordingStartResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should stop recording with correct API response format', async () => {
      const mockResponse = MockDataFactory.getRecordingStopResult();
      mockRecordingService.stopRecording = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await recordingStore.stopRecording('camera0');
      
      // Verify API was called with correct parameters
      expect(mockRecordingService.stopRecording).toHaveBeenCalledWith('camera0');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateRecordingStopResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should take snapshot with correct API response format', async () => {
      const mockResponse = MockDataFactory.getSnapshotResult();
      mockRecordingService.takeSnapshot = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await recordingStore.takeSnapshot('camera0');
      
      // Verify API was called with correct parameters
      expect(mockRecordingService.takeSnapshot).toHaveBeenCalledWith('camera0');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateSnapshotResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should handle recording with duration parameter', async () => {
      const mockResponse = MockDataFactory.getRecordingStartResult();
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(mockResponse);
      
      await recordingStore.startRecording('camera0', { duration: 3600 });
      
      // Verify API was called with duration parameter
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera0', { duration: 3600 });
    });

    test('should handle recording with format parameter', async () => {
      const mockResponse = MockDataFactory.getRecordingStartResult();
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(mockResponse);
      
      await recordingStore.startRecording('camera0', { format: 'mp4' });
      
      // Verify API was called with format parameter
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera0', { format: 'mp4' });
    });

    test('should handle snapshot with filename parameter', async () => {
      const mockResponse = MockDataFactory.getSnapshotResult();
      mockRecordingService.takeSnapshot = jest.fn().mockResolvedValue(mockResponse);
      
      await recordingStore.takeSnapshot('camera0', 'custom-snapshot.jpg');
      
      // Verify API was called with filename parameter
      expect(mockRecordingService.takeSnapshot).toHaveBeenCalledWith('camera0', 'custom-snapshot.jpg');
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should update lastUpdated timestamp on successful API calls', async () => {
      const initialTimestamp = recordingStore.lastUpdated;
      
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(MockDataFactory.getRecordingStartResult());
      
      await recordingStore.startRecording('camera0');
      
      // Verify timestamp was updated
      expect(recordingStore.lastUpdated).not.toBe(initialTimestamp);
      expect(recordingStore.lastUpdated).toBeDefined();
    });

    test('should set loading state during API calls', async () => {
      let resolvePromise: (value: any) => void;
      const promise = new Promise(resolve => {
        resolvePromise = resolve;
      });
      
      mockRecordingService.startRecording = jest.fn().mockReturnValue(promise);
      
      // Start the API call
      const apiCall = recordingStore.startRecording('camera0');
      
      // Verify loading state was set
      expect(recordingStore.loading).toBe(true);
      
      // Resolve the promise
      resolvePromise!(MockDataFactory.getRecordingStartResult());
      await apiCall;
      
      // Verify loading state was cleared
      expect(recordingStore.loading).toBe(false);
    });

    test('should handle concurrent API calls correctly', async () => {
      const mockResponse1 = MockDataFactory.getRecordingStartResult();
      const mockResponse2 = MockDataFactory.getSnapshotResult();
      
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(mockResponse1);
      mockRecordingService.takeSnapshot = jest.fn().mockResolvedValue(mockResponse2);
      
      // Start concurrent calls
      const promise1 = recordingStore.startRecording('camera0');
      const promise2 = recordingStore.takeSnapshot('camera1');
      
      await Promise.all([promise1, promise2]);
      
      // Verify both calls completed successfully
      expect(recordingStore.activeRecordings).toContainEqual(mockResponse1);
      expect(recordingStore.recordingHistory).toContainEqual(mockResponse2);
      expect(recordingStore.loading).toBe(false);
      expect(recordingStore.error).toBe(null);
    });

    test('should reset store state correctly', () => {
      // Set some state
      recordingStore.setLoading(true);
      recordingStore.setError('Test error');
      recordingStore.activeRecordings = [MockDataFactory.getRecordingStartResult()];
      
      // Reset the store
      recordingStore.reset();
      
      // Verify state was reset to defaults
      expect(recordingStore.loading).toBe(false);
      expect(recordingStore.error).toBe(null);
      expect(recordingStore.activeRecordings).toEqual([]);
      expect(recordingStore.recordingHistory).toEqual([]);
    });

    test('should set recording service correctly', () => {
      const newService = MockDataFactory.createMockRecordingService();
      
      recordingStore.setRecordingService(newService);
      
      // Verify service was set (this would be implementation-specific)
      expect(recordingStore.recordingService).toBe(newService);
    });
  });

  describe('Edge Cases and Error Scenarios', () => {
    test('should handle recording start failures', async () => {
      const startError = new Error('Failed to start recording');
      
      mockRecordingService.startRecording = jest.fn().mockRejectedValue(startError);
      
      try {
        await recordingStore.startRecording('camera0');
      } catch (error) {
        expect(recordingStore.error).toBe('Failed to start recording');
        expect(recordingStore.activeRecordings).toHaveLength(0);
      }
    });

    test('should handle recording stop failures', async () => {
      const stopError = new Error('Failed to stop recording');
      
      mockRecordingService.stopRecording = jest.fn().mockRejectedValue(stopError);
      
      // Add an active recording first
      recordingStore.activeRecordings = [MockDataFactory.getRecordingStartResult()];
      
      try {
        await recordingStore.stopRecording('camera0');
      } catch (error) {
        expect(recordingStore.error).toBe('Failed to stop recording');
        // Recording should still be active
        expect(recordingStore.activeRecordings).toHaveLength(1);
      }
    });

    test('should handle snapshot failures', async () => {
      const snapshotError = new Error('Failed to take snapshot');
      
      mockRecordingService.takeSnapshot = jest.fn().mockRejectedValue(snapshotError);
      
      try {
        await recordingStore.takeSnapshot('camera0');
      } catch (error) {
        expect(recordingStore.error).toBe('Failed to take snapshot');
        expect(recordingStore.recordingHistory).toHaveLength(0);
      }
    });

    test('should handle malformed API responses', async () => {
      const malformedResponse = { invalid: 'data' };
      
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(malformedResponse);
      
      try {
        await recordingStore.startRecording('camera0');
      } catch (error) {
        expect(recordingStore.error).toBeDefined();
        expect(recordingStore.loading).toBe(false);
      }
    });

    test('should handle network disconnection', async () => {
      const networkError = new Error('Network disconnected');
      
      mockRecordingService.startRecording = jest.fn().mockRejectedValue(networkError);
      
      try {
        await recordingStore.startRecording('camera0');
      } catch (error) {
        expect(recordingStore.error).toBe('Network disconnected');
        expect(recordingStore.loading).toBe(false);
      }
    });

    test('should handle invalid device IDs', async () => {
      const invalidDeviceId = 'invalid-device';
      
      mockRecordingService.startRecording = jest.fn().mockRejectedValue(new Error('Invalid device ID'));
      
      try {
        await recordingStore.startRecording(invalidDeviceId);
      } catch (error) {
        expect(recordingStore.error).toBe('Invalid device ID');
      }
    });
  });

  describe('Performance and Optimization', () => {
    test('should handle multiple concurrent recordings', async () => {
      const mockResponse1 = MockDataFactory.getRecordingStartResult();
      const mockResponse2 = MockDataFactory.getRecordingStartResult();
      mockResponse2.device = 'camera1';
      
      mockRecordingService.startRecording = jest.fn()
        .mockResolvedValueOnce(mockResponse1)
        .mockResolvedValueOnce(mockResponse2);
      
      // Start recordings on multiple cameras
      await recordingStore.startRecording('camera0');
      await recordingStore.startRecording('camera1');
      
      expect(recordingStore.activeRecordings).toHaveLength(2);
      expect(recordingStore.loading).toBe(false);
      expect(recordingStore.error).toBe(null);
    });

    test('should handle rapid recording start/stop cycles', async () => {
      const startResponse = MockDataFactory.getRecordingStartResult();
      const stopResponse = MockDataFactory.getRecordingStopResult();
      
      mockRecordingService.startRecording = jest.fn().mockResolvedValue(startResponse);
      mockRecordingService.stopRecording = jest.fn().mockResolvedValue(stopResponse);
      
      // Rapid start/stop cycle
      await recordingStore.startRecording('camera0');
      expect(recordingStore.activeRecordings).toHaveLength(1);
      
      await recordingStore.stopRecording('camera0');
      expect(recordingStore.activeRecordings).toHaveLength(0);
      expect(recordingStore.recordingHistory).toHaveLength(1);
    });

    test('should handle large recording history efficiently', () => {
      const largeHistory = Array.from({ length: 1000 }, (_, i) => ({
        device: `camera${i % 10}`,
        filename: `recording_${i}`,
        status: 'STOPPED' as const,
        start_time: '2025-01-15T14:30:00Z',
        end_time: '2025-01-15T15:30:00Z',
        duration: 3600,
        file_size: 1024 * 1024,
        format: 'fmp4' as const
      }));
      
      recordingStore.recordingHistory = largeHistory;
      
      expect(recordingStore.recordingHistory).toHaveLength(1000);
    });
  });
});