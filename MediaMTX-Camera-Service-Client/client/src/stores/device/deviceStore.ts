import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { DeviceService } from '../../services/device/DeviceService';

// Types aligned with architecture section 5.3.1
export interface Camera {
  device: string;
  status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  name: string;
  resolution: string;
  fps: number;
  streams: {
    rtsp: string;
    hls: string;
  };
}

export interface StreamInfo {
  name: string;
  source: string;
  ready: boolean;
  readers: number;
  bytes_sent: number;
}

export interface DeviceState {
  cameras: Camera[];
  streams: StreamInfo[];
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
  handleStreamUpdate: (stream: StreamInfo) => void;
  
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
          if (!deviceService) {
            set({ error: 'Device service not initialized' });
            return;
          }

          set({ loading: true, error: null });
          try {
            const cameras = await deviceService.getCameraList();
            set({ 
              cameras,
              loading: false, 
              lastUpdated: new Date().toISOString(),
              error: null 
            });
          } catch (error) {
            set({ 
              loading: false, 
              error: error instanceof Error ? error.message : 'Failed to get camera list' 
            });
          }
        },

        getStreamUrl: async (device: string) => {
          if (!deviceService) {
            set({ error: 'Device service not initialized' });
            return null;
          }

          try {
            const streamUrl = await deviceService.getStreamUrl(device);
            return streamUrl;
          } catch (error) {
            set({ error: error instanceof Error ? error.message : 'Failed to get stream URL' });
            return null;
          }
        },

        getStreams: async () => {
          if (!deviceService) {
            set({ error: 'Device service not initialized' });
            return;
          }

          set({ loading: true, error: null });
          try {
            const streams = await deviceService.getStreams();
            set({ 
              streams,
              loading: false, 
              lastUpdated: new Date().toISOString(),
              error: null 
            });
          } catch (error) {
            set({ 
              loading: false, 
              error: error instanceof Error ? error.message : 'Failed to get streams' 
            });
          }
        },

      // State management
      setLoading: (loading: boolean) => set({ loading }),
      setError: (error: string | null) => set({ error }),

      updateCameraStatus: (device: string, status: Camera['status']) => {
        set((state) => ({
          cameras: state.cameras.map(camera =>
            camera.device === device ? { ...camera, status } : camera
          )
        }));
      },

      updateStreamStatus: (name: string, ready: boolean, readers: number) => {
        set((state) => ({
          streams: state.streams.map(stream =>
            stream.name === name ? { ...stream, ready, readers } : stream
          )
        }));
      },

      // Real-time updates
      handleCameraStatusUpdate: (camera: Camera) => {
        set((state) => ({
          cameras: state.cameras.map(c =>
            c.device === camera.device ? camera : c
          ).concat(
            state.cameras.find(c => c.device === camera.device) ? [] : [camera]
          )
        }));
      },

      handleStreamUpdate: (stream: StreamInfo) => {
        set((state) => ({
          streams: state.streams.map(s =>
            s.name === stream.name ? stream : s
          ).concat(
            state.streams.find(s => s.name === stream.name) ? [] : [stream]
          )
        }));
      },

        // Reset
        reset: () => set(initialState),
      };
    },
    {
      name: 'device-store',
    }
  )
);
