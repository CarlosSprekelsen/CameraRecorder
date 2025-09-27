/**
 * Ping API Test
 * 
 * Tests the basic ping functionality to validate server API alignment
 */

import { WebSocketService } from '../../src/services/websocket/WebSocketService';

describe('Ping API Test', () => {
  let webSocketService: WebSocketService;

  beforeAll(async () => {
    webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
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
    const response = await webSocketService.sendRPC('ping', {});
    
    expect(response).toBeDefined();
    expect(response).toBe('pong');
  }, 10000);

  test('should handle ping with proper JSON-RPC format', async () => {
    const response = await webSocketService.sendRPC('ping', {});
    
    // Validate response format matches API documentation
    expect(typeof response).toBe('string');
    expect(response).toBe('pong');
  }, 10000);
});
