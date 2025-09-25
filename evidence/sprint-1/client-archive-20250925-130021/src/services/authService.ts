/**
 * Authentication Service
 * Handles JWT token and API key authentication with Go server
 */

import type { WebSocketService } from './websocket';
import { RPC_METHODS, ERROR_CODES } from '../types/rpc';
import type { AuthenticationParams, AuthenticationResponse } from '../types/rpc';
import { logger, loggers } from './loggerService';

export interface AuthConfig {
  jwtToken?: string;
  apiKey?: string;
  autoReauth?: boolean;
  reauthThreshold?: number; // seconds before expiry to reauth
}

export interface AuthState {
  isAuthenticated: boolean;
  role: 'admin' | 'operator' | 'viewer' | null;
  permissions: string[];
  expiresAt: Date | null;
  sessionId: string | null;
  token: string | null;
}

export class AuthenticationService {
  private static instance: AuthenticationService;
  private config: AuthConfig;
  private authState: AuthState;
  private reauthTimer: NodeJS.Timeout | null = null;
  private wsService: WebSocketService | null = null;

  private constructor() {
    this.config = {
      autoReauth: true,
      reauthThreshold: 300, // 5 minutes before expiry
    };
    this.authState = {
      isAuthenticated: false,
      role: null,
      permissions: [],
      expiresAt: null,
      sessionId: null,
      token: null,
    };
  }

  static getInstance(): AuthenticationService {
    if (!AuthenticationService.instance) {
      AuthenticationService.instance = new AuthenticationService();
    }
    return AuthenticationService.instance;
  }

  /**
   * Set WebSocket service reference
   */
  public setWebSocketService(wsService: WebSocketService): void {
    this.wsService = wsService;
  }

  /**
   * Initialize authentication with token or API key
   */
  public async initialize(config: AuthConfig): Promise<void> {
    this.config = { ...this.config, ...config };
    
    if (!this.config.jwtToken && !this.config.apiKey) {
      throw new Error('Either JWT token or API key must be provided');
    }

    const token = this.config.jwtToken || this.config.apiKey!;
    await this.authenticate(token);
  }

  /**
   * Authenticate with the server
   */
  public async authenticate(token: string): Promise<AuthenticationResponse> {
    loggers.service.start('AuthenticationService', 'authenticate');

    if (!this.wsService) {
      throw new Error('WebSocket service not available');
    }

    try {
      const params: AuthenticationParams = { auth_token: token };
      const response = await this.wsService.call(
        RPC_METHODS.AUTHENTICATE,
        params
      ) as AuthenticationResponse;

      // Update auth state
      this.authState = {
        isAuthenticated: response.authenticated,
        role: response.role,
        permissions: response.permissions,
        expiresAt: new Date(response.expires_at),
        sessionId: response.session_id,
        token: token,
      };

      // Set up auto re-authentication if enabled
      if (this.config.autoReauth && response.authenticated) {
        this.scheduleReauth();
      }

      loggers.service.success('AuthenticationService', 'authenticate', {
        role: response.role,
        expiresAt: response.expires_at,
      });

      return response;
    } catch (error) {
      loggers.service.error('AuthenticationService', 'authenticate', error as Error);
      throw error;
    }
  }

  /**
   * Get current authentication state
   */
  public getAuthState(): AuthState {
    return { ...this.authState };
  }

  /**
   * Check if user has required permission
   */
  public hasPermission(permission: string): boolean {
    return this.authState.isAuthenticated && 
           this.authState.permissions.includes(permission);
  }

  /**
   * Check if user has required role or higher
   */
  public hasRole(requiredRole: 'viewer' | 'operator' | 'admin'): boolean {
    if (!this.authState.isAuthenticated || !this.authState.role) {
      return false;
    }

    const roleHierarchy = { viewer: 1, operator: 2, admin: 3 };
    const userLevel = roleHierarchy[this.authState.role];
    const requiredLevel = roleHierarchy[requiredRole];

    return userLevel >= requiredLevel;
  }

  /**
   * Check if authentication is still valid
   */
  public isAuthenticated(): boolean {
    if (!this.authState.isAuthenticated || !this.authState.expiresAt) {
      return false;
    }

    // Check if token is expired
    const now = new Date();
    const timeUntilExpiry = this.authState.expiresAt.getTime() - now.getTime();
    
    return timeUntilExpiry > 0;
  }

  /**
   * Get auth token for API calls
   */
  public getAuthToken(): string | null {
    return this.authState.token;
  }

  /**
   * Add auth token to request parameters
   */
  public addAuthToParams(params: Record<string, unknown> = {}): Record<string, unknown> {
    const token = this.getAuthToken();
    if (token) {
      return { ...params, auth_token: token };
    }
    return params;
  }

  /**
   * Logout and clear authentication state
   */
  public logout(): void {
    loggers.service.info('AuthenticationService', 'logout');
    
    this.authState = {
      isAuthenticated: false,
      role: null,
      permissions: [],
      expiresAt: null,
      sessionId: null,
      token: null,
    };

    if (this.reauthTimer) {
      clearTimeout(this.reauthTimer);
      this.reauthTimer = null;
    }
  }

  /**
   * Schedule automatic re-authentication
   */
  private scheduleReauth(): void {
    if (!this.authState.expiresAt || !this.config.autoReauth) {
      return;
    }

    const now = new Date();
    const timeUntilExpiry = this.authState.expiresAt.getTime() - now.getTime();
    const reauthTime = timeUntilExpiry - (this.config.reauthThreshold! * 1000);

    if (reauthTime > 0) {
      this.reauthTimer = setTimeout(async () => {
        try {
          if (this.authState.token) {
            await this.authenticate(this.authState.token);
          }
        } catch (error) {
          loggers.service.error('AuthenticationService', 'autoReauth', error as Error);
          this.logout();
        }
      }, reauthTime);

      loggers.service.info('AuthenticationService', 'scheduleReauth', {
        reauthIn: Math.round(reauthTime / 1000),
        expiresAt: this.authState.expiresAt.toISOString(),
      });
    }
  }

  /**
   * Handle authentication errors
   */
  public handleAuthError(error: any): boolean {
    if (error?.code === ERROR_CODES.AUTHENTICATION_FAILED) {
      loggers.service.warn('AuthenticationService', 'handleAuthError', 'Authentication failed, logging out');
      this.logout();
      return true; // Error was handled
    }
    return false; // Error was not handled
  }
}

// Export singleton instance
export const authService = AuthenticationService.getInstance();

// Export types
export type { AuthConfig, AuthState, AuthenticationResponse };