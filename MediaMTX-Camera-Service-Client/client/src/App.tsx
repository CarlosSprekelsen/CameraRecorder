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

const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <NotificationProvider maxNotifications={5}>
          <ConnectionManager autoConnect={true} showConnectionUI={true}>
            <Router>
              <Routes>
                <Route path="/" element={<AppShell />}>
                  <Route index element={<Dashboard />} />
                  <Route path="camera/:deviceId" element={<CameraDetail />} />
                  <Route path="files" element={<FileManager />} />
                  <Route path="health" element={<HealthMonitor />} />
                  <Route path="admin" element={<AdminDashboard />} />
                  <Route path="settings" element={<Settings />} />
                </Route>
              </Routes>
            </Router>
          </ConnectionManager>
        </NotificationProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
};

export default App;
