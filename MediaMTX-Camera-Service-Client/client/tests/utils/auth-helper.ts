/**
 * Authentication Helper - THE ONE AND ONLY APPROACH
 * 
 * SINGLE authentication utility for all tests.
 * Architecture compliance: Uses AuthService.authenticate() which calls APIClient.call() which calls WebSocketService.sendRPC()
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-AUTH-001: JWT token generation
 * - REQ-AUTH-002: Role-based access control  
 * - REQ-AUTH-003: Session management (server-managed)
 * 
 * Test Categories: Unit/Integration/Security
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { AuthService } from '../../src/services/auth/AuthService';
import { APIClient } from '../../src/services/abstraction/APIClient';
import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../src/services/logger/LoggerService';
import { AuthResult } from '../../src/types/api';
import jwt from 'jsonwebtoken';

export type UserRole = 'admin' | 'operator' | 'viewer';

export interface UnifiedAuthConfig {
  serverUrl: string;
  timeout?: number;
}

export class AuthHelper {
  private authService: AuthService;
  private apiClient: APIClient;
  private wsService: WebSocketService;
  private logger: LoggerService;

  constructor(config: UnifiedAuthConfig) {
    // Initialize services following architectural hierarchy
    this.logger = new LoggerService();
    this.wsService = new WebSocketService({ url: config.serverUrl });
    this.apiClient = new APIClient(this.wsService, this.logger);
    this.authService = new AuthService(this.apiClient, this.logger);
  }

  // ===== STATIC UTILITY METHODS (from old auth-helper.ts) =====

  /**
   * Generate test JWT token with specified role
   * CRITICAL: Use pre-generated tokens from environment instead of generating new ones
   */
  static generateTestToken(role: UserRole = 'admin'): string {
    const tokenKey = `TEST_${role.toUpperCase()}_TOKEN`;
    const token = process.env[tokenKey];
    
    if (!token) {
      throw new Error(`Missing environment token: ${tokenKey}. Run reinstall-with-tokens.sh to generate tokens.`);
    }
    
    return token;
  }

  /**
   * Validate authentication result against documented schema
   */
  static validateAuthResult(result: any): result is AuthResult {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.authenticated === 'boolean' &&
      typeof result.role === 'string' &&
      ['admin', 'operator', 'viewer'].includes(result.role)
    );
  }

  /**
   * Validate JWT token structure
   */
  static validateTokenStructure(token: string): boolean {
    try {
      const decoded = jwt.decode(token) as any;
      return (
        decoded &&
        typeof decoded.sub === 'string' &&
        typeof decoded.role === 'string' &&
        typeof decoded.exp === 'number' &&
        typeof decoded.iat === 'number'
      );
    } catch {
      return false;
    }
  }

  /**
   * Extract role from token
   */
  static getTokenRole(token: string): UserRole | null {
    try {
      const decoded = jwt.decode(token) as any;
      return decoded?.role || null;
    } catch {
      return null;
    }
  }

  /**
   * Check if token is expired
   */
  static isTokenExpired(token: string): boolean {
    try {
      const decoded = jwt.decode(token) as any;
      return decoded.exp < Math.floor(Date.now() / 1000);
    } catch {
      return true;
    }
  }

  /**
   * THE ONE AND ONLY authentication method
   * Architecture: AuthService.authenticate() -> APIClient.call() -> WebSocketService.sendRPC()
   */
  async authenticateWithToken(token: string): Promise<AuthResult> {
    try {
      // Ensure connection
      if (!this.wsService.isConnected()) {
        await this.wsService.connect();
      }

      // Use the architectural standard: AuthService.authenticate()
      const result = await this.authService.authenticate(token);
      
      return result;
    } catch (error) {
      return {
        authenticated: false,
        role: 'viewer',
        userId: '',
        session_id: '',
        permissions: []
      };
    }
  }

  /**
   * Get authenticated services for use in tests
   */
  getAuthenticatedServices() {
    return {
      authService: this.authService,
      apiClient: this.apiClient,
      wsService: this.wsService,
      logger: this.logger
    };
  }

  /**
   * Cleanup resources
   */
  async disconnect(): Promise<void> {
    if (this.wsService.isConnected()) {
      await this.wsService.disconnect();
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.wsService.isConnected();
  }
}

/**
 * Factory function for easy test setup
 */
export async function createAuthenticatedTestEnvironment(
  serverUrl: string = 'ws://localhost:8002/ws',
  token: string = process.env.TEST_ADMIN_TOKEN || ''
): Promise<AuthHelper> {
  if (!token) {
    throw new Error('No authentication token provided. Set TEST_ADMIN_TOKEN environment variable.');
  }

  const authHelper = new AuthHelper({ serverUrl });
  
  const authResult = await authHelper.authenticateWithToken(token);
  if (!authResult.authenticated) {
    throw new Error(`Authentication failed: ${authResult.error}`);
  }

  return authHelper;
}
