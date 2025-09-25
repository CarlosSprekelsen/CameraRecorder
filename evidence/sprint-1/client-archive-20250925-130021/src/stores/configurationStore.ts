/**
 * Configuration Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around ConfigurationManagerService
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';
import { configurationManagerService } from '../services/configurationManagerService';
import { systemService } from '../services/systemService';

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
          // Fetch server info per API documentation (admin role required)
          const serverInfo = await systemService.getServerInfo();

          const configuration = {
            recordingRotationMinutes: configurationManagerService.getRecordingRotationMinutes(),
            storageWarnPercent: configurationManagerService.getStorageWarnPercent(),
            storageBlockPercent: configurationManagerService.getStorageBlockPercent(),
            webSocketUrl: configurationManagerService.getWebSocketUrl(),
            healthUrl: configurationManagerService.getHealthUrl(),
            apiTimeout: configurationManagerService.getApiTimeout(),
            logLevel: configurationManagerService.getLogLevel(),
            serverInfo,
          };
          set({ configuration, isLoading: false });
          logger.info('Configuration retrieved', undefined, 'configurationStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get configuration';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      updateConfiguration: async (config: any) => {
        set({ isLoading: true, error: null });
        try {
          // Update configuration through service
          if (config.recordingRotationMinutes !== undefined) {
            configurationManagerService.setRecordingRotationMinutes(config.recordingRotationMinutes);
          }
          if (config.storageWarnPercent !== undefined) {
            configurationManagerService.setStorageWarnPercent(config.storageWarnPercent);
          }
          if (config.storageBlockPercent !== undefined) {
            configurationManagerService.setStorageBlockPercent(config.storageBlockPercent);
          }
          set({ configuration: config, isLoading: false });
          logger.info('Configuration updated', undefined, 'configurationStore');
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
