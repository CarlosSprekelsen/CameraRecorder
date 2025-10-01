import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { DeviceService } from '../../services/device/DeviceService';
import { Camera, StreamsListResult } from '../../types/api';

// ARCHITECTURE FIX: Use official API types from types/ directory
// Removed duplicate type definitions - using authoritative types/api.ts

export interface DeviceState {
  cameras: Camera[];
  streams: StreamsListResult[];
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
}

export interface DeviceActions {
  // Discovery methods (I.Discovery interface)
  getCameraList: () => Promise<void>;
  getStreamUrl: (device: string) => Promise<string | null>;
  getStreams: () => Promise<void>;

  // State management
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  updateCameraStatus: (device: string, status: Camera['status']) => void;
  updateStreamStatus: (name: string, ready: boolean, readers: number) => void;

  // Real-time updates
  handleCameraStatusUpdate: (camera: Camera) => void;
  handleStreamUpdate: (stream: StreamsListResult) => void;

  // Service injection
  setDeviceService: (service: DeviceService) => void;

  // Reset
  reset: () => void;
}

const initialState: DeviceState = {
  cameras: [],
  streams: [],
  loading: false,
  error: null,
  lastUpdated: null,
};

export const useDeviceStore = create<DeviceState & DeviceActions>()(
  devtools(
    (set) => {
      let deviceService: DeviceService | null = null;

      return {
        ...initialState,

        // Service injection
        setDeviceService: (service: DeviceService) => {
          deviceService = service;
        },

        // Discovery methods (I.Discovery interface)
        getCameraList: async () => {
          // Synchronous guard - graceful error handling per ADR-002
          if (!deviceService) {
            set({ error: 'Device service not initialized', loading: false });
            return;
          }

          set({ loading: true, error: null });
          try {
            const cameras = await deviceService.getCameraList();
            set({
              cameras,
              loading: false,
              lastUpdated: new Date().toISOString(),
              error: null,
            });
          } catch (error) {
            set({
              loading: false,
              error: error instanceof Error ? error.message : 'Failed to get camera list',
            });
            // No re-throw - graceful degradation per ADR-002
          }
        },

        getStreamUrl: async (device: string) => {
          // Synchronous guard - graceful error handling per ADR-002
          if (!deviceService) {
            set({ error: 'Device service not initialized' });
            return null;
          }

          try {
            const streamUrl = await deviceService.getStreamUrl(device);
            return streamUrl;
          } catch (error) {
            set({ error: error instanceof Error ? error.message : 'Failed to get stream URL' });
            // No re-throw - graceful degradation per ADR-002
            return null;
          }
        },

        getStreams: async () => {
          // Synchronous guard - graceful error handling per ADR-002
          if (!deviceService) {
            set({ error: 'Device service not initialized', loading: false });
            return;
          }

          set({ loading: true, error: null });
          try {
            const streams = await deviceService.getStreams();
            set({
              streams,
              loading: false,
              lastUpdated: new Date().toISOString(),
              error: null,
            });
          } catch (error) {
            set({
              loading: false,
              error: error instanceof Error ? error.message : 'Failed to get streams',
            });
            // No re-throw - graceful degradation per ADR-002
          }
        },

        // State management
        setLoading: (loading: boolean) => set({ loading }),
        setError: (error: string | null) => set({ error }),

        updateCameraStatus: (device: string, status: Camera['status']) => {
          set((state) => ({
            cameras: state.cameras.map((camera) =>
              camera.device === device ? { ...camera, status } : camera,
            ),
          }));
        },

        updateStreamStatus: (name: string, ready: boolean, readers: number) => {
          set((state) => ({
            streams: state.streams.map((stream) =>
              stream.name === name ? { ...stream, ready, readers } : stream,
            ),
          }));
        },

        // Real-time updates
        handleCameraStatusUpdate: (camera: Camera) => {
          set((state) => ({
            cameras: state.cameras
              .map((c) => (c.device === camera.device ? camera : c))
              .concat(state.cameras.find((c) => c.device === camera.device) ? [] : [camera]),
          }));
        },

        handleStreamUpdate: (stream: StreamsListResult) => {
          set((state) => ({
            streams: state.streams
              .map((s) => (s.name === stream.name ? stream : s))
              .concat(state.streams.find((s) => s.name === stream.name) ? [] : [stream]),
          }));
        },


        // Reset
        reset: () => set(initialState),
      };
    },
    {
      name: 'device-store',
    },
  ),
);
