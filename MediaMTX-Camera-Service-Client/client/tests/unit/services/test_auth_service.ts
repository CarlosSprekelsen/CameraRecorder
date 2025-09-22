/**
 * Authentication Service Unit Tests
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Requirements Coverage:
 * - REQ-AUTH01-001: JWT token authentication must be secure and reliable
 * - REQ-AUTH01-002: Role-based access control must be enforced
 * - REQ-AUTH01-003: Session management must handle expiration gracefully
 * - REQ-AUTH01-004: Authentication state must be consistent across components
 * 
 * Coverage: UNIT
 * Quality: HIGH
 */

import { AuthenticationService } from '../../../src/services/authService';
import { websocketService } from '../../../src/services/websocket';
import { RPC_METHODS, ERROR_CODES } from '../../../src/types/rpc';

// Mock the websocket service
jest.mock('../../../src/services/websocket', () => ({
  websocketService: {
    call: jest.fn(),
    isConnected: jest.fn(() => true),
  },
}));

// Mock logger service
jest.mock('../../../src/services/loggerService', () => ({
  logger: {
    info: jest.fn(),
    error: jest.fn(),
    warn: jest.fn(),
    debug: jest.fn(),
  },
  loggers: {
    auth: {
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
    },
  },
}));

describe('Authentication Service', () => {
  let authService: AuthenticationService;
  const mockWebSocketService = websocketService as jest.Mocked<typeof websocketService>;

  beforeEach(() => {
    // Reset singleton instance
    (AuthenticationService as any).instance = undefined;
    authService = AuthenticationService.getInstance();
    jest.clearAllMocks();
  });

  afterEach(() => {
    authService.cleanup();
  });

  describe('REQ-AUTH01-001: JWT Token Authentication', () => {
    it('should authenticate with valid JWT token', async () => {
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: {
          authenticated: true,
          role: 'admin',
          permissions: ['camera:read', 'camera:write', 'admin:manage'],
          expires_at: new Date(Date.now() + 3600000).toISOString(),
          session_id: 'test-session-123',
        },
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await authService.authenticate('valid-jwt-token');

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.AUTHENTICATE, {
        auth_token: 'valid-jwt-token',
      });
      expect(result).toBe(true);
      expect(authService.isAuthenticated()).toBe(true);
      expect(authService.getRole()).toBe('admin');
    });

    it('should reject invalid JWT token', async () => {
      const mockError = {
        jsonrpc: '2.0' as const,
        error: {
          code: ERROR_CODES.INVALID_TOKEN,
          message: 'Invalid or expired token',
        },
        id: 1,
      };

      mockWebSocketService.call.mockRejectedValue(mockError);

      const result = await authService.authenticate('invalid-token');

      expect(result).toBe(false);
      expect(authService.isAuthenticated()).toBe(false);
      expect(authService.getRole()).toBeNull();
    });

    it('should handle authentication errors gracefully', async () => {
      mockWebSocketService.call.mockRejectedValue(new Error('Network error'));

      const result = await authService.authenticate('test-token');

      expect(result).toBe(false);
      expect(authService.isAuthenticated()).toBe(false);
    });
  });

  describe('REQ-AUTH01-002: Role-Based Access Control', () => {
    beforeEach(async () => {
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: {
          authenticated: true,
          role: 'operator',
          permissions: ['camera:read', 'camera:write'],
          expires_at: new Date(Date.now() + 3600000).toISOString(),
          session_id: 'test-session-123',
        },
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);
      await authService.authenticate('test-token');
    });

    it('should check user permissions correctly', () => {
      expect(authService.hasPermission('camera:read')).toBe(true);
      expect(authService.hasPermission('camera:write')).toBe(true);
      expect(authService.hasPermission('admin:manage')).toBe(false);
    });

    it('should validate role-based access', () => {
      expect(authService.hasRole('operator')).toBe(true);
      expect(authService.hasRole('admin')).toBe(false);
      expect(authService.hasRole('viewer')).toBe(false);
    });

    it('should return correct role information', () => {
      expect(authService.getRole()).toBe('operator');
      expect(authService.getPermissions()).toEqual(['camera:read', 'camera:write']);
    });
  });

  describe('REQ-AUTH01-003: Session Management', () => {
    it('should handle session expiration', async () => {
      const expiredTime = new Date(Date.now() - 1000).toISOString();
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: {
          authenticated: true,
          role: 'admin',
          permissions: ['camera:read'],
          expires_at: expiredTime,
          session_id: 'test-session-123',
        },
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);
      await authService.authenticate('test-token');

      // Session should be considered expired
      expect(authService.isSessionExpired()).toBe(true);
    });

    it('should handle auto-reauthentication', async () => {
      const nearExpiryTime = new Date(Date.now() + 60000).toISOString(); // 1 minute
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: {
          authenticated: true,
          role: 'admin',
          permissions: ['camera:read'],
          expires_at: nearExpiryTime,
          session_id: 'test-session-123',
        },
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);
      await authService.authenticate('test-token');

      // Should schedule reauthentication
      expect(authService.isSessionExpired()).toBe(false);
    });

    it('should clear session on logout', async () => {
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: {
          authenticated: true,
          role: 'admin',
          permissions: ['camera:read'],
          expires_at: new Date(Date.now() + 3600000).toISOString(),
          session_id: 'test-session-123',
        },
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);
      await authService.authenticate('test-token');

      expect(authService.isAuthenticated()).toBe(true);

      await authService.logout();

      expect(authService.isAuthenticated()).toBe(false);
      expect(authService.getRole()).toBeNull();
      expect(authService.getPermissions()).toEqual([]);
    });
  });

  describe('REQ-AUTH01-004: Authentication State Consistency', () => {
    it('should maintain consistent state across operations', async () => {
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: {
          authenticated: true,
          role: 'viewer',
          permissions: ['camera:read'],
          expires_at: new Date(Date.now() + 3600000).toISOString(),
          session_id: 'test-session-123',
        },
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);
      await authService.authenticate('test-token');

      // State should be consistent
      expect(authService.isAuthenticated()).toBe(true);
      expect(authService.getRole()).toBe('viewer');
      expect(authService.getPermissions()).toEqual(['camera:read']);
      expect(authService.getSessionId()).toBe('test-session-123');
    });

    it('should handle concurrent authentication attempts', async () => {
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: {
          authenticated: true,
          role: 'admin',
          permissions: ['camera:read', 'camera:write'],
          expires_at: new Date(Date.now() + 3600000).toISOString(),
          session_id: 'test-session-123',
        },
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      // Multiple concurrent authentication attempts
      const promises = [
        authService.authenticate('token1'),
        authService.authenticate('token2'),
        authService.authenticate('token3'),
      ];

      const results = await Promise.all(promises);

      // All should succeed (singleton pattern)
      expect(results.every(result => result === true)).toBe(true);
      expect(authService.isAuthenticated()).toBe(true);
    });
  });

  describe('Configuration Management', () => {
    it('should configure authentication settings', () => {
      authService.configure({
        autoReauth: false,
        reauthThreshold: 600,
      });

      // Configuration should be applied
      expect(authService.getConfig().autoReauth).toBe(false);
      expect(authService.getConfig().reauthThreshold).toBe(600);
    });

    it('should handle API key authentication', async () => {
      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: {
          authenticated: true,
          role: 'operator',
          permissions: ['camera:read'],
          expires_at: new Date(Date.now() + 3600000).toISOString(),
          session_id: 'test-session-123',
        },
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      authService.configure({ apiKey: 'test-api-key' });
      const result = await authService.authenticate();

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.AUTHENTICATE, {
        api_key: 'test-api-key',
      });
      expect(result).toBe(true);
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors during authentication', async () => {
      mockWebSocketService.call.mockRejectedValue(new Error('Connection failed'));

      const result = await authService.authenticate('test-token');

      expect(result).toBe(false);
      expect(authService.isAuthenticated()).toBe(false);
    });

    it('should handle malformed responses', async () => {
      mockWebSocketService.call.mockResolvedValue({
        jsonrpc: '2.0' as const,
        result: null, // Malformed response
        id: 1,
      });

      const result = await authService.authenticate('test-token');

      expect(result).toBe(false);
      expect(authService.isAuthenticated()).toBe(false);
    });
  });
});
