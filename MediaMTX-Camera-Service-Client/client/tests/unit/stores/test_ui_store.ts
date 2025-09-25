/**
 * REQ-UI01-001: UI state management must provide consistent user interface state
 * REQ-UI01-002: UI notifications must provide clear user feedback and alerts
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for UI store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on UI state management and user interaction logic
 * - Test navigation, theme, and notification management
 * - Validate UI state consistency and user experience
 */

import { useUIStore } from '../../../src/stores/uiStore';
import type { ViewMode, ThemeMode, NotificationState, LoadingState, ErrorState } from '../../../src/types/ui';

describe('UI Store', () => {
  let store: ReturnType<typeof useUIStore.getState>;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useUIStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useUIStore.getState();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useUIStore.getState();
      expect(state.selectedCamera).toBeNull();
      expect(state.viewMode).toBe('grid');
      expect(state.theme).toBe('light');
      expect(state.sidebarOpen).toBe(true);
      expect(state.notifications).toEqual([]);
      expect(state.loading).toEqual({
        cameras: false,
        recordings: false,
        snapshots: false,
        system: false
      });
      expect(state.error).toEqual({
        message: null,
        code: null,
        timestamp: null
      });
      expect(state.autoRefresh).toBe(true);
      expect(state.refreshInterval).toBe(30000);
      expect(state.showNotifications).toBe(true);
    });
  });

  describe('Navigation Management', () => {
    it('should select camera', () => {
      store.selectCamera('camera0');
      
      const state = useUIStore.getState();
      expect(state.selectedCamera).toBe('camera0');
    });

    it('should clear camera selection', () => {
      store.selectCamera('camera0');
      store.selectCamera(null);
      
      const state = useUIStore.getState();
      expect(state.selectedCamera).toBeNull();
    });

    it('should set view mode', () => {
      store.setViewMode('list');
      
      const state = useUIStore.getState();
      expect(state.viewMode).toBe('list');
    });

    it('should toggle view mode', () => {
      expect(store.getState().viewMode).toBe('grid');
      
      store.toggleViewMode();
      expect(store.getState().viewMode).toBe('list');
      
      store.toggleViewMode();
      expect(store.getState().viewMode).toBe('grid');
    });

    it('should check if camera is selected', () => {
      expect(store.isCameraSelected()).toBe(false);
      
      store.selectCamera('camera0');
      expect(store.isCameraSelected()).toBe(true);
    });

    it('should get current view mode', () => {
      expect(store.getCurrentViewMode()).toBe('grid');
      
      store.setViewMode('list');
      expect(store.getCurrentViewMode()).toBe('list');
    });
  });

  describe('Theme Management', () => {
    it('should set theme', () => {
      store.setTheme('dark');
      
      const state = useUIStore.getState();
      expect(state.theme).toBe('dark');
    });

    it('should toggle theme', () => {
      expect(store.getState().theme).toBe('light');
      
      store.toggleTheme();
      expect(store.getState().theme).toBe('dark');
      
      store.toggleTheme();
      expect(store.getState().theme).toBe('light');
    });

    it('should check if dark theme is active', () => {
      expect(store.isDarkTheme()).toBe(false);
      
      store.setTheme('dark');
      expect(store.isDarkTheme()).toBe(true);
    });

    it('should get current theme', () => {
      expect(store.getCurrentTheme()).toBe('light');
      
      store.setTheme('dark');
      expect(store.getCurrentTheme()).toBe('dark');
    });
  });

  describe('Layout Management', () => {
    it('should toggle sidebar', () => {
      expect(store.getState().sidebarOpen).toBe(true);
      
      store.toggleSidebar();
      expect(store.getState().sidebarOpen).toBe(false);
      
      store.toggleSidebar();
      expect(store.getState().sidebarOpen).toBe(true);
    });

    it('should set sidebar open state', () => {
      store.setSidebarOpen(false);
      
      const state = useUIStore.getState();
      expect(state.sidebarOpen).toBe(false);
    });

    it('should check if sidebar is open', () => {
      expect(store.isSidebarOpen()).toBe(true);
      
      store.setSidebarOpen(false);
      expect(store.isSidebarOpen()).toBe(false);
    });
  });

  describe('Notification Management', () => {
    it('should add notification', () => {
      const notification: NotificationState = {
        id: 'test-1',
        type: 'info',
        title: 'Test Notification',
        message: 'This is a test notification',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      };

      store.addNotification(notification);
      
      const state = useUIStore.getState();
      expect(state.notifications).toHaveLength(1);
      expect(state.notifications[0]).toEqual(notification);
    });

    it('should remove notification', () => {
      const notification: NotificationState = {
        id: 'test-1',
        type: 'info',
        title: 'Test Notification',
        message: 'This is a test notification',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      };

      store.addNotification(notification);
      store.removeNotification('test-1');
      
      const state = useUIStore.getState();
      expect(state.notifications).toHaveLength(0);
    });

    it('should clear all notifications', () => {
      store.addNotification({
        id: 'test-1',
        type: 'info',
        title: 'Test 1',
        message: 'Message 1',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      store.addNotification({
        id: 'test-2',
        type: 'warning',
        title: 'Test 2',
        message: 'Message 2',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      
      store.clearAllNotifications();
      
      const state = useUIStore.getState();
      expect(state.notifications).toHaveLength(0);
    });

    it('should get notifications by type', () => {
      store.addNotification({
        id: 'info-1',
        type: 'info',
        title: 'Info',
        message: 'Info message',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      store.addNotification({
        id: 'error-1',
        type: 'error',
        title: 'Error',
        message: 'Error message',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      store.addNotification({
        id: 'info-2',
        type: 'info',
        title: 'Info 2',
        message: 'Info message 2',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      
      const infoNotifications = store.getNotificationsByType('info');
      expect(infoNotifications).toHaveLength(2);
      
      const errorNotifications = store.getNotificationsByType('error');
      expect(errorNotifications).toHaveLength(1);
    });

    it('should get notification count', () => {
      expect(store.getNotificationCount()).toBe(0);
      
      store.addNotification({
        id: 'test-1',
        type: 'info',
        title: 'Test',
        message: 'Message',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      store.addNotification({
        id: 'test-2',
        type: 'warning',
        title: 'Test 2',
        message: 'Message 2',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      
      expect(store.getNotificationCount()).toBe(2);
    });

    it('should check if has notifications', () => {
      expect(store.hasNotifications()).toBe(false);
      
      store.addNotification({
        id: 'test-1',
        type: 'info',
        title: 'Test',
        message: 'Message',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      
      expect(store.hasNotifications()).toBe(true);
    });
  });

  describe('Loading State Management', () => {
    it('should set loading state for specific component', () => {
      store.setLoading('cameras', true);
      
      const state = useUIStore.getState();
      expect(state.loading.cameras).toBe(true);
      expect(state.loading.recordings).toBe(false);
    });

    it('should set multiple loading states', () => {
      store.setLoadingStates({
        cameras: true,
        recordings: true,
        snapshots: false,
        system: false
      });
      
      const state = useUIStore.getState();
      expect(state.loading.cameras).toBe(true);
      expect(state.loading.recordings).toBe(true);
      expect(state.loading.snapshots).toBe(false);
      expect(state.loading.system).toBe(false);
    });

    it('should check if any component is loading', () => {
      expect(store.isAnyLoading()).toBe(false);
      
      store.setLoading('cameras', true);
      expect(store.isAnyLoading()).toBe(true);
    });

    it('should check if specific component is loading', () => {
      expect(store.isLoading('cameras')).toBe(false);
      
      store.setLoading('cameras', true);
      expect(store.isLoading('cameras')).toBe(true);
    });

    it('should clear all loading states', () => {
      store.setLoadingStates({
        cameras: true,
        recordings: true,
        snapshots: true,
        system: true
      });
      
      store.clearAllLoading();
      
      const state = useUIStore.getState();
      expect(state.loading.cameras).toBe(false);
      expect(state.loading.recordings).toBe(false);
      expect(state.loading.snapshots).toBe(false);
      expect(state.loading.system).toBe(false);
    });
  });

  describe('Error State Management', () => {
    it('should set error state', () => {
      const error: ErrorState = {
        message: 'Test error',
        code: 1001,
        timestamp: new Date()
      };

      store.setError(error);
      
      const state = useUIStore.getState();
      expect(state.error).toEqual(error);
    });

    it('should clear error state', () => {
      store.setError({
        message: 'Test error',
        code: 1001,
        timestamp: new Date()
      });
      store.clearError();
      
      const state = useUIStore.getState();
      expect(state.error.message).toBeNull();
      expect(state.error.code).toBeNull();
      expect(state.error.timestamp).toBeNull();
    });

    it('should check if has error', () => {
      expect(store.hasError()).toBe(false);
      
      store.setError({
        message: 'Test error',
        code: 1001,
        timestamp: new Date()
      });
      expect(store.hasError()).toBe(true);
    });

    it('should get current error', () => {
      const error: ErrorState = {
        message: 'Test error',
        code: 1001,
        timestamp: new Date()
      };

      store.setError(error);
      expect(store.getCurrentError()).toEqual(error);
    });
  });

  describe('Settings Management', () => {
    it('should set auto refresh', () => {
      store.setAutoRefresh(false);
      
      const state = useUIStore.getState();
      expect(state.autoRefresh).toBe(false);
    });

    it('should set refresh interval', () => {
      store.setRefreshInterval(60000);
      
      const state = useUIStore.getState();
      expect(state.refreshInterval).toBe(60000);
    });

    it('should set show notifications', () => {
      store.setShowNotifications(false);
      
      const state = useUIStore.getState();
      expect(state.showNotifications).toBe(false);
    });

    it('should toggle auto refresh', () => {
      expect(store.getState().autoRefresh).toBe(true);
      
      store.toggleAutoRefresh();
      expect(store.getState().autoRefresh).toBe(false);
      
      store.toggleAutoRefresh();
      expect(store.getState().autoRefresh).toBe(true);
    });

    it('should toggle show notifications', () => {
      expect(store.getState().showNotifications).toBe(true);
      
      store.toggleShowNotifications();
      expect(store.getState().showNotifications).toBe(false);
      
      store.toggleShowNotifications();
      expect(store.getState().showNotifications).toBe(true);
    });
  });

  describe('UI State Analysis', () => {
    it('should get UI state summary', () => {
      store.selectCamera('camera0');
      store.setTheme('dark');
      store.setViewMode('list');
      store.addNotification({
        id: 'test-1',
        type: 'info',
        title: 'Test',
        message: 'Message',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });

      const summary = store.getUIStateSummary();
      
      expect(summary).toHaveProperty('selected_camera', 'camera0');
      expect(summary).toHaveProperty('view_mode', 'list');
      expect(summary).toHaveProperty('theme', 'dark');
      expect(summary).toHaveProperty('sidebar_open', true);
      expect(summary).toHaveProperty('notification_count', 1);
      expect(summary).toHaveProperty('has_loading', false);
      expect(summary).toHaveProperty('has_error', false);
    });

    it('should get navigation state', () => {
      store.selectCamera('camera0');
      store.setViewMode('list');

      const navState = store.getNavigationState();
      
      expect(navState).toEqual({
        selected_camera: 'camera0',
        view_mode: 'list'
      });
    });

    it('should get theme state', () => {
      store.setTheme('dark');

      const themeState = store.getThemeState();
      
      expect(themeState).toEqual({
        theme: 'dark',
        is_dark: true
      });
    });

    it('should get notification summary', () => {
      store.addNotification({
        id: 'info-1',
        type: 'info',
        title: 'Info',
        message: 'Info message',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      store.addNotification({
        id: 'error-1',
        type: 'error',
        title: 'Error',
        message: 'Error message',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });

      const notificationSummary = store.getNotificationSummary();
      
      expect(notificationSummary).toEqual({
        total: 2,
        by_type: {
          info: 1,
          warning: 0,
          error: 1,
          success: 0
        },
        has_notifications: true
      });
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.selectCamera('camera0');
      store.setTheme('dark');
      store.setViewMode('list');
      store.setSidebarOpen(false);
      store.addNotification({
        id: 'test-1',
        type: 'info',
        title: 'Test',
        message: 'Message',
        timestamp: new Date(),
        duration: 5000,
        persistent: false
      });
      store.setLoading('cameras', true);
      store.setError({
        message: 'Test error',
        code: 1001,
        timestamp: new Date()
      });
      store.setAutoRefresh(false);
      
      // Reset
      store.reset();
      
      const state = useUIStore.getState();
      expect(state.selectedCamera).toBeNull();
      expect(state.viewMode).toBe('grid');
      expect(state.theme).toBe('light');
      expect(state.sidebarOpen).toBe(true);
      expect(state.notifications).toEqual([]);
      expect(state.loading).toEqual({
        cameras: false,
        recordings: false,
        snapshots: false,
        system: false
      });
      expect(state.error).toEqual({
        message: null,
        code: null,
        timestamp: null
      });
      expect(state.autoRefresh).toBe(true);
      expect(state.refreshInterval).toBe(30000);
      expect(state.showNotifications).toBe(true);
    });
  });
});
