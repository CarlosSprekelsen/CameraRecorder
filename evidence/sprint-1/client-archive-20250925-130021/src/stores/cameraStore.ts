/**
 * Camera Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around CameraService
 * following the modular store pattern established in connection/
 * 
 * Responsibilities:
 * - Camera list management
 * - Camera status tracking
 * - Camera operations (start/stop recording, snapshots)
 * 
 * Architecture Compliance:
 * - Single responsibility (camera operations only)
 * - Uses service layer abstraction
 * - Provides predictable state interface for components
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';
import { recordingManagerService } from '../services/recordingManagerService';
import type { CameraDevice, CameraStatus } from '../types/camera';

// State interface
interface CameraStoreState {
  // Camera list state
  cameras: CameraDevice[];
  selectedCamera: CameraDevice | null;
  
  // Loading states
  isLoading: boolean;
  isRefreshing: boolean;
  
  // Error state
  error: string | null;
  
  // Camera status
  cameraStatus: Record<string, CameraStatus>;
}

// Actions interface
interface CameraStoreActions {
  // Camera list operations
  getCameraList: () => Promise<void>;
  refreshCameras: () => Promise<void>;
  selectCamera: (device: string) => void;
  
  // Camera operations
  startRecording: (device: string) => Promise<boolean>;
  stopRecording: (device: string) => Promise<boolean>;
  takeSnapshot: (device: string) => Promise<boolean>;
  
  // Status operations
  getCameraStatus: (device: string) => Promise<CameraStatus | null>;
  refreshCameraStatus: (device: string) => Promise<void>;
  
  // Error handling
  clearError: () => void;
  setError: (error: string) => void;
}

// Combined store type
type CameraStore = CameraStoreState & CameraStoreActions;

// Initial state
const initialState: CameraStoreState = {
  cameras: [],
  selectedCamera: null,
  isLoading: false,
  isRefreshing: false,
  error: null,
  cameraStatus: {},
};

// Create store
export const useCameraStore = create<CameraStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      // Camera list operations
      getCameraList: async () => {
        set({ isLoading: true, error: null });
        try {
          const result = await recordingManagerService.getCameraList();
          set({ 
            cameras: result.cameras || [], 
            isLoading: false 
          });
          logger.info('Camera list retrieved', undefined, 'cameraStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get camera list';
          set({ 
            error: errorMessage, 
            isLoading: false 
          });
          logger.error('Failed to get camera list', error, 'cameraStore');
        }
      },
      
      refreshCameras: async () => {
        set({ isRefreshing: true, error: null });
        try {
          await get().getCameraList();
          set({ isRefreshing: false });
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to refresh cameras';
          set({ 
            error: errorMessage, 
            isRefreshing: false 
          });
        }
      },
      
      selectCamera: (device: string) => {
        const { cameras } = get();
        const camera = cameras.find(c => c.device === device);
        set({ selectedCamera: camera || null });
      },
      
      // Camera operations
      startRecording: async (device: string) => {
        try {
          const result = await recordingManagerService.startRecording(device);
          logger.info(`Recording started for camera ${device}`, undefined, 'cameraStore');
          return true;
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to start recording';
          set({ error: errorMessage });
          return false;
        }
      },
      
      stopRecording: async (device: string) => {
        try {
          await recordingManagerService.stopRecording(device);
          logger.info(`Recording stopped for camera ${device}`, undefined, 'cameraStore');
          return true;
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to stop recording';
          set({ error: errorMessage });
          return false;
        }
      },
      
      takeSnapshot: async (device: string) => {
        try {
          const result = await recordingManagerService.takeSnapshot(device);
          logger.info(`Snapshot taken for camera ${device}`, undefined, 'cameraStore');
          return result.status === 'SUCCESS';
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to take snapshot';
          set({ error: errorMessage });
          return false;
        }
      },
      
      // Status operations
      getCameraStatus: async (device: string) => {
        try {
          const status = await recordingManagerService.getCameraStatus(device);
          set(state => ({
            cameraStatus: { ...state.cameraStatus, [device]: status }
          }));
          return status;
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get camera status';
          set({ error: errorMessage });
          return null;
        }
      },
      
      refreshCameraStatus: async (device: string) => {
        try {
          await get().getCameraStatus(device);
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to refresh camera status';
          set({ error: errorMessage });
        }
      },
      
      // Error handling
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    {
      name: 'camera-store',
    }
  )
);

// Export types for components
export type { CameraStoreState, CameraStoreActions };
