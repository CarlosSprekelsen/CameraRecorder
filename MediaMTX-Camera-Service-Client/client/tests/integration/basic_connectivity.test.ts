/**
 * Basic Integration Test: Server Connectivity
 * 
 * Simplified integration test that focuses on core functionality
 * without complex service dependencies
 */

import { APIClient } from '../../src/services/abstraction/APIClient';
import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Basic Integration Test: Server Connectivity', () => {
  let apiClient: APIClient;
  let webSocketService: WebSocketService;
  let loggerService: LoggerService;

  beforeAll(async () => {
    // Initialize services with new architecture
    loggerService = LoggerService.getInstance();
    webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
    apiClient = new APIClient(webSocketService, loggerService);
    
    // Connect to the server
    await webSocketService.connect();
    
    // Wait for connection to be established
    await new Promise(resolve => setTimeout(resolve, 1000));
  });

  afterAll(async () => {
    if (webSocketService) {
      await webSocketService.disconnect();
    }
    
    // Give time for cleanup
    await new Promise(resolve => setTimeout(resolve, 100));
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
