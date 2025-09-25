import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { RecordingService } from '../../services/recording/RecordingService';

export interface RecordingInfo {
  device: string;
  filename?: string;
  status: 'RECORDING' | 'STOPPED' | 'ERROR' | 'STARTED';
  startTime?: string;
  duration?: number;
  format?: string;
}

export interface RecordingState {
  activeRecordings: Record<string, RecordingInfo>;
  history: RecordingInfo[];
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
      (set, get) => {
        let service: RecordingService | null = null;

        return {
          ...initialState,

          setService: (s: RecordingService) => {
            service = s;
          },

          takeSnapshot: async (device: string, filename?: string) => {
            if (!service) {
              set({ error: 'Recording service not initialized' });
              return;
            }
            set({ loading: true, error: null });
            try {
              await service.takeSnapshot(device, filename);
              set({ loading: false });
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Snapshot failed',
              });
            }
          },

          startRecording: async (device: string, duration?: number, format?: string) => {
            if (!service) {
              set({ error: 'Recording service not initialized' });
              return;
            }
            // Concurrency limit: do not start if device already recording
            if (get().activeRecordings[device]) {
              set({ error: `Device ${device} is already recording` });
              return;
            }
            set({ loading: true, error: null });
            try {
              await service.startRecording(device, duration, format);
              set({ loading: false });
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Start recording failed',
              });
            }
          },

          stopRecording: async (device: string) => {
            if (!service) {
              set({ error: 'Recording service not initialized' });
              return;
            }
            set({ loading: true, error: null });
            try {
              await service.stopRecording(device);
              set({ loading: false });
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Stop recording failed',
              });
            }
          },

          handleRecordingStatusUpdate: (info: RecordingInfo) => {
            set((state) => {
              const nextActive = { ...state.activeRecordings };
              if (info.status === 'RECORDING' || info.status === 'STARTED') {
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
