import { create } from 'zustand';
import { ServerState, ServerInfo, SystemStatus, SystemReadinessStatus, StorageInfo } from '../../types/api';
import { ServerService } from '../../services/server/ServerService';

interface ServerStore extends ServerState {
  // Service injection
  setServerService: (service: ServerService) => void;
  
  // State setters
  setInfo: (info: ServerInfo | null) => void;
  setStatus: (status: SystemStatus | null) => void;
  setSystemReadiness: (systemReadiness: SystemReadinessStatus | null) => void;
  setStorage: (storage: StorageInfo | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setLastUpdated: (timestamp: string | null) => void;
  
  // Actions that call services
  loadServerInfo: () => Promise<void>;
  loadSystemStatus: () => Promise<void>;
  loadSystemReadiness: () => Promise<void>;
  loadStorageInfo: () => Promise<void>;
  loadAllServerData: () => Promise<void>;
  
  // Real-time notification handlers
  handleSystemStatusUpdate: (status: any) => void;
  
  // Reset
  reset: () => void;
}

const initialState: ServerState = {
  info: null,
  status: null,
  systemReadiness: null,
  storage: null,
  loading: false,
  error: null,
  lastUpdated: null,
};

export const useServerStore = create<ServerStore>((set) => {
  let serverService: ServerService | null = null;

  return {
    ...initialState,

    // Service injection
    setServerService: (service: ServerService) => {
      serverService = service;
    },

    // State setters
    setInfo: (info: ServerInfo | null) => set((state) => ({ ...state, info })),

    setStatus: (status: SystemStatus | null) => set((state) => ({ ...state, status })),

    setSystemReadiness: (systemReadiness: SystemReadinessStatus | null) => set((state) => ({ ...state, systemReadiness })),

    setStorage: (storage: StorageInfo | null) => set((state) => ({ ...state, storage })),

    setLoading: (loading: boolean) => set((state) => ({ ...state, loading })),

    setError: (error: string | null) => set((state) => ({ ...state, error })),

    setLastUpdated: (timestamp: string | null) =>
      set((state) => ({ ...state, lastUpdated: timestamp })),

    // Actions that call services
    loadServerInfo: async () => {
      if (!serverService) throw new Error('Server service not initialized');
      set({ loading: true, error: null });
      try {
        const info = await serverService.getServerInfo();
        set({ info, loading: false, lastUpdated: new Date().toISOString() });
      } catch (error) {
        set({ 
          loading: false, 
          error: error instanceof Error ? error.message : 'Failed to load server info' 
        });
        throw error;
      }
    },

    loadSystemStatus: async () => {
      if (!serverService) throw new Error('Server service not initialized');
      set({ loading: true, error: null });
      try {
        const status = await serverService.getStatus();
        set({ status, loading: false, lastUpdated: new Date().toISOString() });
      } catch (error) {
        set({ 
          loading: false, 
          error: error instanceof Error ? error.message : 'Failed to load system status' 
        });
      }
    },

    loadSystemReadiness: async () => {
      if (!serverService) throw new Error('Server service not initialized');
      set({ loading: true, error: null });
      try {
        const systemReadiness = await serverService.getSystemStatus();
        set({ systemReadiness, loading: false, lastUpdated: new Date().toISOString() });
      } catch (error) {
        set({ 
          loading: false, 
          error: error instanceof Error ? error.message : 'Failed to load system readiness' 
        });
      }
    },

    loadStorageInfo: async () => {
      if (!serverService) throw new Error('Server service not initialized');
      set({ loading: true, error: null });
      try {
        const storage = await serverService.getStorageInfo();
        set({ storage, loading: false, lastUpdated: new Date().toISOString() });
      } catch (error) {
        set({ 
          loading: false, 
          error: error instanceof Error ? error.message : 'Failed to load storage info' 
        });
      }
    },

    loadAllServerData: async () => {
      if (!serverService) throw new Error('Server service not initialized');
      set({ loading: true, error: null });
      try {
        const [info, status, storage] = await Promise.all([
          serverService.getServerInfo(),
          serverService.getStatus(),
          serverService.getStorageInfo(),
        ]);
        set({ info, status, storage, loading: false, lastUpdated: new Date().toISOString() });
      } catch (error) {
        set({ 
          loading: false, 
          error: error instanceof Error ? error.message : 'Failed to load server data' 
        });
      }
    },

    // Real-time notification handlers
    handleSystemStatusUpdate: (status: any) => {
      console.log('ServerStore: Handling system status update', status);
      set({ status, lastUpdated: new Date().toISOString() });
    },

    reset: () => set(initialState),
  };
});
