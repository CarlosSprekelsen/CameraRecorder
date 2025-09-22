/**
 * Central export point for all stores
 * Provides unified access to all Zustand stores
 */

// Core stores
export { useCameraStore } from './cameraStore';
export { useAdminStore } from './adminStore';
export { useFileStore } from './fileStore';
export { useAuthStore } from './authStore';
export { useSettingsStore } from './settingsStore';

// Modular connection stores (NEW)
export { 
  useConnectionStore, 
  useHealthStore, 
  useMetricsStore,
  useUnifiedConnectionState 
} from './connection';

// Store types
export type { CameraStoreState } from './cameraStore';
export type { AdminStoreState } from './adminStore';
export type { FileStoreState } from './fileStore';
export type { AuthStoreState } from './authStore';
export type { SettingsStoreState } from './settingsStore';

// Modular connection store types (NEW)
export type { 
  ConnectionStoreState, 
  ConnectionStoreActions,
  HealthStoreState,
  HealthStoreActions,
  MetricsStoreState,
  MetricsStoreActions,
  UnifiedConnectionState
} from './connection'; 