/**
 * AuthService unit tests - aligned with refactored architecture
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-AUTH-001: Authentication with JWT tokens
 * - REQ-AUTH-002: Session management (server-managed)
 * - REQ-AUTH-003: Token validation (server-managed)
 * - REQ-AUTH-004: Role-based access control (server-managed)
 * - REQ-AUTH-005: Permission checking (server-managed)
 * 
 * Test Categories: Unit
 * API Documentation Reference: ../mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

import { AuthService } from '../../../src/services/auth/AuthService';
import { IAPIClient } from '../../../src/services/abstraction/IAPIClient';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Use centralized mocks - aligned with refactored architecture
const mockAPIClient = MockDataFactory.createMockAPIClient();
const mockLoggerService = MockDataFactory.createMockLoggerService();

describe('AuthService Unit Tests', () => {
  let authService: AuthService;

  beforeEach(() => {
    jest.clearAllMocks();
    // Mock IAPIClient to be connected
    (mockAPIClient.isConnected as jest.Mock).mockReturnValue(true);
    authService = new AuthService(mockAPIClient, mockLoggerService);
  });

  describe('REQ-AUTH-001: Authentication with JWT tokens', () => {
    test('should authenticate successfully with valid token', async () => {
      const token = 'valid-jwt-token';
      const expectedResult = MockDataFactory.getAuthenticateResult();

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      const result = await authService.authenticate(token);

      expect(mockAPIClient.call).toHaveBeenCalledWith('authenticate', {
        auth_token: token,
      });
      expect(result).toEqual(expectedResult);
      // Architecture requirement: Server manages all authentication state - no client storage
    });

    test('should handle authentication failure', async () => {
      const token = 'invalid-token';
      const authResult = { ...MockDataFactory.getAuthenticateResult(), authenticated: false };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(authResult);

      const result = await authService.authenticate(token);

      expect(result.authenticated).toBe(false);
      // Architecture requirement: Server manages all authentication state - no client storage
    });

    test('should throw error when IAPIClient not connected', async () => {
      (mockAPIClient.isConnected as jest.Mock).mockReturnValue(false);

      await expect(authService.authenticate('token')).rejects.toThrow(
        'WebSocket not connected'
      );
    });

    test('should validate authentication result', async () => {
      const token = 'valid-jwt-token';
      const expectedResult = MockDataFactory.getAuthenticateResult();

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      const result = await authService.authenticate(token);

      expect(APIResponseValidator.validateAuthenticateResult(result)).toBe(true);
    });
  });

  describe('REQ-AUTH-002: Session management (server-managed)', () => {
    test('should authenticate without client-side storage', async () => {
      const token = 'valid-jwt-token';
      const authResult = MockDataFactory.getAuthenticateResult();

      (mockAPIClient.call as jest.Mock).mockResolvedValue(authResult);

      const result = await authService.authenticate(token);

      expect(result).toEqual(authResult);
      // Architecture requirement: Server manages all authentication state - no client storage
    });

    test('should logout without client-side cleanup', () => {
      authService.logout();
      // Architecture requirement: Server manages all authentication state - no client storage to clear
    });
  });

  describe('REQ-AUTH-003: Token validation (server-managed)', () => {
    test('should authenticate with valid token', async () => {
      const token = 'valid-jwt-token';
      const authResult = MockDataFactory.getAuthenticateResult();

      (mockAPIClient.call as jest.Mock).mockResolvedValue(authResult);

      const result = await authService.authenticate(token);

      expect(result.authenticated).toBe(true);
      // Architecture requirement: Server manages token validation and expiration
    });
  });

  describe('REQ-AUTH-004: Role-based access control (server-managed)', () => {
    test('should authenticate and return role from server', async () => {
      const token = 'valid-jwt-token';
      const authResult = MockDataFactory.getAuthenticateResult();

      (mockAPIClient.call as jest.Mock).mockResolvedValue(authResult);

      const result = await authService.authenticate(token);

      expect(result.role).toBe(authResult.role);
      expect(result.permissions).toEqual(authResult.permissions);
      // Architecture requirement: Server manages all role and permission state
    });
  });

  describe('REQ-AUTH-005: Permission checking (server-managed)', () => {
    test('should return permissions from server authentication', async () => {
      const token = 'valid-jwt-token';
      const authResult = MockDataFactory.getAuthenticateResult();

      (mockAPIClient.call as jest.Mock).mockResolvedValue(authResult);

      const result = await authService.authenticate(token);

      expect(result.permissions).toEqual(authResult.permissions);
      // Architecture requirement: Server manages all permission state
    });
  });
});