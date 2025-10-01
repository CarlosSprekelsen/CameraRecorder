import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { StreamingService } from '../../services/streaming/StreamingService';
import { StreamStatusResult } from '../../types/api';
import { logger } from '../../services/logger/LoggerService';

export interface StreamingState {
  activeStreams: Record<string, StreamStatusResult>;
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
}

export interface StreamingActions {
  // Service injection
  setStreamingService: (service: StreamingService) => void;
  
  // Streaming operations
  startStreaming: (device: string) => Promise<void>;
  stopStreaming: (device: string) => Promise<void>;
  getStreamStatus: (device: string) => Promise<void>;
  
  // State management
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  updateStreamStatus: (device: string, status: StreamStatusResult) => void;
  
  
  // Reset
  reset: () => void;
}

const initialState: StreamingState = {
  activeStreams: {},
  loading: false,
  error: null,
  lastUpdated: null,
};

export const useStreamingStore = create<StreamingState & StreamingActions>()(
  devtools(
    (set) => {
      let streamingService: StreamingService | null = null;

      return {
        ...initialState,

        // Service injection
        setStreamingService: (service: StreamingService) => {
          streamingService = service;
        },

        // Streaming operations
        startStreaming: async (device: string) => {
          // Synchronous guard - graceful error handling per ADR-002
          if (!streamingService) {
            set({ error: 'Streaming service not initialized', loading: false });
            return;
          }

          set({ loading: true, error: null });
          try {
            const result = await streamingService.startStreaming(device);
            logger.info('Streaming started', { device, status: result.status });
            set({ loading: false, lastUpdated: new Date().toISOString() });
          } catch (error) {
            set({
              loading: false,
              error: error instanceof Error ? error.message : 'Failed to start streaming',
            });
            // No re-throw - graceful degradation per ADR-002
          }
        },

        stopStreaming: async (device: string) => {
          // Synchronous guard - graceful error handling per ADR-002
          if (!streamingService) {
            set({ error: 'Streaming service not initialized', loading: false });
            return;
          }

          set({ loading: true, error: null });
          try {
            const result = await streamingService.stopStreaming(device);
            logger.info('Streaming stopped', { device, status: result.status });
            set({ loading: false, lastUpdated: new Date().toISOString() });
          } catch (error) {
            set({
              loading: false,
              error: error instanceof Error ? error.message : 'Failed to stop streaming',
            });
            // No re-throw - graceful degradation per ADR-002
          }
        },

        getStreamStatus: async (device: string) => {
          // Synchronous guard - graceful error handling per ADR-002
          if (!streamingService) {
            set({ error: 'Streaming service not initialized', loading: false });
            return;
          }

          set({ loading: true, error: null });
          try {
            const status = await streamingService.getStreamStatus(device);
            set((state) => ({
              activeStreams: { ...state.activeStreams, [device]: status },
              loading: false,
              lastUpdated: new Date().toISOString(),
            }));
          } catch (error) {
            set({
              loading: false,
              error: error instanceof Error ? error.message : 'Failed to get stream status',
            });
            // No re-throw - graceful degradation per ADR-002
          }
        },

        // State management
        setLoading: (loading: boolean) => set({ loading }),
        setError: (error: string | null) => set({ error }),

        updateStreamStatus: (device: string, status: StreamStatusResult) => {
          set((state) => ({
            activeStreams: { ...state.activeStreams, [device]: status },
            lastUpdated: new Date().toISOString(),
          }));
        },


        // Reset
        reset: () => set(initialState),
      };
    },
    {
      name: 'streaming-store',
    },
  ),
);
