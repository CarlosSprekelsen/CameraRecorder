import React from 'react';
import { Navigate } from 'react-router-dom';
import { Box, Typography, Alert } from '@mui/material';
import { usePermissions } from '../../hooks/usePermissions';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: 'admin' | 'operator' | 'viewer';
  requiredPermission?: string;
  fallbackPath?: string;
  showAccessDenied?: boolean;
}

/**
 * ProtectedRoute - Role-based route protection
 * Implements security architecture from section 8.3
 */
const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requiredRole,
  requiredPermission,
  fallbackPath = '/cameras',
  showAccessDenied = true,
}) => {
  const { hasRole, isAuthenticated } = usePermissions();

  // Check authentication
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Check role requirement
  if (requiredRole && !hasRole(requiredRole)) {
    if (showAccessDenied) {
      return (
        <Box sx={{ p: 3, textAlign: 'center' }}>
          <Alert severity="error" sx={{ mb: 2 }}>
            Access Denied
          </Alert>
          <Typography variant="h6" color="text.secondary">
            You don't have permission to access this page.
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            Required role: {requiredRole}
          </Typography>
        </Box>
      );
    }
    return <Navigate to={fallbackPath} replace />;
  }

  // Check permission requirement (if implemented)
  if (requiredPermission) {
    // TODO: Implement specific permission checks when permission system is expanded
    console.warn('Permission-based access control not yet implemented');
  }

  return <>{children}</>;
};

export default ProtectedRoute;
