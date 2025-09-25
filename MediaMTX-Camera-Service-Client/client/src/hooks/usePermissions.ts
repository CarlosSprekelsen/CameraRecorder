import { useAuthStore } from '../stores/auth/authStore';

/**
 * usePermissions - Role-based access control hook
 * Implements security architecture from section 8.3
 */
export const usePermissions = () => {
  const { role, permissions, isAuthenticated } = useAuthStore();

  // Role-based permissions
  const hasRole = (requiredRole: 'admin' | 'operator' | 'viewer'): boolean => {
    if (!isAuthenticated || !role) return false;
    
    const roleHierarchy = {
      'viewer': 1,
      'operator': 2,
      'admin': 3
    };
    
    return roleHierarchy[role] >= roleHierarchy[requiredRole];
  };

  // Specific permission checks
  const canViewCameras = (): boolean => hasRole('viewer');
  const canControlCameras = (): boolean => hasRole('operator');
  const canManageFiles = (): boolean => hasRole('operator');
  const canDeleteFilesPermission = (): boolean => hasRole('admin');
  const canViewSystemStatus = (): boolean => hasRole('viewer');
  const canManageSystem = (): boolean => hasRole('admin');

  // Permission-based UI controls
  const canTakeSnapshot = (): boolean => canControlCameras();
  const canStartRecording = (): boolean => canControlCameras();
  const canStopRecording = (): boolean => canControlCameras();
  const canDownloadFiles = (): boolean => canManageFiles();
  // const canDeleteFiles = (): boolean => canDeleteFilesPermission();
  const canViewAdminPanel = (): boolean => hasRole('admin');

  return {
    // Role checks
    hasRole,
    isAdmin: hasRole('admin'),
    isOperator: hasRole('operator'),
    isViewer: hasRole('viewer'),
    
    // Permission checks
    canViewCameras,
    canControlCameras,
    canManageFiles,
    canDeleteFiles: canDeleteFilesPermission,
    canViewSystemStatus,
    canManageSystem,
    
    // UI controls
    canTakeSnapshot,
    canStartRecording,
    canStopRecording,
    canDownloadFiles,
    canViewAdminPanel,
    
    // Raw data
    role,
    permissions,
    isAuthenticated,
  };
};
