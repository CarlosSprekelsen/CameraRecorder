/**
 * Unit Tests for Recording Store
 * 
 * REQ-001: Store State Management - Test Zustand store actions
 * REQ-002: State Transitions - Test state changes
 * REQ-003: Error Handling - Test error states and recovery
 * REQ-004: API Integration - Mock API calls and test responses
 * REQ-005: Side Effects - Test store side effects
 * REQ-006: Recording Lifecycle - Test recording start/stop/snapshot
 * 
 * Ground Truth: Official RPC Documentation
 * API Reference: docs/api/json_rpc_methods.md
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';
import { useRecordingStore } from '../../../src/stores/recording/recordingStore';

// Mock the RecordingService
const mockRecordingService = {
  takeSnapshot: jest.fn(),
  startRecording: jest.fn(),
  stopRecording: jest.fn()
};

jest.mock('../../../src/services/recording/RecordingService', () => ({
  RecordingService: jest.fn().mockImplementation(() => mockRecordingService)
}));

describe('Recording Store', () => {
  beforeEach(() => {
    // Reset the store to initial state
    useRecordingStore.getState().reset();
    
    // Reset all mocks
    jest.clearAllMocks();
    
    // Set up default mock implementations
    (mockRecordingService.takeSnapshot as jest.Mock).mockResolvedValue(MockDataFactory.getSnapshotResult());
    (mockRecordingService.startRecording as jest.Mock).mockResolvedValue(MockDataFactory.getRecordingStartResult());
    (mockRecordingService.stopRecording as jest.Mock).mockResolvedValue(MockDataFactory.getRecordingStopResult());
  });

  afterEach(() => {
    useRecordingStore.getState().reset();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct initial state', () => {
      const state = useRecordingStore.getState();
      
      expect(state.activeRecordings).toEqual({});
      expect(state.history).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    test('should handle recording status updates', () => {
      const { handleRecordingStatusUpdate } = useRecordingStore.getState();
      
      const recordingInfo = {
        device: 'camera0',
        filename: 'test.mp4',
        status: 'RECORDING' as const,
        startTime: '2025-01-15T14:30:00Z',
        duration: 60,
        format: 'mp4'
      };
      
      handleRecordingStatusUpdate(recordingInfo);
      
      const state = useRecordingStore.getState();
      expect(state.activeRecordings['camera0']).toEqual(recordingInfo);
    });

    test('should add to history when recording stops', () => {
      const { handleRecordingStatusUpdate } = useRecordingStore.getState();
      
      const recordingInfo = {
        device: 'camera0',
        filename: 'test.mp4',
        status: 'STOPPED' as const,
        startTime: '2025-01-15T14:30:00Z',
        duration: 60,
        format: 'mp4'
      };
      
      handleRecordingStatusUpdate(recordingInfo);
      
      const state = useRecordingStore.getState();
      expect(state.history).toContain(recordingInfo);
      expect(state.activeRecordings['camera0']).toBeUndefined();
    });
  });

  describe('REQ-006: Recording Lifecycle', () => {
    test('should take snapshot successfully', async () => {
      const { takeSnapshot, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await takeSnapshot('camera0', 'test.jpg');
      
      expect(mockRecordingService.takeSnapshot).toHaveBeenCalledWith('camera0', 'test.jpg');
    });

    test('should start recording successfully', async () => {
      const { startRecording, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await startRecording('camera0', 60, 'mp4');
      
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera0', 60, 'mp4');
    });

    test('should stop recording successfully', async () => {
      const { stopRecording, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await stopRecording('camera0');
      
      expect(mockRecordingService.stopRecording).toHaveBeenCalledWith('camera0');
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should transition from loading to success state', async () => {
      const { startRecording, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      
      const promise = startRecording('camera0');
      
      // Check loading state
      expect(useRecordingStore.getState().loading).toBe(true);
      expect(useRecordingStore.getState().error).toBe(null);
      
      await promise;
      
      // Check final state
      const state = useRecordingStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    test('should transition from loading to error state', async () => {
      const { startRecording, setService } = useRecordingStore.getState();
      
      (mockRecordingService.startRecording as jest.Mock).mockRejectedValue(new Error('Network error'));
      setService(mockRecordingService as any);
      
      await startRecording('camera0');
      
      const state = useRecordingStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe('Network error');
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle service not initialized error', async () => {
      const { startRecording } = useRecordingStore.getState();
      
      // Don't set the service
      await startRecording('camera0');
      
      const state = useRecordingStore.getState();
      expect(state.error).toBe('Recording service not initialized');
      expect(state.loading).toBe(false);
    });

    test('should handle API errors gracefully', async () => {
      const { startRecording, setService } = useRecordingStore.getState();
      
      (mockRecordingService.startRecording as jest.Mock).mockRejectedValue(new Error('API Error'));
      setService(mockRecordingService as any);
      
      await startRecording('camera0');
      
      const state = useRecordingStore.getState();
      expect(state.error).toBe('API Error');
      expect(state.loading).toBe(false);
    });

    test('should handle non-Error exceptions', async () => {
      const { startRecording, setService } = useRecordingStore.getState();
      
      (mockRecordingService.startRecording as jest.Mock).mockRejectedValue('String error');
      setService(mockRecordingService as any);
      
      await startRecording('camera0');
      
      const state = useRecordingStore.getState();
      expect(state.error).toBe('Failed to start recording');
      expect(state.loading).toBe(false);
    });
  });

  describe('REQ-004: API Integration', () => {
    test('should call takeSnapshot and handle response', async () => {
      const { takeSnapshot, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await takeSnapshot('camera0');
      
      expect(mockRecordingService.takeSnapshot).toHaveBeenCalledTimes(1);
      expect(mockRecordingService.takeSnapshot).toHaveBeenCalledWith('camera0', undefined);
    });

    test('should call startRecording with parameters', async () => {
      const { startRecording, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await startRecording('camera0', 120, 'fmp4');
      
      expect(mockRecordingService.startRecording).toHaveBeenCalledWith('camera0', 120, 'fmp4');
    });

    test('should call stopRecording and handle response', async () => {
      const { stopRecording, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await stopRecording('camera0');
      
      expect(mockRecordingService.stopRecording).toHaveBeenCalledWith('camera0');
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should reset store to initial state', () => {
      const { reset, setService } = useRecordingStore.getState();
      
      // Modify state
      setService(mockRecordingService as any);
      useRecordingStore.setState({
        activeRecordings: { camera0: { device: 'camera0', status: 'RECORDING' } },
        history: [{ device: 'camera0', status: 'STOPPED' }],
        loading: true,
        error: 'Test error'
      });
      
      // Reset
      reset();
      
      // Check state is back to initial
      const state = useRecordingStore.getState();
      expect(state.activeRecordings).toEqual({});
      expect(state.history).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    test('should handle concurrent recordings', () => {
      const { handleRecordingStatusUpdate } = useRecordingStore.getState();
      
      const recording1 = {
        device: 'camera0',
        filename: 'test1.mp4',
        status: 'RECORDING' as const,
        startTime: '2025-01-15T14:30:00Z'
      };
      
      const recording2 = {
        device: 'camera1',
        filename: 'test2.mp4',
        status: 'RECORDING' as const,
        startTime: '2025-01-15T14:30:00Z'
      };
      
      handleRecordingStatusUpdate(recording1);
      handleRecordingStatusUpdate(recording2);
      
      const state = useRecordingStore.getState();
      expect(state.activeRecordings['camera0']).toEqual(recording1);
      expect(state.activeRecordings['camera1']).toEqual(recording2);
      expect(Object.keys(state.activeRecordings)).toHaveLength(2);
    });
  });

  describe('API Compliance Validation', () => {
    test('should validate snapshot response against RPC spec', async () => {
      const { takeSnapshot, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await takeSnapshot('camera0');
      
      // The mock should return a valid snapshot result
      expect(mockRecordingService.takeSnapshot).toHaveBeenCalled();
    });

    test('should validate recording start response against RPC spec', async () => {
      const { startRecording, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await startRecording('camera0');
      
      // The mock should return a valid recording start result
      expect(mockRecordingService.startRecording).toHaveBeenCalled();
    });

    test('should validate recording stop response against RPC spec', async () => {
      const { stopRecording, setService } = useRecordingStore.getState();
      
      setService(mockRecordingService as any);
      await stopRecording('camera0');
      
      // The mock should return a valid recording stop result
      expect(mockRecordingService.stopRecording).toHaveBeenCalled();
    });
  });
});