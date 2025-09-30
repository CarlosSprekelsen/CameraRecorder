/**
 * @fileoverview LoginPage component for user authentication
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React, { useState, memo } from 'react';
// ARCHITECTURE FIX: Removed unused PropTypes import
import { Box } from '../../components/atoms/Box/Box';
import { Card } from '../../components/atoms/Card/Card';
import { CardContent } from '../../components/atoms/CardContent/CardContent';
import { TextField } from '../../components/atoms/TextField/TextField';
import { Button } from '../../components/atoms/Button/Button';
import { Typography } from '../../components/atoms/Typography/Typography';
import { Alert } from '../../components/atoms/Alert/Alert';
import { CircularProgress } from '../../components/atoms/CircularProgress/CircularProgress';
// ARCHITECTURE FIX: Removed direct service import - components must use stores only
import { useAuthStore } from '../../stores/auth/authStore';
import { useConnectionStore } from '../../stores/connection/connectionStore';

interface LoginPageProps {
  // ARCHITECTURE FIX: Removed service props - components only use stores
}

const LoginPage: React.FC<LoginPageProps> = memo(() => {
  const [token, setToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { authenticate } = useAuthStore();
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
      await authenticate(token);
      // Authentication success is handled by the store
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
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        backgroundColor: '#f5f5f5'
      }}
    >
      <Card className="w-full max-w-md">
        <CardContent className="p-6">
          <Box sx={{ textAlign: 'center', marginBottom: 3 }}>
            <Typography variant="h4" component="h1">
              MediaMTX Camera Service
            </Typography>
            <Typography variant="body2" color="secondary">
              Enter your authentication token to continue
            </Typography>
          </Box>

          <Alert
            severity={getConnectionStatusColor() as 'success' | 'error' | 'warning' | 'info'}
            className="mb-2"
          >
            Status: {getConnectionStatusText()}
          </Alert>

          <form onSubmit={handleSubmit}>
            <TextField
              fullWidth
              label="Authentication Token"
              type="password"
              value={token}
              onChange={(e) => setToken(e)}
              disabled={loading || connectionStatus !== 'connected'}
              className="mb-4"
            />

            {error && (
              <Alert severity="error" className="mt-2">
                {error}
              </Alert>
            )}

            <Button
              variant="primary"
              disabled={loading || connectionStatus !== 'connected'}
              className="w-full mt-3 mb-2"
              onClick={handleSubmit}
            >
              {loading ? <CircularProgress size={24} /> : 'Connect'}
            </Button>
          </form>

          <Typography variant="body2" color="secondary" className="text-center">
            Contact your administrator for access credentials
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
});

LoginPage.displayName = 'LoginPage';

// ARCHITECTURE FIX: Removed PropTypes - components use stores, not direct service props

export default LoginPage;
