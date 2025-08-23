/**
 * Settings Store
 * Manages application settings with persistence and validation
 */

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { 
  AppSettings, 
  SettingsValidation, 
  SettingsChangeEvent,
  SettingsCategory 
} from '../types/settings';
import { DEFAULT_SETTINGS } from '../types/settings';

interface SettingsState {
  // Settings data
  settings: AppSettings;
  
  // UI state
  isLoading: boolean;
  isSaving: boolean;
  error: string | null;
  
  // Change tracking
  hasUnsavedChanges: boolean;
  changeHistory: SettingsChangeEvent[];
  
  // Actions
  loadSettings: () => Promise<void>;
  saveSettings: () => Promise<void>;
  resetSettings: () => void;
  updateSettings: <K extends keyof AppSettings>(
    category: K,
    updates: Partial<AppSettings[K]>
  ) => void;
  updateSetting: <K extends keyof AppSettings, P extends keyof AppSettings[K]>(
    category: K,
    key: P,
    value: AppSettings[K][P]
  ) => void;
  
  // Validation
  validateSettings: () => SettingsValidation;
  validateCategory: (category: SettingsCategory) => SettingsValidation;
  
  // Utilities
  getSetting: <K extends keyof AppSettings, P extends keyof AppSettings[K]>(
    category: K,
    key: P
  ) => AppSettings[K][P];
  exportSettings: () => string;
  importSettings: (jsonString: string) => Promise<boolean>;
  
  // UI actions
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
  clearChangeHistory: () => void;
}

/**
 * Settings validation functions
 */
const validateConnectionSettings = (settings: AppSettings['connection']): string[] => {
  const errors: string[] = [];
  
  if (!settings.websocketUrl) {
    errors.push('WebSocket URL is required');
  } else if (!settings.websocketUrl.startsWith('ws://') && !settings.websocketUrl.startsWith('wss://')) {
    errors.push('WebSocket URL must start with ws:// or wss://');
  }
  
  if (!settings.healthUrl) {
    errors.push('Health URL is required');
  } else if (!settings.healthUrl.startsWith('http://') && !settings.healthUrl.startsWith('https://')) {
    errors.push('Health URL must start with http:// or https://');
  }
  
  if (settings.connectionTimeout < 1000) {
    errors.push('Connection timeout must be at least 1000ms');
  }
  
  if (settings.reconnectInterval < 1000) {
    errors.push('Reconnect interval must be at least 1000ms');
  }
  
  if (settings.maxReconnectAttempts < 1) {
    errors.push('Max reconnect attempts must be at least 1');
  }
  
  return errors;
};

const validateRecordingSettings = (settings: AppSettings['recording']): string[] => {
  const errors: string[] = [];
  
  if (!['mp4', 'mkv'].includes(settings.defaultFormat)) {
    errors.push('Default format must be mp4 or mkv');
  }
  
  if (settings.defaultQuality < 1 || settings.defaultQuality > 100) {
    errors.push('Default quality must be between 1 and 100');
  }
  
  if (settings.defaultDuration !== null && settings.defaultDuration < 1) {
    errors.push('Default duration must be at least 1 second or null for unlimited');
  }
  
  if (settings.maxDuration < 1) {
    errors.push('Max duration must be at least 1 second');
  }
  
  if (settings.defaultFrameRate < 1 || settings.defaultFrameRate > 60) {
    errors.push('Default frame rate must be between 1 and 60 fps');
  }
  
  if (settings.defaultBitrate < 100 || settings.defaultBitrate > 10000) {
    errors.push('Default bitrate must be between 100 and 10000 kbps');
  }
  
  if (settings.maxFileSize < 1) {
    errors.push('Max file size must be at least 1MB');
  }
  
  if (settings.maxStorageSize < 1) {
    errors.push('Max storage size must be at least 1GB');
  }
  
  if (settings.autoCleanupAge < 1) {
    errors.push('Auto cleanup age must be at least 1 day');
  }
  
  if (!settings.storageDirectory) {
    errors.push('Storage directory is required');
  }
  
  if (!settings.storagePath) {
    errors.push('Storage path is required');
  }
  
  return errors;
};

