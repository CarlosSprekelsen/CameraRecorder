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
  status: 'RECORDING' | 'STOPPED' | 'ERROR' | 'STARTING';  // Session state
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
  setRecordingService: (service: RecordingService) => void;
  setError: (error: string | null) => void;
  takeSnapshot: (device: string, filename?: string) => Promise<void>;
  startRecording: (device: string, duration?: number, format?: string) => Promise<void>;
  stopRecording: (device: string) => Promise<void>;
  handleRecordingStatusUpdate: (info: RecordingSessionInfo) => void;
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

          setRecordingService: (recordingService: RecordingService) => {
            service = recordingService;
          },

          setError: (error: string | null) => set({ error }),

          takeSnapshot: async (device: string, filename?: string) => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!service) {
              set({ error: 'Recording service not initialized', loading: false });
              return;
            }

            set({ loading: true, error: null });
            try {
              await service.takeSnapshot(device, filename);
              // Update state based on result
              set({ loading: false });
            } catch (error) {
              set({ loading: false, error: error instanceof Error ? error.message : 'Unknown error' });
              // No re-throw - graceful degradation per ADR-002
            }
          },

          startRecording: async (device: string, duration?: number, format?: string) => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!service) {
              set({ error: 'Recording service not initialized', loading: false });
              return;
            }

            set({ loading: true, error: null });
            try {
              const result = await service.startRecording(device, duration, format);
              // Store recording session information from server response
              set((state) => ({
                activeRecordings: {
                  ...state.activeRecordings,
                  [device]: {
                    device: result.device,
                    session_id: `${result.device}_${Date.now()}`, // Generate session ID since server doesn't provide one
                    filename: result.filename,
                    status: 'RECORDING' as const,
                    startTime: result.start_time,
                    duration,
                    format: result.format
                  }
                },
                loading: false
              }));
            } catch (error) {
              set({ loading: false, error: error instanceof Error ? error.message : 'Unknown error' });
              // No re-throw - graceful degradation per ADR-002
            }
          },

          stopRecording: async (device: string) => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!service) {
              set({ error: 'Recording service not initialized', loading: false });
              return;
            }

            set({ loading: true, error: null });
            try {
              const result = await service.stopRecording(device);
              // Remove from active recordings and add to history
              set((state) => {
                const { [device]: stoppedRecording, ...remainingActive } = state.activeRecordings;
                return {
                  activeRecordings: remainingActive,
                  history: [
                    {
                      filename: result.filename,
                      file_size: result.file_size,
                      duration: result.duration,
                      created_time: result.end_time,
                      download_url: '' // Will be populated when file is listed
                    },
                    ...state.history
                  ],
                  loading: false
                };
              });
            } catch (error) {
              set({ loading: false, error: error instanceof Error ? error.message : 'Unknown error' });
              // No re-throw - graceful degradation per ADR-002
            }
          },

          handleRecordingStatusUpdate: (info: RecordingSessionInfo) => {
            set((state) => {
              const nextActive = { ...state.activeRecordings };
              const nextHistory = [...state.history];
              
              if (info.status === 'RECORDING' || info.status === 'STARTING') {
                nextActive[info.device] = info;
              } else if (info.status === 'STOPPED' || info.status === 'ERROR') {
                // Remove from active recordings
                const stoppedRecording = nextActive[info.device];
                delete nextActive[info.device];
                
                // Add to history if recording was active (UX fix: prevent data loss)
                if (stoppedRecording) {
                  nextHistory.unshift({
                    filename: stoppedRecording.filename,
                    file_size: 0, // Will be updated when file is listed
                    duration: stoppedRecording.duration,
                    created_time: new Date().toISOString(),
                    download_url: '' // Will be populated when file is listed
                  });
                }
              }
              
              return { 
                activeRecordings: nextActive,
                history: nextHistory
              };
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

