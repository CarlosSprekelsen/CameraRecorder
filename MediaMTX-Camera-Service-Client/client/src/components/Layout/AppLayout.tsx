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
} from '@mui/material';
import {
  Menu as MenuIcon,
  AccountCircle,
  Logout,
} from '@mui/icons-material';
import { AuthService } from '../../services/auth/AuthService';
import { useConnectionStore } from '../../stores/connection/connectionStore';
import { useAuthStore } from '../../stores/auth/authStore';
import { useServerStore } from '../../stores/server/serverStore';

interface AppLayoutProps {
  children: React.ReactNode;
  authService: AuthService;
}

const AppLayout: React.FC<AppLayoutProps> = ({ 
  children, 
  authService
}) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  
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
      case 'connected': return 'success';
      case 'connecting': return 'warning';
      case 'disconnected': return 'error';
      case 'error': return 'error';
      default: return 'default';
    }
  };

  const getConnectionStatusText = () => {
    switch (connectionStatus) {
      case 'connected': return 'Connected';
      case 'connecting': return 'Connecting...';
      case 'disconnected': return 'Disconnected';
      case 'error': return 'Connection Error';
      default: return 'Unknown';
    }
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin': return 'error';
      case 'operator': return 'warning';
      case 'viewer': return 'info';
      default: return 'default';
    }
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar position="static">
        <Toolbar>
          <IconButton
            size="large"
            edge="start"
            color="inherit"
            aria-label="menu"
            sx={{ mr: 2 }}
          >
            <MenuIcon />
          </IconButton>
          
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            MediaMTX Camera Service
          </Typography>

          {/* Connection Status */}
          <Chip
            label={getConnectionStatusText()}
            color={getConnectionStatusColor() as any}
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
              color={getRoleColor(role || '') as any}
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
