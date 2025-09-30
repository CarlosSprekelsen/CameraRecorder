/**
 * ApplicationShell - Architecture Compliance
 * 
 * Architecture requirement: "ApplicationShell component" (Section 5.2)
 * Main application shell providing navigation and layout structure
 */

import React, { useState, useEffect } from 'react';
import { Box } from '../../atoms/Box/Box';
import { AppBar, Toolbar } from '../../atoms/AppBar/AppBar';
import { Typography } from '../../atoms/Typography/Typography';
import { Button } from '../../atoms/Button/Button';
import { IconButton } from '../../atoms/IconButton/IconButton';
import { Icon } from '../../atoms/Icon/Icon';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth/authStore';
import { usePermissions } from '../../../hooks/usePermissions';
import { logger } from '../../../services/logger/LoggerService';
// ARCHITECTURE FIX: Logger is infrastructure - components can import it directly

interface ApplicationShellProps {
  children: React.ReactNode;
  // ARCHITECTURE FIX: Removed service props - components only use stores
}

const getNavigationItems = (canViewAdminPanel: boolean) => {
  const baseItems = [
    { path: '/', label: 'Dashboard', icon: <Icon name="dashboard" size={20} /> },
    { path: '/cameras', label: 'Cameras', icon: <Icon name="camera" size={20} /> },
    { path: '/files', label: 'Files', icon: <Icon name="folder" size={20} /> },
    { path: '/about', label: 'About', icon: <Icon name="info" size={20} /> },
  ];

  if (canViewAdminPanel) {
    baseItems.splice(3, 0, { path: '/admin', label: 'Admin', icon: <Icon name="admin" size={20} /> });
  }

  return baseItems;
};

export const ApplicationShell: React.FC<ApplicationShellProps> = ({ 
  children 
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { role, logout } = useAuthStore();
  const { canViewAdminPanel } = usePermissions();
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
      const errorRecord = {
        message: err instanceof Error ? err.message : String(err),
        stack: err instanceof Error ? err.stack : undefined,
        name: err instanceof Error ? err.name : 'UnknownError'
      };
      logger.error('Logout failed:', errorRecord);
    }
  };

  const toggleDrawer = () => {
    setDrawerOpen(!drawerOpen);
  };

  return (
    <Box sx={{ display: 'flex', height: '100vh' }}>
      {/* App Bar */}
      <AppBar>
        <Toolbar>
          <IconButton
            color="default"
            aria-label="open drawer"
            onClick={toggleDrawer}
            className="mr-2"
          >
            <Icon name="menu" size={20} />
          </IconButton>
          <Typography variant="h6" component="div" className="flex-grow">
            MediaMTX Camera Service
          </Typography>
          <Typography variant="body2" className="mr-2">
            {role || 'Guest'}
          </Typography>
          <Button variant="secondary" onClick={handleLogout}>
            <Icon name="logout" size={16} className="mr-2" />
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
            {getNavigationItems(canViewAdminPanel()).map((item) => (
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
