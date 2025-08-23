/**
 * Central export point for all stores
 * Provides unified access to all Zustand stores
 */

// Core stores
export { useCameraStore } from './cameraStore';
export { useConnectionStore } from './connectionStore';
export { useHealthStore } from './healthStore';
export { useAdminStore } from './adminStore';
export { useFileStore } from './fileStore';
export { useAuthStore } from './authStore';
export { useSettingsStore } from './settingsStore';

// Store types
export type { CameraStoreState } from './cameraStore';
export type { ConnectionStoreState } from './connectionStore';
export type { HealthStoreState } from './healthStore';
export type { AdminStoreState } from './adminStore';
export type { FileStoreState } from './fileStore';
export type { AuthStoreState } from './authStore';
export type { SettingsStoreState } from './settingsStore'; 