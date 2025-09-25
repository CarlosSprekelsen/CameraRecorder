import { AuthenticateParams, AuthenticateResult } from '../../types/api';
import { WebSocketService } from '../websocket/WebSocketService';

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
  private wsService: WebSocketService;

  constructor(wsService: WebSocketService) {
    this.wsService = wsService;
  }

  async authenticate(token: string): Promise<AuthenticateResult> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    const params: AuthenticateParams = { auth_token: token };
    const result = await this.wsService.sendRPC<AuthenticateResult>('authenticate', params);

    // Store token in session storage
    if (result.authenticated) {
      sessionStorage.setItem('auth_token', token);
      sessionStorage.setItem(
        'auth_session',
        JSON.stringify({
          session_id: result.session_id,
          role: result.role,
          permissions: result.permissions,
          expires_at: result.expires_at,
        }),
      );
    }

    return result;
  }

  async refreshToken(): Promise<void> {
    const token = this.getStoredToken();
    if (!token) {
      throw new Error('No stored token to refresh');
    }

    try {
      await this.authenticate(token);
    } catch (error) {
      this.logout();
      throw error;
    }
  }

  logout(): void {
    sessionStorage.removeItem('auth_token');
    sessionStorage.removeItem('auth_session');
  }

  getStoredToken(): string | null {
    return sessionStorage.getItem('auth_token');
  }

  getStoredSession(): {
    session_id: string;
    role: string;
    permissions: string[];
    expires_at: string;
  } | null {
    const session = sessionStorage.getItem('auth_session');
    if (!session) return null;

    try {
      return JSON.parse(session);
    } catch {
      return null;
    }
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
