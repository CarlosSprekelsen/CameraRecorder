import { useEffect, useState, useCallback } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Box } from '@mui/material';

import { useConnectionStore } from './stores/connection/connectionStore';
import { useAuthStore } from './stores/auth/authStore';
import { useServerStore } from './stores/server/serverStore';
import { serviceFactory } from './services/ServiceFactory';
import { logger } from './services/logger/LoggerService';

import AppLayout from './components/Layout/AppLayout';
import LoginPage from './pages/Login/LoginPage';
import AboutPage from './pages/About/AboutPage';
import CameraPage from './pages/Cameras/CameraPage';
import FilesPage from './pages/Files/FilesPage';
import LoadingSpinner from './components/Layout/LoadingSpinner';
import ErrorBoundary from './components/Error/ErrorBoundary';
import { AccessibilityProvider } from './components/Accessibility/AccessibilityProvider';
import { usePerformanceMonitor } from './hooks/usePerformanceMonitor';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';

// Create theme
const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

// WebSocket configuration
const WS_URL = (import.meta as any).env?.VITE_WS_URL || 'ws://localhost:8002/ws';

function App() {
  const [wsService] = useState(() => serviceFactory.createWebSocketService(WS_URL));
  const [authService] = useState(() => serviceFactory.createAuthService(wsService));
  const [serverService] = useState(() => serviceFactory.createServerService(wsService));
  // Notification service will be used in future sprints
  // const [notificationService] = useState(() => serviceFactory.createNotificationService(wsService));
  const [isInitialized, setIsInitialized] = useState(false);

  // Initialize performance monitoring and keyboard shortcuts
  usePerformanceMonitor();
  useKeyboardShortcuts();

  const {
    status: connectionStatus,
    setStatus: setConnectionStatus,
    setError: setConnectionError,
  } = useConnectionStore();
  const { isAuthenticated, login } = useAuthStore();
  const { setInfo, setStatus, setStorage, setLoading, setError } = useServerStore();

  // Memoized WebSocket event handlers for performance optimization
  const handleWebSocketConnect = useCallback(() => {
    setConnectionStatus('connected');
    setConnectionError(null);
    logger.info('WebSocket connected successfully');
  }, [setConnectionStatus, setConnectionError]);

  const handleWebSocketDisconnect = useCallback(
    (error?: Error) => {
      setConnectionStatus('disconnected');
      if (error) {
        setConnectionError(error.message);
        logger.warn('WebSocket disconnected', { error: error.message });
      }
    },
    [setConnectionStatus, setConnectionError],
  );

  const handleWebSocketError = useCallback(
    (error: Error) => {
      setConnectionStatus('error');
      setConnectionError(error.message);
      logger.error('WebSocket error', { error: error.message }, error);
    },
    [setConnectionStatus, setConnectionError],
  );

  // Initialize WebSocket connection
  useEffect(() => {
    const initializeConnection = async () => {
      try {
        setConnectionStatus('connecting');
        logger.info('Initializing WebSocket connection', { url: WS_URL });

        // Set up WebSocket event handlers with memoized callbacks
        wsService.events = {
          onConnect: handleWebSocketConnect,
          onDisconnect: handleWebSocketDisconnect,
          onError: handleWebSocketError,
        };

        await wsService.connect();

        // Try to restore authentication from session storage
        if (authService.isAuthenticated()) {
          const session = authService.getStoredSession();
          if (session) {
            login(
              authService.getStoredToken()!,
              session.role,
              session.session_id,
              session.expires_at,
              session.permissions,
            );
          }
        }

        setIsInitialized(true);
        logger.info('Application initialized successfully');
      } catch (error) {
        logger.error('Failed to initialize connection', { error }, error as Error);
        setConnectionStatus('error');
        setConnectionError(error instanceof Error ? error.message : 'Connection failed');
        setIsInitialized(true);
      }
    };

    initializeConnection();

    return () => {
      logger.info('Cleaning up WebSocket connection');
      wsService.disconnect();
    };
  }, [wsService, authService, login, setConnectionStatus, setConnectionError]);

  // Load server info when connected and authenticated
  useEffect(() => {
    if (connectionStatus === 'connected' && isAuthenticated && isInitialized) {
      const loadServerData = async () => {
        try {
          setLoading(true);
          setError(null);

          const [info, status, storage] = await Promise.all([
            serverService.getServerInfo(),
            serverService.getStatus(),
            serverService.getStorageInfo(),
          ]);

          setInfo(info);
          setStatus(status);
          setStorage(storage);
        } catch (error) {
          logger.error('Failed to load server data', { error }, error as Error);
          setError(error instanceof Error ? error.message : 'Failed to load server data');
        } finally {
          setLoading(false);
        }
      };

      loadServerData();
    }
  }, [
    connectionStatus,
    isAuthenticated,
    isInitialized,
    serverService,
    setInfo,
    setStatus,
    setStorage,
    setLoading,
    setError,
  ]);

  if (!isInitialized) {
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh">
          <LoadingSpinner />
        </Box>
      </ThemeProvider>
    );
  }

  return (
    <AccessibilityProvider>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <ErrorBoundary>
          <BrowserRouter>
            <Routes>
              <Route
                path="/login"
                element={
                  isAuthenticated ? (
                    <Navigate to="/about" replace />
                  ) : (
                    <LoginPage authService={authService} />
                  )
                }
              />
              <Route
                path="/*"
                element={
                  isAuthenticated ? (
                    <AppLayout authService={authService}>
                      <Routes>
                        <Route path="/" element={<Navigate to="/cameras" replace />} />
                        <Route path="/cameras" element={<CameraPage />} />
                        <Route path="/files" element={<FilesPage />} />
                        <Route path="/about" element={<AboutPage />} />
                        <Route path="*" element={<Navigate to="/cameras" replace />} />
                      </Routes>
                    </AppLayout>
                  ) : (
                    <Navigate to="/login" replace />
                  )
                }
              />
            </Routes>
          </BrowserRouter>
        </ErrorBoundary>
      </ThemeProvider>
    </AccessibilityProvider>
  );
}

export default App;
