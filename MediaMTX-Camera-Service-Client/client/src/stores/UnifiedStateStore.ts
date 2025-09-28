/**
 * Unified State Store - Architecture Compliance
 * 
 * Architecture requirement: "Centralized state store with unidirectional data flow" (ADR-002)
 * Consolidates multiple stores into a single source of truth
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

// Import all state modules
import { Camera } from './device/deviceStore';
import { RecordingInfo } from './recording/recordingStore';
import { FileInfo } from './files/fileStore';
import { ServerStatus } from './server/serverStore';

export interface UnifiedState {
  // Device Module
  devices: {
    cameras: Camera[];
    selectedCamera: string | null;
    loading: boolean;
    error: string | null;
  };

  // Recording Module  
  recordings: {
    activeRecordings: Record<string, RecordingInfo>;
    loading: boolean;
    error: string | null;
  };

  // Files Module
  files: {
    recordings: FileInfo[];
    snapshots: FileInfo[];
    selectedFiles: string[];
    loading: boolean;
    error: string | null;
  };

  // Server Module
  server: {
    status: ServerStatus;
    connected: boolean;
    loading: boolean;
    error: string | null;
  };

  // Authentication Module
  auth: {
    isAuthenticated: boolean;
    user: {
      role: string;
      permissions: string[];
    } | null;
    loading: boolean;
    error: string | null;
  };
}

export interface UnifiedActions {
  // Device Actions
  setCameras: (cameras: Camera[]) => void;
  selectCamera: (deviceId: string) => void;
  setDeviceLoading: (loading: boolean) => void;
  setDeviceError: (error: string | null) => void;

  // Recording Actions
  setActiveRecording: (device: string, info: RecordingInfo) => void;
  removeActiveRecording: (device: string) => void;
  setRecordingLoading: (loading: boolean) => void;
  setRecordingError: (error: string | null) => void;

  // File Actions
  setRecordings: (recordings: FileInfo[]) => void;
  setSnapshots: (snapshots: FileInfo[]) => void;
  selectFiles: (files: string[]) => void;
  setFileLoading: (loading: boolean) => void;
  setFileError: (error: string | null) => void;

  // Server Actions
  setServerStatus: (status: ServerStatus) => void;
  setConnected: (connected: boolean) => void;
  setServerLoading: (loading: boolean) => void;
  setServerError: (error: string | null) => void;

  // Auth Actions
  setAuthenticated: (authenticated: boolean) => void;
  setUser: (user: { role: string; permissions: string[] } | null) => void;
  setAuthLoading: (loading: boolean) => void;
  setAuthError: (error: string | null) => void;

  // Global Actions
  clearAllErrors: () => void;
  reset: () => void;
}

const initialState: UnifiedState = {
  devices: {
    cameras: [],
    selectedCamera: null,
    loading: false,
    error: null,
  },
  recordings: {
    activeRecordings: {},
    loading: false,
    error: null,
  },
  files: {
    recordings: [],
    snapshots: [],
    selectedFiles: [],
    loading: false,
    error: null,
  },
  server: {
    status: 'UNKNOWN',
    connected: false,
    loading: false,
    error: null,
  },
  auth: {
    isAuthenticated: false,
    user: null,
    loading: false,
    error: null,
  },
};

export const useUnifiedStore = create<UnifiedState & UnifiedActions>()(
  devtools(
    (set, get) => ({
      ...initialState,

      // Device Actions
      setCameras: (cameras) => set((state) => ({ 
        devices: { ...state.devices, cameras } 
      })),
      selectCamera: (deviceId) => set((state) => ({ 
        devices: { ...state.devices, selectedCamera: deviceId } 
      })),
      setDeviceLoading: (loading) => set((state) => ({ 
        devices: { ...state.devices, loading } 
      })),
      setDeviceError: (error) => set((state) => ({ 
        devices: { ...state.devices, error } 
      })),

      // Recording Actions
      setActiveRecording: (device, info) => set((state) => ({
        recordings: {
          ...state.recordings,
          activeRecordings: { ...state.recordings.activeRecordings, [device]: info }
        }
      })),
      removeActiveRecording: (device) => set((state) => {
        const { [device]: removed, ...rest } = state.recordings.activeRecordings;
        return {
          recordings: { ...state.recordings, activeRecordings: rest }
        };
      }),
      setRecordingLoading: (loading) => set((state) => ({ 
        recordings: { ...state.recordings, loading } 
      })),
      setRecordingError: (error) => set((state) => ({ 
        recordings: { ...state.recordings, error } 
      })),

      // File Actions
      setRecordings: (recordings) => set((state) => ({ 
        files: { ...state.files, recordings } 
      })),
      setSnapshots: (snapshots) => set((state) => ({ 
        files: { ...state.files, snapshots } 
      })),
      selectFiles: (files) => set((state) => ({ 
        files: { ...state.files, selectedFiles: files } 
      })),
      setFileLoading: (loading) => set((state) => ({ 
        files: { ...state.files, loading } 
      })),
      setFileError: (error) => set((state) => ({ 
        files: { ...state.files, error } 
      })),

      // Server Actions
      setServerStatus: (status) => set((state) => ({ 
        server: { ...state.server, status } 
      })),
      setConnected: (connected) => set((state) => ({ 
        server: { ...state.server, connected } 
      })),
      setServerLoading: (loading) => set((state) => ({ 
        server: { ...state.server, loading } 
      })),
      setServerError: (error) => set((state) => ({ 
        server: { ...state.server, error } 
      })),

      // Auth Actions
      setAuthenticated: (isAuthenticated) => set((state) => ({ 
        auth: { ...state.auth, isAuthenticated } 
      })),
      setUser: (user) => set((state) => ({ 
        auth: { ...state.auth, user } 
      })),
      setAuthLoading: (loading) => set((state) => ({ 
        auth: { ...state.auth, loading } 
      })),
      setAuthError: (error) => set((state) => ({ 
        auth: { ...state.auth, error } 
      })),

      // Global Actions
      clearAllErrors: () => set((state) => ({
        devices: { ...state.devices, error: null },
        recordings: { ...state.recordings, error: null },
        files: { ...state.files, error: null },
        server: { ...state.server, error: null },
        auth: { ...state.auth, error: null },
      })),

      reset: () => set(initialState),
    }),
    {
      name: 'unified-state-store',
    }
  )
);
