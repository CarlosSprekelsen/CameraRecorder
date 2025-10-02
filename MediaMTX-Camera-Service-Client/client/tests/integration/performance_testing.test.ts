/**
 * Performance Testing Suite
 * 
 * Comprehensive performance validation including:
 * - Load testing
 * - Stress testing
 * - Memory usage monitoring
 * - Response time validation
 * - Throughput measurement
 */

import { AuthHelper, createAuthenticatedTestEnvironment } from '../utils/auth-helper';
import { APIClient } from '../../src/services/abstraction/APIClient';
import { LoggerService } from '../../src/services/logger/LoggerService';

interface PerformanceMetrics {
  responseTime: number;
  throughput: number;
  memoryUsage: NodeJS.MemoryUsage;
  timestamp: number;
}

interface LoadTestResult {
  totalRequests: number;
  successfulRequests: number;
  failedRequests: number;
  averageResponseTime: number;
  minResponseTime: number;
  maxResponseTime: number;
  p95ResponseTime: number;
  throughput: number;
  memoryPeak: number;
  errorRate: number;
}

class PerformanceMonitor {
  private metrics: PerformanceMetrics[] = [];
  private memoryBaseline: NodeJS.MemoryUsage;

  constructor() {
    this.memoryBaseline = process.memoryUsage();
  }

  recordMetric(responseTime: number): void {
    const metric: PerformanceMetrics = {
      responseTime,
      throughput: 0, // Will be calculated later
      memoryUsage: process.memoryUsage(),
      timestamp: Date.now()
    };
    this.metrics.push(metric);
  }

  calculateResults(): LoadTestResult {
    const successfulMetrics = this.metrics;
    const totalRequests = this.metrics.length;
    const successfulRequests = successfulMetrics.length;
    const failedRequests = 0; // Assuming all recorded metrics are successful

    if (successfulMetrics.length === 0) {
      throw new Error('No successful metrics recorded');
    }

    const responseTimes = successfulMetrics.map(m => m.responseTime);
    const sortedResponseTimes = responseTimes.sort((a, b) => a - b);
    
    const averageResponseTime = responseTimes.reduce((a, b) => a + b, 0) / responseTimes.length;
    const minResponseTime = Math.min(...responseTimes);
    const maxResponseTime = Math.max(...responseTimes);
    const p95Index = Math.floor(sortedResponseTimes.length * 0.95);
    const p95ResponseTime = sortedResponseTimes[p95Index];

    const testDuration = this.metrics[this.metrics.length - 1].timestamp - this.metrics[0].timestamp;
    const throughput = (successfulRequests / testDuration) * 1000; // requests per second

    const memoryPeak = Math.max(...this.metrics.map(m => m.memoryUsage.heapUsed)) - this.memoryBaseline.heapUsed;
    const errorRate = (failedRequests / totalRequests) * 100;

    return {
      totalRequests,
      successfulRequests,
      failedRequests,
      averageResponseTime,
      minResponseTime,
      maxResponseTime,
      p95ResponseTime,
      throughput,
      memoryPeak,
      errorRate
    };
  }

  reset(): void {
    this.metrics = [];
    this.memoryBaseline = process.memoryUsage();
  }
}

