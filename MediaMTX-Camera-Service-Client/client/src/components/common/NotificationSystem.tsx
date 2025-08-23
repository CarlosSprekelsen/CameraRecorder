/**
 * Notification System Component
 * Provides toast notifications for user feedback
 * Replaces all TODO notification comments
 */

import React, { createContext, useContext, useReducer, useCallback } from 'react';
import {
  Snackbar,
  Alert,
  AlertColor,
  Box,
} from '@mui/material';

/**
 * Notification types
 */
export type NotificationType = 'success' | 'error' | 'warning' | 'info';

/**
 * Notification object
 */
export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message: string;
  duration?: number;
  timestamp: Date;
}

/**
 * Notification state
 */
interface NotificationState {
  notifications: Notification[];
  maxNotifications: number;
}

/**
 * Notification actions
 */
type NotificationAction =
  | { type: 'ADD_NOTIFICATION'; payload: Notification }
  | { type: 'REMOVE_NOTIFICATION'; payload: string }
  | { type: 'CLEAR_ALL' };

/**
 * Notification context
 */
interface NotificationContextType {
  notifications: Notification[];
  showNotification: (type: NotificationType, title: string, message?: string, duration?: number) => void;
  showSuccess: (title: string, message?: string) => void;
  showError: (title: string, message?: string) => void;
  showWarning: (title: string, message?: string) => void;
  showInfo: (title: string, message?: string) => void;
  removeNotification: (id: string) => void;
  clearAll: () => void;
}

/**
 * Notification reducer
 */
const notificationReducer = (state: NotificationState, action: NotificationAction): NotificationState => {
  switch (action.type) {
    case 'ADD_NOTIFICATION':
      const newNotifications = [...state.notifications, action.payload];
      if (newNotifications.length > state.maxNotifications) {
        newNotifications.shift(); // Remove oldest notification
      }
      return {
        ...state,
        notifications: newNotifications,
      };
      
    case 'REMOVE_NOTIFICATION':
      return {
        ...state,
        notifications: state.notifications.filter(n => n.id !== action.payload),
      };
      
    case 'CLEAR_ALL':
      return {
        ...state,
        notifications: [],
      };
      
    default:
      return state;
  }
};

/**
 * Create notification context
 */
const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

/**
 * Notification provider props
 */
interface NotificationProviderProps {
  children: React.ReactNode;
  maxNotifications?: number;
}

/**
 * Notification Provider Component
 */