const validateSnapshotSettings = (settings: AppSettings['snapshot']): string[] => {
  const errors: string[] = [];
  
  if (!['jpg', 'png', 'jpeg', 'bmp', 'webp'].includes(settings.defaultFormat)) {
    errors.push('Default format must be jpg, png, jpeg, bmp, or webp');
  }
  
  if (settings.jpegQuality < 1 || settings.jpegQuality > 100) {
    errors.push('JPEG quality must be between 1 and 100');
  }
  
  if (settings.defaultQuality < 1 || settings.defaultQuality > 100) {
    errors.push('Default quality must be between 1 and 100');
  }
  
  if (settings.defaultWidth < 320 || settings.defaultWidth > 7680) {
    errors.push('Default width must be between 320 and 7680 pixels');
  }
  
  if (settings.defaultHeight < 240 || settings.defaultHeight > 4320) {
    errors.push('Default height must be between 240 and 4320 pixels');
  }
  
  if (!settings.storagePath) {
    errors.push('Storage path is required');
  }
  
  return errors;
};

const validateUISettings = (settings: AppSettings['ui']): string[] => {
  const errors: string[] = [];
  
  if (!['light', 'dark', 'auto'].includes(settings.theme)) {
    errors.push('Theme must be light, dark, or auto');
  }
  
  if (settings.refreshInterval < 1000) {
    errors.push('Refresh interval must be at least 1000ms');
  }
  
  if (settings.notificationDuration < 1000) {
    errors.push('Notification duration must be at least 1000ms');
  }
  
  return errors;
};

const validateSecuritySettings = (settings: AppSettings['security']): string[] => {
  const errors: string[] = [];
  
  if (settings.sessionTimeout < 60000) {
    errors.push('Session timeout must be at least 60000ms (1 minute)');
  }
  
  return errors;
};

const validatePerformanceSettings = (settings: AppSettings['performance']): string[] => {
  const errors: string[] = [];
  
  if (settings.cacheSize < 1) {
    errors.push('Cache size must be at least 1MB');
  }
  
  if (settings.maxConcurrentDownloads < 1) {
    errors.push('Max concurrent downloads must be at least 1');
  }
  
  return errors;
};

/**
 * Settings store implementation
 */
