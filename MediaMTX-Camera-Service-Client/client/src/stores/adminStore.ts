/**
 * Admin Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around AdminService
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';
import { adminService } from '../services/adminService';

interface AdminStoreState {
  isLoading: boolean;
  error: string | null;
  systemInfo: any;
  serverStats: any;
}

interface AdminStoreActions {
  getSystemInfo: () => Promise<void>;
  getServerStats: () => Promise<void>;
  clearError: () => void;
  setError: (error: string) => void;
}

type AdminStore = AdminStoreState & AdminStoreActions;

const initialState: AdminStoreState = {
  isLoading: false,
  error: null,
  systemInfo: null,
  serverStats: null,
};

export const useAdminStore = create<AdminStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      getSystemInfo: async () => {
        set({ isLoading: true, error: null });
        try {
          // TODO: Implement with AdminService
          set({ isLoading: false });
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get system info';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      getServerStats: async () => {
        set({ isLoading: true, error: null });
        try {
          // TODO: Implement with AdminService
          set({ isLoading: false });
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get server stats';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    { name: 'admin-store' }
  )
);
