/**
 * @fileoverview AppLayout component for main application shell
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React from 'react';
import {
  Box,
  AppBar,
  Toolbar,
  Typography,
  IconButton,
  Chip,
  Menu,
  MenuItem,
  Button,
} from '@mui/material';
import {
  Menu as MenuIcon,
  AccountCircle,
  Logout,
  Videocam as CameraIcon,
  Folder as FilesIcon,
  Info as InfoIcon,
} from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import { AuthService } from '../../services/auth/AuthService';
import { useConnectionStore } from '../../stores/connection/connectionStore';
import { useAuthStore } from '../../stores/auth/authStore';
import { useServerStore } from '../../stores/server/serverStore';

interface AppLayoutProps {
  children: React.ReactNode;
  authService: AuthService;
}

/**
 * AppLayout - Main application shell component
 * 
 * Provides the main application layout with navigation, user menu, connection status,
 * and role-based access control. Includes responsive design with drawer navigation
 * and real-time connection status indicators.
 * 
 * @component
 * @param {AppLayoutProps} props - Component props
 * @param {React.ReactNode} props.children - Child components to render
 * @param {AuthService} props.authService - Authentication service instance
 * @returns {JSX.Element} The application layout component
 * 
 * @features
 * - Responsive navigation with drawer
 * - User authentication and role display
 * - Connection status monitoring
 * - Role-based menu items
 * - Server information display
 * - Logout functionality
 * 
 * @example
 * ```tsx
 * <AppLayout authService={authService}>
 *   <Routes>
 *     <Route path="/cameras" element={<CameraPage />} />
 *   </Routes>
 * </AppLayout>
 * ```
 * 
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
const AppLayout: React.FC<AppLayoutProps> = ({ children, authService }) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const navigate = useNavigate();
  const location = useLocation();

  const { status: connectionStatus } = useConnectionStore();
  const { role, logout } = useAuthStore();
  const { info } = useServerStore();

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    authService.logout();
    logout();
    handleMenuClose();
  };

  const getConnectionStatusColor = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'success';
      case 'connecting':
        return 'warning';
      case 'disconnected':
        return 'error';
      case 'error':
        return 'error';
      default:
        return 'default';
    }
  };

  const getConnectionStatusText = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'Connected';
      case 'connecting':
        return 'Connecting...';
      case 'disconnected':
        return 'Disconnected';
      case 'error':
        return 'Connection Error';
      default:
        return 'Unknown';
    }
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin':
        return 'error';
      case 'operator':
        return 'warning';
      case 'viewer':
        return 'info';
      default:
        return 'default';
    }
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar position="static">
        <Toolbar>
          <IconButton size="large" edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }}>
            <MenuIcon />
          </IconButton>

          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            MediaMTX Camera Service
          </Typography>

          {/* Navigation */}
          <Box sx={{ mr: 2 }}>
            <Button
              color="inherit"
              startIcon={<CameraIcon />}
              onClick={() => navigate('/cameras')}
              sx={{
                backgroundColor:
                  location.pathname === '/cameras' ? 'rgba(255,255,255,0.1)' : 'transparent',
                mr: 1,
              }}
            >
              Cameras
            </Button>
            <Button
              color="inherit"
              startIcon={<FilesIcon />}
              onClick={() => navigate('/files')}
              sx={{
                backgroundColor:
                  location.pathname === '/files' ? 'rgba(255,255,255,0.1)' : 'transparent',
                mr: 1,
              }}
            >
              Files
            </Button>
            <Button
              color="inherit"
              startIcon={<InfoIcon />}
              onClick={() => navigate('/about')}
              sx={{
                backgroundColor:
                  location.pathname === '/about' ? 'rgba(255,255,255,0.1)' : 'transparent',
              }}
            >
              About
            </Button>
          </Box>

          {/* Connection Status */}
          <Chip
            label={getConnectionStatusText()}
            color={getConnectionStatusColor() as 'success' | 'error' | 'warning' | 'info'}
            size="small"
            sx={{ mr: 2 }}
          />

          {/* Server Info */}
          {info && (
            <Typography variant="body2" sx={{ mr: 2 }}>
              {info.name} v{info.version}
            </Typography>
          )}

          {/* User Menu */}
          <Box display="flex" alignItems="center">
            <Chip
              label={role?.toUpperCase() || 'UNKNOWN'}
              color={getRoleColor(role || '') as 'success' | 'error' | 'warning' | 'info'}
              size="small"
              sx={{ mr: 1 }}
            />

            <IconButton
              size="large"
              aria-label="account of current user"
              aria-controls="menu-appbar"
              aria-haspopup="true"
              onClick={handleMenuOpen}
              color="inherit"
            >
              <AccountCircle />
            </IconButton>

            <Menu
              id="menu-appbar"
              anchorEl={anchorEl}
              anchorOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              keepMounted
              transformOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              open={Boolean(anchorEl)}
              onClose={handleMenuClose}
            >
              <MenuItem onClick={handleLogout}>
                <Logout sx={{ mr: 1 }} />
                Logout
              </MenuItem>
            </Menu>
          </Box>
        </Toolbar>
      </AppBar>

      <Box component="main" sx={{ flexGrow: 1, p: 0 }}>
        {children}
      </Box>
    </Box>
  );
};

export default AppLayout;
