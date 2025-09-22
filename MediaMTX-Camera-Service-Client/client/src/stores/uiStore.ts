/**
 * UI Store - Architecture Compliant (<200 lines)
 * 
 * This store provides UI state management
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';

interface UIStoreState {
  // UI state
  sidebarOpen: boolean;
  theme: 'light' | 'dark';
  notifications: any[];
  
  // Loading states
  isLoading: boolean;
  
  // Error state
  error: string | null;
}

interface UIStoreActions {
  // UI operations
  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  addNotification: (notification: any) => void;
  removeNotification: (id: string) => void;
  clearNotifications: () => void;
  
  // Error handling
  clearError: () => void;
  setError: (error: string) => void;
}

type UIStore = UIStoreState & UIStoreActions;

const initialState: UIStoreState = {
  sidebarOpen: true,
  theme: 'light',
  notifications: [],
  isLoading: false,
  error: null,
};

export const useUIStore = create<UIStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      toggleSidebar: () => {
        set(state => ({ sidebarOpen: !state.sidebarOpen }));
      },
      
      setTheme: (theme: 'light' | 'dark') => {
        set({ theme });
      },
      
      addNotification: (notification: any) => {
        set(state => ({
          notifications: [...state.notifications, notification]
        }));
      },
      
      removeNotification: (id: string) => {
        set(state => ({
          notifications: state.notifications.filter(n => n.id !== id)
        }));
      },
      
      clearNotifications: () => {
        set({ notifications: [] });
      },
      
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    { name: 'ui-store' }
  )
);
