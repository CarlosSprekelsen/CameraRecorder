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

export interface UnifiedAuthConfig {
  serverUrl: string;
  timeout?: number;
}

export interface AuthResult {
  authenticated: boolean;
  role?: string;
  userId?: string;
  error?: string;
}

export class UnifiedAuthHelper {
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
      
      return {
        authenticated: result.authenticated,
        role: result.role,
        userId: result.userId
      };
    } catch (error) {
      return {
        authenticated: false,
        error: error instanceof Error ? error.message : 'Unknown authentication error'
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
): Promise<UnifiedAuthHelper> {
  if (!token) {
    throw new Error('No authentication token provided. Set TEST_ADMIN_TOKEN environment variable.');
  }

  const authHelper = new UnifiedAuthHelper({ serverUrl });
  
  const authResult = await authHelper.authenticateWithToken(token);
  if (!authResult.authenticated) {
    throw new Error(`Authentication failed: ${authResult.error}`);
  }

  return authHelper;
}
