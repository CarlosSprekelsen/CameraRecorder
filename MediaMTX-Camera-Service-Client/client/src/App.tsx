import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import { CssBaseline } from '@mui/material';
import { theme } from './theme';
import { NotificationProvider } from './components/common/NotificationSystem';
import AppShell from './components/common/AppShell';
import ConnectionManager from './components/common/ConnectionManager';
import Dashboard from './components/Dashboard/Dashboard';
import CameraDetail from './components/CameraDetail/CameraDetail';
import FileManager from './components/FileManager/FileManager';
import HealthMonitor from './components/HealthMonitor/HealthMonitor';
import AdminDashboard from './components/AdminDashboard/AdminDashboard';
import Settings from './components/Settings/Settings';
import ErrorBoundary from './components/common/ErrorBoundary';
import FeatureErrorBoundary from './components/ErrorBoundaries/FeatureErrorBoundary';
import ServiceErrorBoundary from './components/ErrorBoundaries/ServiceErrorBoundary';

const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <NotificationProvider maxNotifications={5}>
          <ServiceErrorBoundary serviceName="ConnectionManager" retryable={true}>
            <ConnectionManager autoConnect={true} showConnectionUI={true}>
              <Router>
                <Routes>
                  <Route path="/" element={<AppShell />}>
                    <Route index element={
                      <FeatureErrorBoundary featureName="Dashboard">
                        <Dashboard />
                      </FeatureErrorBoundary>
                    } />
                    <Route path="camera/:deviceId" element={
                      <FeatureErrorBoundary featureName="CameraDetail">
                        <CameraDetail />
                      </FeatureErrorBoundary>
                    } />
                    <Route path="files" element={
                      <FeatureErrorBoundary featureName="FileManager">
                        <FileManager />
                      </FeatureErrorBoundary>
                    } />
                    <Route path="health" element={
                      <FeatureErrorBoundary featureName="HealthMonitor">
                        <HealthMonitor />
                      </FeatureErrorBoundary>
                    } />
                    <Route path="admin" element={
                      <FeatureErrorBoundary featureName="AdminDashboard">
                        <AdminDashboard />
                      </FeatureErrorBoundary>
                    } />
                    <Route path="settings" element={
                      <FeatureErrorBoundary featureName="Settings">
                        <Settings />
                      </FeatureErrorBoundary>
                    } />
                  </Route>
                </Routes>
              </Router>
            </ConnectionManager>
          </ServiceErrorBoundary>
        </NotificationProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
};

export default App;
