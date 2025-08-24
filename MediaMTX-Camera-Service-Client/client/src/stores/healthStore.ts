/**
 * Health state management store
 * Manages system health status and component health monitoring
 * Aligned with server health endpoints API
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/health-endpoints.md
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

/**
 * Health status types
 */
export type HealthStatus = 'healthy' | 'degraded' | 'unhealthy';

/**
 * Component health status
 */
export interface ComponentHealth {
  status: HealthStatus;
  details: string;
  timestamp: string;
}

/**
 * System health response from server
 */
export interface SystemHealth {
  status: HealthStatus;
  timestamp: string;
  components: {
    mediamtx: ComponentHealth;
    camera_monitor: ComponentHealth;
    service_manager: ComponentHealth;
  };
}

/**
 * Camera system health
 */
export interface CameraHealth {
  status: HealthStatus;
  timestamp: string;
  details: string;
}

/**
 * MediaMTX integration health
 */
export interface MediaMTXHealth {
  status: HealthStatus;
  timestamp: string;
  details: string;
}

/**
 * Kubernetes readiness status
 */
export interface ReadinessStatus {
  status: 'ready' | 'not_ready';
  timestamp: string;
  details?: {
    [component: string]: string;
  };
}

/**
 * Health store state interface
 */
export interface HealthStoreState {
  // System health
  systemHealth: SystemHealth | null;
  cameraHealth: CameraHealth | null;
  mediamtxHealth: MediaMTXHealth | null;
  readinessStatus: ReadinessStatus | null;
  
  // Health monitoring state
  isMonitoring: boolean;
  lastUpdate: Date | null;
  updateInterval: number; // milliseconds
  
  // Health polling state
  isPolling: boolean;
  pollCount: number;
  errorCount: number;
  lastPollTime: Date | null;
  
  // Health history
  healthHistory: SystemHealth[];
  maxHistorySize: number;
}

/**
 * Health store actions interface
 */
interface HealthActions {
  // Health data management
  setSystemHealth: (health: SystemHealth) => void;
  setCameraHealth: (health: CameraHealth) => void;
  setMediaMTXHealth: (health: MediaMTXHealth) => void;
  setReadinessStatus: (status: ReadinessStatus) => void;
  
  // Health monitoring
  startMonitoring: () => void;
  stopMonitoring: () => void;
  setUpdateInterval: (interval: number) => void;
  
  // Health polling
  startPolling: () => void;
  stopPolling: () => void;
  incrementPollCount: () => void;
  incrementErrorCount: () => void;
  setLastPollTime: (time: Date) => void;
  
  // Health history
  addToHistory: (health: SystemHealth) => void;
  clearHistory: () => void;
  setMaxHistorySize: (size: number) => void;
  
  // Utility methods
  getOverallHealth: () => HealthStatus;
  getHealthScore: () => number; // 0-100
  isSystemReady: () => boolean;
}

/**
 * Health store type
 */
type HealthStore = HealthStoreState & HealthActions;

/**
 * Create health store
 */
export const useHealthStore = create<HealthStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      systemHealth: null,
      cameraHealth: null,
      mediamtxHealth: null,
      readinessStatus: null,
      
      isMonitoring: false,
      lastUpdate: null,
      updateInterval: 30000, // 30 seconds
      
      isPolling: false,
      pollCount: 0,
      errorCount: 0,
      lastPollTime: null,
      
      healthHistory: [],
      maxHistorySize: 100,
      
      // Health data management
      setSystemHealth: (health: SystemHealth) => {
        set((state) => {
          const newHistory = [...state.healthHistory, health];
          if (newHistory.length > state.maxHistorySize) {
            newHistory.shift();
          }
          
          return {
            systemHealth: health,
            lastUpdate: new Date(),
            healthHistory: newHistory,
          };
        });
      },
      
      setCameraHealth: (health: CameraHealth) => {
        set({
          cameraHealth: health,
          lastUpdate: new Date(),
        });
      },
      
      setMediaMTXHealth: (health: MediaMTXHealth) => {
        set({
          mediamtxHealth: health,
          lastUpdate: new Date(),
        });
      },
      
      setReadinessStatus: (status: ReadinessStatus) => {
        set({
          readinessStatus: status,
          lastUpdate: new Date(),
        });
      },
      
      // Health monitoring
      startMonitoring: () => {
        set({ isMonitoring: true });
      },
      
      stopMonitoring: () => {
        set({ isMonitoring: false });
      },
      
      setUpdateInterval: (interval: number) => {
        set({ updateInterval: interval });
      },
      
      // Health polling
      startPolling: () => {
        set({ isPolling: true });
      },
      
      stopPolling: () => {
        set({ isPolling: false });
      },
      
      incrementPollCount: () => {
        set((state) => ({ pollCount: state.pollCount + 1 }));
      },
      
      incrementErrorCount: () => {
        set((state) => ({ errorCount: state.errorCount + 1 }));
      },
      
      setLastPollTime: (time: Date) => {
        set({ lastPollTime: time });
      },
      
      // Health history
      addToHistory: (health: SystemHealth) => {
        set((state) => {
          const newHistory = [...state.healthHistory, health];
          if (newHistory.length > state.maxHistorySize) {
            newHistory.shift();
          }
          return { healthHistory: newHistory };
        });
      },
      
      clearHistory: () => {
        set({ healthHistory: [] });
      },
      
      setMaxHistorySize: (size: number) => {
        set({ maxHistorySize: size });
      },
      
      // Utility methods
      getOverallHealth: () => {
        const { systemHealth, cameraHealth, mediamtxHealth } = get();
        
        if (!systemHealth) return 'unhealthy';
        
        // Check if any component is unhealthy
        const components = [systemHealth.components.mediamtx, systemHealth.components.camera_monitor, systemHealth.components.service_manager];
        if (components.some(comp => comp.status === 'unhealthy')) {
          return 'unhealthy';
        }
        
        // Check if any component is degraded
        if (components.some(comp => comp.status === 'degraded')) {
          return 'degraded';
        }
        
        return 'healthy';
      },
      
      getHealthScore: () => {
        const { systemHealth, cameraHealth, mediamtxHealth } = get();
        
        if (!systemHealth) return 0;
        
        let score = 0;
        let totalComponents = 0;
        
        // System health components
        Object.values(systemHealth.components).forEach(component => {
          totalComponents++;
          switch (component.status) {
            case 'healthy':
              score += 100;
              break;
            case 'degraded':
              score += 50;
              break;
            case 'unhealthy':
              score += 0;
              break;
          }
        });
        
        // Additional health checks
        if (cameraHealth) {
          totalComponents++;
          switch (cameraHealth.status) {
            case 'healthy':
              score += 100;
              break;
            case 'degraded':
              score += 50;
              break;
            case 'unhealthy':
              score += 0;
              break;
          }
        }
        
        if (mediamtxHealth) {
          totalComponents++;
          switch (mediamtxHealth.status) {
            case 'healthy':
              score += 100;
              break;
            case 'degraded':
              score += 50;
              break;
            case 'unhealthy':
              score += 0;
              break;
          }
        }
        
        return totalComponents > 0 ? Math.round(score / totalComponents) : 0;
      },
      
      isSystemReady: () => {
        const { readinessStatus } = get();
        return readinessStatus?.status === 'ready';
      },
    }),
    {
      name: 'health-store',
    }
  )
);
