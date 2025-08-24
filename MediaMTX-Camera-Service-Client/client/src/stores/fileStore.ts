/**
 * File management state store
 * Handles file listing, downloading, and file operations
 * 
 * Implements:
 * - File listing with pagination
 * - File download functionality
 * - File metadata management
 * - Real-time file updates
 * - File deletion (admin/operator only)
 * - File info retrieval
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { FileItem, FileType } from '../types';
import { RPC_METHODS } from '../types';
import { createWebSocketService, type WebSocketService } from '../services/websocket';
import { normalizeFileListResponse } from '../services/apiNormalizer';
import { errorRecoveryService } from '../services/errorRecoveryService';

/**
 * File info response from server
 */
export interface FileInfoResponse {
  filename: string;
  file_size: number;
  duration?: number; // Only for recordings
  created_time: string;
  download_url: string;
  resolution?: string; // Only for snapshots
}

/**
 * File deletion response from server
 */
export interface FileDeletionResponse {
  filename: string;
  deleted: boolean;
  message: string;
}

/**
 * File store state interface
 */
export interface FileStoreState {
  // File data
  recordings: FileItem[] | null;
  snapshots: FileItem[] | null;
  
  // File details
  selectedFile: FileInfoResponse | null;
  
  // Loading states
  isLoading: boolean;
  isDownloading: boolean;
  isDeleting: boolean;
  isLoadingFileInfo: boolean;
  
  // Error state
  error: string | null;
  
  // WebSocket service
  wsService: WebSocketService | null;
  
  // Connection state
  isConnected: boolean;
  
  // User permissions
  canDeleteFiles: boolean;
}

/**
 * File store actions interface
 */
interface FileActions {
  // Initialization
  initialize: (wsUrl?: string) => Promise<void>;
  disconnect: () => void;
  
  // File operations
  loadRecordings: (limit?: number, offset?: number) => Promise<void>;
  loadSnapshots: (limit?: number, offset?: number) => Promise<void>;
  downloadFile: (fileType: FileType, filename: string) => Promise<void>;
  
  // File info operations
  getRecordingInfo: (filename: string) => Promise<FileInfoResponse>;
  getSnapshotInfo: (filename: string) => Promise<FileInfoResponse>;
  
  // File deletion operations
  deleteRecording: (filename: string) => Promise<FileDeletionResponse>;
  deleteSnapshot: (filename: string) => Promise<FileDeletionResponse>;
  
  // State management
  setError: (error: string | null) => void;
  clearError: () => void;
  setConnectionStatus: (isConnected: boolean) => void;
  updateFileList: (fileType: FileType, files: FileItem[]) => void;
  setSelectedFile: (file: FileInfoResponse | null) => void;
  setCanDeleteFiles: (canDelete: boolean) => void;
}

/**
 * File store type
 */
type FileStore = FileStoreState & FileActions;

/**
 * Create file store
 */