export const NotificationProvider: React.FC<NotificationProviderProps> = ({
  children,
  maxNotifications = 5,
}) => {
  const [state, dispatch] = useReducer(notificationReducer, {
    notifications: [],
    maxNotifications,
  });

  /**
   * Generate unique notification ID
   */
  const generateId = useCallback(() => {
    return `notification-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }, []);

  /**
   * Show notification
   */
  const showNotification = useCallback((
    type: NotificationType,
    title: string,
    message?: string,
    duration: number = 5000
  ) => {
    const notification: Notification = {
      id: generateId(),
      type,
      title,
      message: message || '',
      duration,
      timestamp: new Date(),
    };

    dispatch({ type: 'ADD_NOTIFICATION', payload: notification });
  }, [generateId]);

  /**
   * Show success notification
   */
  const showSuccess = useCallback((title: string, message?: string) => {
    showNotification('success', title, message);
  }, [showNotification]);

  /**
   * Show error notification
   */
  const showError = useCallback((title: string, message?: string) => {
    showNotification('error', title, message, 8000); // Longer duration for errors
  }, [showNotification]);

  /**
   * Show warning notification
   */
  const showWarning = useCallback((title: string, message?: string) => {
    showNotification('warning', title, message, 6000);
  }, [showNotification]);

  /**
   * Show info notification
   */
  const showInfo = useCallback((title: string, message?: string) => {
    showNotification('info', title, message);
  }, [showNotification]);

  /**
   * Remove notification
   */
  const removeNotification = useCallback((id: string) => {
    dispatch({ type: 'REMOVE_NOTIFICATION', payload: id });
  }, []);

  /**
   * Clear all notifications
   */
  const clearAll = useCallback(() => {
    dispatch({ type: 'CLEAR_ALL' });
  }, []);

  const contextValue: NotificationContextType = {
    notifications: state.notifications,
    showNotification,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    removeNotification,
    clearAll,
  };

  return (
    <NotificationContext.Provider value={contextValue}>
      {children}
      <NotificationDisplay />
    </NotificationContext.Provider>
  );
};

/**
 * Notification Display Component
 */
const NotificationDisplay: React.FC = () => {
  const context = useContext(NotificationContext);
  
  if (!context) {
    return null;
  }

  const { notifications, removeNotification } = context;

  return (
    <Box>
      {notifications.map((notification) => (
        <Snackbar
          key={notification.id}
          open={true}
          autoHideDuration={notification.duration}
          onClose={() => removeNotification(notification.id)}
          anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
          sx={{ mb: 1 }}
        >
          <Alert
            onClose={() => removeNotification(notification.id)}
            severity={notification.type as AlertColor}
            variant="filled"
            sx={{ width: '100%' }}
          >
            <Box>
              <strong>{notification.title}</strong>
              {notification.message && (
                <Box mt={0.5}>
                  {notification.message}
                </Box>
              )}
            </Box>
          </Alert>
        </Snackbar>
      ))}
    </Box>
  );
};

/**
 * Custom hook to use notifications
 */
export const useNotifications = (): NotificationContextType => {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotifications must be used within a NotificationProvider');
  }
  return context;
};

/**
 * Notification utilities for common operations
 */
export const notificationUtils = {
  /**
   * Camera operation notifications
   */
  camera: {
    snapshotTaken: (cameraName: string) => ({
      title: 'Snapshot Captured',
      message: `Successfully captured snapshot from ${cameraName}`,
    }),
    snapshotFailed: (cameraName: string, error?: string) => ({
      title: 'Snapshot Failed',
      message: `Failed to capture snapshot from ${cameraName}${error ? `: ${error}` : ''}`,
    }),
    recordingStarted: (cameraName: string) => ({
      title: 'Recording Started',
      message: `Started recording from ${cameraName}`,
    }),
    recordingStopped: (cameraName: string) => ({
      title: 'Recording Stopped',
      message: `Stopped recording from ${cameraName}`,
    }),
    recordingFailed: (cameraName: string, error?: string) => ({
      title: 'Recording Failed',
      message: `Failed to ${error?.includes('start') ? 'start' : 'stop'} recording from ${cameraName}${error ? `: ${error}` : ''}`,
    }),
  },

  /**
   * File operation notifications
   */
  file: {
    downloadStarted: (filename: string) => ({
      title: 'Download Started',
      message: `Started downloading ${filename}`,
    }),
    downloadCompleted: (filename: string) => ({
      title: 'Download Completed',
      message: `Successfully downloaded ${filename}`,
    }),
    downloadFailed: (filename: string, error?: string) => ({
      title: 'Download Failed',
      message: `Failed to download ${filename}${error ? `: ${error}` : ''}`,
    }),
    fileDeleted: (filename: string) => ({
      title: 'File Deleted',
      message: `Successfully deleted ${filename}`,
    }),
    deleteFailed: (filename: string, error?: string) => ({
      title: 'Delete Failed',
      message: `Failed to delete ${filename}${error ? `: ${error}` : ''}`,
    }),
  },

  /**
   * Connection notifications
   */
  connection: {
    connected: () => ({
      title: 'Connected',
      message: 'Successfully connected to camera service',
    }),
    disconnected: () => ({
      title: 'Disconnected',
      message: 'Disconnected from camera service',
    }),
    connectionFailed: (error?: string) => ({
      title: 'Connection Failed',
      message: `Failed to connect to camera service${error ? `: ${error}` : ''}`,
    }),
  },

  /**
   * Authentication notifications
   */
  auth: {
    loginSuccess: () => ({
      title: 'Login Successful',
      message: 'Successfully authenticated with the service',
    }),
    loginFailed: (error?: string) => ({
      title: 'Login Failed',
      message: `Authentication failed${error ? `: ${error}` : ''}`,
    }),
    logoutSuccess: () => ({
      title: 'Logged Out',
      message: 'Successfully logged out',
    }),
  },
};

export default NotificationProvider;
