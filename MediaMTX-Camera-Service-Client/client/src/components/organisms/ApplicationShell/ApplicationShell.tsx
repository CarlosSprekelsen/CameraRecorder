/**
 * ApplicationShell - Architecture Compliance
 * 
 * Architecture requirement: "ApplicationShell component" (Section 5.2)
 * Main application shell providing navigation and layout structure
 */

import React, { useState, useEffect } from 'react';
import { 
  Box, 
  AppBar, 
  Toolbar, 
  Typography, 
  Button, 
  IconButton,
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Divider
} from '@mui/material';
import { 
  Menu as MenuIcon,
  CameraAlt,
  Folder,
  Info,
  Logout,
  Dashboard
} from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth/authStore';
import { logger } from '../../../services/logger/LoggerService';
// ARCHITECTURE FIX: Logger is infrastructure - components can import it directly

interface ApplicationShellProps {
  children: React.ReactNode;
  // ARCHITECTURE FIX: Removed service props - components only use stores
}

const navigationItems = [
  { path: '/', label: 'Dashboard', icon: <Dashboard /> },
  { path: '/cameras', label: 'Cameras', icon: <CameraAlt /> },
  { path: '/files', label: 'Files', icon: <Folder /> },
  { path: '/about', label: 'About', icon: <Info /> },
];

export const ApplicationShell: React.FC<ApplicationShellProps> = ({ 
  children 
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { role, logout } = useAuthStore();
  // ARCHITECTURE FIX: Use correct auth store for all auth-related data

  useEffect(() => {
    logger.info('ApplicationShell initialized');
  }, [logger]);

  const handleNavigation = (path: string) => {
    navigate(path);
    setDrawerOpen(false);
    logger.info(`Navigation to: ${path}`);
  };

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
      logger.info('User logged out');
    } catch (err) {
      logger.error('Logout failed:', { error: err });
    }
  };

  const toggleDrawer = () => {
    setDrawerOpen(!drawerOpen);
  };

  return (
    <Box sx={{ display: 'flex', height: '100vh' }}>
      {/* App Bar */}
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <IconButton
            color="inherit"
            aria-label="open drawer"
            onClick={toggleDrawer}
            edge="start"
            sx={{ mr: 2 }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            MediaMTX Camera Service
          </Typography>
          <Typography variant="body2" sx={{ mr: 2 }}>
            {role || 'Guest'}
          </Typography>
          <Button color="inherit" onClick={handleLogout} startIcon={<Logout />}>
            Logout
          </Button>
        </Toolbar>
      </AppBar>

      {/* Navigation Drawer */}
      <Drawer
        variant="temporary"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        sx={{
          width: 240,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: 240,
            boxSizing: 'border-box',
          },
        }}
      >
        <Toolbar />
        <Box sx={{ overflow: 'auto' }}>
          <List>
            {navigationItems.map((item) => (
              <ListItem 
                key={item.path}
                onClick={() => handleNavigation(item.path)}
                sx={{ 
                  cursor: 'pointer',
                  backgroundColor: location.pathname === item.path ? 'action.selected' : 'transparent'
                }}
              >
                <ListItemIcon>
                  {item.icon}
                </ListItemIcon>
                <ListItemText primary={item.label} />
              </ListItem>
            ))}
          </List>
          <Divider />
          <List>
            <ListItem>
              <ListItemText 
                primary="Version" 
                secondary="1.0.0"
                secondaryTypographyProps={{ variant: 'caption' }}
              />
            </ListItem>
          </List>
        </Box>
      </Drawer>

      {/* Main Content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - 240px)` },
          ml: { sm: '240px' },
        }}
      >
        <Toolbar />
        {children}
      </Box>
    </Box>
  );
};
