import { useEffect, useState, lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Box } from './components/atoms/Box/Box';

import { useConnectionStore } from './stores/connection/connectionStore';
import { useAuthStore } from './stores/auth/authStore';
import { useServerStore } from './stores/server/serverStore';
import { useDeviceStore } from './stores/device/deviceStore';
import { useFileStore } from './stores/file/fileStore';
import { useRecordingStore } from './stores/recording/recordingStore';
import { useStreamingStore } from './stores/streaming/streamingStore';
import { WebSocketService } from './services/websocket/WebSocketService';
import { APIClient } from './services/abstraction/APIClient';
import { ServiceFactory } from './services/ServiceFactory';
import { logger } from './services/logger/LoggerService';

import AppLayout from './components/Layout/AppLayout';
import LoginPage from './pages/Login/LoginPage';
import AboutPage from './pages/About/AboutPage';
// Lazy load heavy components for code splitting
const CameraPage = lazy(() => import('./pages/Cameras/CameraPage'));
const FilesPage = lazy(() => import('./pages/Files/FilesPage'));
const AdminPage = lazy(() => import('./pages/Admin/AdminPage'));
import LoadingSpinner from './components/Layout/LoadingSpinner';
import ErrorBoundary from './components/Error/ErrorBoundary';
import { AccessibilityProvider } from './components/Accessibility/AccessibilityProvider';
import { usePerformanceMonitor } from './hooks/usePerformanceMonitor';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';

// Global CSS styles for atomic design components
const globalStyles = `
  * {
    box-sizing: border-box;
  }
  
  body {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
      'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
      sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
  
  #root {
    min-height: 100vh;
  }
`;

// WebSocket configuration
const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8002/ws';

function App(): React.JSX.Element {
  // ARCHITECTURE FIX: Services created via dependency injection in stores
  const [isInitialized, setIsInitialized] = useState(false);

  // Initialize performance monitoring and keyboard shortcuts
  usePerformanceMonitor();
  useKeyboardShortcuts();

  // ARCHITECTURE FIX: Service injection removed - services are managed by ServiceFactory

  // ARCHITECTURE FIX: Initialize all services and inject into stores
  useEffect(() => {
    if (!isInitialized) {
      console.log('Initializing services for real-time notifications');
      
      try {
        // Create WebSocket service
        const wsService = new WebSocketService({ url: WS_URL });
        
        // Create APIClient
        const apiClient = new APIClient(wsService, logger);
        
        // Create services using ServiceFactory
        const serviceFactory = ServiceFactory.getInstance();
        const authService = serviceFactory.createAuthService(apiClient);
        const deviceService = serviceFactory.createDeviceService(apiClient);
        const recordingService = serviceFactory.createRecordingService(apiClient);
        const fileService = serviceFactory.createFileService(apiClient);
        const streamingService = serviceFactory.createStreamingService(apiClient);
        const serverService = serviceFactory.createServerService(apiClient);
        
        // Inject services into stores
        useConnectionStore.getState().setWebSocketService(wsService);
        useAuthStore.getState().setAuthService(authService);
        useDeviceStore.getState().setDeviceService(deviceService);
        useRecordingStore.getState().setRecordingService(recordingService);
        useFileStore.getState().setFileService(fileService);
        useStreamingStore.getState().setStreamingService(streamingService);
        useServerStore.getState().setServerService(serverService);
        
        console.log('All services initialized and injected into stores');
        setIsInitialized(true);
      } catch (error) {
        console.error('Failed to initialize services', error);
        setConnectionStatus('error');
        setConnectionError(error instanceof Error ? error.message : 'Service initialization failed');
        setIsInitialized(true);
      }
    }
  }, [isInitialized, setConnectionStatus, setConnectionError]);

  const {
    status: connectionStatus,
    setStatus: setConnectionStatus,
    setError: setConnectionError,
  } = useConnectionStore();
  const { isAuthenticated } = useAuthStore();
  const { loadAllServerData } = useServerStore();
  
  // ARCHITECTURE FIX: Store hooks moved after service injection useEffect

  // ARCHITECTURE FIX: WebSocket handlers managed by connection store

  // ARCHITECTURE FIX: Connection managed by connection store with WebSocket service
  useEffect(() => {
    const initializeConnection = async () => {
      try {
        setConnectionStatus('connecting');
        console.log('Initializing connection', { url: WS_URL });

        // Connect using WebSocket service
        await useConnectionStore.getState().connect();
        
        console.log('Application initialized successfully with real-time notifications');
      } catch (error) {
        console.error('Failed to initialize connection', error);
        setConnectionStatus('error');
        setConnectionError(error instanceof Error ? error.message : 'Connection failed');
      }
    };

    if (isInitialized) {
      initializeConnection();
    }
  }, [isInitialized, setConnectionStatus, setConnectionError]);

  // Load server info when connected and authenticated
  useEffect(() => {
    if (connectionStatus === 'connected' && isAuthenticated && isInitialized) {
      loadAllServerData();
    }
  }, [connectionStatus, isAuthenticated, isInitialized, loadAllServerData]);

  // ARCHITECTURE FIX: Explicitly subscribe to events after authentication
  useEffect(() => {
    if (connectionStatus === 'connected' && isAuthenticated && isInitialized) {
      const subscribeToEvents = async () => {
        try {
          const { subscribeEvents } = useServerStore.getState();
          await subscribeEvents([
            'camera_status_update',
            'recording_status_update',
            'system_health_update'
          ]);
          console.log('Successfully subscribed to real-time events');
        } catch (error) {
          console.error('Failed to subscribe to events:', error);
        }
      };
      
      subscribeToEvents();
    }
  }, [connectionStatus, isAuthenticated, isInitialized]);

  if (!isInitialized) {
    return (
      <>
        <style>{globalStyles}</style>
        <Box className="flex justify-center items-center min-h-screen">
          <LoadingSpinner />
        </Box>
      </>
    );
  }

  return (
    <>
      <style>{globalStyles}</style>
      <AccessibilityProvider>
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
                          <Route path="/admin" element={<AdminPage />} />
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
      </AccessibilityProvider>
    </>
  );
}

export default App;
