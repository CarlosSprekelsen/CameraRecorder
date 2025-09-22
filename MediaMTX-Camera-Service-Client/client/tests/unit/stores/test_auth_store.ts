/**
 * REQ-AUTH01-001: Authentication state management must be secure and reliable
 * REQ-AUTH01-002: Role-based access control must be properly enforced
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for auth store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on authentication state management logic
 * - Test role-based access control functionality
 * - Validate token management and session handling
 */

import { useAuthStore } from '../../../src/stores/authStore';
import type { User, AuthResponse } from '../../../src/stores/authStore';

// Mock the auth service
jest.mock('../../../src/services/authService', () => ({
  authService: {
    authenticate: jest.fn(),
    refreshToken: jest.fn(),
    logout: jest.fn(),
    validateToken: jest.fn(),
    generateToken: jest.fn()
  }
}));

describe('Auth Store', () => {
  let store: ReturnType<typeof useAuthStore.getState>;
  let mockAuthService: any;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useAuthStore.getState();
    currentStore.logout();
    
    // Get fresh store instance after reset
    store = useAuthStore.getState();
    
    // Get mock service
    mockAuthService = require('../../../src/services/authService').authService;
    jest.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.user).toBeNull();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.token).toBeNull();
      expect(state.refreshToken).toBeNull();
      expect(state.tokenExpiry).toBeNull();
      expect(state.sessionId).toBeNull();
    });
  });

  describe('Authentication State Management', () => {
    it('should set authentication status', () => {
      store.setAuthenticated(true);
      let state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(true);

      store.setAuthenticated(false);
      state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
    });

    it('should set user information', () => {
      const user: User = {
        role: 'operator',
        user_id: 'test-user',
        permissions: ['camera:read', 'camera:write'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-123'
      };

      store.setUser(user);
      
      const state = useAuthStore.getState();
      expect(state.user).toEqual(user);
    });

    it('should set loading state', () => {
      store.setLoading(true);
      let state = useAuthStore.getState();
      expect(state.isLoading).toBe(true);

      store.setLoading(false);
      state = useAuthStore.getState();
      expect(state.isLoading).toBe(false);
    });

    it('should set error state', () => {
      store.setError('Authentication failed');
      let state = useAuthStore.getState();
      expect(state.error).toBe('Authentication failed');

      store.setError(null);
      state = useAuthStore.getState();
      expect(state.error).toBeNull();
    });
  });

  describe('Token Management', () => {
    it('should set access token', () => {
      const token = 'access-token-123';
      store.setToken(token);
      
      const state = useAuthStore.getState();
      expect(state.token).toBe(token);
    });

    it('should set refresh token', () => {
      const refreshToken = 'refresh-token-123';
      store.setRefreshToken(refreshToken);
      
      const state = useAuthStore.getState();
      expect(state.refreshToken).toBe(refreshToken);
    });

    it('should set token expiry', () => {
      const expiry = new Date(Date.now() + 3600000); // 1 hour from now
      store.setTokenExpiry(expiry);
      
      const state = useAuthStore.getState();
      expect(state.tokenExpiry).toEqual(expiry);
    });

    it('should set session ID', () => {
      const sessionId = 'session-456';
      store.setSessionId(sessionId);
      
      const state = useAuthStore.getState();
      expect(state.sessionId).toBe(sessionId);
    });

    it('should check if token is expired', () => {
      const pastTime = new Date(Date.now() - 1000); // 1 second ago
      store.setTokenExpiry(pastTime);
      expect(store.isTokenExpired()).toBe(true);

      const futureTime = new Date(Date.now() + 3600000); // 1 hour from now
      store.setTokenExpiry(futureTime);
      expect(store.isTokenExpired()).toBe(false);
    });

    it('should get time until token expiry', () => {
      const futureTime = new Date(Date.now() + 5000); // 5 seconds from now
      store.setTokenExpiry(futureTime);
      
      const timeUntilExpiry = store.getTimeUntilTokenExpiry();
      expect(timeUntilExpiry).toBeGreaterThan(4000);
      expect(timeUntilExpiry).toBeLessThanOrEqual(5000);
    });
  });

  describe('Authentication Operations', () => {
    it('should authenticate successfully', async () => {
      const mockAuthResponse: AuthResponse = {
        authenticated: true,
        role: 'operator',
        permissions: ['camera:read', 'camera:write'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-123'
      };

      mockAuthService.authenticate.mockResolvedValue(mockAuthResponse);

      await store.authenticate('test-token');

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(true);
      expect(state.user).toEqual({
        role: 'operator',
        permissions: ['camera:read', 'camera:write'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-123'
      });
      expect(state.sessionId).toBe('session-123');
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle authentication failure', async () => {
      mockAuthService.authenticate.mockRejectedValue(new Error('Invalid token'));

      await store.authenticate('invalid-token');

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.user).toBeNull();
      expect(state.error).toBe('Invalid token');
      expect(state.isLoading).toBe(false);
    });

    it('should refresh token successfully', async () => {
      const mockAuthResponse: AuthResponse = {
        authenticated: true,
        role: 'operator',
        permissions: ['camera:read', 'camera:write'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-123'
      };

      mockAuthService.refreshToken.mockResolvedValue(mockAuthResponse);

      await store.refreshToken();

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(true);
      expect(state.user).toBeDefined();
      expect(state.error).toBeNull();
    });

    it('should handle refresh token failure', async () => {
      mockAuthService.refreshToken.mockRejectedValue(new Error('Refresh failed'));

      await store.refreshToken();

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.error).toBe('Refresh failed');
    });

    it('should logout successfully', async () => {
      // First authenticate
      const mockAuthResponse: AuthResponse = {
        authenticated: true,
        role: 'operator',
        permissions: ['camera:read'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-123'
      };
      mockAuthService.authenticate.mockResolvedValue(mockAuthResponse);
      await store.authenticate('test-token');

      // Then logout
      mockAuthService.logout.mockResolvedValue(undefined);
      await store.logout();

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
      expect(state.refreshToken).toBeNull();
      expect(state.sessionId).toBeNull();
      expect(state.error).toBeNull();
    });
  });

  describe('Role-Based Access Control', () => {
    beforeEach(async () => {
      // Set up authenticated user
      const user: User = {
        role: 'operator',
        user_id: 'test-user',
        permissions: ['camera:read', 'camera:write', 'recording:start'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-123'
      };
      store.setUser(user);
      store.setAuthenticated(true);
    });

    it('should check if user has specific role', () => {
      expect(store.hasRole('operator')).toBe(true);
      expect(store.hasRole('admin')).toBe(false);
      expect(store.hasRole('viewer')).toBe(false);
    });

    it('should check if user has any of the specified roles', () => {
      expect(store.hasAnyRole(['operator', 'admin'])).toBe(true);
      expect(store.hasAnyRole(['admin', 'viewer'])).toBe(false);
    });

    it('should check if user has specific permission', () => {
      expect(store.hasPermission('camera:read')).toBe(true);
      expect(store.hasPermission('camera:write')).toBe(true);
      expect(store.hasPermission('admin:access')).toBe(false);
    });

    it('should check if user has any of the specified permissions', () => {
      expect(store.hasAnyPermission(['camera:read', 'admin:access'])).toBe(true);
      expect(store.hasAnyPermission(['admin:access', 'system:config'])).toBe(false);
    });

    it('should check if user is admin', () => {
      expect(store.isAdmin()).toBe(false);

      // Change to admin role
      const adminUser: User = {
        role: 'admin',
        user_id: 'admin-user',
        permissions: ['admin:access', 'system:config'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-456'
      };
      store.setUser(adminUser);
      expect(store.isAdmin()).toBe(true);
    });

    it('should check if user is operator or admin', () => {
      expect(store.isOperatorOrAdmin()).toBe(true);

      // Change to viewer role
      const viewerUser: User = {
        role: 'viewer',
        user_id: 'viewer-user',
        permissions: ['camera:read'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-789'
      };
      store.setUser(viewerUser);
      expect(store.isOperatorOrAdmin()).toBe(false);
    });

    it('should return empty permissions for unauthenticated user', () => {
      store.setAuthenticated(false);
      store.setUser(null);

      expect(store.hasPermission('camera:read')).toBe(false);
      expect(store.hasRole('operator')).toBe(false);
      expect(store.isAdmin()).toBe(false);
    });
  });

  describe('Session Management', () => {
    it('should check if session is valid', () => {
      const futureTime = new Date(Date.now() + 3600000); // 1 hour from now
      store.setTokenExpiry(futureTime);
      store.setAuthenticated(true);
      store.setSessionId('session-123');

      expect(store.isSessionValid()).toBe(true);

      // Test with expired token
      const pastTime = new Date(Date.now() - 1000); // 1 second ago
      store.setTokenExpiry(pastTime);
      expect(store.isSessionValid()).toBe(false);

      // Test with no session ID
      store.setTokenExpiry(futureTime);
      store.setSessionId(null);
      expect(store.isSessionValid()).toBe(false);
    });

    it('should get session info', () => {
      const user: User = {
        role: 'operator',
        user_id: 'test-user',
        permissions: ['camera:read'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-123'
      };
      store.setUser(user);
      store.setAuthenticated(true);
      store.setSessionId('session-123');

      const sessionInfo = store.getSessionInfo();
      expect(sessionInfo).toEqual({
        isAuthenticated: true,
        user,
        sessionId: 'session-123',
        isSessionValid: true
      });
    });
  });

  describe('Error Handling', () => {
    it('should clear error', () => {
      store.setError('Test error');
      store.clearError();
      
      const state = useAuthStore.getState();
      expect(state.error).toBeNull();
    });

    it('should handle authentication errors gracefully', async () => {
      mockAuthService.authenticate.mockRejectedValue(new Error('Network error'));

      await store.authenticate('test-token');

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.error).toBe('Network error');
      expect(state.isLoading).toBe(false);
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.setAuthenticated(true);
      store.setUser({
        role: 'operator',
        user_id: 'test-user',
        permissions: ['camera:read'],
        expires_at: '2024-12-31T23:59:59Z',
        session_id: 'session-123'
      });
      store.setToken('test-token');
      store.setError('Test error');
      
      // Reset
      store.reset();
      
      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
      expect(state.error).toBeNull();
      expect(state.isLoading).toBe(false);
    });
  });
});
