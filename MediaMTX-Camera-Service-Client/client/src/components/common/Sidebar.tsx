import React from 'react';
import {
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Divider,
  Typography,
  Box,
  useTheme,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Settings as SettingsIcon,
  Videocam as CameraIcon,
  Folder as FolderIcon,
  MonitorHeart as HealthIcon,
  AdminPanelSettings as AdminIcon,
} from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';

interface SidebarProps {
  open: boolean;
  onClose: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ open, onClose }) => {
  const theme = useTheme();
  const navigate = useNavigate();
  const location = useLocation();

  const drawerWidth = 240;

  const handleNavigation = (path: string) => {
    navigate(path);
    if (window.innerWidth < theme.breakpoints.values.md) {
      onClose();
    }
  };

  const isActive = (path: string) => location.pathname === path;

  return (
    <Drawer
      variant="persistent"
      anchor="left"
      open={open}
      sx={{
        width: drawerWidth,
        flexShrink: 0,
        '& .MuiDrawer-paper': {
          width: drawerWidth,
          boxSizing: 'border-box',
          mt: 8, // Account for header height
        },
      }}
    >
      <Box sx={{ overflow: 'auto' }}>
        <List>
          <ListItem disablePadding>
            <ListItemButton
              onClick={() => handleNavigation('/')}
              selected={isActive('/')}
            >
              <ListItemIcon>
                <DashboardIcon />
              </ListItemIcon>
              <ListItemText primary="Dashboard" />
            </ListItemButton>
          </ListItem>
        </List>
        
        <Divider />
        
        <Box sx={{ p: 2 }}>
          <Typography variant="subtitle2" color="text.secondary">
            Cameras (0)
          </Typography>
        </Box>
        
        <List>
          <ListItem disablePadding>
            <ListItemButton disabled>
              <ListItemIcon>
                <CameraIcon />
              </ListItemIcon>
              <ListItemText 
                primary="No cameras connected"
                secondary="Connect a camera to get started"
              />
            </ListItemButton>
          </ListItem>
        </List>
        
        <Divider />
        
        <List>
          <ListItem disablePadding>
            <ListItemButton
              onClick={() => handleNavigation('/files')}
              selected={isActive('/files')}
            >
              <ListItemIcon>
                <FolderIcon />
              </ListItemIcon>
              <ListItemText primary="File Manager" />
            </ListItemButton>
          </ListItem>
        </List>
        
        <Divider />
        
        <List>
          <ListItem disablePadding>
            <ListItemButton
              onClick={() => handleNavigation('/health')}
              selected={isActive('/health')}
            >
              <ListItemIcon>
                <HealthIcon />
              </ListItemIcon>
              <ListItemText primary="Health Monitor" />
            </ListItemButton>
          </ListItem>
          
          <ListItem disablePadding>
            <ListItemButton
              onClick={() => handleNavigation('/admin')}
              selected={isActive('/admin')}
            >
              <ListItemIcon>
                <AdminIcon />
              </ListItemIcon>
              <ListItemText primary="Admin Dashboard" />
            </ListItemButton>
          </ListItem>
        </List>
        
        <Divider />
        
        <List>
          <ListItem disablePadding>
            <ListItemButton
              onClick={() => handleNavigation('/settings')}
              selected={isActive('/settings')}
            >
              <ListItemIcon>
                <SettingsIcon />
              </ListItemIcon>
              <ListItemText primary="Settings" />
            </ListItemButton>
          </ListItem>
        </List>
      </Box>
    </Drawer>
  );
};

export default Sidebar; 