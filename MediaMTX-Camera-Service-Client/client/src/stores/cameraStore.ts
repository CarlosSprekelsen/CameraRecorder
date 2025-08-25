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
 * - CONSOLIDATION: Merged scaffolding expectations with working implementation
 */

import { create } from 'zustand';
import type {
  CameraDevice,
  CameraListResponse,
  CameraStatus,
  RecordingSession,
  StartRecordingParams,
  StopRecordingParams,
  SnapshotResult,
  TakeSnapshotParams,
  SnapshotFormat,
  FileListResponse,
  FileListParams,
  FileItem,
  StreamInfo,
  StreamListResponse,
  ServerInfo,
  RecordingStatus,
  RecordingProgress,
} from '../types';
import { RPC_METHODS, NOTIFICATION_METHODS } from '../types';
import type { WebSocketService } from '../services/websocket';
import { errorRecoveryService } from '../services/errorRecoveryService';
import { errorHandlerService } from '../services/errorHandlerService';

export interface CameraStoreState {
  // Camera data
  cameras: CameraDevice[];
  selectedCamera: CameraDevice | null;
  
  // Recording state
  activeRecordings: Map<string, RecordingSession>;
  
  // File management
  recordings: FileItem[];
  snapshots: FileItem[];
  
  // Stream management
  streams: StreamInfo[];
  
  // Server info
  serverInfo: ServerInfo | null;
  
  // Standardized loading state
  isLoading: boolean;
  loadingStates: {
    refreshing: boolean;
    connecting: boolean;
    gettingStatus: boolean;
    takingSnapshot: boolean;
    gettingServerInfo: boolean;
    pinging: boolean;
    gettingRecordings: boolean;
    gettingSnapshots: boolean;
  };
  
  // Standardized error state
  error: string | null;
  lastError: string | null;
  errors: Map<string, string>;
  
  // UI state
  lastUpdate: Date | null;
  updateCount: number;
  
  // WebSocket service
  wsService: WebSocketService | null;
  
  // Real-time update state
  realTimeUpdatesEnabled: boolean;
  recordingProgress: Map<string, RecordingProgress>; // device -> comprehensive progress
  lastRecordingUpdate: Date | null;
  notificationCount: number;
  
  // Actions
  setWebSocketService: (service: WebSocketService) => void;
  
  // Camera operations
  getCameraList: () => Promise<CameraListResponse | null>;
  getCameraStatus: (device: string) => Promise<CameraDevice | null>;
  
  // Stream operations
  getStreams: () => Promise<StreamListResponse | null>;
  
  // Recording operations
  startRecording: (device: string, duration?: number, format?: string) => Promise<RecordingSession | null>;
  stopRecording: (device: string) => Promise<RecordingSession | null>;
  
  // Snapshot operations
  takeSnapshot: (
    device: string, 
    format?: SnapshotFormat, 
    quality?: number,
    filename?: string
  ) => Promise<SnapshotResult | null>;
  
  // Server operations
  getServerInfo: () => Promise<ServerInfo | null>;
  pingServer: () => Promise<boolean>;
  
  // File operations
  getRecordings: () => Promise<FileListResponse | null>;
  getSnapshots: () => Promise<FileListResponse | null>;
  
  // Standardized state management
  setLoading: (loading: boolean) => void;
  setOperationLoading: (operation: string, loading: boolean) => void;
  setError: (error: string | null) => void;
  setLastError: (error: string | null) => void;
  setOperationError: (operation: string, error: string | null) => void;
  clearErrors: () => void;
  reset: () => void;
  cleanup: () => void;
  
  // Legacy compatibility
  loading: boolean; // Alias for isLoading
  isRefreshing: boolean; // Alias for loadingStates.refreshing
  isConnecting: boolean; // Alias for loadingStates.connecting
  connect: (url?: string) => Promise<void>;
  disconnect: () => void;
  refreshCameras: () => Promise<void>;
  selectCamera: (device: string) => void;
}

