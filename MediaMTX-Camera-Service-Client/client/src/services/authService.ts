/**
 * Authentication Service for MediaMTX Camera Service Client
 * Implements JWT authentication flow based on server implementation
 * 
 * Authentication Flow:
 * 1. Client obtains JWT token externally (admin/configuration)
 * 2. Client authenticates WebSocket connection using authenticate method
 * 3. Client includes JWT in all protected JSON-RPC calls
 * 4. Client handles token refresh/expiry
 */

import type { AuthenticateParams, AuthenticateResponse } from '../types/camera';
import type { JSONRPCRequest, JSONRPCResponse } from '../types/rpc';
import { RPC_METHODS } from '../types/rpc';

/**
 * Login credentials interface
 */
export interface LoginCredentials {
  token: string;
  auth_type?: 'jwt' | 'api_key' | 'auto';
}

/**
 * Authentication state interface
 */
export interface AuthState {
  authenticated: boolean;
  user_id?: string;
  role?: string;
  auth_method?: string;
  token?: string;
  expires_at?: number;
}

/**
 * Authentication service class
 * Handles JWT authentication flow for MediaMTX Camera Service
 */
export class AuthService {
  private authState: AuthState = {
    authenticated: false
  };
  
  private tokenRefreshTimer?: NodeJS.Timeout;
  private readonly tokenRefreshThreshold = 5 * 60 * 1000; // 5 minutes before expiry
  private readonly maxTokenAge = 24 * 60 * 60 * 1000; // 24 hours max

  /**
   * Login with JWT token
   * @param credentials Login credentials containing JWT token
   * @returns Promise<string> JWT token if authentication successful
   */
  async login(credentials: LoginCredentials): Promise<string> {
    try {
      // Store token for authentication
      this.authState.token = credentials.token;
      
      // Validate token format (basic JWT structure check)
      if (!this.isValidJWTFormat(credentials.token)) {
        throw new Error('Invalid JWT token format');
      }

      // Check if token is expired
      if (this.isTokenExpired(credentials.token)) {
        throw new Error('JWT token is expired');
      }

      // Set up token refresh timer
      this.setupTokenRefresh(credentials.token);
      
      return credentials.token;
    } catch (error) {
      this.clearAuthState();
      throw error;
    }
  }

  /**
   * Include authentication token in JSON-RPC parameters
   * @param params Original JSON-RPC parameters
   * @returns Parameters with authentication token included
   */
  includeAuth(params: any = {}): any {
    if (!this.authState.authenticated || !this.authState.token) {
      throw new Error('Not authenticated. Call login() first.');
    }

    // Add auth_token to parameters for protected methods
    return {
      ...params,
      auth_token: this.authState.token
    };
  }

  /**
   * Handle token expiry by attempting to refresh or re-authenticate
   * @returns Promise<void>
   */
  async handleTokenExpiry(): Promise<void> {
    if (!this.authState.token) {
      throw new Error('No token available for refresh');
    }

    // Check if token is expired or about to expire
    if (this.isTokenExpired(this.authState.token) || this.isTokenExpiringSoon(this.authState.token)) {
      // For now, we require manual token refresh since server doesn't provide refresh endpoint
      // In production, this would typically call a refresh endpoint
      throw new Error('Token expired. Please obtain a new token and call login() again.');
    }
  }

  /**
   * Authenticate WebSocket connection using server's authenticate method
   * @param sendRequest Function to send JSON-RPC request
   * @returns Promise<AuthenticateResponse> Authentication result
   */
  async authenticateConnection(
    sendRequest: (request: JSONRPCRequest) => Promise<JSONRPCResponse>
  ): Promise<AuthenticateResponse> {
    if (!this.authState.token) {
      throw new Error('No token available for authentication');
    }

    const authParams: AuthenticateParams = {
      token: this.authState.token
    };

    const request: JSONRPCRequest = {
      jsonrpc: '2.0',
      method: RPC_METHODS.AUTHENTICATE,
      params: authParams as Record<string, unknown>,
      id: Date.now()
    };

    try {
      const response = await sendRequest(request);
      
      if (response.error) {
        throw new Error(`Authentication failed: ${response.error.message}`);
      }

      const result = response.result as AuthenticateResponse;
      
      if (result.authenticated) {
        // Update authentication state
        this.authState.authenticated = true;
        this.authState.user_id = result.user_id;
        this.authState.role = result.role;
        this.authState.auth_method = result.auth_method;
        
        // Set up token refresh timer
        this.setupTokenRefresh(this.authState.token!);
      } else {
        this.clearAuthState();
      }

      return result;
    } catch (error) {
      this.clearAuthState();
      throw error;
    }
  }

