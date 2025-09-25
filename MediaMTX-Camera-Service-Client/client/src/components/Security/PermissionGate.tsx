import React from 'react';
import { usePermissions } from '../../hooks/usePermissions';

interface PermissionGateProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  requireRole?: 'admin' | 'operator' | 'viewer';
  requirePermission?: 'viewCameras' | 'controlCameras' | 'manageFiles' | 'deleteFiles' | 'viewSystem' | 'manageSystem';
  requireAny?: boolean; // If true, any of the requirements must be met
}

/**
 * PermissionGate - Fine-grained permission-based UI control
 * Implements security architecture from section 8.3
 */
const PermissionGate: React.FC<PermissionGateProps> = ({
  children,
  fallback = null,
  requireRole,
  requirePermission,
  requireAny = false,
}) => {
  const {
    hasRole,
    canViewCameras,
    canControlCameras,
    canManageFiles,
    canDeleteFiles,
    canViewSystemStatus,
    canManageSystem,
  } = usePermissions();

  // Check role requirement
  const hasRequiredRole = requireRole ? hasRole(requireRole) : true;

  // Check permission requirement
  const hasRequiredPermission = (() => {
    if (!requirePermission) return true;
    
    switch (requirePermission) {
      case 'viewCameras':
        return canViewCameras();
      case 'controlCameras':
        return canControlCameras();
      case 'manageFiles':
        return canManageFiles();
      case 'deleteFiles':
        return canDeleteFiles();
      case 'viewSystem':
        return canViewSystemStatus();
      case 'manageSystem':
        return canManageSystem();
      default:
        return false;
    }
  })();

  // Determine access based on requirements
  const hasAccess = requireAny 
    ? (hasRequiredRole || hasRequiredPermission)
    : (hasRequiredRole && hasRequiredPermission);

  return hasAccess ? <>{children}</> : <>{fallback}</>;
};

export default PermissionGate;
