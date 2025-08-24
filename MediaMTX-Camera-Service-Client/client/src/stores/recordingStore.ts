import { create } from 'zustand';
import { recordingManagerService } from '../services/recordingManagerService';
import type {
  RecordingState,
  RecordingSession,
  RecordingProgress,
  RecordingStatus
} from '../types/camera';
import type { JSONRPCError } from '../types/rpc';

/**
 * Recording Store State Interface
 */
interface RecordingStoreState {
  // Recording states for each camera
  recordingStates: Map<string, RecordingStatus>;
  
  // Active recording sessions
  activeSessions: Map<string, RecordingSession>;
  
  // Recording errors
  errors: Map<string, JSONRPCError>;
  
  // Recording progress
  progress: Map<string, RecordingProgress>;
  
  // Loading states
  isLoading: boolean;
  isStarting: boolean;
  isStopping: boolean;
  
  // Error states
  error: string | null;
  lastError: string | null;
}

/**
 * Recording Store Actions Interface
 */
interface RecordingStoreActions {
  // State management
  setRecordingState: (device: string, state: RecordingStatus) => void;
  setActiveSessions: (sessions: Map<string, RecordingSession>) => void;
  addActiveSession: (device: string, session: RecordingSession) => void;
  removeActiveSession: (device: string) => void;
  setProgress: (device: string, progress: RecordingProgress) => void;
  clearProgress: (device: string) => void;
  setError: (device: string, error: JSONRPCError) => void;
  clearError: (device: string) => void;
  
  // Loading states
  setLoading: (loading: boolean) => void;
  setStarting: (starting: boolean) => void;
  setStopping: (stopping: boolean) => void;
  
  // Error states
  setError: (error: string | null) => void;
  setLastError: (error: string | null) => void;
  clearErrors: () => void;
  
  // Recording operations
  startRecording: (device: string) => Promise<void>;
  stopRecording: (device: string) => Promise<void>;
  
  // State queries
  isRecording: (device: string) => boolean;
  getRecordingState: (device: string) => RecordingStatus;
  getActiveSessions: () => Map<string, RecordingSession>;
  getRecordingProgress: (device: string) => RecordingProgress | null;
  getActiveSession: (device: string) => RecordingSession | null;
  
  // Service integration
  initialize: () => void;
  cleanup: () => void;
}

/**
 * Recording Store Type
 */
type RecordingStore = RecordingStoreState & RecordingStoreActions;

/**
 * Recording Store Implementation
 */
export const useRecordingStore = create<RecordingStore>((set, get) => ({
  // Initial state
  recordingStates: new Map(),
  activeSessions: new Map(),
  errors: new Map(),
  progress: new Map(),
  isLoading: false,
  isStarting: false,
  isStopping: false,
  error: null,
  lastError: null,

  // State management actions
  setRecordingState: (device: string, state: RecordingStatus) => {
    set((storeState) => {
      const newStates = new Map(storeState.recordingStates);
      newStates.set(device, state);
      return { recordingStates: newStates };
    });
  },

  setActiveSessions: (sessions: Map<string, RecordingSession>) => {
    set({ activeSessions: sessions });
  },

  addActiveSession: (device: string, session: RecordingSession) => {
    set((storeState) => {
      const newSessions = new Map(storeState.activeSessions);
      newSessions.set(device, session);
      return { activeSessions: newSessions };
    });
  },

  removeActiveSession: (device: string) => {
    set((storeState) => {
      const newSessions = new Map(storeState.activeSessions);
      newSessions.delete(device);
      return { activeSessions: newSessions };
    });
  },

  setProgress: (device: string, progress: RecordingProgress) => {
    set((storeState) => {
      const newProgress = new Map(storeState.progress);
      newProgress.set(device, progress);
      return { progress: newProgress };
    });
  },

  clearProgress: (device: string) => {
    set((storeState) => {
      const newProgress = new Map(storeState.progress);
      newProgress.delete(device);
      return { progress: newProgress };
    });
  },

  setError: (device: string, error: JSONRPCError) => {
    set((storeState) => {
      const newErrors = new Map(storeState.errors);
      newErrors.set(device, error);
      return { errors: newErrors };
    });
  },

  clearError: (device: string) => {
    set((storeState) => {
      const newErrors = new Map(storeState.errors);
      newErrors.delete(device);
      return { errors: newErrors };
    });
  },

  // Loading state actions
  setLoading: (loading: boolean) => {
    set({ isLoading: loading });
  },

  setStarting: (starting: boolean) => {
    set({ isStarting: starting });
  },

  setStopping: (stopping: boolean) => {
    set({ isStopping: stopping });
  },

  // Error state actions
  setError: (error: string | null) => {
    set({ error });
  },

  setLastError: (error: string | null) => {
    set({ lastError: error });
  },

  clearErrors: () => {
    set({ error: null, lastError: null });
  },

  // Recording operations
  startRecording: async (device: string) => {
    const { setStarting, setError, setLastError } = get();
    
    try {
      setStarting(true);
      setError(null);
      
      const session = await recordingManagerService.startRecording(device);
      
      // Update store state
      const { setRecordingState, addActiveSession } = get();
      setRecordingState(device, 'RECORDING');
      addActiveSession(device, session);
      
    } catch (error: any) {
      const errorMessage = error.message || 'Failed to start recording';
      setError(errorMessage);
      setLastError(errorMessage);
      
      // Handle recording errors
      if (error.code === -1006) { // CAMERA_ALREADY_RECORDING
        const { setError: setDeviceError } = get();
        setDeviceError(device, error);
      }
      
      throw error;
    } finally {
      setStarting(false);
    }
  },

  stopRecording: async (device: string) => {
    const { setStopping, setError, setLastError } = get();
    
    try {
      setStopping(true);
      setError(null);
      
      await recordingManagerService.stopRecording(device);
      
      // Update store state
      const { setRecordingState, removeActiveSession, clearProgress, clearError } = get();
      setRecordingState(device, 'STOPPED');
      removeActiveSession(device);
      clearProgress(device);
      clearError(device);
      
    } catch (error: any) {
      const errorMessage = error.message || 'Failed to stop recording';
      setError(errorMessage);
      setLastError(errorMessage);
      throw error;
    } finally {
      setStopping(false);
    }
  },

  // State queries
  isRecording: (device: string) => {
    return get().activeSessions.has(device);
  },

  getRecordingState: (device: string) => {
    return get().recordingStates.get(device) || 'STOPPED';
  },

  getActiveSessions: () => {
    return get().activeSessions;
  },

  getRecordingProgress: (device: string) => {
    return get().progress.get(device) || null;
  },

  getActiveSession: (device: string) => {
    return get().activeSessions.get(device) || null;
  },

  // Service integration
  initialize: () => {
    // Initialize recording manager service
    recordingManagerService.initialize();
  },

  cleanup: () => {
    // Cleanup recording manager service
    recordingManagerService.cleanup();
  },
}));
