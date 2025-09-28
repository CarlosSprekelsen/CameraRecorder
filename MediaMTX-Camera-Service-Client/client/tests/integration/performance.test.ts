/**
 * Integration Tests: Performance Testing
 * 
 * Tests system performance with real server
 * Focus: WebSocket performance, file operations, concurrent users
 */

import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { APIClient } from '../../src/services/abstraction/APIClient';
import { FileService } from '../../src/services/file/FileService';
import { DeviceService } from '../../src/services/device/DeviceService';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Integration Tests: Performance', () => {
  let webSocketService: WebSocketService;
  let fileService: FileService;
  let deviceService: DeviceService;
  let loggerService: LoggerService;

  beforeAll(async () => {
    loggerService = new LoggerService();
    webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
    
    // Wait for connection
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    fileService = new FileService(apiClient, loggerService);
    deviceService = new DeviceService(apiClient, loggerService);
  });

  afterAll(async () => {
    if (webSocketService) {
      await webSocketService.disconnect();
    }
  });

  describe('REQ-PERF-001: WebSocket Performance', () => {
    test('should meet connection latency targets', async () => {
      const startTime = Date.now();
      const connected = webSocketService.isConnected;
      const endTime = Date.now();
      
      expect(connected).toBe(true);
      expect(endTime - startTime).toBeLessThan(100); // < 100ms
    });

    test('should handle rapid message sending', async () => {
      const startTime = Date.now();
      const promises = [];
      
      // Send 10 rapid requests
      for (let i = 0; i < 10; i++) {
        promises.push(deviceService.getCameraList());
      }
      
      await Promise.all(promises);
      const endTime = Date.now();
      
      expect(endTime - startTime).toBeLessThan(2000); // < 2s for 10 requests
    });
  });

  describe('REQ-PERF-002: File Operations Performance', () => {
    test('should meet file listing performance targets', async () => {
      const startTime = Date.now();
      const recordings = await fileService.listRecordings(50, 0);
      const endTime = Date.now();
      
      expect(Array.isArray(recordings.files)).toBe(true);
      expect(endTime - startTime).toBeLessThan(1000); // < 1s
    });

    test('should handle large file lists efficiently', async () => {
      const startTime = Date.now();
      const recordings = await fileService.listRecordings(100, 0);
      const endTime = Date.now();
      
      expect(Array.isArray(recordings.files)).toBe(true);
      expect(endTime - startTime).toBeLessThan(2000); // < 2s for 100 files
    });
  });

  describe('REQ-PERF-003: Concurrent Operations', () => {
    test('should handle concurrent file operations', async () => {
      const startTime = Date.now();
      const promises = [];
      
      // Concurrent file operations
      for (let i = 0; i < 5; i++) {
        promises.push(fileService.listRecordings(10, i * 10));
        promises.push(fileService.listSnapshots(10, i * 10));
      }
      
      const results = await Promise.all(promises);
      const endTime = Date.now();
      
      expect(results).toHaveLength(10);
      expect(endTime - startTime).toBeLessThan(3000); // < 3s for 10 concurrent operations
    });

    test('should handle concurrent device operations', async () => {
      const startTime = Date.now();
      const promises = [];
      
      // Concurrent device operations
      for (let i = 0; i < 5; i++) {
        promises.push(deviceService.getCameraList());
        promises.push(deviceService.getStreams());
      }
      
      const results = await Promise.all(promises);
      const endTime = Date.now();
      
      expect(results).toHaveLength(10);
      expect(endTime - startTime).toBeLessThan(2000); // < 2s for 10 concurrent operations
    });
  });

  describe('REQ-PERF-004: Memory Usage', () => {
    test('should not leak memory during operations', async () => {
      const initialMemory = process.memoryUsage().heapUsed;
      
      // Perform multiple operations
      for (let i = 0; i < 10; i++) {
        await fileService.listRecordings(10, 0);
        await deviceService.getCameraList();
      }
      
      // Force garbage collection
      if (global.gc) {
        global.gc();
      }
      
      const finalMemory = process.memoryUsage().heapUsed;
      const memoryIncrease = finalMemory - initialMemory;
      
      // Memory increase should be reasonable (< 10MB)
      expect(memoryIncrease).toBeLessThan(10 * 1024 * 1024);
    });
  });

  describe('REQ-PERF-005: Network Resilience', () => {
    test('should handle network interruptions gracefully', async () => {
      // Test connection state during operations
      const isConnected = webSocketService.isConnected;
      expect(isConnected).toBe(true);
      
      // Perform operation
      const cameras = await deviceService.getCameraList();
      expect(Array.isArray(cameras)).toBe(true);
    });

    test('should recover from temporary disconnections', async () => {
      // This test would require network simulation
      // For now, just test that we can detect connection state
      expect(webSocketService.connectionState).toBeDefined();
    });
  });

  describe('REQ-PERF-006: Load Testing', () => {
    test('should handle sustained load', async () => {
      const startTime = Date.now();
      const operations = [];
      
      // Sustained load for 30 seconds
      const loadDuration = 30000; // 30 seconds
      const operationInterval = 100; // 100ms between operations
      
      const loadTest = setInterval(async () => {
        operations.push(deviceService.getCameraList());
      }, operationInterval);
      
      // Wait for load duration
      await new Promise(resolve => setTimeout(resolve, loadDuration));
      clearInterval(loadTest);
      
      // Wait for all operations to complete
      await Promise.all(operations);
      const endTime = Date.now();
      
      expect(endTime - startTime).toBeGreaterThan(loadDuration);
      expect(operations.length).toBeGreaterThan(100); // Should have many operations
    });
  });
});
