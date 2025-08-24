import { create } from 'zustand';
import { configurationManagerService } from '../services/configurationManagerService';
import type { AppConfig, ValidationResult } from '../types/settings';

/**
 * Configuration Store State Interface
 */
interface ConfigurationStoreState {
  // Current configuration
  currentConfig: AppConfig | null;
  
  // Default configuration
  defaultConfig: AppConfig | null;
  
  // Environment variables
  environmentVariables: Record<string, string>;
  
  // Validation state
  validationResult: ValidationResult | null;
  validationErrors: string[];
  
  // Loading states
  isLoading: boolean;
  isReloading: boolean;
  isValidating: boolean;
  
  // Error states
  error: string | null;
  lastError: string | null;
  
  // Configuration state
  isConfigured: boolean;
  hasValidationErrors: boolean;
}

/**
 * Configuration Store Actions Interface
 */
interface ConfigurationStoreActions {
  // State management
  setCurrentConfig: (config: AppConfig) => void;
  setDefaultConfig: (config: AppConfig) => void;
  setEnvironmentVariables: (vars: Record<string, string>) => void;
  setValidationResult: (result: ValidationResult) => void;
  setValidationErrors: (errors: string[]) => void;
  
  // Loading states
  setLoading: (loading: boolean) => void;
  setReloading: (reloading: boolean) => void;
  setValidating: (validating: boolean) => void;
  
  // Error states
  setError: (error: string | null) => void;
  setLastError: (error: string | null) => void;
  clearErrors: () => void;
  
  // Configuration states
  setConfigured: (configured: boolean) => void;
  setHasValidationErrors: (hasErrors: boolean) => void;
  
  // Configuration operations
  loadConfiguration: () => Promise<void>;
  reloadConfiguration: () => Promise<void>;
  updateConfiguration: (config: Partial<AppConfig>) => void;
  validateConfiguration: () => Promise<void>;
  resetToDefaults: () => void;
  
  // State queries
  getCurrentConfig: () => AppConfig | null;
  getDefaultConfig: () => AppConfig | null;
  getEnvironmentVariables: () => Record<string, string>;
  getValidationResult: () => ValidationResult | null;
  getValidationErrors: () => string[];
  isConfigurationValid: () => boolean;
  isConfigurationLoaded: () => boolean;
  
  // Configuration getters
  getRecordingRotationMinutes: () => number;
  getStorageWarnPercent: () => number;
  getStorageBlockPercent: () => number;
  getWebSocketUrl: () => string;
  getHealthUrl: () => string;
  getApiTimeout: () => number;
  getLogLevel: () => string;
  
  // Service integration
  initialize: () => void;
  cleanup: () => void;
}

/**
 * Configuration Store Type
 */
type ConfigurationStore = ConfigurationStoreState & ConfigurationStoreActions;

/**
 * Configuration Store Implementation
 */
