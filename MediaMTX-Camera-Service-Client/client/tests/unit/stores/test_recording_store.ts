/**
 * REQ-REC01-001: Recording state management must be reliable and consistent
 * REQ-REC01-002: Recording operations must handle errors gracefully
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for recording store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on recording state management logic
 * - Test recording session management and progress tracking
 * - Validate error handling and recovery mechanisms
 */

import { useRecordingStore } from '../../../src/stores/recordingStore';
import type { RecordingSession, RecordingProgress, RecordingStatus } from '../../../src/types/camera';
import type { JSONRPCError } from '../../../src/types/rpc';

// Mock the recording manager service
jest.mock('../../../src/services/recordingManagerService', () => ({
  recordingManagerService: {
    startRecording: jest.fn(),
    stopRecording: jest.fn(),
    getRecordingStatus: jest.fn(),
    getActiveRecordings: jest.fn()
  }
}));

describe('Recording Store', () => {
  let store: ReturnType<typeof useRecordingStore.getState>;
  let mockRecordingService: any;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useRecordingStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useRecordingStore.getState();
    
    // Get mock service
    mockRecordingService = require('../../../src/services/recordingManagerService').recordingManagerService;
    jest.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useRecordingStore.getState();
      expect(state.recordingStates).toEqual(new Map());
      expect(state.activeSessions).toEqual(new Map());
      expect(state.errors).toEqual(new Map());
      expect(state.progress).toEqual(new Map());
      expect(state.isLoading).toBe(false);
      expect(state.isStarting).toBe(false);
      expect(state.isStopping).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
    });
  });

  describe('Recording State Management', () => {
    it('should set recording state for device', () => {
      const device = '/dev/video0';
      const status: RecordingStatus = 'recording';
      
      store.setRecordingState(device, status);
      
      const state = useRecordingStore.getState();
      expect(state.recordingStates.get(device)).toBe(status);
    });

    it('should get recording state for device', () => {
      const device = '/dev/video0';
      const status: RecordingStatus = 'idle';
      
      store.setRecordingState(device, status);
      
      expect(store.getRecordingState(device)).toBe(status);
    });

    it('should return idle for unknown device', () => {
      expect(store.getRecordingState('/dev/unknown')).toBe('idle');
    });

    it('should check if device is recording', () => {
      const device = '/dev/video0';
      
      store.setRecordingState(device, 'recording');
      expect(store.isRecording(device)).toBe(true);
      
      store.setRecordingState(device, 'idle');
      expect(store.isRecording(device)).toBe(false);
    });

    it('should get all recording states', () => {
      store.setRecordingState('/dev/video0', 'recording');
      store.setRecordingState('/dev/video1', 'idle');
      store.setRecordingState('/dev/video2', 'error');
      
      const states = store.getAllRecordingStates();
      expect(states.size).toBe(3);
      expect(states.get('/dev/video0')).toBe('recording');
      expect(states.get('/dev/video1')).toBe('idle');
      expect(states.get('/dev/video2')).toBe('error');
    });
  });

  describe('Active Session Management', () => {
    it('should add active session', () => {
      const device = '/dev/video0';
      const session: RecordingSession = {
        session_id: 'session-123',
        device,
        filename: 'recording.mp4',
        status: 'recording',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 0,
        file_size: 0
      };
      
      store.addActiveSession(device, session);
      
      const state = useRecordingStore.getState();
      expect(state.activeSessions.get(device)).toEqual(session);
    });

    it('should remove active session', () => {
      const device = '/dev/video0';
      const session: RecordingSession = {
        session_id: 'session-123',
        device,
        filename: 'recording.mp4',
        status: 'recording',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 0,
        file_size: 0
      };
      
      store.addActiveSession(device, session);
      store.removeActiveSession(device);
      
      const state = useRecordingStore.getState();
      expect(state.activeSessions.get(device)).toBeUndefined();
    });

    it('should get active session for device', () => {
      const device = '/dev/video0';
      const session: RecordingSession = {
        session_id: 'session-123',
        device,
        filename: 'recording.mp4',
        status: 'recording',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 0,
        file_size: 0
      };
      
      store.addActiveSession(device, session);
      
      expect(store.getActiveSession(device)).toEqual(session);
    });

    it('should return undefined for unknown device', () => {
      expect(store.getActiveSession('/dev/unknown')).toBeUndefined();
    });

    it('should get all active sessions', () => {
      const session1: RecordingSession = {
        session_id: 'session-1',
        device: '/dev/video0',
        filename: 'recording1.mp4',
        status: 'recording',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 0,
        file_size: 0
      };
      
      const session2: RecordingSession = {
        session_id: 'session-2',
        device: '/dev/video1',
        filename: 'recording2.mp4',
        status: 'recording',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 0,
        file_size: 0
      };
      
      store.addActiveSession('/dev/video0', session1);
      store.addActiveSession('/dev/video1', session2);
      
      const sessions = store.getAllActiveSessions();
      expect(sessions.size).toBe(2);
      expect(sessions.get('/dev/video0')).toEqual(session1);
      expect(sessions.get('/dev/video1')).toEqual(session2);
    });

    it('should check if device has active session', () => {
      const device = '/dev/video0';
      const session: RecordingSession = {
        session_id: 'session-123',
        device,
        filename: 'recording.mp4',
        status: 'recording',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 0,
        file_size: 0
      };
      
      expect(store.hasActiveSession(device)).toBe(false);
      
      store.addActiveSession(device, session);
      expect(store.hasActiveSession(device)).toBe(true);
    });
  });

  describe('Progress Tracking', () => {
    it('should set progress for device', () => {
      const device = '/dev/video0';
      const progress: RecordingProgress = {
        session_id: 'session-123',
        duration: 30,
        file_size: 1024000,
        bitrate: 1000,
        fps: 30
      };
      
      store.setProgress(device, progress);
      
      const state = useRecordingStore.getState();
      expect(state.progress.get(device)).toEqual(progress);
    });

    it('should get progress for device', () => {
      const device = '/dev/video0';
      const progress: RecordingProgress = {
        session_id: 'session-123',
        duration: 30,
        file_size: 1024000,
        bitrate: 1000,
        fps: 30
      };
      
      store.setProgress(device, progress);
      
      expect(store.getProgress(device)).toEqual(progress);
    });

    it('should return undefined for unknown device', () => {
      expect(store.getProgress('/dev/unknown')).toBeUndefined();
    });

    it('should clear progress for device', () => {
      const device = '/dev/video0';
      const progress: RecordingProgress = {
        session_id: 'session-123',
        duration: 30,
        file_size: 1024000,
        bitrate: 1000,
        fps: 30
      };
      
      store.setProgress(device, progress);
      store.clearProgress(device);
      
      const state = useRecordingStore.getState();
      expect(state.progress.get(device)).toBeUndefined();
    });

    it('should get all progress', () => {
      const progress1: RecordingProgress = {
        session_id: 'session-1',
        duration: 30,
        file_size: 1024000,
        bitrate: 1000,
        fps: 30
      };
      
      const progress2: RecordingProgress = {
        session_id: 'session-2',
        duration: 60,
        file_size: 2048000,
        bitrate: 2000,
        fps: 30
      };
      
      store.setProgress('/dev/video0', progress1);
      store.setProgress('/dev/video1', progress2);
      
      const allProgress = store.getAllProgress();
      expect(allProgress.size).toBe(2);
      expect(allProgress.get('/dev/video0')).toEqual(progress1);
      expect(allProgress.get('/dev/video1')).toEqual(progress2);
    });
  });

  describe('Error Management', () => {
    it('should set error for device', () => {
      const device = '/dev/video0';
      const error: JSONRPCError = {
        code: -32001,
        message: 'Camera not found',
        data: { device }
      };
      
      store.setError(device, error);
      
      const state = useRecordingStore.getState();
      expect(state.errors.get(device)).toEqual(error);
    });

    it('should get error for device', () => {
      const device = '/dev/video0';
      const error: JSONRPCError = {
        code: -32001,
        message: 'Camera not found',
        data: { device }
      };
      
      store.setError(device, error);
      
      expect(store.getError(device)).toEqual(error);
    });

    it('should return undefined for unknown device', () => {
      expect(store.getError('/dev/unknown')).toBeUndefined();
    });

    it('should clear error for device', () => {
      const device = '/dev/video0';
      const error: JSONRPCError = {
        code: -32001,
        message: 'Camera not found',
        data: { device }
      };
      
      store.setError(device, error);
      store.clearError(device);
      
      const state = useRecordingStore.getState();
      expect(state.errors.get(device)).toBeUndefined();
    });

    it('should get all errors', () => {
      const error1: JSONRPCError = {
        code: -32001,
        message: 'Camera not found',
        data: { device: '/dev/video0' }
      };
      
      const error2: JSONRPCError = {
        code: -32002,
        message: 'Recording failed',
        data: { device: '/dev/video1' }
      };
      
      store.setError('/dev/video0', error1);
      store.setError('/dev/video1', error2);
      
      const allErrors = store.getAllErrors();
      expect(allErrors.size).toBe(2);
      expect(allErrors.get('/dev/video0')).toEqual(error1);
      expect(allErrors.get('/dev/video1')).toEqual(error2);
    });
  });

  describe('Loading State Management', () => {
    it('should set loading state', () => {
      store.setLoading(true);
      let state = useRecordingStore.getState();
      expect(state.isLoading).toBe(true);

      store.setLoading(false);
      state = useRecordingStore.getState();
      expect(state.isLoading).toBe(false);
    });

    it('should set starting state', () => {
      store.setStarting(true);
      let state = useRecordingStore.getState();
      expect(state.isStarting).toBe(true);

      store.setStarting(false);
      state = useRecordingStore.getState();
      expect(state.isStarting).toBe(false);
    });

    it('should set stopping state', () => {
      store.setStopping(true);
      let state = useRecordingStore.getState();
      expect(state.isStopping).toBe(true);

      store.setStopping(false);
      state = useRecordingStore.getState();
      expect(state.isStopping).toBe(false);
    });
  });

  describe('Global Error Management', () => {
    it('should set global error', () => {
      store.setGlobalError('Global recording error');
      let state = useRecordingStore.getState();
      expect(state.error).toBe('Global recording error');

      store.setGlobalError(null);
      state = useRecordingStore.getState();
      expect(state.error).toBeNull();
    });

    it('should set last error', () => {
      store.setLastError('Last recording error');
      let state = useRecordingStore.getState();
      expect(state.lastError).toBe('Last recording error');

      store.setLastError(null);
      state = useRecordingStore.getState();
      expect(state.lastError).toBeNull();
    });

    it('should clear all errors', () => {
      store.setGlobalError('Global error');
      store.setLastError('Last error');
      store.setError('/dev/video0', {
        code: -32001,
        message: 'Device error',
        data: {}
      });
      
      store.clearAllErrors();
      
      const state = useRecordingStore.getState();
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
      expect(state.errors.size).toBe(0);
    });
  });

  describe('Recording Operations', () => {
    it('should start recording successfully', async () => {
      const device = '/dev/video0';
      const mockSession: RecordingSession = {
        session_id: 'session-123',
        device,
        filename: 'recording.mp4',
        status: 'recording',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 0,
        file_size: 0
      };

      mockRecordingService.startRecording.mockResolvedValue(mockSession);

      await store.startRecording(device, 'mp4', 60);

      const state = useRecordingStore.getState();
      expect(state.activeSessions.get(device)).toEqual(mockSession);
      expect(state.recordingStates.get(device)).toBe('recording');
      expect(state.isStarting).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle start recording failure', async () => {
      const device = '/dev/video0';
      const error = new Error('Recording failed');

      mockRecordingService.startRecording.mockRejectedValue(error);

      await store.startRecording(device, 'mp4', 60);

      const state = useRecordingStore.getState();
      expect(state.recordingStates.get(device)).toBe('error');
      expect(state.isStarting).toBe(false);
      expect(state.error).toBe('Recording failed');
    });

    it('should stop recording successfully', async () => {
      const device = '/dev/video0';
      const mockSession: RecordingSession = {
        session_id: 'session-123',
        device,
        filename: 'recording.mp4',
        status: 'stopped',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 60,
        file_size: 1024000
      };

      mockRecordingService.stopRecording.mockResolvedValue(mockSession);

      await store.stopRecording(device);

      const state = useRecordingStore.getState();
      expect(state.recordingStates.get(device)).toBe('idle');
      expect(state.activeSessions.get(device)).toBeUndefined();
      expect(state.isStopping).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle stop recording failure', async () => {
      const device = '/dev/video0';
      const error = new Error('Stop recording failed');

      mockRecordingService.stopRecording.mockRejectedValue(error);

      await store.stopRecording(device);

      const state = useRecordingStore.getState();
      expect(state.recordingStates.get(device)).toBe('error');
      expect(state.isStopping).toBe(false);
      expect(state.error).toBe('Stop recording failed');
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.setRecordingState('/dev/video0', 'recording');
      store.addActiveSession('/dev/video0', {
        session_id: 'session-123',
        device: '/dev/video0',
        filename: 'recording.mp4',
        status: 'recording',
        start_time: new Date().toISOString(),
        format: 'mp4',
        duration: 0,
        file_size: 0
      });
      store.setGlobalError('Test error');
      store.setLoading(true);
      
      // Reset
      store.reset();
      
      const state = useRecordingStore.getState();
      expect(state.recordingStates.size).toBe(0);
      expect(state.activeSessions.size).toBe(0);
      expect(state.errors.size).toBe(0);
      expect(state.progress.size).toBe(0);
      expect(state.isLoading).toBe(false);
      expect(state.isStarting).toBe(false);
      expect(state.isStopping).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
    });
  });
});
