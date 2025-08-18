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
import type {
  CameraDevice,
  CameraListResponse,
  CameraStatus,
  RecordingSession,
  RecordingStatus,
  StartRecordingParams,
  StopRecordingParams,
  SnapshotResult,
  TakeSnapshotParams,
  SnapshotFormat,
  FileListResponse,
  FileListParams,
  FileType,
  FileItem,
  CameraStatusUpdateParams,
  RecordingStatusUpdateParams,
  ServerInfo,
} from '../types';
import { RPC_METHODS, NOTIFICATION_METHODS } from '../types';
import type { WebSocketService } from '../services/websocket';

interface CameraState {
  // Camera data
  cameras: CameraDevice[];
  selectedCamera: CameraDevice | null;
  
  // Recording state
  activeRecordings: Map<string, RecordingSession>;
  
  // File management
  recordings: FileItem[];
  snapshots: FileItem[];
  
  // Server info
  serverInfo: ServerInfo | null;
  
  // UI state
  loading: boolean;
  error: string | null;
  lastUpdate: Date | null;
  updateCount: number;
  
  // WebSocket service
  wsService: WebSocketService | null;
  
  // Real-time update state
  realTimeUpdatesEnabled: boolean;
  recordingProgress: Map<string, number>; // device -> progress percentage
  lastRecordingUpdate: Date | null;
  notificationCount: number;
  
  // Actions
  setWebSocketService: (service: WebSocketService) => void;
  
  // Camera operations
  getCameraList: () => Promise<CameraListResponse | null>;
  getCameraStatus: (device: string) => Promise<CameraDevice | null>;
  
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
  
  // State management
  setError: (error: string | null) => void;
  clearError: () => void;
  updateCameraStatus: (device: string, status: CameraStatus) => void;
  addRecording: (device: string, recording: RecordingSession) => void;
  removeRecording: (device: string) => void;
  
  // File operations
  getRecordings: (params?: FileListParams) => Promise<FileListResponse | null>;
  getSnapshots: (params?: FileListParams) => Promise<FileListResponse | null>;
  
  // Notification handling
  handleNotification: (notification: any) => void;
  
  // Real-time update management
  enableRealTimeUpdates: () => void;
  disableRealTimeUpdates: () => void;
  updateRecordingProgress: (device: string, progress: number) => void;
  getRecordingProgress: (device: string) => number;
  clearRecordingProgress: (device: string) => void;
}