export const useConfigurationStore = create<ConfigurationStore>((set, get) => ({
  // Initial state
  currentConfig: null,
  defaultConfig: null,
  environmentVariables: {},
  validationResult: null,
  validationErrors: [],
  isLoading: false,
  isReloading: false,
  isValidating: false,
  error: null,
  lastError: null,
  isConfigured: false,
  hasValidationErrors: false,

  // State management actions
  setCurrentConfig: (config: AppConfig) => {
    set({ currentConfig: config });
  },

  setDefaultConfig: (config: AppConfig) => {
    set({ defaultConfig: config });
  },

  setEnvironmentVariables: (vars: Record<string, string>) => {
    set({ environmentVariables: vars });
  },

  setValidationResult: (result: ValidationResult) => {
    set({ 
      validationResult: result,
      hasValidationErrors: !result.valid
    });
  },

  setValidationErrors: (errors: string[]) => {
    set({ validationErrors: errors });
  },

  // Loading state actions
  setLoading: (loading: boolean) => {
    set({ isLoading: loading });
  },

  setReloading: (reloading: boolean) => {
    set({ isReloading: reloading });
  },

  setValidating: (validating: boolean) => {
    set({ isValidating: validating });
  },

  // Error state actions
  setError: (error: string | null) => {
    set({ error });
  },

  setLastError: (error: string | null) => {
    set({ lastError: error });
  },

  clearErrors: () => {
    set({ error: null, lastError: null });
  },

  // Configuration state actions
  setConfigured: (configured: boolean) => {
    set({ isConfigured: configured });
  },

  setHasValidationErrors: (hasErrors: boolean) => {
    set({ hasValidationErrors: hasErrors });
  },

  // Configuration operations
  loadConfiguration: async () => {
    const { setLoading, setError, setLastError } = get();
    
    try {
      setLoading(true);
      setError(null);
      
      const currentConfig = configurationManagerService.getCurrentConfiguration();
      const defaultConfig = configurationManagerService.getDefaultConfiguration();
      const environmentVariables = configurationManagerService.getEnvironmentVariables();
      const validationResult = configurationManagerService.validateConfiguration();
      const validationErrors = configurationManagerService.getConfigurationErrors();
      
      const { 
        setCurrentConfig, 
        setDefaultConfig, 
        setEnvironmentVariables, 
        setValidationResult, 
        setValidationErrors,
        setConfigured 
      } = get();
      
      setCurrentConfig(currentConfig);
      setDefaultConfig(defaultConfig);
      setEnvironmentVariables(environmentVariables);
      setValidationResult(validationResult);
      setValidationErrors(validationErrors);
      setConfigured(true);
      
    } catch (error: any) {
      const errorMessage = error.message || 'Failed to load configuration';
      setError(errorMessage);
      setLastError(errorMessage);
      throw error;
    } finally {
      setLoading(false);
    }
  },

  reloadConfiguration: async () => {
    const { setReloading, setError, setLastError } = get();
    
    try {
      setReloading(true);
      setError(null);
      
      await configurationManagerService.reloadConfiguration();
      await get().loadConfiguration();
      
    } catch (error: any) {
      const errorMessage = error.message || 'Failed to reload configuration';
      setError(errorMessage);
      setLastError(errorMessage);
      throw error;
    } finally {
      setReloading(false);
    }
  },

  updateConfiguration: (config: Partial<AppConfig>) => {
    const { setCurrentConfig, setValidationResult, setValidationErrors } = get();
    
    configurationManagerService.updateConfiguration(config);
    
    const currentConfig = configurationManagerService.getCurrentConfiguration();
    const validationResult = configurationManagerService.validateConfiguration();
    const validationErrors = configurationManagerService.getConfigurationErrors();
    
    setCurrentConfig(currentConfig);
    setValidationResult(validationResult);
    setValidationErrors(validationErrors);
  },

  validateConfiguration: async () => {
    const { setValidating, setError, setLastError } = get();
    
    try {
      setValidating(true);
      setError(null);
      
      const validationResult = configurationManagerService.validateConfiguration();
      const validationErrors = configurationManagerService.getConfigurationErrors();
      
      const { setValidationResult, setValidationErrors } = get();
      setValidationResult(validationResult);
      setValidationErrors(validationErrors);
      
    } catch (error: any) {
      const errorMessage = error.message || 'Failed to validate configuration';
      setError(errorMessage);
      setLastError(errorMessage);
      throw error;
    } finally {
      setValidating(false);
    }
  },

  resetToDefaults: () => {
    const { setCurrentConfig, setValidationResult, setValidationErrors } = get();
    
    const defaultConfig = configurationManagerService.getDefaultConfiguration();
    const validationResult = configurationManagerService.validateConfiguration();
    const validationErrors = configurationManagerService.getConfigurationErrors();
    
    setCurrentConfig(defaultConfig);
    setValidationResult(validationResult);
    setValidationErrors(validationErrors);
  },

  // State queries
  getCurrentConfig: () => {
    return get().currentConfig;
  },

  getDefaultConfig: () => {
    return get().defaultConfig;
  },

  getEnvironmentVariables: () => {
    return get().environmentVariables;
  },

  getValidationResult: () => {
    return get().validationResult;
  },

  getValidationErrors: () => {
    return get().validationErrors;
  },

  isConfigurationValid: () => {
    const result = get().validationResult;
    return result ? result.valid : false;
  },

  isConfigurationLoaded: () => {
    return get().isConfigured;
  },

  // Configuration getters
  getRecordingRotationMinutes: () => {
    return configurationManagerService.getRecordingRotationMinutes();
  },

  getStorageWarnPercent: () => {
    return configurationManagerService.getStorageWarnPercent();
  },

  getStorageBlockPercent: () => {
    return configurationManagerService.getStorageBlockPercent();
  },

  getWebSocketUrl: () => {
    return configurationManagerService.getWebSocketUrl();
  },

  getHealthUrl: () => {
    return configurationManagerService.getHealthUrl();
  },

  getApiTimeout: () => {
    return configurationManagerService.getApiTimeout();
  },

  getLogLevel: () => {
    return configurationManagerService.getLogLevel();
  },

  // Service integration
  initialize: () => {
    // Set up event handlers
    configurationManagerService.onConfigurationChange((config) => {
      get().setCurrentConfig(config);
    });

    configurationManagerService.onValidationError((errors) => {
      get().setValidationErrors(errors);
      get().setHasValidationErrors(errors.length > 0);
    });

    // Initial load
    get().loadConfiguration().catch(error => {
      console.error('Failed to initialize configuration store:', error);
    });
  },

  cleanup: () => {
    configurationManagerService.cleanup();
    set({
      currentConfig: null,
      defaultConfig: null,
      environmentVariables: {},
      validationResult: null,
      validationErrors: [],
      isLoading: false,
      isReloading: false,
      isValidating: false,
      error: null,
      lastError: null,
      isConfigured: false,
      hasValidationErrors: false
    });
  }
}));