export const useCameraStore = create<CameraStoreState>((set, get) => ({
  // Initial state
  cameras: [],
  selectedCamera: null,
  activeRecordings: new Map(),
  recordings: [],
  snapshots: [],
  streams: [],
  serverInfo: null,
  
  // Standardized loading state
  isLoading: false,
  loadingStates: {
    refreshing: false,
    connecting: false,
    gettingStatus: false,
    takingSnapshot: false,
    gettingServerInfo: false,
    pinging: false,
    gettingRecordings: false,
    gettingSnapshots: false,
  },
  
  // Standardized error state
  error: null,
  lastError: null,
  errors: new Map(),
  
  // UI state
  lastUpdate: null,
  updateCount: 0,
  wsService: null,
  
  // Real-time update state
  realTimeUpdatesEnabled: true,
  recordingProgress: new Map<string, RecordingProgress>(),
  lastRecordingUpdate: null,
  notificationCount: 0,
  
  // Legacy compatibility
  loading: false, // Alias for isLoading
  isRefreshing: false, // Alias for loadingStates.refreshing
  isConnecting: false, // Alias for loadingStates.connecting

  // Standardized state management
  setLoading: (loading: boolean) => {
    set({ isLoading: loading });
  },
  
  setOperationLoading: (operation: string, loading: boolean) => {
    set((state) => ({
      loadingStates: {
        ...state.loadingStates,
        [operation]: loading,
      },
    }));
  },
  
  setError: (error: string | null) => {
    set({ error });
  },
  
  setLastError: (error: string | null) => {
    set({ lastError: error });
  },
  
  setOperationError: (operation: string, error: string | null) => {
    set((state) => {
      const newErrors = new Map(state.errors);
      if (error) {
        newErrors.set(operation, error);
      } else {
        newErrors.delete(operation);
      }
      return { errors: newErrors };
    });
  },
  
  clearErrors: () => {
    set({ error: null, lastError: null, errors: new Map() });
  },
  
  reset: () => {
    set({
      cameras: [],
      selectedCamera: null,
      activeRecordings: new Map(),
      recordings: [],
      snapshots: [],
      streams: [],
      serverInfo: null,
      isLoading: false,
      loadingStates: {
        refreshing: false,
        connecting: false,
        gettingStatus: false,
        takingSnapshot: false,
        gettingServerInfo: false,
        pinging: false,
        gettingRecordings: false,
        gettingSnapshots: false,
      },
      error: null,
      lastError: null,
      errors: new Map(),
      lastUpdate: null,
      updateCount: 0,
      wsService: null,
      realTimeUpdatesEnabled: true,
      recordingProgress: new Map(),
      lastRecordingUpdate: null,
      notificationCount: 0,
      loading: false,
      isRefreshing: false,
      isConnecting: false,
    });
  },
  
  cleanup: () => {
    const { wsService } = get();
    if (wsService) {
      wsService.disconnect();
    }
    get().reset();
  },
  
  // WebSocket service management
  setWebSocketService: (service: WebSocketService) => {
    set({ wsService: service });
  },

  // CONSOLIDATION: Added missing scaffolding methods
  initialize: async () => {
    const { wsService } = get();
    
    if (wsService) {
      return; // Already initialized
    }

    try {
      set({ isConnecting: true, error: null });
      
      // Initialize WebSocket service
      const { createWebSocketService } = await import('../services/websocket');
      const newWsService = await createWebSocketService({
        url: 'ws://localhost:8002/ws',
        reconnectInterval: 5000,
        maxReconnectAttempts: 5,
      });

      set({ wsService: newWsService });

      // Set up event handlers
      newWsService.onConnect(() => {
        set({ isConnected: true, error: null });
      });

      newWsService.onDisconnect(() => {
        set({ isConnected: false });
      });

      newWsService.onError((error) => {
        set({ error: error.message, isConnected: false });
      });

      // Connect to WebSocket
      await newWsService.connect();
      
      // Load initial data
      await get().refreshCameras();
      
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : 'Failed to initialize camera store',
        isConnected: false 
      });
    } finally {
      set({ isConnecting: false });
    }
  },

  refreshCameras: async () => {
    try {
      set({ isRefreshing: true, error: null });
      await get().getCameraList();
      await get().getStreams();
      set({ isRefreshing: false });
    } catch (error) {
      set({ 
        isRefreshing: false,
        error: error instanceof Error ? error.message : 'Failed to refresh cameras'
      });
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
      cameras: [],
      selectedCamera: null,
      activeRecordings: new Map(),
      streams: []
    });
  },

  selectCamera: (device: string) => {
    const { cameras } = get();
    const camera = cameras.find(c => c.device === device);
    set({ selectedCamera: camera || null });
  },

  // Camera operations
  getCameraList: async (): Promise<CameraListResponse | null> => {
    const { wsService } = get();
    
    if (!wsService) {
      throw new Error('WebSocket service not initialized');
    }

    if (!wsService.isConnected()) {
      throw new Error('WebSocket not connected');
    }

          console.log('Getting camera list with error recovery');
    
    const result = await errorRecoveryService.executeWithRetry(
      async () => {
        const response = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as CameraListResponse;
        set({ cameras: response.cameras });
        return response;
      },
      'getCameraList'
    );
    return result as unknown as CameraListResponse;
  },

  getCameraStatus: async (device: string): Promise<CameraDevice | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected()) {
        throw new Error('WebSocket not connected');
      }

      console.log(`Getting camera status for ${device}`);
      const result = await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device }) as CameraDevice;
      
      // Update camera in list
      set((state) => ({
        cameras: state.cameras.map(camera => 
          camera.device === device ? result : camera
        )
      }));
      
      return result;
      
    } catch (error) {
      console.error(`Failed to get camera status for ${device}:`, error);
      set({ 
        error: error instanceof Error ? error.message : 'Failed to get camera status' 
      });
      return null;
    }
  },

  // Stream operations
  getStreams: async (): Promise<StreamListResponse | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected()) {
        throw new Error('WebSocket not connected');
      }

      console.log('Getting stream list');
      const result = await wsService.call(RPC_METHODS.GET_STREAMS, {}) as StreamListResponse;
      set({ streams: result.streams });
      return result;
      
    } catch (error) {
      console.error('Failed to get stream list:', error);
      set({ 
        error: error instanceof Error ? error.message : 'Failed to get stream list' 
      });
      return null;
    }
  },

  // Recording operations
  startRecording: async (device: string, duration?: number, format?: string): Promise<RecordingSession | null> => {
    const { wsService } = get();
    
    if (!wsService) {
      throw new Error('WebSocket service not initialized');
    }

    if (!wsService.isConnected()) {
      throw new Error('WebSocket not connected');
    }

    const params: StartRecordingParams = {
      device,
      ...(duration && { duration_seconds: duration }),
      ...(format && { format: format as any })
    };

    console.log(`Starting recording for camera ${device} with error recovery`);
    
    const result = await errorRecoveryService.executeWithRetry(
      async () => {
        const response = await wsService.call(RPC_METHODS.START_RECORDING, params as unknown as Record<string, unknown>) as RecordingSession;
        console.log(`Recording started for camera ${device}`);
        return response;
      },
      'startRecording'
    );

    // Add to active recordings
    set((state) => {
      const newRecordings = new Map(state.activeRecordings);
      newRecordings.set(device, result as unknown as RecordingSession);
      return { activeRecordings: newRecordings };
    });

    return result as unknown as RecordingSession;
  },

  stopRecording: async (device: string): Promise<RecordingSession | null> => {
    const { wsService } = get();
    
    if (!wsService) {
      throw new Error('WebSocket service not initialized');
    }

    if (!wsService.isConnected()) {
      throw new Error('WebSocket not connected');
    }

    const params: StopRecordingParams = { device };

    console.log(`Stopping recording for camera ${device} with error recovery`);
    
    const result = await errorRecoveryService.executeWithRetry(
      async () => {
        const response = await wsService.call(RPC_METHODS.STOP_RECORDING, params as unknown as Record<string, unknown>) as RecordingSession;
        console.log(`Recording stopped for camera ${device}`);
        return response;
      },
      'stopRecording'
    );

    // Remove from active recordings
    set((state) => {
      const newRecordings = new Map(state.activeRecordings);
      newRecordings.delete(device);
      return { activeRecordings: newRecordings };
    });

    return result as unknown as RecordingSession;
  },

  // Snapshot operations
  takeSnapshot: async (
    device: string, 
    format: SnapshotFormat = 'jpg', 
    quality: number = 85,
    filename?: string
  ): Promise<SnapshotResult | null> => {
    const { wsService } = get();
    
    if (!wsService) {
      throw new Error('WebSocket service not initialized');
    }

    if (!wsService.isConnected()) {
      throw new Error('WebSocket not connected');
    }

    // Validate quality parameter
    if (quality < 1 || quality > 100) {
      throw new Error('Quality must be between 1 and 100');
    }

    // Validate format parameter
    if (format !== 'jpg' && format !== 'png') {
      throw new Error('Format must be either "jpg" or "png"');
    }

          console.log(`Taking snapshot for camera ${device} with error recovery (format: ${format}, quality: ${quality})`);
    
    const params: TakeSnapshotParams = {
      device,
      format,
      quality,
      ...(filename && { filename })
    };

    const result = await errorRecoveryService.executeWithRetry(
      async () => {
        const response = await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, params as unknown as Record<string, unknown>) as SnapshotResult;
        console.log(`Snapshot taken for camera ${device}:`, response);
        return response;
      },
      'takeSnapshot'
    );
    return result as unknown as SnapshotResult;
  },

  // Server operations
  getServerInfo: async (): Promise<ServerInfo | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected()) {
        throw new Error('WebSocket not connected');
      }

      console.log('Getting server information');
      const result = await wsService.call('get_server_info', {}) as ServerInfo;
      set({ serverInfo: result });
      return result;
      
    } catch (error) {
      console.error('Failed to get server info:', error);
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

      if (!wsService.isConnected()) {
        return false;
      }

      console.log('Pinging server');
      const result = await wsService.call('ping', {});
      return result === 'pong';
      
    } catch (error) {
      console.error('Server ping failed:', error);
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
          console.log(`Camera status update: ${device} -> ${status}`);
    set((state) => ({
      cameras: state.cameras.map(camera => 
        camera.device === device ? { ...camera, status } : camera
      ),
      lastUpdate: new Date(),
      updateCount: state.updateCount + 1
    }));
  },

  addRecording: (device: string, recording: RecordingSession) => {
          console.log(`Adding recording for camera ${device}`);
    set((state) => {
      const newRecordings = new Map(state.activeRecordings);
      newRecordings.set(device, recording);
      return { activeRecordings: newRecordings };
    });
  },

  removeRecording: (device: string) => {
          console.log(`Removing recording for camera ${device}`);
    set((state) => {
      const newRecordings = new Map(state.activeRecordings);
      newRecordings.delete(device);
      return { activeRecordings: newRecordings };
    });
  },

  // File operations
  getRecordings: async (params?: FileListParams): Promise<FileListResponse | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected()) {
        throw new Error('WebSocket not connected');
      }

      console.log('Getting recordings list');
      const result = await wsService.call(RPC_METHODS.LIST_RECORDINGS, (params || {}) as Record<string, unknown>) as FileListResponse;
      
      // Normalize file data
      const normalized = {
        ...result,
        files: result.files.map((file: any) => ({
          ...file,
          created_at: file.modified_time // Use modified_time as created_at fallback
        }))
      };
      
      set({ recordings: normalized.files });
      return normalized;
      
    } catch (error) {
      console.error('Failed to get recordings:', error);
      set({ 
        error: error instanceof Error ? error.message : 'Failed to get recordings' 
      });
      return null;
    }
  },

  getSnapshots: async (params?: FileListParams): Promise<FileListResponse | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected()) {
        throw new Error('WebSocket not connected');
      }

      console.log('📸 Getting snapshots list');
      const result = await wsService.call(RPC_METHODS.LIST_SNAPSHOTS, (params || {}) as Record<string, unknown>) as FileListResponse;
      
      // Normalize file data
      const normalized = {
        ...result,
        files: result.files.map((file: any) => ({
          ...file,
          created_at: file.modified_time // Use modified_time as created_at fallback
        }))
      };
      
      set({ snapshots: normalized.files });
      return normalized;
      
    } catch (error) {
      console.error('❌ Failed to get snapshots:', error);
      set({ 
        error: error instanceof Error ? error.message : 'Failed to get snapshots' 
      });
      return null;
    }
  },

  // Notification handling
  handleNotification: (notification: unknown) => {
    console.log('📡 Handling notification:', notification);
    
    // Type guard for notification structure
    if (notification && typeof notification === 'object' && 'method' in notification) {
      const notif = notification as { method: string; params?: unknown };
      
      switch (notif.method) {
        case NOTIFICATION_METHODS.CAMERA_STATUS_UPDATE:
          if (notif.params && typeof notif.params === 'object' && 'device' in notif.params) {
            const params = notif.params as { device: string; status: CameraStatus };
            get().updateCameraStatus(params.device, params.status);
          }
          break;
          
        case NOTIFICATION_METHODS.RECORDING_STATUS_UPDATE:
          if (notif.params && typeof notif.params === 'object' && 'device' in notif.params) {
            const params = notif.params as { device: string; session_id: string; status: RecordingStatus };
            // Handle recording status update
            console.log('📹 Recording status update:', params);
          }
          break;
          
        default:
          console.log('📡 Unknown notification method:', notif.method);
      }
    }
    
    set({ 
      notificationCount: get().notificationCount + 1,
      lastRecordingUpdate: new Date()
    });
  },

  // Real-time update management
  enableRealTimeUpdates: () => {
    set({ realTimeUpdatesEnabled: true });
    console.log('🔄 Camera store real-time updates enabled');
  },

  disableRealTimeUpdates: () => {
    set({ realTimeUpdatesEnabled: false });
    console.log('⏸️ Camera store real-time updates disabled');
  },

  updateRecordingProgress: (device: string, progress: RecordingProgress) => {
    const { recordingProgress } = get();
    const newProgress = new Map(recordingProgress);
    newProgress.set(device, progress);
    set({ 
      recordingProgress: newProgress,
      lastRecordingUpdate: new Date()
    });
  },

  getRecordingProgress: (device: string) => {
    const { recordingProgress } = get();
    return recordingProgress.get(device) || null;
  },

  clearRecordingProgress: (device: string) => {
    const { recordingProgress } = get();
    const newProgress = new Map(recordingProgress);
    newProgress.delete(device);
    set({ recordingProgress: newProgress });
  },
})); 