export const useFileStore = create<FileStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      recordings: null,
      snapshots: null,
      selectedFile: null,
      isLoading: false,
      isDownloading: false,
      isDeleting: false,
      isLoadingFileInfo: false,
      error: null,
      wsService: null,
      isConnected: false,
      canDeleteFiles: false,

      // Initialization
      initialize: async (wsUrl = 'ws://localhost:8002/ws') => {
        const { wsService } = get();
        
        if (wsService) {
          return; // Already initialized
        }

        try {
          const newWsService = await createWebSocketService({
            url: wsUrl,
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
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to initialize file store',
            isConnected: false 
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
          recordings: null,
          snapshots: null,
          selectedFile: null
        });
      },

      // File operations
      loadRecordings: async (limit = 20, offset = 0) => {
        const { wsService } = get();
        
        if (!wsService) {
          set({ error: 'WebSocket not connected' });
          return;
        }

        set({ isLoading: true, error: null });

        try {
          const response = await errorRecoveryService.executeWithRetry(
            async () => {
              return await wsService.call(RPC_METHODS.LIST_RECORDINGS, {
                limit,
                offset
              });
            },
            'loadRecordings'
          );

          const normalized = normalizeFileListResponse(response);
          set({ recordings: normalized.files as FileItem[] });

        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to load recordings',
            recordings: null 
          });
        } finally {
          set({ isLoading: false });
        }
      },

      loadSnapshots: async (limit = 20, offset = 0) => {
        const { wsService } = get();
        
        if (!wsService) {
          set({ error: 'WebSocket not connected' });
          return;
        }

        set({ isLoading: true, error: null });

        try {
          const response = await errorRecoveryService.executeWithRetry(
            async () => {
              return await wsService.call(RPC_METHODS.LIST_SNAPSHOTS, {
                limit,
                offset
              });
            },
            'loadSnapshots'
          );

          const normalized = normalizeFileListResponse(response);
          set({ snapshots: normalized.files as FileItem[] });

        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to load snapshots',
            snapshots: null 
          });
        } finally {
          set({ isLoading: false });
        }
      },

      downloadFile: async (fileType: FileType, filename: string) => {
        set({ isDownloading: true, error: null });

        try {
          // Construct download URL based on file type
          // File downloads are served by the health server on port 8003
          const baseUrl = window.location.protocol === 'https:' 
            ? 'https://localhost:8003' 
            : 'http://localhost:8003';
          const downloadUrl = `${baseUrl}/files/${fileType}/${encodeURIComponent(filename)}`;

          // First check if file exists by making a HEAD request
          const headResponse = await fetch(downloadUrl, { method: 'HEAD' });
          
          if (!headResponse.ok) {
            if (headResponse.status === 404) {
              throw new Error('File not found');
            } else {
              throw new Error(`Download failed: ${headResponse.status} ${headResponse.statusText}`);
            }
          }

          // File exists, proceed with download
          const response = await fetch(downloadUrl);
          
          if (!response.ok) {
            throw new Error(`Download failed: ${response.status} ${response.statusText}`);
          }

          // Create blob and download
          const blob = await response.blob();
          const url = window.URL.createObjectURL(blob);
          
          const link = document.createElement('a');
          link.href = url;
          link.download = filename;
          link.target = '_blank';
          
          // Append to body, click, and remove
          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);
          
          // Clean up blob URL
          window.URL.revokeObjectURL(url);
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Download failed' 
          });
        } finally {
          set({ isDownloading: false });
        }
      },

      // File info operations
      getRecordingInfo: async (filename: string): Promise<FileInfoResponse> => {
        const { wsService } = get();
        
        if (!wsService) {
          throw new Error('WebSocket not connected');
        }

        set({ isLoadingFileInfo: true, error: null });

        try {
          const response = await wsService.call('get_recording_info', {
            filename
          });

          const fileInfo: FileInfoResponse = {
            filename: (response as any).filename,
            file_size: (response as any).file_size,
            duration: (response as any).duration,
            created_time: (response as any).created_time,
            download_url: (response as any).download_url,
          };

          set({ selectedFile: fileInfo });
          return fileInfo;

        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get recording info';
          set({ error: errorMessage });
          throw new Error(errorMessage);
        } finally {
          set({ isLoadingFileInfo: false });
        }
      },

      getSnapshotInfo: async (filename: string): Promise<FileInfoResponse> => {
        const { wsService } = get();
        
        if (!wsService) {
          throw new Error('WebSocket not connected');
        }

        set({ isLoadingFileInfo: true, error: null });

        try {
          const response = await wsService.call('get_snapshot_info', {
            filename
          });

          const fileInfo: FileInfoResponse = {
            filename: (response as any).filename,
            file_size: (response as any).file_size,
            created_time: (response as any).created_time,
            download_url: (response as any).download_url,
            resolution: (response as any).resolution,
          };

          set({ selectedFile: fileInfo });
          return fileInfo;

        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get snapshot info';
          set({ error: errorMessage });
          throw new Error(errorMessage);
        } finally {
          set({ isLoadingFileInfo: false });
        }
      },

      // File deletion operations
      deleteRecording: async (filename: string): Promise<FileDeletionResponse> => {
        const { wsService, canDeleteFiles } = get();
        
        if (!canDeleteFiles) {
          throw new Error('Insufficient permissions to delete files');
        }
        
        if (!wsService) {
          throw new Error('WebSocket not connected');
        }

        set({ isDeleting: true, error: null });

        try {
          const response = await wsService.call('delete_recording', {
            filename
          });

          const deletionResponse: FileDeletionResponse = {
            filename: (response as any).filename,
            deleted: (response as any).deleted,
            message: (response as any).message,
          };

          // Refresh recordings list after deletion
          if (deletionResponse.deleted) {
            const { recordings } = get();
            if (recordings) {
              const updatedRecordings = recordings.filter(file => file.filename !== filename);
              set({ recordings: updatedRecordings });
            }
          }

          return deletionResponse;

        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to delete recording';
          set({ error: errorMessage });
          throw new Error(errorMessage);
        } finally {
          set({ isDeleting: false });
        }
      },

      deleteSnapshot: async (filename: string): Promise<FileDeletionResponse> => {
        const { wsService, canDeleteFiles } = get();
        
        if (!canDeleteFiles) {
          throw new Error('Insufficient permissions to delete files');
        }
        
        if (!wsService) {
          throw new Error('WebSocket not connected');
        }

        set({ isDeleting: true, error: null });

        try {
          const response = await wsService.call('delete_snapshot', {
            filename
          });

          const deletionResponse: FileDeletionResponse = {
            filename: (response as any).filename,
            deleted: (response as any).deleted,
            message: (response as any).message,
          };

          // Refresh snapshots list after deletion
          if (deletionResponse.deleted) {
            const { snapshots } = get();
            if (snapshots) {
              const updatedSnapshots = snapshots.filter(file => file.filename !== filename);
              set({ snapshots: updatedSnapshots });
            }
          }

          return deletionResponse;

        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to delete snapshot';
          set({ error: errorMessage });
          throw new Error(errorMessage);
        } finally {
          set({ isDeleting: false });
        }
      },

      // State management
      setError: (error: string | null) => {
        set({ error });
      },

      clearError: () => {
        set({ error: null });
      },

      setConnectionStatus: (isConnected: boolean) => {
        set({ isConnected });
      },

      updateFileList: (fileType: FileType, files: FileItem[]) => {
        if (fileType === 'recordings') {
          set({ recordings: files });
        } else {
          set({ snapshots: files });
        }
      },

      setSelectedFile: (file: FileInfoResponse | null) => {
        set({ selectedFile: file });
      },

      setCanDeleteFiles: (canDelete: boolean) => {
        set({ canDeleteFiles: canDelete });
      },
    }),
    {
      name: 'file-store',
    }
  )
);
