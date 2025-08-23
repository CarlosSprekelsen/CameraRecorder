/**
 * Authentication UI Component
 * Handles user login and role-based access control
 * Aligned with server authentication API
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Alert,
  CircularProgress,
  Chip,
  Stack,
  Divider,
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import {
  Login as LoginIcon,
  Logout as LogoutIcon,
  Person as PersonIcon,
  Security as SecurityIcon,
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
} from '@mui/icons-material';
import { useAuthStore } from '../../stores/authStore';
import { authService } from '../../services/authService';

/**
 * Authentication UI Component Props
 */
interface AuthUIProps {
  showLoginForm?: boolean;
  onLoginSuccess?: () => void;
  onLogout?: () => void;
}

/**
 * Authentication UI Component
 */
const AuthUI: React.FC<AuthUIProps> = ({
  showLoginForm = true,
  onLoginSuccess,
  onLogout,
}) => {
  const {
    isAuthenticated,
    user,
    isLoading,
    error,
    login,
    logout,
    clearError,
  } = useAuthStore();

  const [showPassword, setShowPassword] = useState(false);
  const [showRoleDialog, setShowRoleDialog] = useState(false);
  const [authMethod, setAuthMethod] = useState<'jwt' | 'api_key'>('jwt');
  const [token, setToken] = useState('');
  const [apiKey, setApiKey] = useState('');

  // Auto-login if token exists in storage
  useEffect(() => {
    const storedToken = authService.getToken();
    if (storedToken && !isAuthenticated) {
      handleAutoLogin(storedToken);
    }
  }, [isAuthenticated]);

  const handleAutoLogin = async (storedToken: string) => {
    try {
      await login(storedToken);
      onLoginSuccess?.();
    } catch (error) {
      // Auto-login failed, clear invalid token
      authService.clearToken();
    }
  };

  const handleLogin = async () => {
    const authToken = authMethod === 'jwt' ? token : apiKey;
    
    if (!authToken.trim()) {
      return;
    }

    try {
      await login(authToken);
      onLoginSuccess?.();
    } catch (error) {
      // Error is handled by the store
    }
  };

  const handleLogout = () => {
    logout();
    setToken('');
    setApiKey('');
    onLogout?.();
  };

  const handleRoleInfo = () => {
    setShowRoleDialog(true);
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin':
        return 'error';
      case 'operator':
        return 'warning';
      case 'viewer':
        return 'success';
      default:
        return 'default';
    }
  };

  const getRoleDescription = (role: string) => {
    switch (role) {
      case 'admin':
        return 'Full access to all features including system management';
      case 'operator':
        return 'Camera control and file management capabilities';
      case 'viewer':
        return 'Read-only access to camera status and file listings';
      default:
        return 'Unknown role';
    }
  };

  if (isAuthenticated && user) {
    return (
      <Card>
        <CardContent>
          <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
            <Stack direction="row" spacing={2} alignItems="center">
              <PersonIcon color="primary" />
              <Box>
                <Typography variant="h6" gutterBottom>
                  Authenticated
                </Typography>
                <Stack direction="row" spacing={1} alignItems="center">
                  <Chip
                    label={user.role}
                    color={getRoleColor(user.role)}
                    size="small"
                  />
                  <Tooltip title="Role information">
                    <IconButton size="small" onClick={handleRoleInfo}>
                      <SecurityIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </Stack>
                {user.user_id && (
                  <Typography variant="body2" color="text.secondary">
                    User ID: {user.user_id}
                  </Typography>
                )}
              </Box>
            </Stack>
            
            <Button
              variant="outlined"
              startIcon={<LogoutIcon />}
              onClick={handleLogout}
              color="error"
            >
              Logout
            </Button>
          </Stack>
        </CardContent>

        {/* Role Information Dialog */}
        <Dialog open={showRoleDialog} onClose={() => setShowRoleDialog(false)}>
          <DialogTitle>Role Information</DialogTitle>
          <DialogContent>
            <Typography variant="body1" gutterBottom>
              <strong>Current Role:</strong> {user.role}
            </Typography>
            <Typography variant="body2" color="text.secondary" paragraph>
              {getRoleDescription(user.role)}
            </Typography>
            
            <Divider sx={{ my: 2 }} />
            
            <Typography variant="h6" gutterBottom>
              Available Roles:
            </Typography>
            
            <Stack spacing={1}>
              <Box>
                <Chip label="admin" color="error" size="small" sx={{ mr: 1 }} />
                <Typography variant="body2" color="text.secondary">
                  Full access to all features including system management, metrics, and configuration
                </Typography>
              </Box>
              
              <Box>
                <Chip label="operator" color="warning" size="small" sx={{ mr: 1 }} />
                <Typography variant="body2" color="text.secondary">
                  Camera control operations (snapshots, recording) and file management
                </Typography>
              </Box>
              
              <Box>
                <Chip label="viewer" color="success" size="small" sx={{ mr: 1 }} />
                <Typography variant="body2" color="text.secondary">
                  Read-only access to camera status, file listings, and basic information
                </Typography>
              </Box>
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setShowRoleDialog(false)}>Close</Button>
          </DialogActions>
        </Dialog>
      </Card>
    );
  }

  if (!showLoginForm) {
    return null;
  }

  return (
    <Card>
      <CardContent>
        <Stack spacing={3}>
          <Box textAlign="center">
            <LoginIcon color="primary" sx={{ fontSize: 48, mb: 1 }} />
            <Typography variant="h5" gutterBottom>
              Authentication Required
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Please authenticate to access the camera service
            </Typography>
          </Box>

          {error && (
            <Alert severity="error" onClose={clearError}>
              {error}
            </Alert>
          )}

          <FormControl fullWidth>
            <InputLabel>Authentication Method</InputLabel>
            <Select
              value={authMethod}
              onChange={(e) => setAuthMethod(e.target.value as 'jwt' | 'api_key')}
              label="Authentication Method"
            >
              <MenuItem value="jwt">JWT Token</MenuItem>
              <MenuItem value="api_key">API Key</MenuItem>
            </Select>
          </FormControl>

          {authMethod === 'jwt' ? (
            <TextField
              fullWidth
              label="JWT Token"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
              multiline
              rows={3}
              helperText="Enter your JWT authentication token"
            />
          ) : (
            <TextField
              fullWidth
              label="API Key"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="your-api-key-here"
              helperText="Enter your API key"
            />
          )}

          <Button
            fullWidth
            variant="contained"
            size="large"
            onClick={handleLogin}
            disabled={isLoading || (!token.trim() && !apiKey.trim())}
            startIcon={isLoading ? <CircularProgress size={20} /> : <LoginIcon />}
          >
            {isLoading ? 'Authenticating...' : 'Login'}
          </Button>

          <Divider>
            <Typography variant="body2" color="text.secondary">
              Authentication Information
            </Typography>
          </Divider>

          <Alert severity="info">
            <Typography variant="body2" gutterBottom>
              <strong>Available Roles:</strong>
            </Typography>
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
              <Chip label="viewer" color="success" size="small" />
              <Chip label="operator" color="warning" size="small" />
              <Chip label="admin" color="error" size="small" />
            </Stack>
            <Typography variant="body2" sx={{ mt: 1 }}>
              Contact your system administrator to obtain authentication credentials.
            </Typography>
          </Alert>
        </Stack>
      </CardContent>
    </Card>
  );
};

export default AuthUI;
