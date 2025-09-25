/**
 * File Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around FileService
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';
import { RPC_METHODS } from '../types/rpc';
import { errorRecoveryService } from '../services/errorRecoveryService';
import { filesService } from '../services/filesService';

interface FileStoreState {
  files: any[];
  isLoading: boolean;
  error: string | null;
}

interface FileStoreActions {
  getFiles: () => Promise<void>;
  downloadFile: (fileId: string) => Promise<void>;
  deleteFile: (fileId: string) => Promise<void>;
  clearError: () => void;
  setError: (error: string) => void;
}

type FileStore = FileStoreState & FileStoreActions;

const initialState: FileStoreState = {
  files: [],
  isLoading: false,
  error: null,
};

export const useFileStore = create<FileStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      getFiles: async () => {
        set({ isLoading: true, error: null });
        try {
          // Get both recordings and snapshots via filesService (JSON-RPC per API doc)
          const [recordings, snapshots] = await Promise.all([
            errorRecoveryService.executeWithRetry(
              async () => {
                return filesService.listRecordings({ limit: 50, offset: 0 });
              },
              'getFiles_recordings'
            ),
            errorRecoveryService.executeWithRetry(
              async () => {
                return filesService.listSnapshots({ limit: 50, offset: 0 });
              },
              'getFiles_snapshots'
            )
          ]);
          
          const allFiles = [
            ...recordings.files.map((f: any) => ({ ...f, type: 'recording' })),
            ...snapshots.files.map((f: any) => ({ ...f, type: 'snapshot' }))
          ];
          
          set({ files: allFiles, isLoading: false });
          logger.info('Files retrieved', undefined, 'fileStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get files';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      downloadFile: async (fileId: string) => {
        try {
          // File download would be handled by direct HTTP request to download URL
          // For now, just log the action
          logger.info(`Downloading file ${fileId}`, undefined, 'fileStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to download file';
          set({ error: errorMessage });
        }
      },
      
      deleteFile: async (fileId: string) => {
        try {
          // File deletion would use RPC methods DELETE_RECORDING or DELETE_SNAPSHOT
          // For now, just log the action
          logger.info(`Deleting file ${fileId}`, undefined, 'fileStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to delete file';
          set({ error: errorMessage });
        }
      },
      
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    { name: 'file-store' }
  )
);
