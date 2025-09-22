/**
 * REQ-CONF01-001: Configuration management must provide reliable settings management
 * REQ-CONF01-002: Configuration validation must ensure settings integrity
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for configuration store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on configuration management and validation logic
 * - Test configuration loading, saving, and validation
 * - Validate environment variable handling
 */

import { useConfigurationStore } from '../../../src/stores/configurationStore';
import type { AppConfig, ValidationResult } from '../../../src/types/settings';

// Mock the configuration manager service
jest.mock('../../../src/services/configurationManagerService', () => ({
  configurationManagerService: {
    loadConfiguration: jest.fn(),
    saveConfiguration: jest.fn(),
    validateConfiguration: jest.fn(),
    resetToDefaults: jest.fn(),
    getEnvironmentVariables: jest.fn()
  }
}));

// Mock the logger service
jest.mock('../../../src/services/loggerService', () => ({
  logger: {
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn()
  },
  loggers: {
    config: {
      info: jest.fn(),
      warn: jest.fn(),
      error: jest.fn()
    }
  }
}));

describe('Configuration Store', () => {
  let store: ReturnType<typeof useConfigurationStore.getState>;
  let mockConfigService: any;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useConfigurationStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useConfigurationStore.getState();
    
    // Get mock service
    mockConfigService = require('../../../src/services/configurationManagerService').configurationManagerService;
    jest.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useConfigurationStore.getState();
      expect(state.currentConfig).toBeNull();
      expect(state.defaultConfig).toBeNull();
      expect(state.environmentVariables).toEqual({});
      expect(state.validationResult).toBeNull();
      expect(state.validationErrors).toEqual([]);
      expect(state.isLoading).toBe(false);
      expect(state.isReloading).toBe(false);
      expect(state.isValidating).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
      expect(state.isConfigured).toBe(false);
      expect(state.hasValidationErrors).toBe(false);
    });
  });

  describe('Configuration Management', () => {
    it('should set current configuration', () => {
      const config: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      store.setCurrentConfig(config);
      
      const state = useConfigurationStore.getState();
      expect(state.currentConfig).toEqual(config);
      expect(state.isConfigured).toBe(true);
    });

    it('should get current configuration', () => {
      const config: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      store.setCurrentConfig(config);
      
      expect(store.getCurrentConfig()).toEqual(config);
    });

    it('should set default configuration', () => {
      const defaultConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'medium'
        },
        ui: {
          theme: 'light',
          language: 'en',
          auto_refresh: false
        }
      };

      store.setDefaultConfig(defaultConfig);
      
      const state = useConfigurationStore.getState();
      expect(state.defaultConfig).toEqual(defaultConfig);
    });

    it('should get default configuration', () => {
      const defaultConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'medium'
        },
        ui: {
          theme: 'light',
          language: 'en',
          auto_refresh: false
        }
      };

      store.setDefaultConfig(defaultConfig);
      
      expect(store.getDefaultConfig()).toEqual(defaultConfig);
    });

    it('should check if configuration is loaded', () => {
      expect(store.isConfigurationLoaded()).toBe(false);

      const config: AppConfig = {
        server: { host: 'localhost', port: 8002, protocol: 'ws' },
        camera: { default_format: 'mp4', max_duration: 3600, quality: 'high' },
        ui: { theme: 'dark', language: 'en', auto_refresh: true }
      };

      store.setCurrentConfig(config);
      expect(store.isConfigurationLoaded()).toBe(true);
    });
  });

  describe('Environment Variables Management', () => {
    it('should set environment variables', () => {
      const envVars = {
        NODE_ENV: 'development',
        API_URL: 'ws://localhost:8002',
        LOG_LEVEL: 'debug'
      };

      store.setEnvironmentVariables(envVars);
      
      const state = useConfigurationStore.getState();
      expect(state.environmentVariables).toEqual(envVars);
    });

    it('should get environment variables', () => {
      const envVars = {
        NODE_ENV: 'development',
        API_URL: 'ws://localhost:8002',
        LOG_LEVEL: 'debug'
      };

      store.setEnvironmentVariables(envVars);
      
      expect(store.getEnvironmentVariables()).toEqual(envVars);
    });

    it('should get specific environment variable', () => {
      const envVars = {
        NODE_ENV: 'development',
        API_URL: 'ws://localhost:8002',
        LOG_LEVEL: 'debug'
      };

      store.setEnvironmentVariables(envVars);
      
      expect(store.getEnvironmentVariable('NODE_ENV')).toBe('development');
      expect(store.getEnvironmentVariable('API_URL')).toBe('ws://localhost:8002');
      expect(store.getEnvironmentVariable('UNKNOWN_VAR')).toBeUndefined();
    });
  });

  describe('Configuration Validation', () => {
    it('should set validation result', () => {
      const validationResult: ValidationResult = {
        isValid: true,
        errors: [],
        warnings: [],
        validatedAt: new Date()
      };

      store.setValidationResult(validationResult);
      
      const state = useConfigurationStore.getState();
      expect(state.validationResult).toEqual(validationResult);
      expect(state.hasValidationErrors).toBe(false);
    });

    it('should set validation errors', () => {
      const errors = ['Invalid server port', 'Missing camera format'];
      store.setValidationErrors(errors);
      
      const state = useConfigurationStore.getState();
      expect(state.validationErrors).toEqual(errors);
      expect(state.hasValidationErrors).toBe(true);
    });

    it('should get validation result', () => {
      const validationResult: ValidationResult = {
        isValid: true,
        errors: [],
        warnings: [],
        validatedAt: new Date()
      };

      store.setValidationResult(validationResult);
      
      expect(store.getValidationResult()).toEqual(validationResult);
    });

    it('should get validation errors', () => {
      const errors = ['Invalid server port', 'Missing camera format'];
      store.setValidationErrors(errors);
      
      expect(store.getValidationErrors()).toEqual(errors);
    });

    it('should check if configuration is valid', () => {
      expect(store.isConfigurationValid()).toBe(false);

      const validResult: ValidationResult = {
        isValid: true,
        errors: [],
        warnings: [],
        validatedAt: new Date()
      };
      store.setValidationResult(validResult);
      expect(store.isConfigurationValid()).toBe(true);

      const invalidResult: ValidationResult = {
        isValid: false,
        errors: ['Invalid config'],
        warnings: [],
        validatedAt: new Date()
      };
      store.setValidationResult(invalidResult);
      expect(store.isConfigurationValid()).toBe(false);
    });
  });

  describe('Loading State Management', () => {
    it('should set loading state', () => {
      store.setLoading(true);
      let state = useConfigurationStore.getState();
      expect(state.isLoading).toBe(true);

      store.setLoading(false);
      state = useConfigurationStore.getState();
      expect(state.isLoading).toBe(false);
    });

    it('should set reloading state', () => {
      store.setReloading(true);
      let state = useConfigurationStore.getState();
      expect(state.isReloading).toBe(true);

      store.setReloading(false);
      state = useConfigurationStore.getState();
      expect(state.isReloading).toBe(false);
    });

    it('should set validating state', () => {
      store.setValidating(true);
      let state = useConfigurationStore.getState();
      expect(state.isValidating).toBe(true);

      store.setValidating(false);
      state = useConfigurationStore.getState();
      expect(state.isValidating).toBe(false);
    });
  });

  describe('Error Management', () => {
    it('should set error', () => {
      store.setError('Configuration load failed');
      let state = useConfigurationStore.getState();
      expect(state.error).toBe('Configuration load failed');

      store.setError(null);
      state = useConfigurationStore.getState();
      expect(state.error).toBeNull();
    });

    it('should set last error', () => {
      store.setLastError('Last configuration error');
      let state = useConfigurationStore.getState();
      expect(state.lastError).toBe('Last configuration error');

      store.setLastError(null);
      state = useConfigurationStore.getState();
      expect(state.lastError).toBeNull();
    });

    it('should clear all errors', () => {
      store.setError('Current error');
      store.setLastError('Last error');
      
      store.clearAllErrors();
      
      const state = useConfigurationStore.getState();
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
    });
  });

  describe('Configuration Operations', () => {
    it('should load configuration successfully', async () => {
      const mockConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      mockConfigService.loadConfiguration.mockResolvedValue(mockConfig);

      await store.loadConfiguration();

      const state = useConfigurationStore.getState();
      expect(state.currentConfig).toEqual(mockConfig);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.isConfigured).toBe(true);
    });

    it('should handle load configuration failure', async () => {
      const error = new Error('Configuration file not found');
      mockConfigService.loadConfiguration.mockRejectedValue(error);

      await store.loadConfiguration();

      const state = useConfigurationStore.getState();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe('Configuration file not found');
      expect(state.isConfigured).toBe(false);
    });

    it('should save configuration successfully', async () => {
      const config: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      mockConfigService.saveConfiguration.mockResolvedValue(undefined);

      await store.saveConfiguration(config);

      const state = useConfigurationStore.getState();
      expect(state.currentConfig).toEqual(config);
      expect(state.error).toBeNull();
    });

    it('should validate configuration successfully', async () => {
      const config: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      const mockValidationResult: ValidationResult = {
        isValid: true,
        errors: [],
        warnings: [],
        validatedAt: new Date()
      };

      mockConfigService.validateConfiguration.mockResolvedValue(mockValidationResult);

      await store.validateConfiguration(config);

      const state = useConfigurationStore.getState();
      expect(state.validationResult).toEqual(mockValidationResult);
      expect(state.isValidating).toBe(false);
      expect(state.error).toBeNull();
      expect(state.hasValidationErrors).toBe(false);
    });

    it('should handle validation failure', async () => {
      const config: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      const mockValidationResult: ValidationResult = {
        isValid: false,
        errors: ['Invalid server port'],
        warnings: [],
        validatedAt: new Date()
      };

      mockConfigService.validateConfiguration.mockResolvedValue(mockValidationResult);

      await store.validateConfiguration(config);

      const state = useConfigurationStore.getState();
      expect(state.validationResult).toEqual(mockValidationResult);
      expect(state.isValidating).toBe(false);
      expect(state.hasValidationErrors).toBe(true);
    });

    it('should reset to defaults successfully', async () => {
      const mockDefaultConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'medium'
        },
        ui: {
          theme: 'light',
          language: 'en',
          auto_refresh: false
        }
      };

      mockConfigService.resetToDefaults.mockResolvedValue(mockDefaultConfig);

      await store.resetToDefaults();

      const state = useConfigurationStore.getState();
      expect(state.currentConfig).toEqual(mockDefaultConfig);
      expect(state.error).toBeNull();
    });

    it('should reload configuration successfully', async () => {
      const mockConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      mockConfigService.loadConfiguration.mockResolvedValue(mockConfig);

      await store.reloadConfiguration();

      const state = useConfigurationStore.getState();
      expect(state.currentConfig).toEqual(mockConfig);
      expect(state.isReloading).toBe(false);
      expect(state.error).toBeNull();
    });
  });

  describe('Configuration Analysis', () => {
    it('should get configuration summary', () => {
      const config: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      store.setCurrentConfig(config);

      const summary = store.getConfigurationSummary();
      expect(summary).toEqual({
        is_configured: true,
        is_valid: false, // No validation result set
        has_errors: false,
        server_host: 'localhost',
        server_port: 8002,
        camera_format: 'mp4',
        ui_theme: 'dark'
      });
    });

    it('should check if configuration has changed from defaults', () => {
      const defaultConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'medium'
        },
        ui: {
          theme: 'light',
          language: 'en',
          auto_refresh: false
        }
      };

      const currentConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high' // Changed from medium
        },
        ui: {
          theme: 'dark', // Changed from light
          language: 'en',
          auto_refresh: true // Changed from false
        }
      };

      store.setDefaultConfig(defaultConfig);
      store.setCurrentConfig(currentConfig);

      expect(store.hasConfigurationChanged()).toBe(true);
    });

    it('should get configuration differences', () => {
      const defaultConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'medium'
        },
        ui: {
          theme: 'light',
          language: 'en',
          auto_refresh: false
        }
      };

      const currentConfig: AppConfig = {
        server: {
          host: 'localhost',
          port: 8002,
          protocol: 'ws'
        },
        camera: {
          default_format: 'mp4',
          max_duration: 3600,
          quality: 'high'
        },
        ui: {
          theme: 'dark',
          language: 'en',
          auto_refresh: true
        }
      };

      store.setDefaultConfig(defaultConfig);
      store.setCurrentConfig(currentConfig);

      const differences = store.getConfigurationDifferences();
      expect(differences).toEqual({
        'camera.quality': { default: 'medium', current: 'high' },
        'ui.theme': { default: 'light', current: 'dark' },
        'ui.auto_refresh': { default: false, current: true }
      });
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      const config: AppConfig = {
        server: { host: 'localhost', port: 8002, protocol: 'ws' },
        camera: { default_format: 'mp4', max_duration: 3600, quality: 'high' },
        ui: { theme: 'dark', language: 'en', auto_refresh: true }
      };
      store.setCurrentConfig(config);
      store.setValidationErrors(['Test error']);
      store.setError('Test error');
      
      // Reset
      store.reset();
      
      const state = useConfigurationStore.getState();
      expect(state.currentConfig).toBeNull();
      expect(state.defaultConfig).toBeNull();
      expect(state.environmentVariables).toEqual({});
      expect(state.validationResult).toBeNull();
      expect(state.validationErrors).toEqual([]);
      expect(state.isLoading).toBe(false);
      expect(state.isReloading).toBe(false);
      expect(state.isValidating).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
      expect(state.isConfigured).toBe(false);
      expect(state.hasValidationErrors).toBe(false);
    });
  });
});
