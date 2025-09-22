/**
 * Configuration Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around ConfigurationManagerService
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';

interface ConfigurationStoreState {
  configuration: any;
  isLoading: boolean;
  error: string | null;
}

interface ConfigurationStoreActions {
  getConfiguration: () => Promise<void>;
  updateConfiguration: (config: any) => Promise<void>;
  clearError: () => void;
  setError: (error: string) => void;
}

type ConfigurationStore = ConfigurationStoreState & ConfigurationStoreActions;

const initialState: ConfigurationStoreState = {
  configuration: {},
  isLoading: false,
  error: null,
};

export const useConfigurationStore = create<ConfigurationStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      getConfiguration: async () => {
        set({ isLoading: true, error: null });
        try {
          // TODO: Implement with ConfigurationManagerService
          set({ configuration: {}, isLoading: false });
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get configuration';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      updateConfiguration: async (config: any) => {
        set({ isLoading: true, error: null });
        try {
          // TODO: Implement with ConfigurationManagerService
          set({ configuration: config, isLoading: false });
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to update configuration';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    { name: 'configuration-store' }
  )
);
