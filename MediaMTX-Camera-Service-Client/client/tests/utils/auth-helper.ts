/**
 * SINGLE authentication utility for all tests
 * Dynamic token generation - NO hardcoded credentials
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-AUTH-001: JWT token generation
 * - REQ-AUTH-002: Role-based access control
 * - REQ-AUTH-003: Session management
 * 
 * Test Categories: Unit/Integration/Security
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import jwt from 'jsonwebtoken';
import { AuthResult } from '@/types/api';

export type UserRole = 'admin' | 'operator' | 'viewer';

export interface TestTokenPayload {
  sub: string;
  role: UserRole;
  permissions: string[];
  exp: number;
  iat: number;
}

export class AuthHelper {
  private static readonly TEST_SECRET = process.env.TEST_JWT_SECRET || 'test-secret-key';
  private static readonly TOKEN_EXPIRY = 3600; // 1 hour

  /**
   * Generate test JWT token with specified role
   * MANDATORY: Use this method for all authentication tests
   * NO hardcoded credentials allowed
   */
  static async generateTestToken(role: UserRole = 'admin'): Promise<string> {
    const payload: TestTokenPayload = {
      sub: `test-user-${Date.now()}`,
      role,
      permissions: this.getRolePermissions(role),
      exp: Math.floor(Date.now() / 1000) + this.TOKEN_EXPIRY,
      iat: Math.floor(Date.now() / 1000)
    };

    return jwt.sign(payload, this.TEST_SECRET);
  }

  /**
   * Generate test API key
   * MANDATORY: Use this method for API key authentication tests
   */
  static async generateTestApiKey(role: UserRole = 'admin'): Promise<string> {
    const token = await this.generateTestToken(role);
    return `api_${Buffer.from(token).toString('base64')}`;
  }

  /**
   * Validate authentication result against documented schema
   * MANDATORY: Use this validation for all auth result tests
   */
  static validateAuthResult(result: any): result is AuthResult {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.authenticated === 'boolean' &&
      typeof result.role === 'string' &&
      ['admin', 'operator', 'viewer'].includes(result.role) &&
      Array.isArray(result.permissions) &&
      typeof result.session_id === 'string'
    );
  }

  /**
   * Validate JWT token structure
   * MANDATORY: Use this validation for token tests
   */
  static validateTokenStructure(token: string): boolean {
    try {
      const decoded = jwt.decode(token) as any;
      return (
        decoded &&
        typeof decoded.sub === 'string' &&
        typeof decoded.role === 'string' &&
        Array.isArray(decoded.permissions) &&
        typeof decoded.exp === 'number' &&
        typeof decoded.iat === 'number'
      );
    } catch {
      return false;
    }
  }

  /**
   * Check if token is expired
   * MANDATORY: Use this method for token expiry tests
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
   * Extract role from token
   * MANDATORY: Use this method for role-based tests
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
   * Extract permissions from token
   * MANDATORY: Use this method for permission tests
   */
  static getTokenPermissions(token: string): string[] {
    try {
      const decoded = jwt.decode(token) as any;
      return decoded?.permissions || [];
    } catch {
      return [];
    }
  }

  /**
   * Get role-specific permissions
   * MANDATORY: Use this method for role permission tests
   */
  private static getRolePermissions(role: UserRole): string[] {
    switch (role) {
      case 'admin':
        return ['read', 'write', 'delete', 'admin'];
      case 'operator':
        return ['read', 'write'];
      case 'viewer':
        return ['read'];
      default:
        return [];
    }
  }

  /**
   * Create test authentication context
   * MANDATORY: Use this method for complete auth flow tests
   */
  static async createTestAuthContext(role: UserRole = 'admin'): Promise<{
    token: string;
    apiKey: string;
    role: UserRole;
    permissions: string[];
  }> {
    const token = await this.generateTestToken(role);
    const apiKey = await this.generateTestApiKey(role);
    const permissions = this.getRolePermissions(role);

    return {
      token,
      apiKey,
      role,
      permissions
    };
  }

  /**
   * Validate role-based access
   * MANDATORY: Use this method for authorization tests
   */
  static hasPermission(token: string, requiredPermission: string): boolean {
    const permissions = this.getTokenPermissions(token);
    return permissions.includes(requiredPermission);
  }

  /**
   * Validate admin access
   * MANDATORY: Use this method for admin-only tests
   */
  static isAdmin(token: string): boolean {
    return this.getTokenRole(token) === 'admin';
  }

  /**
   * Validate operator access
   * MANDATORY: Use this method for operator tests
   */
  static isOperator(token: string): boolean {
    const role = this.getTokenRole(token);
    return role === 'admin' || role === 'operator';
  }

  /**
   * Validate viewer access
   * MANDATORY: Use this method for viewer tests
   */
  static isViewer(token: string): boolean {
    const role = this.getTokenRole(token);
    return role === 'admin' || role === 'operator' || role === 'viewer';
  }
}
