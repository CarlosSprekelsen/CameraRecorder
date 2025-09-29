/**
 * AuthService unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-AUTH-001: Authentication with JWT tokens
 * - REQ-AUTH-002: Session management and storage
 * - REQ-AUTH-003: Token validation and expiration
 * - REQ-AUTH-004: Role-based access control
 * - REQ-AUTH-005: Permission checking
 * 
 * Test Categories: Unit
 * API Documentation Reference: ../mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

import { AuthService } from '../../../src/services/auth/AuthService';
import { APIClient } from '../../../src/services/abstraction/APIClient';
import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Use centralized mocks - eliminates duplication
const mockWebSocketService = MockDataFactory.createMockWebSocketService();
const mockLoggerService = MockDataFactory.createMockLoggerService();
const mockAPIClient = new APIClient(mockWebSocketService, mockLoggerService);
const mockSessionStorage = MockDataFactory.createMockSessionStorage();

// Mock sessionStorage for jsdom environment
if (typeof window === 'undefined') {
  (global as any).window = {
    sessionStorage: mockSessionStorage,
  };
} else {
  Object.defineProperty(window, 'sessionStorage', {
    value: mockSessionStorage,
    writable: true,
  });
}

describe('AuthService Unit Tests', () => {
  let authService: AuthService;

  beforeEach(() => {
    jest.clearAllMocks();
    mockWebSocketService.isConnected = true;
    authService = new AuthService(mockAPIClient, mockLoggerService);
  });

  describe('REQ-AUTH-001: Authentication with JWT tokens', () => {
    test('should authenticate successfully with valid token', async () => {
      const token = 'valid-jwt-token';
      const expectedResult = MockDataFactory.getAuthenticateResult();

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await authService.authenticate(token);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('authenticate', {
        auth_token: token,
      });
      expect(result).toEqual(expectedResult);
      expect(mockSessionStorage.setItem).toHaveBeenCalledWith('auth_token', token);
      expect(mockSessionStorage.setItem).toHaveBeenCalledWith(
        'auth_session',
        JSON.stringify({
          session_id: expectedResult.session_id,
          role: expectedResult.role,
          permissions: expectedResult.permissions,
          expires_at: expectedResult.expires_at,
        })
      );
    });

    test('should handle authentication failure', async () => {
      const token = 'invalid-token';
      const authResult = { ...MockDataFactory.getAuthenticateResult(), authenticated: false };

      mockWebSocketService.sendRPC.mockResolvedValue(authResult);

      const result = await authService.authenticate(token);

      expect(result.authenticated).toBe(false);
      expect(mockSessionStorage.setItem).not.toHaveBeenCalled();
    });

    test('should throw error when WebSocket not connected', async () => {
      mockWebSocketService.isConnected = false;

      await expect(authService.authenticate('token')).rejects.toThrow(
        'WebSocket not connected'
      );
    });

    test('should validate authentication result', async () => {
      const token = 'valid-jwt-token';
      const expectedResult = MockDataFactory.getAuthenticateResult();

      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await authService.authenticate(token);

      expect(APIResponseValidator.validateAuthenticateResult(result)).toBe(true);
    });
  });

  describe('REQ-AUTH-002: Session management and storage', () => {
    test('should store session data on successful authentication', async () => {
      const token = 'valid-jwt-token';
      const authResult = MockDataFactory.getAuthenticateResult();

      mockWebSocketService.sendRPC.mockResolvedValue(authResult);

      await authService.authenticate(token);

      expect(mockSessionStorage.setItem).toHaveBeenCalledWith('auth_token', token);
      expect(mockSessionStorage.setItem).toHaveBeenCalledWith(
        'auth_session',
        JSON.stringify({
          session_id: authResult.session_id,
          role: authResult.role,
          permissions: authResult.permissions,
          expires_at: authResult.expires_at,
        })
      );
    });

    test('should get stored token', () => {
      const token = 'stored-token';
      mockSessionStorage.getItem.mockReturnValue(token);

      const result = authService.getStoredToken();

      expect(result).toBe(token);
      expect(mockSessionStorage.getItem).toHaveBeenCalledWith('auth_token');
    });

    test('should get stored session', () => {
      const session = {
        session_id: 'session-123',
        role: 'admin',
        permissions: ['read', 'write', 'delete', 'admin'],
        expires_at: new Date(Date.now() + 3600000).toISOString(),
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(session));

      const result = authService.getStoredSession();

      expect(result).toEqual(session);
      expect(mockSessionStorage.getItem).toHaveBeenCalledWith('auth_session');
    });

    test('should return null for invalid session JSON', () => {
      mockSessionStorage.getItem.mockReturnValue('invalid-json');

      const result = authService.getStoredSession();

      expect(result).toBeNull();
    });

    test('should logout and clear session', () => {
      authService.logout();

      expect(mockSessionStorage.removeItem).toHaveBeenCalledWith('auth_token');
      expect(mockSessionStorage.removeItem).toHaveBeenCalledWith('auth_session');
    });
  });

  describe('REQ-AUTH-003: Token validation and expiration', () => {
    test('should detect expired token', () => {
      const expiredSession = {
        session_id: 'session-123',
        role: 'admin',
        permissions: ['read', 'write', 'delete', 'admin'],
        expires_at: new Date(Date.now() - 60000).toISOString(), // 1 minute ago
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(expiredSession));

      const result = authService.isTokenExpired();

      expect(result).toBe(true);
    });

    test('should detect valid token', () => {
      const validSession = {
        session_id: 'session-123',
        role: 'admin',
        permissions: ['read', 'write', 'delete', 'admin'],
        expires_at: new Date(Date.now() + 600000).toISOString(), // 10 minutes from now
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(validSession));

      const result = authService.isTokenExpired();

      expect(result).toBe(false);
    });

    test('should consider token expired if no session', () => {
      mockSessionStorage.getItem.mockReturnValue(null);

      const result = authService.isTokenExpired();

      expect(result).toBe(true);
    });

    test('should consider token expired if expires within 5 minutes', () => {
      const soonToExpireSession = {
        session_id: 'session-123',
        role: 'admin',
        permissions: ['read', 'write', 'delete', 'admin'],
        expires_at: new Date(Date.now() + 240000).toISOString(), // 4 minutes from now
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(soonToExpireSession));

      const result = authService.isTokenExpired();

      expect(result).toBe(true);
    });
  });

  describe('REQ-AUTH-004: Role-based access control', () => {
    test('should check if user is authenticated', () => {
      const token = 'valid-token';
      const session = {
        session_id: 'session-123',
        role: 'admin',
        permissions: ['read', 'write', 'delete', 'admin'],
        expires_at: new Date(Date.now() + 600000).toISOString(),
      };

      // Mock the sessionStorage calls
      mockSessionStorage.getItem.mockImplementation((key: string) => {
        if (key === 'auth_token') return token;
        if (key === 'auth_session') return JSON.stringify(session);
        return null;
      });

      const result = authService.isAuthenticated();

      expect(result).toBe(true);
    });

    test('should return false if not authenticated', () => {
      mockSessionStorage.getItem.mockReturnValue(null);

      const result = authService.isAuthenticated();

      expect(result).toBe(false);
    });

    test('should get user role', () => {
      const session = {
        session_id: 'session-123',
        role: 'operator',
        permissions: ['read', 'write'],
        expires_at: new Date(Date.now() + 600000).toISOString(),
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(session));

      const result = authService.getRole();

      expect(result).toBe('operator');
    });

    test('should return null role if no session', () => {
      mockSessionStorage.getItem.mockReturnValue(null);

      const result = authService.getRole();

      expect(result).toBeNull();
    });

    test('should check specific role', () => {
      const session = {
        session_id: 'session-123',
        role: 'admin',
        permissions: ['read', 'write', 'delete', 'admin'],
        expires_at: new Date(Date.now() + 600000).toISOString(),
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(session));

      expect(authService.hasRole('admin')).toBe(true);
      expect(authService.hasRole('operator')).toBe(false);
    });

    test('should check any role from list', () => {
      const session = {
        session_id: 'session-123',
        role: 'operator',
        permissions: ['read', 'write'],
        expires_at: new Date(Date.now() + 600000).toISOString(),
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(session));

      expect(authService.hasAnyRole(['admin', 'operator'])).toBe(true);
      expect(authService.hasAnyRole(['admin', 'viewer'])).toBe(false);
    });
  });

  describe('REQ-AUTH-005: Permission checking', () => {
    test('should get user permissions', () => {
      const session = {
        session_id: 'session-123',
        role: 'admin',
        permissions: ['read', 'write', 'delete', 'admin'],
        expires_at: new Date(Date.now() + 600000).toISOString(),
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(session));

      const result = authService.getPermissions();

      expect(result).toEqual(['read', 'write', 'delete', 'admin']);
    });

    test('should return empty array if no session', () => {
      mockSessionStorage.getItem.mockReturnValue(null);

      const result = authService.getPermissions();

      expect(result).toEqual([]);
    });

    test('should check specific permission', () => {
      const session = {
        session_id: 'session-123',
        role: 'admin',
        permissions: ['read', 'write', 'delete', 'admin'],
        expires_at: new Date(Date.now() + 600000).toISOString(),
      };
      mockSessionStorage.getItem.mockReturnValue(JSON.stringify(session));

      expect(authService.hasPermission('read')).toBe(true);
      expect(authService.hasPermission('write')).toBe(true);
      expect(authService.hasPermission('delete')).toBe(true);
      expect(authService.hasPermission('admin')).toBe(true);
      expect(authService.hasPermission('nonexistent')).toBe(false);
    });

    test('should return false for permission if no session', () => {
      mockSessionStorage.getItem.mockReturnValue(null);

      expect(authService.hasPermission('read')).toBe(false);
    });
  });

  describe('Token refresh', () => {
    test('should refresh token successfully', async () => {
      const token = 'valid-token';
      const authResult = MockDataFactory.getAuthenticateResult();

      mockSessionStorage.getItem.mockReturnValue(token);
      mockWebSocketService.sendRPC.mockResolvedValue(authResult);

      await authService.refreshToken();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('authenticate', {
        auth_token: token,
      });
    });

    test('should logout on refresh failure', async () => {
      const token = 'invalid-token';

      mockSessionStorage.getItem.mockReturnValue(token);
      mockWebSocketService.sendRPC.mockRejectedValue(new Error('Auth failed'));

      await expect(authService.refreshToken()).rejects.toThrow('Auth failed');
      expect(mockSessionStorage.removeItem).toHaveBeenCalledWith('auth_token');
      expect(mockSessionStorage.removeItem).toHaveBeenCalledWith('auth_session');
    });

    test('should throw error if no stored token', async () => {
      mockSessionStorage.getItem.mockReturnValue(null);

      await expect(authService.refreshToken()).rejects.toThrow(
        'No stored token to refresh'
      );
    });
  });

  describe('REQ-AUTH-006: Parameter validation', () => {
    test('should validate JWT token format', async () => {
      const invalidToken = 'invalid-token-format';
      const errorResponse = {
        code: -32001,
        message: 'Invalid token format',
        data: { reason: 'Invalid token format' }
      };

      mockWebSocketService.sendRPC.mockRejectedValue(new Error('Invalid token format'));

      await expect(authService.authenticate(invalidToken)).rejects.toThrow('Invalid token format');
    });

    test('should validate role permissions', () => {
      const adminResult = MockDataFactory.getAuthenticateResult();
      const viewerResult = MockDataFactory.getAuthenticateResult();

      expect(adminResult.permissions).toContain('view');
      expect(adminResult.permissions).toContain('control');
      expect(viewerResult.permissions).toContain('view');
      expect(viewerResult.permissions).toContain('control');
    });
  });
});
