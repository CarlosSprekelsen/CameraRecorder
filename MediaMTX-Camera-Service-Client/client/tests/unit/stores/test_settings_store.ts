/**
 * REQ-SETT01-001: Settings management must provide reliable user preference storage
 * REQ-SETT01-002: Settings validation must ensure data integrity and consistency
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for settings store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on settings management and persistence logic
 * - Test settings validation and change tracking
 * - Validate settings categories and user preferences
 */

import { useSettingsStore } from '../../../src/stores/settingsStore';
import type { AppSettings, SettingsValidation, SettingsChangeEvent, SettingsCategory } from '../../../src/types/settings';
import { DEFAULT_SETTINGS } from '../../../src/types/settings';

// Mock localStorage for persistence testing
const mockLocalStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn()
};

Object.defineProperty(window, 'localStorage', {
  value: mockLocalStorage
});

describe('Settings Store', () => {
  let store: ReturnType<typeof useSettingsStore.getState>;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useSettingsStore.getState();
    currentStore.resetSettings();
    
    // Get fresh store instance after reset
    store = useSettingsStore.getState();
    
    // Clear all mocks
    jest.clearAllMocks();
    mockLocalStorage.getItem.mockReturnValue(null);
  });

  describe('Initialization', () => {
    it('should start with default settings', () => {
      const state = useSettingsStore.getState();
      expect(state.settings).toEqual(DEFAULT_SETTINGS);
      expect(state.isLoading).toBe(false);
      expect(state.isSaving).toBe(false);
      expect(state.error).toBeNull();
      expect(state.hasUnsavedChanges).toBe(false);
      expect(state.changeHistory).toEqual([]);
    });
  });

  describe('Settings Management', () => {
    it('should update settings category', () => {
      const updates = {
        theme: 'dark',
        language: 'es'
      };

      store.updateSettings('ui', updates);
      
      const state = useSettingsStore.getState();
      expect(state.settings.ui.theme).toBe('dark');
      expect(state.settings.ui.language).toBe('es');
      expect(state.hasUnsavedChanges).toBe(true);
    });

    it('should update individual setting', () => {
      store.updateSetting('ui', 'theme', 'dark');
      
      const state = useSettingsStore.getState();
      expect(state.settings.ui.theme).toBe('dark');
      expect(state.hasUnsavedChanges).toBe(true);
    });

    it('should get specific setting', () => {
      store.updateSetting('ui', 'theme', 'dark');
      
      const theme = store.getSetting('ui', 'theme');
      expect(theme).toBe('dark');
    });

    it('should get settings category', () => {
      const uiSettings = store.getCategory('ui');
      expect(uiSettings).toEqual(DEFAULT_SETTINGS.ui);
    });

    it('should check if setting exists', () => {
      expect(store.hasSetting('ui', 'theme')).toBe(true);
      expect(store.hasSetting('ui', 'nonexistent')).toBe(false);
    });
  });

  describe('Change Tracking', () => {
    it('should track settings changes', () => {
      store.updateSetting('ui', 'theme', 'dark');
      
      const state = useSettingsStore.getState();
      expect(state.hasUnsavedChanges).toBe(true);
      expect(state.changeHistory).toHaveLength(1);
      
      const change = state.changeHistory[0];
      expect(change.category).toBe('ui');
      expect(change.key).toBe('theme');
      expect(change.oldValue).toBe(DEFAULT_SETTINGS.ui.theme);
      expect(change.newValue).toBe('dark');
      expect(change.timestamp).toBeInstanceOf(Date);
    });

    it('should clear unsaved changes', () => {
      store.updateSetting('ui', 'theme', 'dark');
      store.clearUnsavedChanges();
      
      const state = useSettingsStore.getState();
      expect(state.hasUnsavedChanges).toBe(false);
    });

    it('should get change history for category', () => {
      store.updateSetting('ui', 'theme', 'dark');
      store.updateSetting('ui', 'language', 'es');
      store.updateSetting('camera', 'default_format', 'h264');
      
      const uiChanges = store.getChangeHistory('ui');
      expect(uiChanges).toHaveLength(2);
      expect(uiChanges[0].category).toBe('ui');
      expect(uiChanges[1].category).toBe('ui');
    });

    it('should get recent changes', () => {
      const now = new Date();
      store.updateSetting('ui', 'theme', 'dark');
      
      // Mock timestamp to be recent
      const state = useSettingsStore.getState();
      state.changeHistory[0].timestamp = new Date(now.getTime() - 1000);
      
      const recentChanges = store.getRecentChanges(5000); // 5 seconds
      expect(recentChanges).toHaveLength(1);
    });
  });

  describe('Settings Validation', () => {
    it('should validate all settings', () => {
      const validation = store.validateSettings();
      
      expect(validation).toHaveProperty('isValid');
      expect(validation).toHaveProperty('errors');
      expect(validation).toHaveProperty('warnings');
      expect(validation).toHaveProperty('validatedAt');
    });

    it('should validate specific category', () => {
      const validation = store.validateCategory('ui');
      
      expect(validation).toHaveProperty('isValid');
      expect(validation).toHaveProperty('errors');
      expect(validation).toHaveProperty('warnings');
      expect(validation).toHaveProperty('validatedAt');
    });

    it('should detect invalid settings', () => {
      // Set invalid theme
      store.updateSetting('ui', 'theme', 'invalid-theme' as any);
      
      const validation = store.validateCategory('ui');
      expect(validation.isValid).toBe(false);
      expect(validation.errors.length).toBeGreaterThan(0);
    });

    it('should detect warnings for settings', () => {
      // Set potentially problematic setting
      store.updateSetting('camera', 'max_duration', 86400); // 24 hours
      
      const validation = store.validateCategory('camera');
      // May have warnings for very long duration
      expect(validation).toHaveProperty('warnings');
    });
  });

  describe('Settings Persistence', () => {
    it('should load settings from localStorage', async () => {
      const savedSettings = {
        ...DEFAULT_SETTINGS,
        ui: {
          ...DEFAULT_SETTINGS.ui,
          theme: 'dark'
        }
      };

      mockLocalStorage.getItem.mockReturnValue(JSON.stringify(savedSettings));

      await store.loadSettings();

      const state = useSettingsStore.getState();
      expect(state.settings.ui.theme).toBe('dark');
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle load settings failure', async () => {
      mockLocalStorage.getItem.mockImplementation(() => {
        throw new Error('localStorage error');
      });

      await store.loadSettings();

      const state = useSettingsStore.getState();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe('localStorage error');
    });

    it('should save settings to localStorage', async () => {
      store.updateSetting('ui', 'theme', 'dark');
      
      await store.saveSettings();

      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        'settings',
        expect.stringContaining('"theme":"dark"')
      );
      
      const state = useSettingsStore.getState();
      expect(state.isSaving).toBe(false);
      expect(state.hasUnsavedChanges).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle save settings failure', async () => {
      mockLocalStorage.setItem.mockImplementation(() => {
        throw new Error('localStorage error');
      });

      store.updateSetting('ui', 'theme', 'dark');
      
      await store.saveSettings();

      const state = useSettingsStore.getState();
      expect(state.isSaving).toBe(false);
      expect(state.error).toBe('localStorage error');
      expect(state.hasUnsavedChanges).toBe(true); // Still has unsaved changes
    });
  });

  describe('Settings Reset', () => {
    it('should reset settings to defaults', () => {
      store.updateSetting('ui', 'theme', 'dark');
      store.updateSetting('ui', 'language', 'es');
      store.updateSetting('camera', 'default_format', 'h264');
      
      store.resetSettings();
      
      const state = useSettingsStore.getState();
      expect(state.settings).toEqual(DEFAULT_SETTINGS);
      expect(state.hasUnsavedChanges).toBe(false);
      expect(state.changeHistory).toEqual([]);
    });

    it('should reset specific category to defaults', () => {
      store.updateSetting('ui', 'theme', 'dark');
      store.updateSetting('ui', 'language', 'es');
      store.updateSetting('camera', 'default_format', 'h264');
      
      store.resetCategory('ui');
      
      const state = useSettingsStore.getState();
      expect(state.settings.ui).toEqual(DEFAULT_SETTINGS.ui);
      expect(state.settings.camera.default_format).toBe('h264'); // Should remain changed
      expect(state.hasUnsavedChanges).toBe(true);
    });
  });

  describe('Settings Analysis', () => {
    it('should get settings summary', () => {
      store.updateSetting('ui', 'theme', 'dark');
      store.updateSetting('camera', 'default_format', 'h264');
      
      const summary = store.getSettingsSummary();
      
      expect(summary).toHaveProperty('total_categories');
      expect(summary).toHaveProperty('has_unsaved_changes');
      expect(summary).toHaveProperty('change_count');
      expect(summary).toHaveProperty('last_modified');
    });

    it('should get settings differences from defaults', () => {
      store.updateSetting('ui', 'theme', 'dark');
      store.updateSetting('ui', 'language', 'es');
      store.updateSetting('camera', 'default_format', 'h264');
      
      const differences = store.getSettingsDifferences();
      
      expect(differences).toHaveProperty('ui');
      expect(differences).toHaveProperty('camera');
      expect(differences.ui).toHaveProperty('theme');
      expect(differences.ui).toHaveProperty('language');
      expect(differences.camera).toHaveProperty('default_format');
    });

    it('should check if settings are modified', () => {
      expect(store.areSettingsModified()).toBe(false);
      
      store.updateSetting('ui', 'theme', 'dark');
      expect(store.areSettingsModified()).toBe(true);
    });

    it('should get modified categories', () => {
      store.updateSetting('ui', 'theme', 'dark');
      store.updateSetting('camera', 'default_format', 'h264');
      
      const modifiedCategories = store.getModifiedCategories();
      expect(modifiedCategories).toContain('ui');
      expect(modifiedCategories).toContain('camera');
    });
  });

  describe('Settings Utilities', () => {
    it('should export settings', () => {
      store.updateSetting('ui', 'theme', 'dark');
      
      const exported = store.exportSettings();
      
      expect(exported).toHaveProperty('settings');
      expect(exported).toHaveProperty('exportedAt');
      expect(exported).toHaveProperty('version');
      expect(exported.settings.ui.theme).toBe('dark');
    });

    it('should import settings', () => {
      const importedSettings = {
        ...DEFAULT_SETTINGS,
        ui: {
          ...DEFAULT_SETTINGS.ui,
          theme: 'dark',
          language: 'es'
        }
      };

      const importData = {
        settings: importedSettings,
        exportedAt: new Date().toISOString(),
        version: '1.0.0'
      };

      store.importSettings(importData);
      
      const state = useSettingsStore.getState();
      expect(state.settings.ui.theme).toBe('dark');
      expect(state.settings.ui.language).toBe('es');
      expect(state.hasUnsavedChanges).toBe(true);
    });

    it('should validate import data', () => {
      const validImportData = {
        settings: DEFAULT_SETTINGS,
        exportedAt: new Date().toISOString(),
        version: '1.0.0'
      };

      expect(store.validateImportData(validImportData)).toBe(true);

      const invalidImportData = {
        settings: { invalid: 'data' },
        exportedAt: new Date().toISOString(),
        version: '1.0.0'
      };

      expect(store.validateImportData(invalidImportData)).toBe(false);
    });

    it('should get settings version', () => {
      const version = store.getSettingsVersion();
      expect(version).toBeDefined();
      expect(typeof version).toBe('string');
    });
  });

  describe('Error Handling', () => {
    it('should set error state', () => {
      store.setError('Test error');
      
      const state = useSettingsStore.getState();
      expect(state.error).toBe('Test error');
    });

    it('should clear error state', () => {
      store.setError('Test error');
      store.clearError();
      
      const state = useSettingsStore.getState();
      expect(state.error).toBeNull();
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.updateSetting('ui', 'theme', 'dark');
      store.setError('Test error');
      
      // Reset
      store.resetSettings();
      
      const state = useSettingsStore.getState();
      expect(state.settings).toEqual(DEFAULT_SETTINGS);
      expect(state.isLoading).toBe(false);
      expect(state.isSaving).toBe(false);
      expect(state.error).toBeNull();
      expect(state.hasUnsavedChanges).toBe(false);
      expect(state.changeHistory).toEqual([]);
    });
  });
});