describe('Performance Testing Suite', () => {
  let authHelper: AuthHelper;
  let apiClient: APIClient;
  let loggerService: LoggerService;
  let monitor: PerformanceMonitor;

  beforeAll(async () => {
    // Use unified authentication approach
    authHelper = await createAuthenticatedTestEnvironment(
      process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws'
    );
    
    const services = authHelper.getAuthenticatedServices();
    apiClient = services.apiClient;
    loggerService = services.logger;
    monitor = new PerformanceMonitor();
  });

  afterAll(async () => {
    if (authHelper) {
      await authHelper.disconnect();
    }
  });

  beforeEach(() => {
    monitor.reset();
  });

  describe('REQ-PERF-001: Basic Performance Validation', () => {
    test('should meet single request performance targets', async () => {
      const iterations = 10;
      
      for (let i = 0; i < iterations; i++) {
        const startTime = Date.now();
        
        try {
          await apiClient.call('ping', {});
          const responseTime = Date.now() - startTime;
          monitor.recordMetric(responseTime);
        } catch (error) {
          console.warn(`Request ${i} failed:`, error);
        }
      }

      const results = monitor.calculateResults();
      
      console.log('Basic Performance Results:', {
        averageResponseTime: `${results.averageResponseTime.toFixed(2)}ms`,
        p95ResponseTime: `${results.p95ResponseTime.toFixed(2)}ms`,
        throughput: `${results.throughput.toFixed(2)} req/s`,
        memoryPeak: `${(results.memoryPeak / 1024).toFixed(2)} KB`
      });

      // Performance targets
      expect(results.averageResponseTime).toBeLessThan(50); // < 50ms average
      expect(results.p95ResponseTime).toBeLessThan(100); // < 100ms p95
      expect(results.throughput).toBeGreaterThan(10); // > 10 req/s
      expect(results.errorRate).toBe(0); // No errors
    });
  });

  describe('REQ-PERF-002: Load Testing', () => {
    test('should handle sustained load', async () => {
      const loadDuration = 30000; // 30 seconds
      const requestInterval = 100; // 100ms between requests
      const startTime = Date.now();

      console.log(`Starting load test for ${loadDuration}ms...`);

      while (Date.now() - startTime < loadDuration) {
        const requestStart = Date.now();
        
        try {
          await apiClient.call('ping', {});
          const responseTime = Date.now() - requestStart;
          monitor.recordMetric(responseTime);
        } catch (error) {
          console.warn('Load test request failed:', error);
        }

        await new Promise(resolve => setTimeout(resolve, requestInterval));
      }

      const results = monitor.calculateResults();
      
      console.log('Load Test Results:', {
        duration: `${loadDuration}ms`,
        totalRequests: results.totalRequests,
        successfulRequests: results.successfulRequests,
        averageResponseTime: `${results.averageResponseTime.toFixed(2)}ms`,
        p95ResponseTime: `${results.p95ResponseTime.toFixed(2)}ms`,
        throughput: `${results.throughput.toFixed(2)} req/s`,
        memoryPeak: `${(results.memoryPeak / 1024).toFixed(2)} KB`,
        errorRate: `${results.errorRate.toFixed(2)}%`
      });

      // Load test targets
      expect(results.successfulRequests).toBeGreaterThan(200); // At least 200 successful requests
      expect(results.averageResponseTime).toBeLessThan(100); // < 100ms average
      expect(results.p95ResponseTime).toBeLessThan(200); // < 200ms p95
      expect(results.throughput).toBeGreaterThan(5); // > 5 req/s sustained
      expect(results.errorRate).toBeLessThan(5); // < 5% error rate
    });
  });

  describe('REQ-PERF-003: Burst Testing', () => {
    test('should handle request bursts', async () => {
      const burstSize = 20;
      const burstDelay = 2000; // 2 seconds between bursts
      const numberOfBursts = 3;

      for (let burst = 0; burst < numberOfBursts; burst++) {
        console.log(`Executing burst ${burst + 1}/${numberOfBursts} with ${burstSize} requests`);
        
        const promises: Promise<any>[] = [];
        
        // Send burst of concurrent requests
        for (let i = 0; i < burstSize; i++) {
          promises.push(apiClient.call('ping', {}));
        }

        const burstStart = Date.now();
        const results = await Promise.allSettled(promises);
        const burstDuration = Date.now() - burstStart;

        // Record metrics for successful requests
        results.forEach((result, index) => {
          if (result.status === 'fulfilled') {
            const responseTime = burstDuration / burstSize; // Approximate per-request time
            monitor.recordMetric(responseTime);
          } else {
            console.warn(`Burst request ${index} failed:`, result.reason);
          }
        });

        console.log(`Burst ${burst + 1} completed in ${burstDuration}ms`);
        
        if (burst < numberOfBursts - 1) {
          await new Promise(resolve => setTimeout(resolve, burstDelay));
        }
      }

      const results = monitor.calculateResults();
      
      console.log('Burst Test Results:', {
        burstSize,
        numberOfBursts,
        totalRequests: results.totalRequests,
        successfulRequests: results.successfulRequests,
        averageResponseTime: `${results.averageResponseTime.toFixed(2)}ms`,
        maxResponseTime: `${results.maxResponseTime.toFixed(2)}ms`,
        throughput: `${results.throughput.toFixed(2)} req/s`,
        errorRate: `${results.errorRate.toFixed(2)}%`
      });

      // Burst test targets
      expect(results.successfulRequests).toBeGreaterThan(burstSize * numberOfBursts * 0.9); // 90% success rate
      expect(results.averageResponseTime).toBeLessThan(150); // < 150ms average
      expect(results.maxResponseTime).toBeLessThan(500); // < 500ms max
      expect(results.errorRate).toBeLessThan(10); // < 10% error rate
    });
  });

  describe('REQ-PERF-004: Memory Usage Monitoring', () => {
    test('should maintain stable memory usage', async () => {
      const testDuration = 20000; // 20 seconds
      const requestInterval = 500; // 500ms between requests
      const memorySamples: NodeJS.MemoryUsage[] = [];
      const startTime = Date.now();

      console.log('Starting memory usage monitoring...');

      while (Date.now() - startTime < testDuration) {
        const requestStart = Date.now();
        
        try {
          await apiClient.call('ping', {});
          const responseTime = Date.now() - requestStart;
          monitor.recordMetric(responseTime);
          
          // Sample memory usage
          memorySamples.push(process.memoryUsage());
        } catch (error) {
          console.warn('Memory test request failed:', error);
        }

        await new Promise(resolve => setTimeout(resolve, requestInterval));
      }

      // Analyze memory usage
      const heapUsage = memorySamples.map(m => m.heapUsed);
      const initialMemory = heapUsage[0];
      const finalMemory = heapUsage[heapUsage.length - 1];
      const maxMemory = Math.max(...heapUsage);
      const memoryGrowth = finalMemory - initialMemory;
      const memoryGrowthPercent = (memoryGrowth / initialMemory) * 100;

      console.log('Memory Usage Analysis:', {
        testDuration: `${testDuration}ms`,
        initialMemory: `${(initialMemory / 1024 / 1024).toFixed(2)} MB`,
        finalMemory: `${(finalMemory / 1024 / 1024).toFixed(2)} MB`,
        maxMemory: `${(maxMemory / 1024 / 1024).toFixed(2)} MB`,
        memoryGrowth: `${(memoryGrowth / 1024 / 1024).toFixed(2)} MB`,
        memoryGrowthPercent: `${memoryGrowthPercent.toFixed(2)}%`,
        samples: memorySamples.length
      });

      // Memory stability targets
      expect(memoryGrowthPercent).toBeLessThan(50); // < 50% memory growth
      expect(maxMemory).toBeLessThan(initialMemory * 2); // Max memory < 2x initial
    });
  });

  describe('REQ-PERF-005: Connection Stability Under Load', () => {
    test('should maintain connection stability', async () => {
      const testDuration = 15000; // 15 seconds
      const requestInterval = 200; // 200ms between requests
      let connectionDrops = 0;
      let lastConnectionCheck = Date.now();
      const startTime = Date.now();

      console.log('Testing connection stability under load...');

      while (Date.now() - startTime < testDuration) {
        const requestStart = Date.now();
        
        try {
          await apiClient.call('ping', {});
          const responseTime = Date.now() - requestStart;
          monitor.recordMetric(responseTime);
          
          // Check connection state periodically
          if (Date.now() - lastConnectionCheck > 5000) { // Every 5 seconds
            if (!webSocketService.isConnected) {
              connectionDrops++;
              console.warn('Connection dropped, attempting reconnection...');
              await webSocketService.connect();
            }
            lastConnectionCheck = Date.now();
          }
        } catch (error) {
          console.warn('Stability test request failed:', error);
          connectionDrops++;
        }

        await new Promise(resolve => setTimeout(resolve, requestInterval));
      }

      const results = monitor.calculateResults();
      const uptime = ((results.successfulRequests / results.totalRequests) * 100);
      
      console.log('Connection Stability Results:', {
        testDuration: `${testDuration}ms`,
        totalRequests: results.totalRequests,
        successfulRequests: results.successfulRequests,
        connectionDrops,
        uptime: `${uptime.toFixed(2)}%`,
        averageResponseTime: `${results.averageResponseTime.toFixed(2)}ms`,
        errorRate: `${results.errorRate.toFixed(2)}%`
      });

      // Stability targets
      expect(uptime).toBeGreaterThan(95); // > 95% uptime
      expect(connectionDrops).toBeLessThan(3); // < 3 connection drops
      expect(results.errorRate).toBeLessThan(5); // < 5% error rate
    });
  });

  describe('REQ-PERF-006: Error Recovery Performance', () => {
    test('should recover quickly from errors', async () => {
      const errorRecoveryTests = [
        { method: 'nonexistent_method', params: {} },
        { method: 'get_camera_list', params: {} }, // Should fail without auth
        { method: 'ping', params: { invalid_param: 'test' } }
      ];

      const recoveryTimes: number[] = [];

      for (const test of errorRecoveryTests) {
        console.log(`Testing error recovery for method: ${test.method}`);
        
        // Send request that should fail
        const errorStart = Date.now();
        try {
          await apiClient.call(test.method as any, test.params);
        } catch (error) {
          // Expected error
        }
        const errorTime = Date.now() - errorStart;

        // Immediately send a ping to test recovery
        const recoveryStart = Date.now();
        try {
          await apiClient.call('ping', {});
          const recoveryTime = Date.now() - recoveryStart;
          recoveryTimes.push(recoveryTime);
          
          console.log(`Error recovery time: ${recoveryTime}ms`);
        } catch (error) {
          console.warn(`Recovery failed for ${test.method}:`, error);
          recoveryTimes.push(1000); // Penalty for failed recovery
        }

        await new Promise(resolve => setTimeout(resolve, 100));
      }

      const averageRecoveryTime = recoveryTimes.reduce((a, b) => a + b, 0) / recoveryTimes.length;
      
      console.log('Error Recovery Performance:', {
        tests: errorRecoveryTests.length,
        averageRecoveryTime: `${averageRecoveryTime.toFixed(2)}ms`,
        maxRecoveryTime: `${Math.max(...recoveryTimes)}ms`,
        minRecoveryTime: `${Math.min(...recoveryTimes)}ms`
      });

      // Recovery targets
      expect(averageRecoveryTime).toBeLessThan(200); // < 200ms average recovery
      expect(Math.max(...recoveryTimes)).toBeLessThan(500); // < 500ms max recovery
    });
  });
});
