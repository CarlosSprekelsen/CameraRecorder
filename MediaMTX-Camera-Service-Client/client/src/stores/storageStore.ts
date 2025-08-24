import { create } from 'zustand';
import { storageMonitorService } from '../services/storageMonitorService';
import type {
  StorageInfo,
  StorageUsage,
  ThresholdStatus
} from '../types/camera';

/**
 * Storage Store State Interface
 */
interface StorageStoreState {
  // Current storage information
  storageInfo: StorageInfo | null;
  
  // Storage usage statistics
  storageUsage: StorageUsage | null;
  
  // Threshold status
  thresholdStatus: ThresholdStatus | null;
  
  // Monitoring state
  isMonitoring: boolean;
  monitoringInterval: number;
  
  // Loading states
  isLoading: boolean;
  isCheckingThresholds: boolean;
  
  // Error states
  error: string | null;
  lastError: string | null;
  
  // Warning states
  hasWarnings: boolean;
  hasCriticalIssues: boolean;
  
  // Warnings array for component compatibility
  warnings: string[]; // Added for component compatibility
}

/**
 * Storage Store Actions Interface
 */
interface StorageStoreActions {
  // State management
  setStorageInfo: (info: StorageInfo) => void;
  setStorageUsage: (usage: StorageUsage) => void;
  setThresholdStatus: (status: ThresholdStatus) => void;
  
  // Monitoring states
  setMonitoring: (monitoring: boolean) => void;
  setMonitoringInterval: (interval: number) => void;
  
  // Loading states
  setLoading: (loading: boolean) => void;
  setCheckingThresholds: (checking: boolean) => void;
  
  // Error states
  setError: (error: string | null) => void;
  setLastError: (error: string | null) => void;
  clearErrors: () => void;
  
  // Warning states
  setHasWarnings: (hasWarnings: boolean) => void;
  setHasCriticalIssues: (hasCritical: boolean) => void;
  
  // Storage operations
  refreshStorageInfo: () => Promise<void>;
  refreshStorage: () => Promise<void>; // Added for component compatibility
  checkThresholds: () => Promise<void>;
  startMonitoring: (interval?: number) => void;
  stopMonitoring: () => void;
  
  // State queries
  getStorageInfo: () => StorageInfo | null;
  getStorageUsage: () => StorageUsage | null;
  getThresholdStatus: () => ThresholdStatus | null;
  isStorageAvailable: () => boolean;
  isStorageCritical: () => boolean;
  isStorageWarning: () => boolean;
  
  // Service integration
  initialize: () => void;
  cleanup: () => void;
}

/**
 * Storage Store Type
 */
type StorageStore = StorageStoreState & StorageStoreActions;

/**
 * Storage Store Implementation
 */
