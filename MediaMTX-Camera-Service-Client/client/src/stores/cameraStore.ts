/**
 * Camera state management store
 * Handles camera list, selected camera, and camera operations
 * 
 * Sprint 3 Updates:
 * - Real server integration
 * - Improved error handling
 * - Better loading states
 * - Real-time camera status updates
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type {
  CameraDevice,
  CameraStatus,
  RecordingResponse,
  SnapshotResponse,
  ServerInfo,
} from '../types';
import { RPC_METHODS } from '../types';
import { createWebSocketService, type WebSocketService } from '../services/websocket';

/**
 * Camera store state interface
 */
interface CameraState {
  // Camera data
  cameras: CameraDevice[];
  selectedCamera: string | null;
  serverInfo: ServerInfo | null;
  
  // Loading states
  isLoading: boolean;
  isRefreshing: boolean;
  isConnecting: boolean;
  
  // Error state
  error: string | null;
  
  // Recording state
  activeRecordings: Map<string, RecordingResponse>;
  
  // WebSocket service
  wsService: WebSocketService | null;
  
  // Connection state
  isConnected: boolean;
}

/**
 * Camera store actions interface
 */
interface CameraActions {
  // Initialization
  initialize: (wsUrl?: string) => Promise<void>;
  disconnect: () => void;
  
  // Camera operations
  refreshCameras: () => Promise<void>;
  selectCamera: (device: string | null) => void;
  getCameraStatus: (device: string) => Promise<CameraDevice | null>;
  
  // Recording operations
  startRecording: (device: string, duration?: number, format?: string) => Promise<RecordingResponse | null>;
  stopRecording: (device: string) => Promise<RecordingResponse | null>;
  
  // Snapshot operations
  takeSnapshot: (device: string, format?: string, quality?: number) => Promise<SnapshotResponse | null>;
  
  // Server operations
  getServerInfo: () => Promise<ServerInfo | null>;
  pingServer: () => Promise<boolean>;
  
  // State management
  setError: (error: string | null) => void;
  clearError: () => void;
  updateCameraStatus: (device: string, status: CameraStatus) => void;
  addRecording: (device: string, recording: RecordingResponse) => void;
  removeRecording: (device: string) => void;
  setConnectionStatus: (isConnected: boolean) => void;
}

/**
 * Camera store type
 */
type CameraStore = CameraState & CameraActions;

/**
 * Create camera store
 */
