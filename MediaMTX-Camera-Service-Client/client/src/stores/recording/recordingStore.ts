import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { RecordingService } from '../../services/recording/RecordingService';
import { RecordingInfo } from '../../types/api';

// FIXED: Separate operation results from session state
export interface RecordingOperationResult {
  device: string;
  filename?: string;
  status: 'SUCCESS' | 'FAILED';  // FIXED: API spec uses SUCCESS/FAILED for operations
  error?: string;
}

export interface RecordingSessionInfo {
  device: string;
  session_id: string;
  filename?: string;
  status: 'RECORDING' | 'STOPPED' | 'ERROR';  // Session state
  startTime?: string;
  duration?: number;
  format?: string;
}

// Legacy interface removed - use RecordingSessionInfo directly

export interface RecordingState {
  activeRecordings: Record<string, RecordingSessionInfo>;
  history: RecordingInfo[];  // Use API RecordingInfo for file history
  loading: boolean;
  error: string | null;
}

export interface RecordingActions {
  setService: (service: RecordingService) => void;
  takeSnapshot: (device: string, filename?: string) => Promise<void>;
  startRecording: (device: string, duration?: number, format?: string) => Promise<void>;
  stopRecording: (device: string) => Promise<void>;
  handleRecordingStatusUpdate: (info: RecordingInfo) => void;
  reset: () => void;
}

const initialState: RecordingState = {
  activeRecordings: {},
  history: [],
  loading: false,
  error: null,
};

export const useRecordingStore = create<RecordingState & RecordingActions>()(
  devtools(
    persist(
      (set) => {
        // ARCHITECTURE FIX: Remove direct service injection
        // Use action dispatchers instead of direct service calls
        // Architecture requirement: Unidirectional data flow (ADR-002)

        let service: RecordingService | null = null;

        return {
          ...initialState,

          setService: (recordingService: RecordingService) => {
            service = recordingService;
          },

          takeSnapshot: async (device: string, filename?: string) => {
            if (!service) throw new Error('Recording service not initialized');
            set({ loading: true, error: null });
            try {
              await service.takeSnapshot(device, filename);
              // Update state based on result
              set({ loading: false });
            } catch (error) {
              set({ loading: false, error: error instanceof Error ? error.message : 'Unknown error' });
            }
          },

          startRecording: async (device: string, duration?: number, format?: string) => {
            if (!service) throw new Error('Recording service not initialized');
            set({ loading: true, error: null });
            try {
              await service.startRecording(device, duration, format);
              // Update state based on result
              set({ loading: false });
            } catch (error) {
              set({ loading: false, error: error instanceof Error ? error.message : 'Unknown error' });
            }
          },

          stopRecording: async (device: string) => {
            if (!service) throw new Error('Recording service not initialized');
            set({ loading: true, error: null });
            try {
              await service.stopRecording(device);
              // Update state based on result
              set({ loading: false });
            } catch (error) {
              set({ loading: false, error: error instanceof Error ? error.message : 'Unknown error' });
            }
          },

          handleRecordingStatusUpdate: (info: RecordingInfo) => {
            set((state) => {
              const nextActive = { ...state.activeRecordings };
              if (info.status === 'RECORDING' || info.status === 'STARTING') {
                nextActive[info.device] = info;
              } else if (info.status === 'STOPPED' || info.status === 'ERROR') {
                delete nextActive[info.device];
              }
              const nextHistory = [...state.history];
              nextHistory.unshift(info);
              return { activeRecordings: nextActive, history: nextHistory };
            });
          },

          reset: () => set(initialState),
        };
      },
      {
        name: 'recording-store',
        partialize: (state) => ({
          activeRecordings: state.activeRecordings,
          history: state.history,
        }),
      },
    ),
  ),
);

