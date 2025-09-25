import React, { useState, memo } from 'react';
import PropTypes from 'prop-types';
import {
  Box,
  Card,
  CardContent,
  TextField,
  Button,
  Typography,
  Alert,
  CircularProgress,
} from '@mui/material';
import { AuthService } from '../../services/auth/AuthService';
import { useAuthStore } from '../../stores/auth/authStore';
import { useConnectionStore } from '../../stores/connection/connectionStore';

interface LoginPageProps {
  authService: AuthService;
}

const LoginPage: React.FC<LoginPageProps> = memo(({ authService }) => {
  const [token, setToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { login } = useAuthStore();
  const { status: connectionStatus } = useConnectionStore();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!token.trim()) {
      setError('Please enter a token');
      return;
    }

    if (connectionStatus !== 'connected') {
      setError('WebSocket not connected. Please check your connection.');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const result = await authService.authenticate(token);

      if (result.authenticated) {
        login(token, result.role, result.session_id, result.expires_at, result.permissions);
      } else {
        setError('Authentication failed. Please check your token.');
      }
    } catch (error) {
      // Authentication error - handled by error boundary
      setError(error instanceof Error ? error.message : 'Authentication failed');
    } finally {
      setLoading(false);
    }
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
        return 'info';
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

  return (
    <Box
      display="flex"
      justifyContent="center"
      alignItems="center"
      minHeight="100vh"
      bgcolor="grey.100"
    >
      <Card sx={{ width: '100%', maxWidth: 400 }}>
        <CardContent sx={{ p: 4 }}>
          <Box textAlign="center" mb={3}>
            <Typography variant="h4" component="h1" gutterBottom>
              MediaMTX Camera Service
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Enter your authentication token to continue
            </Typography>
          </Box>

          <Alert severity={getConnectionStatusColor() as 'success' | 'error' | 'warning' | 'info'} sx={{ mb: 2 }}>
            Status: {getConnectionStatusText()}
          </Alert>

          <form onSubmit={handleSubmit}>
            <TextField
              fullWidth
              label="Authentication Token"
              type="password"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              disabled={loading || connectionStatus !== 'connected'}
              margin="normal"
              required
              autoFocus
            />

            {error && (
              <Alert severity="error" sx={{ mt: 2 }}>
                {error}
              </Alert>
            )}

            <Button
              type="submit"
              fullWidth
              variant="contained"
              disabled={loading || connectionStatus !== 'connected'}
              sx={{ mt: 3, mb: 2 }}
            >
              {loading ? <CircularProgress size={24} /> : 'Connect'}
            </Button>
          </form>

          <Typography variant="body2" color="text.secondary" textAlign="center">
            Contact your administrator for access credentials
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
});

LoginPage.displayName = 'LoginPage';

LoginPage.propTypes = {
  authService: PropTypes.instanceOf(AuthService).isRequired,
};

export default LoginPage;
