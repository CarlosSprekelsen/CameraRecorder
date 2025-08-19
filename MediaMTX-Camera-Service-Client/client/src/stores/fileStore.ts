/**
 * File management state store
 * Handles file listing, downloading, and file operations
 * 
 * Implements:
 * - File listing with pagination
 * - File download functionality
 * - File metadata management
 * - Real-time file updates
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { FileItem, FileType } from '../types';
import { RPC_METHODS } from '../types';
import { createWebSocketService, type WebSocketService } from '../services/websocket';
import { normalizeFileListResponse } from '../services/apiNormalizer';

/**
 * File store state interface
 */
interface FileState {
  // File data
  recordings: FileItem[] | null;
  snapshots: FileItem[] | null;
  
  // Loading states
  isLoading: boolean;
  isDownloading: boolean;
  
  // Error state
  error: string | null;
  
  // WebSocket service
  wsService: WebSocketService | null;
  
  // Connection state
  isConnected: boolean;
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
  
  // State management
  setError: (error: string | null) => void;
  clearError: () => void;
  setConnectionStatus: (isConnected: boolean) => void;
  updateFileList: (fileType: FileType, files: FileItem[]) => void;
}

/**
 * File store type
 */
type FileStore = FileState & FileActions;

/**
 * Create file store
 */
export const useFileStore = create<FileStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      recordings: null,
      snapshots: null,
      isLoading: false,
      isDownloading: false,
      error: null,
      wsService: null,
      isConnected: false,

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
          snapshots: null 
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
          const response = await wsService.call(RPC_METHODS.LIST_RECORDINGS, {
            limit,
            offset
          });

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
          const response = await wsService.call(RPC_METHODS.LIST_SNAPSHOTS, {
            limit,
            offset
          });

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

          // Create a temporary anchor element to trigger download
          const link = document.createElement('a');
          link.href = downloadUrl;
          link.download = filename;
          link.target = '_blank';
          
          // Append to body, click, and remove
          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);

          // TODO: Add progress tracking for large files
          // TODO: Add authentication headers if required
          
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to download file' 
          });
        } finally {
          set({ isDownloading: false });
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
    }),
    {
      name: 'file-store',
    }
  )
);
