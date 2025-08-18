/**
 * Camera state management store
 * Sprint 3: Enhanced for real server integration
 * Handles camera list, selected camera, and camera operations
 * 
 * Sprint 3 Updates:
 * - Real server integration with MediaMTX Camera Service
 * - Improved error handling and recovery
 * - Better loading states and user feedback
 * - Real-time camera status updates via WebSocket notifications
 * - Enhanced connection state management
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type {
  CameraDevice,
  CameraStatus,
  RecordingResponse,
  SnapshotResponse,
  ServerInfo,
  CameraStatusNotification,
  RecordingStatusNotification,
} from '../types';
import { RPC_METHODS, NOTIFICATION_METHODS } from '../types';
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
  
  // Real-time update tracking
  lastUpdate: Date | null;
  updateCount: number;
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
  
  // Real-time updates
  handleNotification: (notification: CameraStatusNotification | RecordingStatusNotification) => void;
  incrementUpdateCount: () => void;
}

/**
 * Camera store type
 */
type CameraStore = CameraState & CameraActions;

/**
 * Create camera store
 * Sprint 3: Enhanced for real server integration
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
      lastUpdate: null,
      updateCount: 0,

      // Initialization
      initialize: async (wsUrl = 'ws://localhost:8002/ws') => {
        try {
          set({ isLoading: true, isConnecting: true, error: null });
          
          console.log('🚀 Initializing camera store with real server integration');
          
          // Real server integration - no mock data
          console.log('🔌 Connecting to real MediaMTX server for camera list integration');
          
          console.log('🔌 Creating WebSocket service for real server connection');
          const wsService = createWebSocketService({
            url: wsUrl,
            maxReconnectAttempts: 10,
            reconnectInterval: 1000,
            maxDelay: 30000,
            requestTimeout: 15000,
          });

          // Set up connection event handlers
          wsService.onConnect(() => {
            console.log('✅ WebSocket connected to MediaMTX server');
            set({ isConnected: true, isConnecting: false });
            get().clearError();
          });

          wsService.onDisconnect(() => {
            console.log('🔌 WebSocket disconnected from MediaMTX server');
            set({ isConnected: false, isConnecting: false });
          });

          wsService.onError((error) => {
            console.error('❌ WebSocket error:', error);
            set({ 
              error: error.message,
              isConnecting: false,
              isConnected: false 
            });
          });

          // Connect to WebSocket
          console.log('🔌 Connecting to MediaMTX server...');
          await wsService.connect();
          
          // Set up message handler for real-time updates
          wsService.onMessage((message) => {
            if ('method' in message) {
              console.log('📢 Received notification:', message.method);
              get().handleNotification(message as CameraStatusNotification | RecordingStatusNotification);
            }
          });

          set({ wsService });
          
          // Load initial data
          console.log('📋 Loading initial camera data...');
          await get().refreshCameras();
          await get().getServerInfo();
          
          console.log('✅ Camera store initialization complete');
          
        } catch (error) {
          console.error('❌ Camera store initialization failed:', error);
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
        console.log('🔌 Disconnecting camera store');
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

          console.log('📋 Refreshing camera list from MediaMTX server');
          const result = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as {
            cameras: CameraDevice[];
            total: number;
            connected: number;
          };

          console.log(`📊 Found ${result.cameras.length} cameras (${result.connected} connected)`);
          set({ cameras: result.cameras });
          
        } catch (error) {
          console.error('❌ Failed to refresh cameras:', error);
          set({ 
            error: error instanceof Error ? error.message : 'Failed to refresh cameras',
            cameras: [] // Clear cameras on error
          });
        } finally {
          set({ isRefreshing: false });
        }
      },

      selectCamera: (device: string | null) => {
        console.log('📷 Selected camera:', device);
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

          console.log(`📷 Getting status for camera: ${device}`);
          const result = await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device }) as CameraDevice;
          
          // Update camera in list if it exists
          set((state) => ({
            cameras: state.cameras.map(camera => 
              camera.device === device ? result : camera
            )
          }));

          return result;
          
        } catch (error) {
          console.error(`❌ Failed to get status for camera ${device}:`, error);
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

          console.log(`🎬 Starting recording for camera ${device} (duration: ${duration}s, format: ${format})`);
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
            console.log(`✅ Recording started for camera ${device}`);
          }

          return result;
          
        } catch (error) {
          console.error(`❌ Failed to start recording for camera ${device}:`, error);
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

          console.log(`⏹️ Stopping recording for camera ${device}`);
          const result = await wsService.call(RPC_METHODS.STOP_RECORDING, { device }, true) as RecordingResponse; // Require authentication for protected operations

          if (result.success) {
            set((state) => {
              const newRecordings = new Map(state.activeRecordings);
              newRecordings.delete(device);
              return { activeRecordings: newRecordings };
            });
            console.log(`✅ Recording stopped for camera ${device}`);
          }

          return result;
          
        } catch (error) {
          console.error(`❌ Failed to stop recording for camera ${device}:`, error);
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

          console.log(`📸 Taking snapshot for camera ${device} (format: ${format}, quality: ${quality})`);
          const result = await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, {
            device,
            format,
            quality,
          }, true) as SnapshotResponse; // Require authentication for protected operations

          console.log(`✅ Snapshot taken for camera ${device}`);
          return result;
          
        } catch (error) {
          console.error(`❌ Failed to take snapshot for camera ${device}:`, error);
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

          console.log('ℹ️ Getting server information');
          const result = await wsService.call('get_server_info', {}) as ServerInfo;
          set({ serverInfo: result });
          return result;
          
        } catch (error) {
          console.error('❌ Failed to get server info:', error);
          set({ 
            error: error instanceof Error ? error.message : 'Failed to get server information' 
          });
          return null;
        }
      },

      pingServer: async (): Promise<boolean> => {
        try {
          const { wsService } = get();
          
          if (!wsService) {
            return false;
          }

          if (!wsService.isConnected) {
            return false;
          }

          console.log('🏓 Pinging server');
          const result = await wsService.call('ping', {});
          return result === 'pong';
          
        } catch (error) {
          console.error('❌ Server ping failed:', error);
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
        console.log(`📷 Camera status update: ${device} -> ${status}`);
        set((state) => ({
          cameras: state.cameras.map(camera => 
            camera.device === device ? { ...camera, status } : camera
          ),
          lastUpdate: new Date(),
          updateCount: state.updateCount + 1
        }));
      },

      addRecording: (device: string, recording: RecordingResponse) => {
        console.log(`🎬 Adding recording for camera ${device}`);
        set((state) => {
          const newRecordings = new Map(state.activeRecordings);
          newRecordings.set(device, recording);
          return { 
            activeRecordings: newRecordings,
            lastUpdate: new Date(),
            updateCount: state.updateCount + 1
          };
        });
      },

      removeRecording: (device: string) => {
        console.log(`🎬 Removing recording for camera ${device}`);
        set((state) => {
          const newRecordings = new Map(state.activeRecordings);
          newRecordings.delete(device);
          return { 
            activeRecordings: newRecordings,
            lastUpdate: new Date(),
            updateCount: state.updateCount + 1
          };
        });
      },

      setConnectionStatus: (isConnected: boolean) => {
        set({ isConnected });
      },

      // Real-time updates
      handleNotification: (notification: CameraStatusNotification | RecordingStatusNotification) => {
        console.log('📢 Handling notification:', notification.method);
        
        switch (notification.method) {
          case NOTIFICATION_METHODS.CAMERA_STATUS_UPDATE: {
            const { device, status } = notification.params as { device: string; status: CameraStatus };
            get().updateCameraStatus(device, status);
            break;
          }
          case NOTIFICATION_METHODS.RECORDING_STATUS_UPDATE: {
            const recordingUpdate = notification.params as { device: string; recording: RecordingResponse };
            get().addRecording(recordingUpdate.device, recordingUpdate.recording);
            break;
          }
          default:
            console.warn('⚠️ Unknown notification method:', notification.method);
        }
      },

      incrementUpdateCount: () => {
        set((state) => ({ 
          updateCount: state.updateCount + 1,
          lastUpdate: new Date()
        }));
      },
    }),
    {
      name: 'camera-store',
    }
  )
); 