  /**
   * Check if user has required role permission
   * @param requiredRole Minimum required role
   * @returns boolean True if user has permission
   */
  hasPermission(requiredRole: string): boolean {
    if (!this.authState.authenticated || !this.authState.role) {
      return false;
    }

    const roleHierarchy = {
      'viewer': 1,
      'operator': 2,
      'admin': 3
    };

    const userLevel = roleHierarchy[this.authState.role as keyof typeof roleHierarchy] || 0;
    const requiredLevel = roleHierarchy[requiredRole as keyof typeof roleHierarchy] || 0;

    return userLevel >= requiredLevel;
  }

  /**
   * Get current authentication state
   * @returns AuthState Current authentication state
   */
  getAuthState(): AuthState {
    return { ...this.authState };
  }

  /**
   * Logout and clear authentication state
   */
  logout(): void {
    this.clearAuthState();
  }

  /**
   * Check if JWT token format is valid
   * @param token JWT token string
   * @returns boolean True if format is valid
   */
  private isValidJWTFormat(token: string): boolean {
    // Basic JWT format validation (3 parts separated by dots)
    const parts = token.split('.');
    return parts.length === 3 && parts.every(part => part.length > 0);
  }

  /**
   * Check if JWT token is expired
   * @param token JWT token string
   * @returns boolean True if token is expired
   */
  private isTokenExpired(token: string): boolean {
    try {
      const payload = this.decodeJWTPayload(token);
      const exp = payload.exp;
      
      if (!exp) {
        return true; // No expiry claim means expired
      }

      const currentTime = Math.floor(Date.now() / 1000);
      return currentTime > exp;
    } catch (error) {
      return true; // Invalid token considered expired
    }
  }

  /**
   * Check if JWT token is expiring soon
   * @param token JWT token string
   * @returns boolean True if token expires within threshold
   */
  private isTokenExpiringSoon(token: string): boolean {
    try {
      const payload = this.decodeJWTPayload(token);
      const exp = payload.exp;
      
      if (!exp) {
        return true;
      }

      const currentTime = Math.floor(Date.now() / 1000);
      const threshold = Math.floor(this.tokenRefreshThreshold / 1000);
      
      return (exp - currentTime) < threshold;
    } catch (error) {
      return true;
    }
  }

  /**
   * Decode JWT payload without verification (for expiry checking)
   * @param token JWT token string
   * @returns any Decoded payload
   */
  private decodeJWTPayload(token: string): any {
    try {
      const parts = token.split('.');
      if (parts.length !== 3) {
        throw new Error('Invalid JWT format');
      }

      const payload = parts[1];
      const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
      return JSON.parse(decoded);
    } catch (error) {
      throw new Error('Failed to decode JWT payload');
    }
  }

  /**
   * Set up token refresh timer
   * @param token JWT token string
   */
  private setupTokenRefresh(token: string): void {
    // Clear existing timer
    if (this.tokenRefreshTimer) {
      clearTimeout(this.tokenRefreshTimer);
    }

    try {
      const payload = this.decodeJWTPayload(token);
      const exp = payload.exp;
      
      if (exp) {
        const currentTime = Math.floor(Date.now() / 1000);
        const timeUntilExpiry = (exp - currentTime) * 1000; // Convert to milliseconds
        const timeUntilRefresh = Math.max(timeUntilExpiry - this.tokenRefreshThreshold, 0);

        // Set timer to refresh token before expiry
        this.tokenRefreshTimer = setTimeout(() => {
          this.handleTokenExpiry().catch(error => {
            console.warn('Token refresh failed:', error.message);
          });
        }, timeUntilRefresh);
      }
    } catch (error) {
      console.warn('Failed to setup token refresh:', error);
    }
  }

  /**
   * Clear authentication state and timers
   */
  private clearAuthState(): void {
    if (this.tokenRefreshTimer) {
      clearTimeout(this.tokenRefreshTimer);
      this.tokenRefreshTimer = undefined;
    }

    this.authState = {
      authenticated: false
    };
  }
}

/**
 * Singleton instance of authentication service
 */
export const authService = new AuthService();

/**
 * Export types for external use
 */
// Types already exported above
