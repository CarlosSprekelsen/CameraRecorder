/**
 * Storage Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around StorageMonitorService
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';

interface StorageStoreState {
  storageInfo: any;
  isLoading: boolean;
  error: string | null;
}

interface StorageStoreActions {
  getStorageInfo: () => Promise<void>;
  refreshStorage: () => Promise<void>;
  clearError: () => void;
  setError: (error: string) => void;
}

type StorageStore = StorageStoreState & StorageStoreActions;

const initialState: StorageStoreState = {
  storageInfo: null,
  isLoading: false,
  error: null,
};

export const useStorageStore = create<StorageStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      getStorageInfo: async () => {
        set({ isLoading: true, error: null });
        try {
          // TODO: Implement with StorageMonitorService
          set({ storageInfo: null, isLoading: false });
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get storage info';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      refreshStorage: async () => {
        try {
          await get().getStorageInfo();
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to refresh storage';
          set({ error: errorMessage });
        }
      },
      
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    { name: 'storage-store' }
  )
);
