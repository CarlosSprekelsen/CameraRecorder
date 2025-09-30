/**
 * @fileoverview AppLayout component for main application shell
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React from 'react';
import { Box } from '../atoms/Box/Box';
import { Typography } from '../atoms/Typography/Typography';
import { Chip } from '../atoms/Chip/Chip';
import { Menu, MenuItem } from '../atoms/Menu/Menu';
import { Button } from '../atoms/Button/Button';
import { Icon } from '../atoms/Icon/Icon';
import { IconButton } from '../atoms/IconButton/IconButton';
import { AppBar, Toolbar } from '../atoms/AppBar/AppBar';
import { useNavigate, useLocation } from 'react-router-dom';
// ARCHITECTURE FIX: Removed direct service import - components must use stores only
import { useConnectionStore } from '../../stores/connection/connectionStore';
import { useAuthStore } from '../../stores/auth/authStore';
import { useServerStore } from '../../stores/server/serverStore';

interface AppLayoutProps {
  children: React.ReactNode;
  // ARCHITECTURE FIX: Removed service props - components only use stores
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
 * <AppLayout>
 *   <Routes>
 *     <Route path="/cameras" element={<CameraPage />} />
 *   </Routes>
 * </AppLayout>
 * ```
 *
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
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
          <IconButton size="large" color="inherit" className="mr-2">
            <Icon name="menu" size={20} />
          </IconButton>

          <Typography variant="h6" component="div" className="flex-grow">
            MediaMTX Camera Service
          </Typography>

          {/* Navigation */}
          <Box className="mr-2">
            <Button
              variant="secondary"
              onClick={() => navigate('/cameras')}
              className={`mr-1 ${location.pathname === '/cameras' ? 'bg-white bg-opacity-10' : 'bg-transparent'}`}
            >
              <Icon name="camera" size={16} />
              Cameras
            </Button>
            <Button
              variant="secondary"
              onClick={() => navigate('/files')}
              className={`mr-1 ${location.pathname === '/files' ? 'bg-white bg-opacity-10' : 'bg-transparent'}`}
            >
              <Icon name="folder" size={16} />
              Files
            </Button>
            <Button
              variant="secondary"
              onClick={() => navigate('/about')}
              className={location.pathname === '/about' ? 'bg-white bg-opacity-10' : 'bg-transparent'}
            >
              <Icon name="info" size={16} />
              About
            </Button>
          </Box>

          {/* Connection Status */}
          <Chip
            label={getConnectionStatusText()}
            color={getConnectionStatusColor() as 'success' | 'error' | 'warning' | 'primary'}
            size="small"
            className="mr-2"
          />

          {/* Server Info */}
          {info && (
            <Typography variant="body2" className="mr-2">
              {info.name} v{info.version}
            </Typography>
          )}

          {/* User Menu */}
          <Box className="flex items-center">
            <Chip
              label={role?.toUpperCase() || 'UNKNOWN'}
              color={getRoleColor(role || '') as 'primary' | 'secondary' | 'success' | 'warning' | 'error'}
              size="small"
              className="mr-1"
            />

            <IconButton
              size="large"
              aria-label="account of current user"
              aria-controls="menu-appbar"
              aria-haspopup="true"
              onClick={handleMenuOpen}
              color="default"
            >
              <Icon name="user" size={20} />
            </IconButton>

            <Menu
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={handleMenuClose}
            >
              <MenuItem onClick={handleLogout}>
                <Icon name="logout" size={16} className="mr-1" />
                Logout
              </MenuItem>
            </Menu>
          </Box>
        </Toolbar>
      </AppBar>

      <Box component="main" sx={{ flexGrow: 1, padding: 0 }}>
        {children}
      </Box>
    </Box>
  );
};

export default AppLayout;
