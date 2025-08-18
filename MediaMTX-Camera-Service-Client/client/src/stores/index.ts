/**
 * Central export point for all stores
 * Provides unified access to camera, connection, and UI stores
 */

// Export individual stores
export { useCameraStore } from './cameraStore';
export { useConnectionStore } from './connectionStore';
export { useUIStore, initializeUIStore } from './uiStore';

// TODO: Fix store initialization in Sprint 3
// Temporarily disabled to complete Sprint 2 