export const useSettingsStore = create<SettingsState>()(
  persist(
    (set, get) => ({
      // Initial state
      settings: DEFAULT_SETTINGS,
      isLoading: false,
      isSaving: false,
      error: null,
      hasUnsavedChanges: false,
      changeHistory: [],

      // Load settings from storage
      loadSettings: async () => {
        set({ isLoading: true, error: null });
        try {
          // Settings are automatically loaded by persist middleware
          set({ isLoading: false });
        } catch (error) {
          set({ 
            isLoading: false, 
            error: error instanceof Error ? error.message : 'Failed to load settings' 
          });
        }
      },

      // Save settings to storage
      saveSettings: async () => {
        set({ isSaving: true, error: null });
        try {
          // Validate settings before saving
          const validation = get().validateSettings();
          if (!validation.isValid) {
            throw new Error(`Settings validation failed: ${validation.errors.join(', ')}`);
          }
          
          // Settings are automatically saved by persist middleware
          set({ 
            isSaving: false, 
            hasUnsavedChanges: false,
            changeHistory: []
          });
        } catch (error) {
          set({ 
            isSaving: false, 
            error: error instanceof Error ? error.message : 'Failed to save settings' 
          });
        }
      },

      // Reset settings to defaults
      resetSettings: () => {
        const oldSettings = get().settings;
        set({ 
          settings: DEFAULT_SETTINGS,
          hasUnsavedChanges: true,
          changeHistory: [
            {
              category: 'connection',
              key: 'all',
              oldValue: oldSettings,
              newValue: DEFAULT_SETTINGS,
              timestamp: new Date(),
            }
          ]
        });
      },

      // Update multiple settings in a category
      updateSettings: <K extends keyof AppSettings>(
        category: K,
        updates: Partial<AppSettings[K]>
      ) => {
        const { settings, changeHistory } = get();
        const oldCategorySettings = settings[category];
        const newCategorySettings = { ...oldCategorySettings, ...updates };
        
        const newChangeHistory: SettingsChangeEvent[] = Object.keys(updates).map(key => ({
          category,
          key,
          oldValue: oldCategorySettings[key as keyof AppSettings[K]],
          newValue: updates[key as keyof AppSettings[K]],
          timestamp: new Date(),
        }));
        
        set({
          settings: {
            ...settings,
            [category]: newCategorySettings,
            lastUpdated: new Date(),
          },
          hasUnsavedChanges: true,
          changeHistory: [...changeHistory, ...newChangeHistory],
        });
      },

      // Update a single setting
      updateSetting: <K extends keyof AppSettings, P extends keyof AppSettings[K]>(
        category: K,
        key: P,
        value: AppSettings[K][P]
      ) => {
        const { settings, changeHistory } = get();
        const oldValue = settings[category][key];
        
        if (oldValue !== value) {
          set({
            settings: {
              ...settings,
              [category]: {
                ...settings[category],
                [key]: value,
              },
              lastUpdated: new Date(),
            },
            hasUnsavedChanges: true,
            changeHistory: [
              ...changeHistory,
              {
                category,
                key: String(key),
                oldValue,
                newValue: value,
                timestamp: new Date(),
              },
            ],
          });
        }
      },

      // Validate all settings
      validateSettings: (): SettingsValidation => {
        const { settings } = get();
        const errors: string[] = [];
        const warnings: string[] = [];

        // Validate each category
        errors.push(...validateConnectionSettings(settings.connection));
        errors.push(...validateRecordingSettings(settings.recording));
        errors.push(...validateSnapshotSettings(settings.snapshot));
        errors.push(...validateUISettings(settings.ui));
        errors.push(...validateSecuritySettings(settings.security));
        errors.push(...validatePerformanceSettings(settings.performance));

        // Add warnings
        if (settings.connection.maxReconnectAttempts > 10) {
          warnings.push('High reconnect attempts may cause performance issues');
        }
        
        if (settings.performance.cacheSize > 1000) {
          warnings.push('Large cache size may consume significant memory');
        }

        return {
          isValid: errors.length === 0,
          errors,
          warnings,
        };
      },

      // Validate a specific category
      validateCategory: (category: SettingsCategory): SettingsValidation => {
        const { settings } = get();
        const errors: string[] = [];
        const warnings: string[] = [];

        switch (category) {
          case 'connection':
            errors.push(...validateConnectionSettings(settings.connection));
            break;
          case 'recording':
            errors.push(...validateRecordingSettings(settings.recording));
            break;
          case 'snapshot':
            errors.push(...validateSnapshotSettings(settings.snapshot));
            break;
          case 'ui':
            errors.push(...validateUISettings(settings.ui));
            break;
          case 'security':
            errors.push(...validateSecuritySettings(settings.security));
            break;
          case 'performance':
            errors.push(...validatePerformanceSettings(settings.performance));
            break;
        }

        return {
          isValid: errors.length === 0,
          errors,
          warnings,
        };
      },

      // Get a specific setting value
      getSetting: <K extends keyof AppSettings, P extends keyof AppSettings[K]>(
        category: K,
        key: P
      ): AppSettings[K][P] => {
        const { settings } = get();
        return settings[category][key];
      },

      // Export settings as JSON
      exportSettings: (): string => {
        const { settings } = get();
        return JSON.stringify(settings, null, 2);
      },

      // Import settings from JSON
      importSettings: async (jsonString: string): Promise<boolean> => {
        try {
          const importedSettings = JSON.parse(jsonString) as AppSettings;
          
          // Validate imported settings
          const tempStore = { settings: importedSettings };
          const validation = validateConnectionSettings(importedSettings.connection);
          if (validation.length > 0) {
            throw new Error(`Invalid settings: ${validation.join(', ')}`);
          }
          
          set({
            settings: {
              ...importedSettings,
              lastUpdated: new Date(),
            },
            hasUnsavedChanges: true,
          });
          
          return true;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to import settings' 
          });
          return false;
        }
      },

      // UI actions
      setLoading: (loading: boolean) => set({ isLoading: loading }),
      setError: (error: string | null) => set({ error }),
      clearError: () => set({ error: null }),
      clearChangeHistory: () => set({ changeHistory: [] }),
    }),
    {
      name: 'camera-app-settings',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ settings: state.settings }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.loadSettings();
        }
      },
    }
  )
);
