/**
 * Settings Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around ConfigurationManagerService
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';
import { configurationManagerService } from '../services/configurationManagerService';

interface SettingsStoreState {
  settings: any;
  isLoading: boolean;
  error: string | null;
}

interface SettingsStoreActions {
  getSettings: () => Promise<void>;
  updateSettings: (settings: any) => Promise<void>;
  clearError: () => void;
  setError: (error: string) => void;
}

type SettingsStore = SettingsStoreState & SettingsStoreActions;

const initialState: SettingsStoreState = {
  settings: {},
  isLoading: false,
  error: null,
};

export const useSettingsStore = create<SettingsStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      getSettings: async () => {
        set({ isLoading: true, error: null });
        try {
          const settings = {
            recordingRotationMinutes: configurationManagerService.getRecordingRotationMinutes(),
            storageWarnPercent: configurationManagerService.getStorageWarnPercent(),
            storageBlockPercent: configurationManagerService.getStorageBlockPercent(),
            webSocketUrl: configurationManagerService.getWebSocketUrl(),
            healthUrl: configurationManagerService.getHealthUrl(),
            apiTimeout: configurationManagerService.getApiTimeout(),
            logLevel: configurationManagerService.getLogLevel()
          };
          set({ settings, isLoading: false });
          logger.info('Settings retrieved', undefined, 'settingsStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get settings';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      updateSettings: async (settings: any) => {
        set({ isLoading: true, error: null });
        try {
          // Update configuration through service
          if (settings.recordingRotationMinutes !== undefined) {
            configurationManagerService.setRecordingRotationMinutes(settings.recordingRotationMinutes);
          }
          if (settings.storageWarnPercent !== undefined) {
            configurationManagerService.setStorageWarnPercent(settings.storageWarnPercent);
          }
          if (settings.storageBlockPercent !== undefined) {
            configurationManagerService.setStorageBlockPercent(settings.storageBlockPercent);
          }
          set({ settings, isLoading: false });
          logger.info('Settings updated', undefined, 'settingsStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to update settings';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    { name: 'settings-store' }
  )
);
