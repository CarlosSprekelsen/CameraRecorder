/**
 * File Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around FileService
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';

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
          // TODO: Implement with FileService
          set({ files: [], isLoading: false });
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get files';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      downloadFile: async (fileId: string) => {
        try {
          // TODO: Implement with FileService
          logger.info(`Downloading file ${fileId}`, undefined, 'fileStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to download file';
          set({ error: errorMessage });
        }
      },
      
      deleteFile: async (fileId: string) => {
        try {
          // TODO: Implement with FileService
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
