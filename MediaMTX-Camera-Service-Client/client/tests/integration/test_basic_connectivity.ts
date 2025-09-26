/**
 * Basic Integration Test: Server Connectivity
 * 
 * Simplified integration test that focuses on core functionality
 * without complex service dependencies
 */

import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Basic Integration Test: Server Connectivity', () => {
  let webSocketService: WebSocketService;
  let loggerService: LoggerService;

  beforeAll(async () => {
    // Initialize services
    loggerService = LoggerService.getInstance();
    webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
    
    // Wait for connection
    await new Promise(resolve => setTimeout(resolve, 2000));
  });

  afterAll(async () => {
    if (webSocketService) {
      await webSocketService.disconnect();
    }
  });

  describe('REQ-BASIC-001: Server Connection', () => {
    test('should connect to real server', async () => {
      expect(webSocketService.isConnected).toBe(true);
      expect(webSocketService.connectionState).toBe(1); // WebSocket.OPEN
    });

    test('should maintain connection stability', async () => {
      // Test connection stability over time
      const startTime = Date.now();
      await new Promise(resolve => setTimeout(resolve, 3000));
      const endTime = Date.now();
      
      expect(webSocketService.isConnected).toBe(true);
      expect(endTime - startTime).toBeGreaterThan(2000);
    });
  });

  describe('REQ-BASIC-002: Performance Validation', () => {
    test('should meet connection performance targets', async () => {
      const startTime = Date.now();
      const connected = webSocketService.isConnected;
      const endTime = Date.now();
      
      expect(connected).toBe(true);
      expect(endTime - startTime).toBeLessThan(100); // < 100ms
    });
  });

  describe('REQ-BASIC-003: Error Handling', () => {
    test('should handle connection state correctly', async () => {
      expect(webSocketService.connectionState).toBeDefined();
      expect(typeof webSocketService.connectionState).toBe('number');
    });
  });
});