export const useCameraStore = create<CameraStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      cameras: [],
      selectedCamera: null,
      serverInfo: null,
      isLoading: false,
      isRefreshing: false,
      isConnecting: false,
      error: null,
      activeRecordings: new Map(),
      wsService: null,
      isConnected: false,

      // Initialization
      initialize: async (wsUrl = 'ws://localhost:8002/ws') => {
        try {
          set({ isLoading: true, isConnecting: true, error: null });
          
          // For testing without server, use mock data
          if (process.env.NODE_ENV === 'development' && !wsUrl.includes('localhost:8002')) {
            // Mock data for testing
            setTimeout(() => {
              set({
                cameras: [
                  {
                    device: '/dev/video0',
                    name: 'Test Camera 1',
                    status: 'CONNECTED',
                    capabilities: {
                      resolutions: ['1920x1080', '1280x720'],
                      fps: 30,
                      validation_status: 'confirmed',
                      formats: ['MJPEG', 'YUYV']
                    }
                  },
                  {
                    device: '/dev/video1', 
                    name: 'Test Camera 2',
                    status: 'CONNECTED',
                    capabilities: {
                      resolutions: ['1280x720', '640x480'],
                      fps: 25,
                      validation_status: 'confirmed',
                      formats: ['MJPEG', 'YUYV']
                    }
                  }
                ],
                isConnected: true,
                isConnecting: false,
                isLoading: false,
                error: null
              });
            }, 1000);
            return;
          }
          
          const wsService = createWebSocketService({
            url: wsUrl,
            maxReconnectAttempts: 10,
            reconnectInterval: 1000,
            maxDelay: 30000,
            requestTimeout: 15000,
          });

          // Set up connection event handlers
          wsService.onConnect(() => {
            set({ isConnected: true, isConnecting: false });
            get().clearError();
          });

          wsService.onDisconnect(() => {
            set({ isConnected: false, isConnecting: false });
          });

          wsService.onError((error) => {
            set({ 
              error: error.message,
              isConnecting: false,
              isConnected: false 
            });
          });

          // Connect to WebSocket
          await wsService.connect();
          
          // Set up message handler for real-time updates
          wsService.onMessage((message) => {
            if ('method' in message) {
              switch (message.method) {
                case 'camera_status_update': {
                  const { device, status } = message.params as { device: string; status: CameraStatus };
                  get().updateCameraStatus(device, status);
                  break;
                }
                case 'recording_status_update': {
                  const recordingUpdate = message.params as { device: string; recording: RecordingResponse };
                  get().addRecording(recordingUpdate.device, recordingUpdate.recording);
                  break;
                }
                case 'recording_completed': {
                  const completedDevice = (message.params as { device: string }).device;
                  get().removeRecording(completedDevice);
                  break;
                }
              }
            }
          });

          set({ wsService });
          
          // Load initial data
          await get().refreshCameras();
          await get().getServerInfo();
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to initialize camera store',
            isLoading: false,
            isConnecting: false,
            isConnected: false
          });
        } finally {
          set({ isLoading: false });
        }
      },

      disconnect: () => {
        const { wsService } = get();
        if (wsService) {
          wsService.disconnect();
        }
        set({ 
          wsService: null,
          isConnected: false,
          isConnecting: false
        });
      },

      // Camera operations
      refreshCameras: async () => {
        try {
          set({ isRefreshing: true, error: null });
          const { wsService } = get();
          
          if (!wsService) {
            throw new Error('WebSocket service not initialized');
          }

          if (!wsService.isConnected) {
            throw new Error('WebSocket not connected');
          }

          const result = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as {
            cameras: CameraDevice[];
            total: number;
            connected: number;
          };

          set({ cameras: result.cameras });
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to refresh cameras',
            cameras: [] // Clear cameras on error
          });
        } finally {
          set({ isRefreshing: false });
        }
      },

      selectCamera: (device: string | null) => {
        set({ selectedCamera: device });
      },

      getCameraStatus: async (device: string): Promise<CameraDevice | null> => {
        try {
          const { wsService } = get();
          
          if (!wsService) {
            throw new Error('WebSocket service not initialized');
          }

          if (!wsService.isConnected) {
            throw new Error('WebSocket not connected');
          }

          const result = await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device }) as CameraDevice;
          
          // Update camera in list if it exists
          set((state) => ({
            cameras: state.cameras.map(camera => 
              camera.device === device ? result : camera
            )
          }));

          return result;
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : `Failed to get status for camera ${device}` 
          });
          return null;
        }
      },

      // Recording operations
      startRecording: async (device: string, duration = 60, format = 'mp4'): Promise<RecordingResponse | null> => {
        try {
          const { wsService } = get();
          
          if (!wsService) {
            throw new Error('WebSocket service not initialized');
          }

          if (!wsService.isConnected) {
            throw new Error('WebSocket not connected');
          }

          const result = await wsService.call(RPC_METHODS.START_RECORDING, {
            device,
            duration,
            format,
          }, true) as RecordingResponse; // Require authentication for protected operations

          if (result.success) {
            set((state) => {
              const newRecordings = new Map(state.activeRecordings);
              newRecordings.set(device, result);
              return { activeRecordings: newRecordings };
            });
          }

          return result;
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : `Failed to start recording for camera ${device}` 
          });
          return null;
        }
      },

      stopRecording: async (device: string): Promise<RecordingResponse | null> => {
        try {
          const { wsService } = get();
          
          if (!wsService) {
            throw new Error('WebSocket service not initialized');
          }

          if (!wsService.isConnected) {
            throw new Error('WebSocket not connected');
          }

          const result = await wsService.call(RPC_METHODS.STOP_RECORDING, { device }, true) as RecordingResponse; // Require authentication for protected operations

          if (result.success) {
            set((state) => {
              const newRecordings = new Map(state.activeRecordings);
              newRecordings.delete(device);
              return { activeRecordings: newRecordings };
            });
          }

          return result;
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : `Failed to stop recording for camera ${device}` 
          });
          return null;
        }
      },

      // Snapshot operations
      takeSnapshot: async (device: string, format = 'jpg', quality = 85): Promise<SnapshotResponse | null> => {
        try {
          const { wsService } = get();
          
          if (!wsService) {
            throw new Error('WebSocket service not initialized');
          }

          if (!wsService.isConnected) {
            throw new Error('WebSocket not connected');
          }

          const result = await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, {
            device,
            format,
            quality,
          }, true) as SnapshotResponse; // Require authentication for protected operations

          return result;
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : `Failed to take snapshot for camera ${device}` 
          });
          return null;
        }
      },

      // Server operations
      getServerInfo: async (): Promise<ServerInfo | null> => {
        try {
          const { wsService } = get();
          
          if (!wsService) {
            throw new Error('WebSocket service not initialized');
          }

          if (!wsService.isConnected) {
            throw new Error('WebSocket not connected');
          }

          // Server info not implemented in current API - using ping for health check
          const result = await wsService.call(RPC_METHODS.PING, {}) as string;
                      set({ serverInfo: { version: '1.0', uptime: 0, cameras_connected: 0, total_recordings: 0, total_snapshots: 0 } });
          return result;
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to get server info' 
          });
          return null;
        }
      },

      pingServer: async (): Promise<boolean> => {
        try {
          const { wsService } = get();
          
          if (!wsService) {
            throw new Error('WebSocket service not initialized');
          }

          if (!wsService.isConnected) {
            throw new Error('WebSocket not connected');
          }

          const result = await wsService.call(RPC_METHODS.PING, {});
          return result === 'pong';
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to ping server' 
          });
          return false;
        }
      },

      // State management
      setError: (error: string | null) => {
        set({ error });
      },

      clearError: () => {
        set({ error: null });
      },

      updateCameraStatus: (device: string, status: CameraStatus) => {
        set((state) => ({
          cameras: state.cameras.map(camera => 
            camera.device === device 
              ? { ...camera, status }
              : camera
          )
        }));
      },

      addRecording: (device: string, recording: RecordingResponse) => {
        set((state) => {
          const newRecordings = new Map(state.activeRecordings);
          newRecordings.set(device, recording);
          return { activeRecordings: newRecordings };
        });
      },

      removeRecording: (device: string) => {
        set((state) => {
          const newRecordings = new Map(state.activeRecordings);
          newRecordings.delete(device);
          return { activeRecordings: newRecordings };
        });
      },

      setConnectionStatus: (isConnected: boolean) => {
        set({ isConnected });
      },
    }),
    {
      name: 'camera-store',
    }
  )
); 