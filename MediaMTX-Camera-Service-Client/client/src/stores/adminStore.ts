/**
 * Admin state management store
 * Manages system administration and management functionality
 * Aligned with server JSON-RPC methods for admin operations
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

/**
 * System metrics from server
 */
export interface SystemMetrics {
  active_connections: number;
  total_requests: number;
  average_response_time: number;
  error_rate: number;
  memory_usage: number;
  cpu_usage: number;
}

/**
 * System status from server
 */
export interface SystemStatus {
  status: 'healthy' | 'degraded' | 'unhealthy';
  uptime: number;
  version: string;
  components: {
    websocket_server: string;
    camera_monitor: string;
    mediamtx_controller: string;
  };
}

/**
 * Server information from server
 */
export interface ServerInfo {
  name: string;
  version: string;
  capabilities: string[];
  supported_formats: string[];
  max_cameras: number;
}

/**
 * Storage information from server
 */
export interface StorageInfo {
  total_space: number;
  used_space: number;
  available_space: number;
  usage_percentage: number;
  recordings_size: number;
  snapshots_size: number;
  low_space_warning: boolean;
}

/**
 * Retention policy configuration
 */
export interface RetentionPolicy {
  policy_type: 'age' | 'size' | 'manual';
  max_age_days?: number;
  max_size_gb?: number;
  enabled: boolean;
}

/**
 * Cleanup results
 */
export interface CleanupResults {
  cleanup_executed: boolean;
  files_deleted: number;
  space_freed: number;
  message: string;
}

/**
 * Admin store state interface
 */
interface AdminState {
  // System information
  systemMetrics: SystemMetrics | null;
  systemStatus: SystemStatus | null;
  serverInfo: ServerInfo | null;
  storageInfo: StorageInfo | null;
  
  // Configuration
  retentionPolicy: RetentionPolicy | null;
  
  // Operations
  isPerformingCleanup: boolean;
  lastCleanupResults: CleanupResults | null;
  
  // Admin state
  isAdmin: boolean;
  hasAdminPermissions: boolean;
  
  // Loading states
  isLoadingMetrics: boolean;
  isLoadingStatus: boolean;
  isLoadingStorage: boolean;
  
  // Error states
  error: string | null;
  errorTimestamp: Date | null;
}

/**
 * Admin store actions interface
 */
interface AdminActions {
  // System information management
  setSystemMetrics: (metrics: SystemMetrics) => void;
  setSystemStatus: (status: SystemStatus) => void;
  setServerInfo: (info: ServerInfo) => void;
  setStorageInfo: (info: StorageInfo) => void;
  
  // Configuration management
  setRetentionPolicy: (policy: RetentionPolicy) => void;
  
  // Operations
  setCleanupResults: (results: CleanupResults) => void;
  setPerformingCleanup: (performing: boolean) => void;
  
  // Admin state
  setAdminStatus: (isAdmin: boolean, hasPermissions: boolean) => void;
  
  // Loading states
  setLoadingMetrics: (loading: boolean) => void;
  setLoadingStatus: (loading: boolean) => void;
  setLoadingStorage: (loading: boolean) => void;
  
  // Error handling
  setError: (error: string | null) => void;
  clearError: () => void;
  
  // Utility methods
  getStorageUsagePercentage: () => number;
  getStorageUsageColor: () => 'success' | 'warning' | 'error';
  isLowSpace: () => boolean;
  formatBytes: (bytes: number) => string;
  formatUptime: (seconds: number) => string;
}

/**
 * Admin store type
 */
type AdminStore = AdminState & AdminActions;

/**
 * Create admin store
 */
export const useAdminStore = create<AdminStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      systemMetrics: null,
      systemStatus: null,
      serverInfo: null,
      storageInfo: null,
      
      retentionPolicy: null,
      
      isPerformingCleanup: false,
      lastCleanupResults: null,
      
      isAdmin: false,
      hasAdminPermissions: false,
      
      isLoadingMetrics: false,
      isLoadingStatus: false,
      isLoadingStorage: false,
      
      error: null,
      errorTimestamp: null,
      
      // System information management
      setSystemMetrics: (metrics: SystemMetrics) => {
        set({
          systemMetrics: metrics,
          isLoadingMetrics: false,
        });
      },
      
      setSystemStatus: (status: SystemStatus) => {
        set({
          systemStatus: status,
          isLoadingStatus: false,
        });
      },
      
      setServerInfo: (info: ServerInfo) => {
        set({ serverInfo: info });
      },
      
      setStorageInfo: (info: StorageInfo) => {
        set({
          storageInfo: info,
          isLoadingStorage: false,
        });
      },
      
      // Configuration management
      setRetentionPolicy: (policy: RetentionPolicy) => {
        set({ retentionPolicy: policy });
      },
      
      // Operations
      setCleanupResults: (results: CleanupResults) => {
        set({
          lastCleanupResults: results,
          isPerformingCleanup: false,
        });
      },
      
      setPerformingCleanup: (performing: boolean) => {
        set({ isPerformingCleanup: performing });
      },
      
      // Admin state
      setAdminStatus: (isAdmin: boolean, hasPermissions: boolean) => {
        set({
          isAdmin,
          hasAdminPermissions: hasPermissions,
        });
      },
      
      // Loading states
      setLoadingMetrics: (loading: boolean) => {
        set({ isLoadingMetrics: loading });
      },
      
      setLoadingStatus: (loading: boolean) => {
        set({ isLoadingStatus: loading });
      },
      
      setLoadingStorage: (loading: boolean) => {
        set({ isLoadingStorage: loading });
      },
      
      // Error handling
      setError: (error: string | null) => {
        set({
          error,
          errorTimestamp: error ? new Date() : null,
        });
      },
      
      clearError: () => {
        set({
          error: null,
          errorTimestamp: null,
        });
      },
      
      // Utility methods
      getStorageUsagePercentage: () => {
        const { storageInfo } = get();
        return storageInfo?.usage_percentage || 0;
      },
      
      getStorageUsageColor: () => {
        const usage = get().getStorageUsagePercentage();
        if (usage < 70) return 'success';
        if (usage < 90) return 'warning';
        return 'error';
      },
      
      isLowSpace: () => {
        const { storageInfo } = get();
        return storageInfo?.low_space_warning || false;
      },
      
      formatBytes: (bytes: number) => {
        if (bytes === 0) return '0 B';
        
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
      },
      
      formatUptime: (seconds: number) => {
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        if (days > 0) {
          return `${days}d ${hours}h ${minutes}m`;
        } else if (hours > 0) {
          return `${hours}h ${minutes}m`;
        } else {
          return `${minutes}m`;
        }
      },
    }),
    {
      name: 'admin-store',
    }
  )
);
