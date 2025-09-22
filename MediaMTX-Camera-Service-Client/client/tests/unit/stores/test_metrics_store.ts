/**
 * REQ-PERF01-001: Performance metrics must track connection and message performance
 * REQ-PERF01-002: Metrics must provide accurate performance analysis and reporting
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for metrics store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on performance metrics tracking and analysis
 * - Test metrics calculations and aggregations
 * - Validate performance reporting functionality
 */

import { useMetricsStore } from '../../../src/stores/connection/metricsStore';

describe('Metrics Store', () => {
  let store: ReturnType<typeof useMetricsStore.getState>;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useMetricsStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useMetricsStore.getState();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useMetricsStore.getState();
      expect(state.messageCount).toBe(0);
      expect(state.errorCount).toBe(0);
      expect(state.lastMessageTime).toBeNull();
      expect(state.responseTimes).toEqual([]);
      expect(state.averageResponseTime).toBe(0);
      expect(state.maxResponseTime).toBe(0);
      expect(state.minResponseTime).toBe(0);
      expect(state.connectionUptime).toBeNull();
      expect(state.totalConnections).toBe(0);
      expect(state.successfulConnections).toBe(0);
      expect(state.failedConnections).toBe(0);
      expect(state.bytesSent).toBe(0);
      expect(state.bytesReceived).toBe(0);
      expect(state.messagesPerSecond).toBe(0);
      expect(state.maxResponseTimeHistory).toBe(100);
      expect(state.metricsResetTime).toBeNull();
    });
  });

  describe('Message Metrics', () => {
    it('should increment message count', () => {
      store.incrementMessageCount();
      let state = useMetricsStore.getState();
      expect(state.messageCount).toBe(1);

      store.incrementMessageCount();
      state = useMetricsStore.getState();
      expect(state.messageCount).toBe(2);
    });

    it('should increment error count', () => {
      store.incrementErrorCount();
      let state = useMetricsStore.getState();
      expect(state.errorCount).toBe(1);

      store.incrementErrorCount();
      state = useMetricsStore.getState();
      expect(state.errorCount).toBe(2);
    });

    it('should update last message time', () => {
      const now = new Date();
      store.updateLastMessageTime(now);
      
      const state = useMetricsStore.getState();
      expect(state.lastMessageTime).toEqual(now);
    });

    it('should record message with timestamp', () => {
      const timestamp = new Date();
      store.recordMessage(timestamp);
      
      const state = useMetricsStore.getState();
      expect(state.messageCount).toBe(1);
      expect(state.lastMessageTime).toEqual(timestamp);
    });
  });

  describe('Response Time Metrics', () => {
    it('should add response time', () => {
      store.addResponseTime(100);
      let state = useMetricsStore.getState();
      expect(state.responseTimes).toEqual([100]);
      expect(state.averageResponseTime).toBe(100);
      expect(state.maxResponseTime).toBe(100);
      expect(state.minResponseTime).toBe(100);

      store.addResponseTime(200);
      state = useMetricsStore.getState();
      expect(state.responseTimes).toEqual([100, 200]);
      expect(state.averageResponseTime).toBe(150);
      expect(state.maxResponseTime).toBe(200);
      expect(state.minResponseTime).toBe(100);
    });

    it('should limit response time history size', () => {
      store.setMaxResponseTimeHistory(3);
      
      // Add more response times than the limit
      store.addResponseTime(100);
      store.addResponseTime(200);
      store.addResponseTime(300);
      store.addResponseTime(400);
      store.addResponseTime(500);
      
      const state = useMetricsStore.getState();
      expect(state.responseTimes).toHaveLength(3);
      // Should keep the most recent ones
      expect(state.responseTimes).toEqual([300, 400, 500]);
    });

    it('should calculate average response time correctly', () => {
      store.addResponseTime(100);
      store.addResponseTime(200);
      store.addResponseTime(300);
      
      const state = useMetricsStore.getState();
      expect(state.averageResponseTime).toBe(200);
    });

    it('should track min and max response times', () => {
      store.addResponseTime(300);
      store.addResponseTime(100);
      store.addResponseTime(200);
      
      const state = useMetricsStore.getState();
      expect(state.minResponseTime).toBe(100);
      expect(state.maxResponseTime).toBe(300);
    });

    it('should clear response times', () => {
      store.addResponseTime(100);
      store.addResponseTime(200);
      store.clearResponseTimes();
      
      const state = useMetricsStore.getState();
      expect(state.responseTimes).toEqual([]);
      expect(state.averageResponseTime).toBe(0);
      expect(state.maxResponseTime).toBe(0);
      expect(state.minResponseTime).toBe(0);
    });
  });

  describe('Connection Metrics', () => {
    it('should start connection uptime tracking', () => {
      const startTime = new Date();
      store.startConnectionUptime(startTime);
      
      const state = useMetricsStore.getState();
      expect(state.connectionUptime).toEqual(startTime);
    });

    it('should stop connection uptime tracking', () => {
      const startTime = new Date(Date.now() - 5000);
      store.startConnectionUptime(startTime);
      store.stopConnectionUptime();
      
      const state = useMetricsStore.getState();
      expect(state.connectionUptime).toBeNull();
    });

    it('should get current connection duration', () => {
      const startTime = new Date(Date.now() - 5000);
      store.startConnectionUptime(startTime);
      
      const duration = store.getConnectionDuration();
      expect(duration).toBeGreaterThanOrEqual(5000);
      expect(duration).toBeLessThan(6000); // Allow some tolerance
    });

    it('should increment total connections', () => {
      store.incrementTotalConnections();
      let state = useMetricsStore.getState();
      expect(state.totalConnections).toBe(1);

      store.incrementTotalConnections();
      state = useMetricsStore.getState();
      expect(state.totalConnections).toBe(2);
    });

    it('should increment successful connections', () => {
      store.incrementSuccessfulConnections();
      let state = useMetricsStore.getState();
      expect(state.successfulConnections).toBe(1);

      store.incrementSuccessfulConnections();
      state = useMetricsStore.getState();
      expect(state.successfulConnections).toBe(2);
    });

    it('should increment failed connections', () => {
      store.incrementFailedConnections();
      let state = useMetricsStore.getState();
      expect(state.failedConnections).toBe(1);

      store.incrementFailedConnections();
      state = useMetricsStore.getState();
      expect(state.failedConnections).toBe(2);
    });

    it('should record connection attempt', () => {
      store.recordConnectionAttempt(true);
      let state = useMetricsStore.getState();
      expect(state.totalConnections).toBe(1);
      expect(state.successfulConnections).toBe(1);
      expect(state.failedConnections).toBe(0);

      store.recordConnectionAttempt(false);
      state = useMetricsStore.getState();
      expect(state.totalConnections).toBe(2);
      expect(state.successfulConnections).toBe(1);
      expect(state.failedConnections).toBe(1);
    });
  });

  describe('Data Transfer Metrics', () => {
    it('should add bytes sent', () => {
      store.addBytesSent(1024);
      let state = useMetricsStore.getState();
      expect(state.bytesSent).toBe(1024);

      store.addBytesSent(2048);
      state = useMetricsStore.getState();
      expect(state.bytesSent).toBe(3072);
    });

    it('should add bytes received', () => {
      store.addBytesReceived(512);
      let state = useMetricsStore.getState();
      expect(state.bytesReceived).toBe(512);

      store.addBytesReceived(1024);
      state = useMetricsStore.getState();
      expect(state.bytesReceived).toBe(1536);
    });

    it('should record data transfer', () => {
      store.recordDataTransfer(1024, 512);
      
      const state = useMetricsStore.getState();
      expect(state.bytesSent).toBe(1024);
      expect(state.bytesReceived).toBe(512);
    });

    it('should calculate total data transfer', () => {
      store.addBytesSent(1024);
      store.addBytesReceived(512);
      
      const total = store.getTotalDataTransfer();
      expect(total).toBe(1536);
    });
  });

  describe('Performance Calculations', () => {
    it('should calculate messages per second', () => {
      const startTime = new Date(Date.now() - 10000); // 10 seconds ago
      store.startConnectionUptime(startTime);
      
      // Record 50 messages over 10 seconds
      for (let i = 0; i < 50; i++) {
        store.incrementMessageCount();
      }
      
      store.calculateMessagesPerSecond();
      
      const state = useMetricsStore.getState();
      expect(state.messagesPerSecond).toBeCloseTo(5, 1); // 5 messages per second
    });

    it('should calculate error rate', () => {
      store.incrementMessageCount();
      store.incrementMessageCount();
      store.incrementMessageCount();
      store.incrementErrorCount();
      
      const errorRate = store.getErrorRate();
      expect(errorRate).toBeCloseTo(0.33, 2); // 1 error out of 3 messages
    });

    it('should calculate success rate', () => {
      store.incrementMessageCount();
      store.incrementMessageCount();
      store.incrementMessageCount();
      store.incrementErrorCount();
      
      const successRate = store.getSuccessRate();
      expect(successRate).toBeCloseTo(0.67, 2); // 2 successes out of 3 messages
    });

    it('should get performance summary', () => {
      store.addResponseTime(100);
      store.addResponseTime(200);
      store.addResponseTime(300);
      store.incrementMessageCount();
      store.incrementMessageCount();
      store.incrementErrorCount();
      store.addBytesSent(1024);
      store.addBytesReceived(512);
      
      const summary = store.getPerformanceSummary();
      
      expect(summary).toHaveProperty('messageCount', 2);
      expect(summary).toHaveProperty('errorCount', 1);
      expect(summary).toHaveProperty('averageResponseTime', 200);
      expect(summary).toHaveProperty('maxResponseTime', 300);
      expect(summary).toHaveProperty('minResponseTime', 100);
      expect(summary).toHaveProperty('bytesSent', 1024);
      expect(summary).toHaveProperty('bytesReceived', 512);
      expect(summary).toHaveProperty('errorRate');
      expect(summary).toHaveProperty('successRate');
    });
  });

  describe('Configuration', () => {
    it('should set max response time history', () => {
      store.setMaxResponseTimeHistory(50);
      
      const state = useMetricsStore.getState();
      expect(state.maxResponseTimeHistory).toBe(50);
    });
  });

  describe('Metrics Reset', () => {
    it('should reset all metrics', () => {
      // Set some metrics
      store.incrementMessageCount();
      store.incrementErrorCount();
      store.addResponseTime(100);
      store.addBytesSent(1024);
      store.incrementTotalConnections();
      
      store.reset();
      
      const state = useMetricsStore.getState();
      expect(state.messageCount).toBe(0);
      expect(state.errorCount).toBe(0);
      expect(state.responseTimes).toEqual([]);
      expect(state.bytesSent).toBe(0);
      expect(state.bytesReceived).toBe(0);
      expect(state.totalConnections).toBe(0);
      expect(state.successfulConnections).toBe(0);
      expect(state.failedConnections).toBe(0);
      expect(state.metricsResetTime).toBeInstanceOf(Date);
    });

    it('should reset specific metrics', () => {
      store.incrementMessageCount();
      store.incrementErrorCount();
      store.addResponseTime(100);
      
      store.resetMessageMetrics();
      
      const state = useMetricsStore.getState();
      expect(state.messageCount).toBe(0);
      expect(state.errorCount).toBe(0);
      expect(state.lastMessageTime).toBeNull();
      expect(state.responseTimes).toEqual([]);
    });

    it('should reset connection metrics', () => {
      store.incrementTotalConnections();
      store.incrementSuccessfulConnections();
      store.incrementFailedConnections();
      store.startConnectionUptime(new Date());
      
      store.resetConnectionMetrics();
      
      const state = useMetricsStore.getState();
      expect(state.totalConnections).toBe(0);
      expect(state.successfulConnections).toBe(0);
      expect(state.failedConnections).toBe(0);
      expect(state.connectionUptime).toBeNull();
    });

    it('should reset data transfer metrics', () => {
      store.addBytesSent(1024);
      store.addBytesReceived(512);
      
      store.resetDataTransferMetrics();
      
      const state = useMetricsStore.getState();
      expect(state.bytesSent).toBe(0);
      expect(state.bytesReceived).toBe(0);
    });
  });

  describe('Metrics Queries', () => {
    it('should check if metrics are available', () => {
      expect(store.hasMetrics()).toBe(false);
      
      store.incrementMessageCount();
      expect(store.hasMetrics()).toBe(true);
    });

    it('should get metrics age', () => {
      const oldTime = new Date(Date.now() - 60000); // 1 minute ago
      store.updateLastMessageTime(oldTime);
      
      const age = store.getMetricsAge();
      expect(age).toBeGreaterThanOrEqual(60000);
      expect(age).toBeLessThan(61000); // Allow some tolerance
    });

    it('should check if metrics are stale', () => {
      const oldTime = new Date(Date.now() - 300000); // 5 minutes ago
      store.updateLastMessageTime(oldTime);
      
      expect(store.areMetricsStale(240000)).toBe(true); // 4 minute threshold
      expect(store.areMetricsStale(360000)).toBe(false); // 6 minute threshold
    });
  });
});
