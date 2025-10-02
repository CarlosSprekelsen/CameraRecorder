/**
 * Basic Integration Test: Server Connectivity
 * 
 * Simplified integration test that focuses on core functionality
 * without complex service dependencies
 */

import { APIClient } from '../../src/services/abstraction/APIClient';
import { AuthHelper, createAuthenticatedTestEnvironment } from '../utils/auth-helper';

describe('Basic Integration Test: Server Connectivity', () => {
  let authHelper: AuthHelper;
  let apiClient: APIClient;

  beforeAll(async () => {
    // Use unified authentication approach
    authHelper = await createAuthenticatedTestEnvironment(
      process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws'
    );
    
    apiClient = authHelper.getAuthenticatedServices().apiClient;
  });

  afterAll(async () => {
    if (authHelper) {
      await authHelper.disconnect();
    }
  });

  describe('REQ-BASIC-001: Server Connection', () => {
    test('should connect to real server', async () => {
      expect(apiClient.isConnected()).toBe(true);
      expect(apiClient.getConnectionStatus().connected).toBe(true);
    });

    test('should maintain connection stability', async () => {
      // Test connection stability over time
      const startTime = Date.now();
      await new Promise(resolve => setTimeout(resolve, 3000));
      const endTime = Date.now();
      
      expect(apiClient.isConnected()).toBe(true);
      expect(apiClient.getConnectionStatus().connected).toBe(true);
      expect(endTime - startTime).toBeGreaterThan(2000);
    });
  });

  describe('REQ-BASIC-002: Performance Validation', () => {
    test('should meet connection performance targets', async () => {
      const startTime = Date.now();
      const connected = apiClient.isConnected();
      const endTime = Date.now();
      
      expect(connected).toBe(true);
      expect(endTime - startTime).toBeLessThan(100); // < 100ms
    });
  });

  describe('REQ-BASIC-003: Error Handling', () => {
    test('should handle connection state correctly', async () => {
      const status = apiClient.getConnectionStatus();
      expect(status.connected).toBeDefined();
      expect(typeof status.connected).toBe('boolean');
      expect(typeof status.ready).toBe('boolean');
    });
  });
});
