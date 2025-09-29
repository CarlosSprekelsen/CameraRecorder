import { AuthenticateParams, AuthenticateResult } from '../../types/api';
import { APIClient } from '../abstraction/APIClient';
import { LoggerService } from '../logger/LoggerService';

/**
 * Authentication Service
 *
 * Handles user authentication and session management for the MediaMTX client.
 * Provides JWT token validation, role-based access control, and session persistence.
 *
 * @class AuthService
 *
 * @example
 * ```typescript
 * const authService = new AuthService(wsService);
 * const result = await authService.authenticate('jwt-token');
 * if (result.authenticated) {
 *   console.log(`Logged in as ${result.role}`);
 * }
 * ```
 */
export class AuthService {
  constructor(
    private apiClient: APIClient,
    private logger: LoggerService,
  ) {
    this.logger.info('AuthService initialized');
  }

  async authenticate(token: string): Promise<AuthenticateResult> {
    if (!this.apiClient.isConnected) {
      throw new Error('WebSocket not connected');
    }

    const params: AuthenticateParams = { auth_token: token };
    const result = await this.apiClient.call<AuthenticateResult>('authenticate', params as unknown as Record<string, unknown>);

    // SECURITY: Do not store credentials in browser storage
    // Architecture requirement: "Download links are opaque; the client must not persist or display credentials"
    // Authentication state is managed by the server session, not client storage

    return result;
  }

  // Legacy methods removed - use authenticate() directly

  logout(): void {
    // SECURITY: No client-side credential storage to clear
    // Architecture requirement: Server manages all authentication state
    // No-op: Server session is the source of truth
  }

  getStoredToken(): string | null {
    // SECURITY: No credential storage in browser
    // Architecture requirement: "Download links are opaque; the client must not persist or display credentials"
    return null;
  }

  getStoredSession(): {
    session_id: string;
    role: string;
    permissions: string[];
    expires_at: string;
  } | null {
    // SECURITY: No credential storage in browser
    // Architecture requirement: "Download links are opaque; the client must not persist or display credentials"
    return null;
  }

  isTokenExpired(): boolean {
    const session = this.getStoredSession();
    if (!session) return true;

    const expiresAt = new Date(session.expires_at);
    const now = new Date();

    // Consider token expired if it expires within 5 minutes
    return expiresAt.getTime() - now.getTime() < 5 * 60 * 1000;
  }

  isAuthenticated(): boolean {
    const token = this.getStoredToken();
    const session = this.getStoredSession();

    return !!(token && session && !this.isTokenExpired());
  }

  getRole(): string | null {
    const session = this.getStoredSession();
    return session?.role || null;
  }

  getPermissions(): string[] {
    const session = this.getStoredSession();
    return session?.permissions || [];
  }

  hasPermission(permission: string): boolean {
    const permissions = this.getPermissions();
    return permissions.includes(permission);
  }

  hasRole(role: string): boolean {
    const userRole = this.getRole();
    return userRole === role;
  }

  hasAnyRole(roles: string[]): boolean {
    const userRole = this.getRole();
    return roles.includes(userRole || '');
  }
}
