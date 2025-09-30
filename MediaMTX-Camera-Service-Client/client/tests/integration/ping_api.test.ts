/**
 * Ping API Test
 * 
 * Tests the basic ping functionality to validate server API alignment
 */

import { APIClient } from '../../src/services/abstraction/APIClient';
import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Ping API Test', () => {
  let apiClient: APIClient;
  let webSocketService: WebSocketService;

  beforeAll(async () => {
    const loggerService = LoggerService.getInstance();
    webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
    apiClient = new APIClient(webSocketService, loggerService);
    await webSocketService.connect();
    await new Promise(resolve => setTimeout(resolve, 1000));
  });

  afterAll(async () => {
    if (webSocketService) {
      await webSocketService.disconnect();
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