export const useStorageStore = create<StorageStore>((set, get) => ({
  // Initial state
  storageInfo: null,
  storageUsage: null,
  thresholdStatus: null,
  isMonitoring: false,
  monitoringInterval: 30000,
  isLoading: false,
  isCheckingThresholds: false,
  warnings: [], // Added for component compatibility
  error: null,
  lastError: null,
  hasWarnings: false,
  hasCriticalIssues: false,

  // State management actions
  setStorageInfo: (info: StorageInfo) => {
    set({ storageInfo: info });
  },

  setStorageUsage: (usage: StorageUsage) => {
    set({ storageUsage: usage });
  },

  setThresholdStatus: (status: ThresholdStatus) => {
    set({ 
      thresholdStatus: status,
      hasWarnings: status.is_warning,
      hasCriticalIssues: status.is_critical
    });
  },

  // Monitoring state actions
  setMonitoring: (monitoring: boolean) => {
    set({ isMonitoring: monitoring });
  },

  setMonitoringInterval: (interval: number) => {
    set({ monitoringInterval: interval });
  },

  // Loading state actions
  setLoading: (loading: boolean) => {
    set({ isLoading: loading });
  },

  setCheckingThresholds: (checking: boolean) => {
    set({ isCheckingThresholds: checking });
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

  // Warning state actions
  setHasWarnings: (hasWarnings: boolean) => {
    set({ hasWarnings });
  },

  setHasCriticalIssues: (hasCritical: boolean) => {
    set({ hasCriticalIssues: hasCritical });
  },

  // Storage operations
  refreshStorageInfo: async () => {
    const { setLoading, setError, setLastError } = get();
    
    try {
      setLoading(true);
      setError(null);
      
      const info = await storageMonitorService.getStorageInfo();
      const usage = await storageMonitorService.getStorageUsage();
      
      const { setStorageInfo, setStorageUsage } = get();
      setStorageInfo(info);
      setStorageUsage(usage);
      
    } catch (error: any) {
      const errorMessage = error.message || 'Failed to refresh storage info';
      setError(errorMessage);
      setLastError(errorMessage);
      throw error;
    } finally {
      setLoading(false);
    }
  },

  checkThresholds: async () => {
    const { setCheckingThresholds, setError, setLastError } = get();
    
    try {
      setCheckingThresholds(true);
      setError(null);
      
      const status = await storageMonitorService.checkStorageThresholds();
      
      const { setThresholdStatus } = get();
      setThresholdStatus(status);
      
    } catch (error: any) {
      const errorMessage = error.message || 'Failed to check storage thresholds';
      setError(errorMessage);
      setLastError(errorMessage);
      throw error;
    } finally {
      setCheckingThresholds(false);
    }
  },

  startMonitoring: (interval?: number) => {
    const { setMonitoring, setMonitoringInterval } = get();
    const monitoringInterval = interval || get().monitoringInterval;
    
    setMonitoringInterval(monitoringInterval);
    storageMonitorService.startStorageMonitoring(monitoringInterval);
    setMonitoring(true);
  },

  stopMonitoring: () => {
    const { setMonitoring } = get();
    storageMonitorService.stopStorageMonitoring();
    setMonitoring(false);
  },

  // State queries
  getStorageInfo: () => {
    return get().storageInfo;
  },

  getStorageUsage: () => {
    return get().storageUsage;
  },

  getThresholdStatus: () => {
    return get().thresholdStatus;
  },

  isStorageAvailable: () => {
    const status = get().thresholdStatus;
    return status ? !status.is_critical : true;
  },

  isStorageCritical: () => {
    const status = get().thresholdStatus;
    return status ? status.is_critical : false;
  },

  isStorageWarning: () => {
    const status = get().thresholdStatus;
    return status ? status.is_warning : false;
  },

  // Storage refresh method for component compatibility
  refreshStorage: async () => {
    const { setLoading, setError } = get();
    setLoading(true);
    setError(null);
    
    try {
      // Refresh storage info and check thresholds
      await get().refreshStorageInfo();
      await get().checkThresholds();
      
      // Update warnings based on threshold status
      const status = get().thresholdStatus;
      const warnings: string[] = [];
      
      if (status) {
        if (status.current_status === 'critical') {
          warnings.push('Storage space is critical');
        } else if (status.current_status === 'warning') {
          warnings.push('Storage space is low');
        }
      }
      
      // Update warnings in state
      set({ warnings });
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to refresh storage';
      setError(errorMessage);
      throw new Error(errorMessage);
    } finally {
      setLoading(false);
    }
  },

  // Service integration
  initialize: () => {
    // Set up event handlers
    storageMonitorService.onStorageUpdate((info) => {
      get().setStorageInfo(info);
    });

    storageMonitorService.onStorageThresholdExceeded((threshold) => {
      get().setThresholdStatus(threshold);
    });

    storageMonitorService.onStorageCritical((info) => {
      get().setStorageInfo(info);
      get().setHasCriticalIssues(true);
    });

    // Initial load
    get().refreshStorageInfo().catch(error => {
      console.error('Failed to initialize storage store:', error);
    });
  },

  cleanup: () => {
    storageMonitorService.cleanup();
    set({
      storageInfo: null,
      storageUsage: null,
      thresholdStatus: null,
      isMonitoring: false,
      monitoringInterval: 30000,
      isLoading: false,
      isCheckingThresholds: false,
      error: null,
      lastError: null,
      hasWarnings: false,
      hasCriticalIssues: false
    });
  }
}));
