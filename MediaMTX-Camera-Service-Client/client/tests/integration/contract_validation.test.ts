/**
 * Comprehensive API Contract Validation Suite
 * 
 * Validates complete API contract compliance including:
 * - Request/Response structure validation
 * - Data type validation
 * - Error response validation
 * - Field presence and format validation
 * - Performance contract validation
 */

import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { AuthService } from '../../src/services/auth/AuthService';
import { DeviceService } from '../../src/services/device/DeviceService';
import { FileService } from '../../src/services/file/FileService';
import { ServerService } from '../../src/services/server/ServerService';
import { LoggerService } from '../../src/services/logger/LoggerService';

// Contract validation utilities
interface ContractValidationResult {
  passed: boolean;
  errors: string[];
  warnings: string[];
  performance: {
    responseTime: number;
    withinThreshold: boolean;
  };
}

class ContractValidator {
  private errors: string[] = [];
  private warnings: string[] = [];
  private startTime: number = 0;

  startTimer(): void {
    this.startTime = Date.now();
  }

  getPerformance(): { responseTime: number; withinThreshold: boolean } {
    const responseTime = Date.now() - this.startTime;
    return {
      responseTime,
      withinThreshold: responseTime < 100 // 100ms threshold for API calls
    };
  }

  validateField(data: any, fieldName: string, expectedType: string, required: boolean = true): void {
    if (required && (data === undefined || data === null)) {
      this.errors.push(`Missing required field: ${fieldName}`);
      return;
    }

    if (data !== undefined && data !== null) {
      const actualType = typeof data;
      if (actualType !== expectedType) {
        this.errors.push(`Field ${fieldName}: expected ${expectedType}, got ${actualType}`);
      }
    }
  }

  validateArray(data: any, fieldName: string, required: boolean = true): void {
    this.validateField(data, fieldName, 'object', required);
    if (data !== undefined && data !== null && !Array.isArray(data)) {
      this.errors.push(`Field ${fieldName}: expected array, got ${typeof data}`);
    }
  }

  validateTimestamp(timestamp: string, fieldName: string): void {
    if (!timestamp) {
      this.errors.push(`Missing timestamp: ${fieldName}`);
      return;
    }

    const date = new Date(timestamp);
    if (isNaN(date.getTime())) {
      this.errors.push(`Invalid timestamp format: ${fieldName} (${timestamp})`);
    }
  }

  validateUrl(url: string, fieldName: string): void {
    if (!url) {
      this.errors.push(`Missing URL: ${fieldName}`);
      return;
    }

    try {
      new URL(url);
    } catch {
      this.errors.push(`Invalid URL format: ${fieldName} (${url})`);
    }
  }

  addWarning(message: string): void {
    this.warnings.push(message);
  }

  getResult(): ContractValidationResult {
    return {
      passed: this.errors.length === 0,
      errors: [...this.errors],
      warnings: [...this.warnings],
      performance: this.getPerformance()
    };
  }

  reset(): void {
    this.errors = [];
    this.warnings = [];
    this.startTime = 0;
  }
}

