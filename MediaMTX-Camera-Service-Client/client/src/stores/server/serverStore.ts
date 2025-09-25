import { create } from 'zustand';
import { ServerState, ServerInfo, SystemStatus, StorageInfo } from '../../types/api';

interface ServerStore extends ServerState {
  setInfo: (info: ServerInfo | null) => void;
  setStatus: (status: SystemStatus | null) => void;
  setStorage: (storage: StorageInfo | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setLastUpdated: (timestamp: string | null) => void;
  reset: () => void;
}

const initialState: ServerState = {
  info: null,
  status: null,
  storage: null,
  loading: false,
  error: null,
  lastUpdated: null,
};

export const useServerStore = create<ServerStore>((set) => ({
  ...initialState,

  setInfo: (info: ServerInfo | null) => set((state) => ({ ...state, info })),

  setStatus: (status: SystemStatus | null) => set((state) => ({ ...state, status })),

  setStorage: (storage: StorageInfo | null) => set((state) => ({ ...state, storage })),

  setLoading: (loading: boolean) => set((state) => ({ ...state, loading })),

  setError: (error: string | null) => set((state) => ({ ...state, error })),

  setLastUpdated: (timestamp: string | null) =>
    set((state) => ({ ...state, lastUpdated: timestamp })),

  reset: () => set(initialState),
}));
