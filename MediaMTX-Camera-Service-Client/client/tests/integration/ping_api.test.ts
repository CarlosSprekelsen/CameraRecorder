/**
 * Ping API Test
 * 
 * Tests the basic ping functionality to validate server API alignment
 */

import { AuthHelper, createAuthenticatedTestEnvironment } from '../utils/auth-helper';
import { APIClient } from '../../src/services/abstraction/APIClient';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Ping API Test', () => {
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
    await new Promise(resolve => setTimeout(resolve, 100));
  });

  test('should respond to ping method', async () => {
    const response = await apiClient.call('ping', {});
    
    expect(response).toBeDefined();
    expect(response).toBe('pong');
  }, 10000);

  test('should handle ping with proper JSON-RPC format', async () => {
    const response = await apiClient.call('ping', {});
    
    // Validate response format matches API documentation
    expect(typeof response).toBe('string');
    expect(response).toBe('pong');
  }, 10000);
});
