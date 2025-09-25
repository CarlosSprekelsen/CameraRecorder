/**
 * usePermissions hook unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * 
 * Requirements Coverage:
 * - REQ-HOOK-001: Role-based access control functionality
 * - REQ-HOOK-002: Permission checking for different roles
 * - REQ-HOOK-003: Authentication state handling
 * - REQ-HOOK-004: UI permission controls
 * 
 * Test Categories: Unit
 * API Documentation Reference: docs/api/json-rpc-methods.md
 */

import { renderHook } from '@testing-library/react';
import { usePermissions } from '../../../src/hooks/usePermissions';
import { useAuthStore } from '../../../src/stores/auth/authStore';
import { MockDataFactory } from '../../utils/mocks';

// Mock the auth store - centralized pattern
jest.mock('../../../src/stores/auth/authStore');

// Mock logger service - use centralized pattern
jest.mock('../../../src/services/logger/LoggerService', () => ({
  logger: MockDataFactory.createMockLoggerService()
}));

describe('usePermissions Hook Unit Tests', () => {
  const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>;

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('REQ-HOOK-001: Should return correct permissions for admin role', () => {
    // Arrange
    mockUseAuthStore.mockReturnValue(MockDataFactory.createMockAuthStore({
      role: 'admin',
      permissions: ['read', 'write', 'delete', 'admin']
    }));

    // Act
    const { result } = renderHook(() => usePermissions());

    // Assert
    expect(result.current.hasRole('admin')).toBe(true);
    expect(result.current.hasRole('operator')).toBe(true);
    expect(result.current.hasRole('viewer')).toBe(true);
    expect(result.current.isAdmin).toBe(true);
    expect(result.current.isOperator).toBe(true);
    expect(result.current.isViewer).toBe(true);
    expect(result.current.canViewCameras()).toBe(true);
    expect(result.current.canControlCameras()).toBe(true);
    expect(result.current.canManageFiles()).toBe(true);
    expect(result.current.canDeleteFiles()).toBe(true);
    expect(result.current.canViewSystemStatus()).toBe(true);
    expect(result.current.canManageSystem()).toBe(true);
    expect(result.current.canTakeSnapshot()).toBe(true);
    expect(result.current.canStartRecording()).toBe(true);
    expect(result.current.canStopRecording()).toBe(true);
    expect(result.current.canDownloadFiles()).toBe(true);
    expect(result.current.canViewAdminPanel()).toBe(true);
  });

  test('REQ-HOOK-002: Should return correct permissions for operator role', () => {
    // Arrange
    mockUseAuthStore.mockReturnValue(MockDataFactory.createMockAuthStore({
      role: 'operator',
      permissions: ['read', 'write']
    }));

    // Act
    const { result } = renderHook(() => usePermissions());

    // Assert
    expect(result.current.hasRole('admin')).toBe(false);
    expect(result.current.hasRole('operator')).toBe(true);
    expect(result.current.hasRole('viewer')).toBe(true);
    expect(result.current.isAdmin).toBe(false);
    expect(result.current.isOperator).toBe(true);
    expect(result.current.isViewer).toBe(true);
    expect(result.current.canViewCameras()).toBe(true);
    expect(result.current.canControlCameras()).toBe(true);
    expect(result.current.canManageFiles()).toBe(true);
    expect(result.current.canDeleteFiles()).toBe(false);
    expect(result.current.canViewSystemStatus()).toBe(false);
    expect(result.current.canManageSystem()).toBe(false);
    expect(result.current.canTakeSnapshot()).toBe(true);
    expect(result.current.canStartRecording()).toBe(true);
    expect(result.current.canStopRecording()).toBe(true);
    expect(result.current.canDownloadFiles()).toBe(true);
    expect(result.current.canViewAdminPanel()).toBe(false);
  });

  test('REQ-HOOK-003: Should return correct permissions for viewer role', () => {
    // Arrange
    mockUseAuthStore.mockReturnValue(MockDataFactory.createMockAuthStore({
      role: 'viewer',
      permissions: ['read']
    }));

    // Act
    const { result } = renderHook(() => usePermissions());

    // Assert
    expect(result.current.hasRole('admin')).toBe(false);
    expect(result.current.hasRole('operator')).toBe(false);
    expect(result.current.hasRole('viewer')).toBe(true);
    expect(result.current.isAdmin).toBe(false);
    expect(result.current.isOperator).toBe(false);
    expect(result.current.isViewer).toBe(true);
    expect(result.current.canViewCameras()).toBe(true);
    expect(result.current.canControlCameras()).toBe(false);
    expect(result.current.canManageFiles()).toBe(false);
    expect(result.current.canDeleteFiles()).toBe(false);
    expect(result.current.canViewSystemStatus()).toBe(false);
    expect(result.current.canManageSystem()).toBe(false);
    expect(result.current.canTakeSnapshot()).toBe(false);
    expect(result.current.canStartRecording()).toBe(false);
    expect(result.current.canStopRecording()).toBe(false);
    expect(result.current.canDownloadFiles()).toBe(false);
    expect(result.current.canViewAdminPanel()).toBe(false);
  });

  test('REQ-HOOK-004: Should handle unauthenticated state correctly', () => {
    // Arrange
    mockUseAuthStore.mockReturnValue(MockDataFactory.createMockAuthStore({
      role: null,
      permissions: [],
      isAuthenticated: false
    }));

    // Act
    const { result } = renderHook(() => usePermissions());

    // Assert
    expect(result.current.hasRole('admin')).toBe(false);
    expect(result.current.hasRole('operator')).toBe(false);
    expect(result.current.hasRole('viewer')).toBe(false);
    expect(result.current.isAdmin).toBe(false);
    expect(result.current.isOperator).toBe(false);
    expect(result.current.isViewer).toBe(false);
    expect(result.current.canViewCameras()).toBe(false);
    expect(result.current.canControlCameras()).toBe(false);
    expect(result.current.canManageFiles()).toBe(false);
    expect(result.current.canDeleteFiles()).toBe(false);
    expect(result.current.canViewSystemStatus()).toBe(false);
    expect(result.current.canManageSystem()).toBe(false);
    expect(result.current.canTakeSnapshot()).toBe(false);
    expect(result.current.canStartRecording()).toBe(false);
    expect(result.current.canStopRecording()).toBe(false);
    expect(result.current.canDownloadFiles()).toBe(false);
    expect(result.current.canViewAdminPanel()).toBe(false);
    expect(result.current.role).toBe(null);
    expect(result.current.permissions).toEqual([]);
    expect(result.current.isAuthenticated).toBe(false);
  });

  test('REQ-HOOK-005: Should handle null role correctly', () => {
    // Arrange
    mockUseAuthStore.mockReturnValue(MockDataFactory.createMockAuthStore({
      role: null,
      permissions: []
    }));

    // Act
    const { result } = renderHook(() => usePermissions());

    // Assert
    expect(result.current.hasRole('admin')).toBe(false);
    expect(result.current.hasRole('operator')).toBe(false);
    expect(result.current.hasRole('viewer')).toBe(false);
    expect(result.current.isAdmin).toBe(false);
    expect(result.current.isOperator).toBe(false);
    expect(result.current.isViewer).toBe(false);
    expect(result.current.canViewCameras()).toBe(false);
    expect(result.current.canControlCameras()).toBe(false);
    expect(result.current.canManageFiles()).toBe(false);
    expect(result.current.canDeleteFiles()).toBe(false);
    expect(result.current.canViewSystemStatus()).toBe(false);
    expect(result.current.canManageSystem()).toBe(false);
    expect(result.current.canTakeSnapshot()).toBe(false);
    expect(result.current.canStartRecording()).toBe(false);
    expect(result.current.canStopRecording()).toBe(false);
    expect(result.current.canDownloadFiles()).toBe(false);
    expect(result.current.canViewAdminPanel()).toBe(false);
  });

  test('REQ-HOOK-006: Should return raw data correctly', () => {
    // Arrange
    const mockAuthData = MockDataFactory.createMockAuthStore({
      role: 'admin' as const,
      permissions: ['read', 'write', 'delete', 'admin']
    });
    mockUseAuthStore.mockReturnValue(mockAuthData);

    // Act
    const { result } = renderHook(() => usePermissions());

    // Assert
    expect(result.current.role).toBe('admin');
    expect(result.current.permissions).toEqual(['read', 'write', 'delete', 'admin']);
    expect(result.current.isAuthenticated).toBe(true);
  });

  test('REQ-HOOK-007: Should handle role hierarchy correctly', () => {
    // Test admin role hierarchy
    mockUseAuthStore.mockReturnValue(MockDataFactory.createMockAuthStore({
      role: 'admin',
      permissions: ['read', 'write', 'delete', 'admin']
    }));

    const { result: adminResult } = renderHook(() => usePermissions());
    expect(adminResult.current.hasRole('admin')).toBe(true);
    expect(adminResult.current.hasRole('operator')).toBe(true);
    expect(adminResult.current.hasRole('viewer')).toBe(true);

    // Test operator role hierarchy
    mockUseAuthStore.mockReturnValue(MockDataFactory.createMockAuthStore({
      role: 'operator',
      permissions: ['read', 'write']
    }));

    const { result: operatorResult } = renderHook(() => usePermissions());
    expect(operatorResult.current.hasRole('admin')).toBe(false);
    expect(operatorResult.current.hasRole('operator')).toBe(true);
    expect(operatorResult.current.hasRole('viewer')).toBe(true);

    // Test viewer role hierarchy
    mockUseAuthStore.mockReturnValue(MockDataFactory.createMockAuthStore({
      role: 'viewer',
      permissions: ['read']
    }));

    const { result: viewerResult } = renderHook(() => usePermissions());
    expect(viewerResult.current.hasRole('admin')).toBe(false);
    expect(viewerResult.current.hasRole('operator')).toBe(false);
    expect(viewerResult.current.hasRole('viewer')).toBe(true);
  });
});