export const useCameraStore = create<CameraState>((set, get) => ({
  // Initial state
  cameras: [],
  selectedCamera: null,
  activeRecordings: new Map(),
  recordings: [],
  snapshots: [],
  serverInfo: null,
  loading: false,
  error: null,
  lastUpdate: null,
  updateCount: 0,
  wsService: null,
  
  // Real-time update state
  realTimeUpdatesEnabled: true,
  recordingProgress: new Map(),
  lastRecordingUpdate: null,
  notificationCount: 0,

  // WebSocket service management
  setWebSocketService: (service: WebSocketService) => {
    set({ wsService: service });
  },

  // Camera operations
  getCameraList: async (): Promise<CameraListResponse | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected) {
        throw new Error('WebSocket not connected');
      }

      console.log('📷 Getting camera list');
      const result = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}) as CameraListResponse;
      
      set({ 
        cameras: result.cameras,
        lastUpdate: new Date(),
        updateCount: get().updateCount + 1
      });
      
      return result;
      
    } catch (error) {
      console.error('❌ Failed to get camera list:', error);
      set({ 
        error: error instanceof Error ? error.message : 'Failed to get camera list' 
      });
      return null;
    }
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

      console.log(`📷 Getting status for camera ${device}`);
      const result = await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device }) as CameraDevice;
      
      // Update camera in list
      set((state) => ({
        cameras: state.cameras.map(camera => 
          camera.device === device ? result : camera
        ),
        lastUpdate: new Date(),
        updateCount: state.updateCount + 1
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
  startRecording: async (device: string, duration?: number, format?: string): Promise<RecordingSession | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected) {
        throw new Error('WebSocket not connected');
      }

      const params: StartRecordingParams = {
        device,
        ...(duration && { duration_seconds: duration }),
        ...(format && { format: format as any })
      };

      console.log(`🎬 Starting recording for camera ${device}`);
      const result = await wsService.call(RPC_METHODS.START_RECORDING, params as Record<string, unknown>, true) as RecordingSession;

      // Add to active recordings
      set((state) => {
        const newRecordings = new Map(state.activeRecordings);
        newRecordings.set(device, result);
        return { activeRecordings: newRecordings };
      });

      console.log(`✅ Recording started for camera ${device}`);
      return result;
      
    } catch (error) {
      console.error(`❌ Failed to start recording for camera ${device}:`, error);
      set({ 
        error: error instanceof Error ? error.message : `Failed to start recording for camera ${device}` 
      });
      return null;
    }
  },

  stopRecording: async (device: string): Promise<RecordingSession | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected) {
        throw new Error('WebSocket not connected');
      }

      const params: StopRecordingParams = { device };

      console.log(`⏹️ Stopping recording for camera ${device}`);
      const result = await wsService.call(RPC_METHODS.STOP_RECORDING, params as Record<string, unknown>, true) as RecordingSession;

      // Remove from active recordings
      set((state) => {
        const newRecordings = new Map(state.activeRecordings);
        newRecordings.delete(device);
        return { activeRecordings: newRecordings };
      });
      console.log(`✅ Recording stopped for camera ${device}`);

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
  takeSnapshot: async (
    device: string, 
    format: SnapshotFormat = 'jpg', 
    quality: number = 85,
    filename?: string
  ): Promise<SnapshotResult | null> => {
    try {
      const { wsService } = get();
      
      if (!wsService) {
        throw new Error('WebSocket service not initialized');
      }

      if (!wsService.isConnected) {
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

      console.log(`📸 Taking snapshot for camera ${device} (format: ${format}, quality: ${quality})`);
      
      const params: TakeSnapshotParams = {
        device,
        format,
        quality,
        ...(filename && { filename })
      };

      const result = await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, params as Record<string, unknown>, true) as SnapshotResult;

      console.log(`✅ Snapshot taken for camera ${device}:`, result);
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

  addRecording: (device: string, recording: RecordingSession) => {
    console.log(`🎬 Adding recording for camera ${device}`);
    set((state) => {
      const newRecordings = new Map(state.activeRecordings);
      newRecordings.set(device, recording);
      return { activeRecordings: newRecordings };
    });
  },

  removeRecording: (device: string) => {
    console.log(`🎬 Removing recording for camera ${device}`);
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

      if (!wsService.isConnected) {
        throw new Error('WebSocket not connected');
      }

      console.log('📁 Getting recordings list');
      const result = await wsService.call(RPC_METHODS.LIST_RECORDINGS, (params || {}) as Record<string, unknown>) as FileListResponse;
      
      // Normalize file data
      const normalized = {
        ...result,
        files: result.files.map(file => ({
          ...file,
          created_at: file.modified_time // Use modified_time as created_at fallback
        }))
      };
      
      set({ recordings: normalized.files });
      return normalized;
      
    } catch (error) {
      console.error('❌ Failed to get recordings:', error);
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

      if (!wsService.isConnected) {
        throw new Error('WebSocket not connected');
      }

      console.log('📸 Getting snapshots list');
      const result = await wsService.call(RPC_METHODS.LIST_SNAPSHOTS, (params || {}) as Record<string, unknown>) as FileListResponse;
      
      // Normalize file data
      const normalized = {
        ...result,
        files: result.files.map(file => ({
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
  handleNotification: (notification: any) => {
    console.log('📡 Handling notification:', notification);
    
    if (!get().realTimeUpdatesEnabled) {
      return;
    }
    
    // Update notification count
    set(state => ({ notificationCount: state.notificationCount + 1 }));
    
    if (notification.method === NOTIFICATION_METHODS.CAMERA_STATUS_UPDATE) {
      const statusUpdate = notification.params as CameraStatusUpdateParams;
      get().updateCameraStatus(statusUpdate.device, statusUpdate.status as CameraStatus);
    } else if (notification.method === NOTIFICATION_METHODS.RECORDING_STATUS_UPDATE) {
      const recordingUpdate = notification.params as RecordingStatusUpdateParams;
      
      if (recordingUpdate.status === 'STARTED') {
        // Add to active recordings
        const recording: RecordingSession = {
          device: recordingUpdate.device,
          session_id: `session_${Date.now()}`,
          filename: recordingUpdate.filename,
          status: 'STARTED',
          start_time: new Date().toISOString(),
          duration: recordingUpdate.duration,
          format: 'mp4'
        };
        get().addRecording(recordingUpdate.device, recording);
        
        // Initialize recording progress
        get().updateRecordingProgress(recordingUpdate.device, 0);
      } else if (recordingUpdate.status === 'STOPPED') {
        // Remove from active recordings
        get().removeRecording(recordingUpdate.device);
        
        // Clear recording progress
        get().clearRecordingProgress(recordingUpdate.device);
      } else if (recordingUpdate.status === 'RECORDING') {
        // Update recording progress based on duration
        const progress = Math.min(100, (recordingUpdate.duration / 60) * 100); // Assuming 1 minute = 100%
        get().updateRecordingProgress(recordingUpdate.device, progress);
      }
    } else {
      console.warn('⚠️ Unknown notification method:', notification.method);
    }
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

  updateRecordingProgress: (device: string, progress: number) => {
    const { recordingProgress } = get();
    const newProgress = new Map(recordingProgress);
    newProgress.set(device, Math.max(0, Math.min(100, progress)));
    set({ 
      recordingProgress: newProgress,
      lastRecordingUpdate: new Date()
    });
  },

  getRecordingProgress: (device: string) => {
    const { recordingProgress } = get();
    return recordingProgress.get(device) || 0;
  },

  clearRecordingProgress: (device: string) => {
    const { recordingProgress } = get();
    const newProgress = new Map(recordingProgress);
    newProgress.delete(device);
    set({ recordingProgress: newProgress });
  },
})); 