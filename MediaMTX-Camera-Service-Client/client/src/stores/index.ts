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
export type { CameraState } from './cameraStore';
export type { ConnectionState } from './connectionStore';
export type { HealthState } from './healthStore';
export type { AdminState } from './adminStore';
export type { FileState } from './fileStore';
export type { AuthState } from './authStore';
export type { SettingsState } from './settingsStore'; 