import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import { CssBaseline } from '@mui/material';
import { theme } from './theme';
import AppShell from './components/common/AppShell';
import Dashboard from './components/Dashboard/Dashboard';
import CameraDetail from './components/CameraDetail/CameraDetail';
import ErrorBoundary from './components/common/ErrorBoundary';

const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Router>
          <Routes>
            <Route path="/" element={<AppShell />}>
              <Route index element={<Dashboard />} />
              <Route path="camera/:deviceId" element={<CameraDetail />} />
              {/* TODO: Add Settings route when implemented */}
              <Route path="settings" element={<div>Settings (Coming Soon)</div>} />
            </Route>
          </Routes>
        </Router>
      </ThemeProvider>
    </ErrorBoundary>
  );
};

export default App;
