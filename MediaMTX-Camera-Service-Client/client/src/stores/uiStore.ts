/**
 * UI state management store
 * Handles UI state like selected camera, view modes, theme, notifications
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type {
  ViewMode,
  ThemeMode,
  NotificationState,
  LoadingState,
  ErrorState,
} from '../types/ui';

/**
 * UI store state interface
 */
interface UIState {
  // Navigation
  selectedCamera: string | null;
  viewMode: ViewMode;
  
  // Theme
  theme: ThemeMode;
  
  // Layout
  sidebarOpen: boolean;
  
  // Notifications
  notifications: NotificationState[];
  
  // Loading states
  loading: LoadingState;
  
  // Error state
  error: ErrorState;
  
  // Settings
  autoRefresh: boolean;
  refreshInterval: number;
  showNotifications: boolean;
}

/**
 * UI store actions interface
 */
interface UIActions {
  // Navigation
  selectCamera: (device: string | null) => void;
  setViewMode: (mode: ViewMode) => void;
  
  // Theme
  setTheme: (theme: ThemeMode) => void;
  toggleTheme: () => void;
  
  // Layout
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  
  // Notifications
  addNotification: (notification: Omit<NotificationState, 'id' | 'timestamp'>) => void;
  removeNotification: (id: string) => void;
  clearNotifications: () => void;
  clearNotification: (id: string) => void;
  
  // Loading
  setLoading: (loading: boolean, message?: string) => void;
  clearLoading: () => void;
  
  // Error handling
  setError: (error: string | Error | null) => void;
  clearError: () => void;
  
  // Settings
  setAutoRefresh: (enabled: boolean) => void;
  setRefreshInterval: (interval: number) => void;
  setShowNotifications: (enabled: boolean) => void;
  
  // Utility
  resetUI: () => void;
}

/**
 * UI store type
 */
type UIStore = UIState & UIActions;

/**
 * Generate unique notification ID
 */
const generateNotificationId = (): string => {
  return `notification-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
};

/**
 * Create UI store
 */
export const useUIStore = create<UIStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      selectedCamera: null,
      viewMode: 'dashboard',
      theme: 'light',
      sidebarOpen: false,
      notifications: [],
      loading: { isLoading: false },
      error: { hasError: false },
      autoRefresh: true,
      refreshInterval: 30000, // 30 seconds
      showNotifications: true,

      // Navigation
      selectCamera: (device: string | null) => {
        set({ selectedCamera: device });
      },

      setViewMode: (mode: ViewMode) => {
        set({ viewMode: mode });
      },

      // Theme
      setTheme: (theme: ThemeMode) => {
        set({ theme });
        // Store theme preference in localStorage
        if (typeof window !== 'undefined') {
          localStorage.setItem('theme', theme);
        }
      },

      toggleTheme: () => {
        const { theme } = get();
        const newTheme: ThemeMode = theme === 'light' ? 'dark' : 'light';
        get().setTheme(newTheme);
      },

      // Layout
      toggleSidebar: () => {
        set((state) => ({ sidebarOpen: !state.sidebarOpen }));
      },

      setSidebarOpen: (open: boolean) => {
        set({ sidebarOpen: open });
      },

      // Notifications
      addNotification: (notification) => {
        const newNotification: NotificationState = {
          id: generateNotificationId(),
          timestamp: new Date(),
          ...notification,
        };

        set((state) => ({
          notifications: [...state.notifications, newNotification],
        }));

        // Auto-remove notification after duration (default: 5 seconds)
        const duration = notification.duration || 5000;
        if (notification.autoClose !== false && duration > 0) {
          setTimeout(() => {
            get().removeNotification(newNotification.id);
          }, duration);
        }
      },

      removeNotification: (id: string) => {
        set((state) => ({
          notifications: state.notifications.filter(n => n.id !== id),
        }));
      },

      clearNotifications: () => {
        set({ notifications: [] });
      },

      clearNotification: (id: string) => {
        get().removeNotification(id);
      },

      // Loading
      setLoading: (isLoading: boolean, message?: string) => {
        set({
          loading: {
            isLoading,
            message,
          },
        });
      },

      clearLoading: () => {
        set({
          loading: {
            isLoading: false,
          },
        });
      },

      // Error handling
      setError: (error: string | Error | null) => {
        if (error === null) {
          set({
            error: {
              hasError: false,
            },
          });
        } else {
          const errorMessage = error instanceof Error ? error.message : error;
          set({
            error: {
              hasError: true,
              error: errorMessage,
              timestamp: new Date(),
            },
          });
        }
      },

      clearError: () => {
        set({
          error: {
            hasError: false,
          },
        });
      },

      // Settings
      setAutoRefresh: (enabled: boolean) => {
        set({ autoRefresh: enabled });
        // Store setting in localStorage
        if (typeof window !== 'undefined') {
          localStorage.setItem('autoRefresh', enabled.toString());
        }
      },

      setRefreshInterval: (interval: number) => {
        set({ refreshInterval: interval });
        // Store setting in localStorage
        if (typeof window !== 'undefined') {
          localStorage.setItem('refreshInterval', interval.toString());
        }
      },

      setShowNotifications: (enabled: boolean) => {
        set({ showNotifications: enabled });
        // Store setting in localStorage
        if (typeof window !== 'undefined') {
          localStorage.setItem('showNotifications', enabled.toString());
        }
      },

      // Utility
      resetUI: () => {
        set({
          selectedCamera: null,
          viewMode: 'dashboard',
          sidebarOpen: false,
          notifications: [],
          loading: { isLoading: false },
          error: { hasError: false },
        });
      },
    }),
    {
      name: 'ui-store',
    }
  )
);

/**
 * Initialize UI store with saved preferences
 */
export const initializeUIStore = () => {
  if (typeof window === 'undefined') return;

  const savedTheme = localStorage.getItem('theme') as ThemeMode;
  const savedAutoRefresh = localStorage.getItem('autoRefresh');
  const savedRefreshInterval = localStorage.getItem('refreshInterval');
  const savedShowNotifications = localStorage.getItem('showNotifications');

  if (savedTheme) {
    useUIStore.getState().setTheme(savedTheme);
  }

  if (savedAutoRefresh !== null) {
    useUIStore.getState().setAutoRefresh(savedAutoRefresh === 'true');
  }

  if (savedRefreshInterval !== null) {
    useUIStore.getState().setRefreshInterval(parseInt(savedRefreshInterval, 10));
  }

  if (savedShowNotifications !== null) {
    useUIStore.getState().setShowNotifications(savedShowNotifications === 'true');
  }
}; 