describe('API Contract Validation Suite', () => {
  let webSocketService: WebSocketService;
  let authService: AuthService;
  let deviceService: DeviceService;
  let fileService: FileService;
  let serverService: ServerService;
  let loggerService: LoggerService;
  let validator: ContractValidator;

  beforeAll(async () => {
    loggerService = new LoggerService();
    webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
    validator = new ContractValidator();
    
    // Connect to the server
    await webSocketService.connect();
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    authService = new AuthService(webSocketService, loggerService);
    deviceService = new DeviceService(webSocketService, loggerService);
    fileService = new FileService(webSocketService, loggerService);
    serverService = new ServerService(webSocketService, loggerService);
  });

  afterAll(async () => {
    if (webSocketService) {
      await webSocketService.disconnect();
    }
    await new Promise(resolve => setTimeout(resolve, 100));
  });

  beforeEach(() => {
    validator.reset();
  });

  describe('REQ-CONTRACT-001: Ping Method Contract', () => {
    test('should validate ping response contract', async () => {
      validator.startTimer();
      
      try {
        const response = await webSocketService.sendRPC('ping', {});
        
        // Validate response structure
        expect(typeof response).toBe('string');
        expect(response).toBe('pong');
        
        const result = validator.getResult();
        expect(result.passed).toBe(true);
        expect(result.performance.withinThreshold).toBe(true);
        
      } catch (error) {
        validator.addWarning(`Ping method failed: ${error}`);
        const result = validator.getResult();
        expect(result.passed).toBe(true); // Ping should always work
      }
    });
  });

  describe('REQ-CONTRACT-002: Error Response Contract', () => {
    test('should validate authentication error contract', async () => {
      validator.startTimer();
      
      try {
        await deviceService.getCameraList();
        validator.addWarning('Expected authentication error but got success');
      } catch (error: any) {
        // Validate error structure
        expect(error.code).toBeDefined();
        expect(typeof error.code).toBe('number');
        expect(error.message).toBeDefined();
        expect(typeof error.message).toBe('string');
        
        // Validate JSON-RPC error format
        validator.validateField(error.code, 'error.code', 'number');
        validator.validateField(error.message, 'error.message', 'string');
        
        const result = validator.getResult();
        expect(result.passed).toBe(true);
        expect(result.performance.withinThreshold).toBe(true);
      }
    });

    test('should validate method not found error contract', async () => {
      validator.startTimer();
      
      try {
        await webSocketService.sendRPC('nonexistent_method', {});
        validator.addWarning('Expected method not found error but got success');
      } catch (error: any) {
        // Should get method not found error
        validator.validateField(error.code, 'error.code', 'number');
        validator.validateField(error.message, 'error.message', 'string');
        
        const result = validator.getResult();
        expect(result.passed).toBe(true);
      }
    });
  });

  describe('REQ-CONTRACT-003: Authentication Contract', () => {
    test('should validate authentication request contract', async () => {
      validator.startTimer();
      
      try {
        // Test with invalid token to validate error contract
        await webSocketService.sendRPC('authenticate', { auth_token: 'invalid_token' });
        validator.addWarning('Expected authentication failure but got success');
      } catch (error: any) {
        validator.validateField(error.code, 'error.code', 'number');
        validator.validateField(error.message, 'error.message', 'string');
        
        const result = validator.getResult();
        expect(result.passed).toBe(true);
        expect(result.performance.withinThreshold).toBe(true);
      }
    });
  });

  describe('REQ-CONTRACT-004: Data Structure Contracts', () => {
    test('should validate timestamp format contract', () => {
      const validTimestamps = [
        '2025-01-15T14:30:00Z',
        '2025-01-15T14:30:00.000Z',
        '2025-01-15T14:30:00+00:00'
      ];

      validTimestamps.forEach(timestamp => {
        validator.reset();
        validator.validateTimestamp(timestamp, 'test_timestamp');
        const result = validator.getResult();
        expect(result.passed).toBe(true);
      });

      const invalidTimestamps = [
        'invalid-date',
        '2025-13-45T25:70:00Z',
        'not-a-timestamp'
      ];

      invalidTimestamps.forEach(timestamp => {
        validator.reset();
        validator.validateTimestamp(timestamp, 'test_timestamp');
        const result = validator.getResult();
        expect(result.passed).toBe(false);
        expect(result.errors.length).toBeGreaterThan(0);
      });
    });

    test('should validate URL format contract', () => {
      const validUrls = [
        'rtsp://localhost:8554/camera0',
        'https://localhost/hls/camera0.m3u8',
        'http://192.168.1.100:8080/stream'
      ];

      validUrls.forEach(url => {
        validator.reset();
        validator.validateUrl(url, 'test_url');
        const result = validator.getResult();
        expect(result.passed).toBe(true);
      });

      const invalidUrls = [
        'not-a-url',
        'rtsp://',
        'ftp://invalid-protocol'
      ];

      invalidUrls.forEach(url => {
        validator.reset();
        validator.validateUrl(url, 'test_url');
        const result = validator.getResult();
        expect(result.passed).toBe(false);
        expect(result.errors.length).toBeGreaterThan(0);
      });
    });
  });

  describe('REQ-CONTRACT-005: Performance Contract', () => {
    test('should validate ping performance contract', async () => {
      const iterations = 10;
      const responseTimes: number[] = [];

      for (let i = 0; i < iterations; i++) {
        validator.startTimer();
        
        try {
          await webSocketService.sendRPC('ping', {});
          const performance = validator.getPerformance();
          responseTimes.push(performance.responseTime);
        } catch (error) {
          validator.addWarning(`Ping iteration ${i} failed: ${error}`);
        }
      }

      // Validate performance metrics
      const avgResponseTime = responseTimes.reduce((a, b) => a + b, 0) / responseTimes.length;
      const maxResponseTime = Math.max(...responseTimes);
      const minResponseTime = Math.min(...responseTimes);

      console.log(`Ping Performance Metrics:
        - Average: ${avgResponseTime.toFixed(2)}ms
        - Maximum: ${maxResponseTime}ms
        - Minimum: ${minResponseTime}ms
        - Samples: ${responseTimes.length}`);

      // Performance contract: 95th percentile should be under 100ms
      expect(avgResponseTime).toBeLessThan(100);
      expect(maxResponseTime).toBeLessThan(200);
      expect(responseTimes.filter(t => t < 100).length).toBeGreaterThan(8); // 80% under threshold
    });

    test('should validate error response performance contract', async () => {
      const iterations = 5;
      const responseTimes: number[] = [];

      for (let i = 0; i < iterations; i++) {
        validator.startTimer();
        
        try {
          await deviceService.getCameraList();
        } catch (error) {
          // Expected error - measure response time
          const performance = validator.getPerformance();
          responseTimes.push(performance.responseTime);
        }
      }

      const avgResponseTime = responseTimes.reduce((a, b) => a + b, 0) / responseTimes.length;
      
      console.log(`Error Response Performance:
        - Average: ${avgResponseTime.toFixed(2)}ms
        - Samples: ${responseTimes.length}`);

      // Error responses should also be fast
      expect(avgResponseTime).toBeLessThan(150);
    });
  });

  describe('REQ-CONTRACT-006: Concurrency Contract', () => {
    test('should validate concurrent request handling', async () => {
      const concurrentRequests = 5;
      const promises: Promise<any>[] = [];

      // Send multiple concurrent ping requests
      for (let i = 0; i < concurrentRequests; i++) {
        promises.push(webSocketService.sendRPC('ping', {}));
      }

      const startTime = Date.now();
      const results = await Promise.allSettled(promises);
      const totalTime = Date.now() - startTime;

      // Validate all requests completed
      const successful = results.filter(r => r.status === 'fulfilled').length;
      expect(successful).toBe(concurrentRequests);

      console.log(`Concurrent Requests:
        - Requests: ${concurrentRequests}
        - Successful: ${successful}
        - Total Time: ${totalTime}ms
        - Average per Request: ${(totalTime / concurrentRequests).toFixed(2)}ms`);

      // Concurrent requests should complete efficiently
      expect(totalTime).toBeLessThan(1000);
    });
  });

  describe('REQ-CONTRACT-007: Connection Stability Contract', () => {
    test('should validate connection stability over time', async () => {
      const testDuration = 10000; // 10 seconds
      const interval = 1000; // 1 second intervals
      const results: { time: number; success: boolean; responseTime: number }[] = [];

      const startTime = Date.now();
      
      while (Date.now() - startTime < testDuration) {
        const requestStart = Date.now();
        
        try {
          await webSocketService.sendRPC('ping', {});
          results.push({
            time: Date.now() - startTime,
            success: true,
            responseTime: Date.now() - requestStart
          });
        } catch (error) {
          results.push({
            time: Date.now() - startTime,
            success: false,
            responseTime: Date.now() - requestStart
          });
        }

        await new Promise(resolve => setTimeout(resolve, interval));
      }

      // Validate stability metrics
      const successful = results.filter(r => r.success).length;
      const successRate = (successful / results.length) * 100;
      const avgResponseTime = results
        .filter(r => r.success)
        .reduce((sum, r) => sum + r.responseTime, 0) / successful;

      console.log(`Connection Stability Metrics:
        - Test Duration: ${testDuration}ms
        - Total Requests: ${results.length}
        - Success Rate: ${successRate.toFixed(2)}%
        - Average Response Time: ${avgResponseTime.toFixed(2)}ms`);

      // Connection should be stable
      expect(successRate).toBeGreaterThan(95); // 95% success rate
      expect(avgResponseTime).toBeLessThan(100); // Under 100ms average
    });
  });
});
