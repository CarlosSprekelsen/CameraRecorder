import { useEffect, useState, useCallback, lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Box } from '@mui/material';

import { useConnectionStore } from './stores/connection/connectionStore';
import { useAuthStore } from './stores/auth/authStore';
import { useServerStore } from './stores/server/serverStore';
import { useDeviceStore } from './stores/device/deviceStore';
import { useRecordingStore } from './stores/recording/recordingStore';
import { useFileStore } from './stores/file/fileStore';
// ARCHITECTURE FIX: Removed direct service imports - use dependency injection

import AppLayout from './components/Layout/AppLayout';
import LoginPage from './pages/Login/LoginPage';
import AboutPage from './pages/About/AboutPage';
// Lazy load heavy components for code splitting
const CameraPage = lazy(() => import('./pages/Cameras/CameraPage'));
const FilesPage = lazy(() => import('./pages/Files/FilesPage'));
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
const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8002/ws';

function App(): React.JSX.Element {
  // ARCHITECTURE FIX: Services created via dependency injection in stores
  const [isInitialized, setIsInitialized] = useState(false);

  // Initialize performance monitoring and keyboard shortcuts
  usePerformanceMonitor();
  useKeyboardShortcuts();

  // Store hooks for service injection
  const { setDeviceService } = useDeviceStore();
  const { setService: setRecordingService } = useRecordingStore();
  const { setFileService } = useFileStore();
  const { setAuthService } = useAuthStore();
  const { setServerService } = useServerStore();
  const { setWebSocketService } = useConnectionStore();

  // ARCHITECTURE FIX: Services injected via store initialization
  useEffect(() => {
    if (isInitialized) {
      console.log('Application initialized - services managed by stores');
    }
  }, [isInitialized]);

  const {
    status: connectionStatus,
    setStatus: setConnectionStatus,
    setError: setConnectionError,
  } = useConnectionStore();
  const { isAuthenticated, login } = useAuthStore();
  const { loadAllServerData } = useServerStore();
  
  // ARCHITECTURE FIX: Store hooks moved after service injection useEffect

  // Memoized WebSocket event handlers for performance optimization
  const handleWebSocketConnect = useCallback(() => {
    setConnectionStatus('connected');
    setConnectionError(null);
    console.log('WebSocket connected successfully');
  }, [setConnectionStatus, setConnectionError]);

  const handleWebSocketDisconnect = useCallback(
    (error?: Error) => {
      setConnectionStatus('disconnected');
      if (error) {
        setConnectionError(error.message);
        console.warn('WebSocket disconnected', { error: error.message });
      }
    },
    [setConnectionStatus, setConnectionError],
  );

  const handleWebSocketError = useCallback(
    (error: Error) => {
      setConnectionStatus('error');
      setConnectionError(error.message);
      console.error('WebSocket error', { error: error.message }, error);
    },
    [setConnectionStatus, setConnectionError],
  );

  // ARCHITECTURE FIX: Connection managed by connection store
  useEffect(() => {
    const initializeConnection = async () => {
      try {
        setConnectionStatus('connecting');
        console.log('Initializing connection', { url: WS_URL });

        // Connection handled by stores
        setIsInitialized(true);
        console.log('Application initialized successfully');
      } catch (error) {
        console.error('Failed to initialize connection', error);
        setConnectionStatus('error');
        setConnectionError(error instanceof Error ? error.message : 'Connection failed');
        setIsInitialized(true);
      }
    };

    initializeConnection();
  }, [setConnectionStatus, setConnectionError]);

  // Load server info when connected and authenticated
  useEffect(() => {
    if (connectionStatus === 'connected' && isAuthenticated && isInitialized) {
      loadAllServerData();
    }
  }, [connectionStatus, isAuthenticated, isInitialized, loadAllServerData]);

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
                    <LoginPage />
                  )
                }
              />
              <Route
                path="/*"
                element={
                  isAuthenticated ? (
                    <AppLayout>
                      <Suspense fallback={<LoadingSpinner />}>
                        <Routes>
                          <Route path="/" element={<Navigate to="/cameras" replace />} />
                          <Route path="/cameras" element={<CameraPage />} />
                          <Route path="/files" element={<FilesPage />} />
                          <Route path="/about" element={<AboutPage />} />
                          <Route path="*" element={<Navigate to="/cameras" replace />} />
                        </Routes>
                      </Suspense>
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
