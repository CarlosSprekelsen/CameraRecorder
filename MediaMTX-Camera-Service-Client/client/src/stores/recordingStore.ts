/**
 * Recording Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around RecordingManagerService
 * following the modular store pattern established in connection/
 * 
 * Responsibilities:
 * - Recording session management
 * - Recording status tracking
 * - Recording operations
 * 
 * Architecture Compliance:
 * - Single responsibility (recording operations only)
 * - Uses service layer abstraction
 * - Provides predictable state interface for components
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';
import type { RecordingSession, RecordingStatus } from '../types/camera';

// State interface
interface RecordingStoreState {
  // Recording sessions
  sessions: RecordingSession[];
  activeSessions: string[]; // device IDs with active recordings
  
  // Loading states
  isLoading: boolean;
  isStarting: boolean;
  isStopping: boolean;
  
  // Error state
  error: string | null;
  
  // Recording status
  recordingStatus: Record<string, RecordingStatus>;
}

// Actions interface
interface RecordingStoreActions {
  // Recording operations
  startRecording: (device: string) => Promise<boolean>;
  stopRecording: (device: string) => Promise<boolean>;
  
  // Session management
  getRecordingSessions: () => Promise<void>;
  refreshSessions: () => Promise<void>;
  
  // Status operations
  getRecordingStatus: (device: string) => Promise<RecordingStatus | null>;
  refreshRecordingStatus: (device: string) => Promise<void>;
  
  // Error handling
  clearError: () => void;
  setError: (error: string) => void;
}

// Combined store type
type RecordingStore = RecordingStoreState & RecordingStoreActions;

// Initial state
const initialState: RecordingStoreState = {
  sessions: [],
  activeSessions: [],
  isLoading: false,
  isStarting: false,
  isStopping: false,
  error: null,
  recordingStatus: {},
};

// Create store
export const useRecordingStore = create<RecordingStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      // Recording operations
      startRecording: async (device: string) => {
        set({ isStarting: true, error: null });
        try {
          // TODO: Implement with RecordingManagerService
          // const recordingService = new RecordingManagerService();
          // const result = await recordingService.startRecording(device);
          // if (result.success) {
          //   set(state => ({
          //     activeSessions: [...state.activeSessions, device],
          //     isStarting: false
          //   }));
          // }
          // return result.success;
          
          // Temporary mock
          set(state => ({
            activeSessions: [...state.activeSessions, device],
            isStarting: false
          }));
          logger.info(`Recording started for camera ${device}`, undefined, 'recordingStore');
          return true;
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to start recording';
          set({ 
            error: errorMessage, 
            isStarting: false 
          });
          return false;
        }
      },
      
      stopRecording: async (device: string) => {
        set({ isStopping: true, error: null });
        try {
          // TODO: Implement with RecordingManagerService
          // const recordingService = new RecordingManagerService();
          // const result = await recordingService.stopRecording(device);
          // if (result.success) {
          //   set(state => ({
          //     activeSessions: state.activeSessions.filter(id => id !== device),
          //     isStopping: false
          //   }));
          // }
          // return result.success;
          
          // Temporary mock
          set(state => ({
            activeSessions: state.activeSessions.filter(id => id !== device),
            isStopping: false
          }));
          logger.info(`Recording stopped for camera ${device}`, undefined, 'recordingStore');
          return true;
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to stop recording';
          set({ 
            error: errorMessage, 
            isStopping: false 
          });
          return false;
        }
      },
      
      // Session management
      getRecordingSessions: async () => {
        set({ isLoading: true, error: null });
        try {
          // TODO: Implement with RecordingManagerService
          // const recordingService = new RecordingManagerService();
          // const sessions = await recordingService.getSessions();
          // set({ sessions, isLoading: false });
          
          // Temporary mock
          set({ 
            sessions: [], 
            isLoading: false 
          });
          logger.info('Recording sessions retrieved', undefined, 'recordingStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get recording sessions';
          set({ 
            error: errorMessage, 
            isLoading: false 
          });
        }
      },
      
      refreshSessions: async () => {
        try {
          await get().getRecordingSessions();
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to refresh sessions';
          set({ error: errorMessage });
        }
      },
      
      // Status operations
      getRecordingStatus: async (device: string) => {
        try {
          // TODO: Implement with RecordingManagerService
          // const recordingService = new RecordingManagerService();
          // const status = await recordingService.getRecordingStatus(device);
          // set(state => ({
          //   recordingStatus: { ...state.recordingStatus, [device]: status }
          // }));
          // return status;
          
          // Temporary mock
          return null;
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get recording status';
          set({ error: errorMessage });
          return null;
        }
      },
      
      refreshRecordingStatus: async (device: string) => {
        try {
          await get().getRecordingStatus(device);
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to refresh recording status';
          set({ error: errorMessage });
        }
      },
      
      // Error handling
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    {
      name: 'recording-store',
    }
  )
);

// Export types for components
export type { RecordingStoreState, RecordingStoreActions };
