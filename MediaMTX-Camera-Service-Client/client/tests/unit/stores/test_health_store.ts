/**
 * REQ-NET01-003: Health monitoring must provide accurate connection quality assessment
 * REQ-NET01-004: Health metrics must track connection performance over time
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for health store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on health monitoring and quality assessment logic
 * - Test health score calculations and quality metrics
 * - Validate health history tracking
 */

import { useHealthStore } from '../../../src/stores/connection/healthStore';

describe('Health Store', () => {
  let store: ReturnType<typeof useHealthStore.getState>;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useHealthStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useHealthStore.getState();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useHealthStore.getState();
      expect(state.isHealthy).toBe(false);
      expect(state.healthScore).toBe(0);
      expect(state.connectionQuality).toBe('unstable');
      expect(state.lastHeartbeat).toBeNull();
      expect(state.heartbeatInterval).toBe(30000);
      expect(state.missedHeartbeats).toBe(0);
      expect(state.maxMissedHeartbeats).toBe(3);
      expect(state.latency).toBeNull();
      expect(state.packetLoss).toBeNull();
      expect(state.jitter).toBeNull();
      expect(state.healthHistory).toEqual([]);
      expect(state.maxHistorySize).toBe(100);
    });
  });

  describe('Health Status Management', () => {
    it('should set healthy status', () => {
      store.setHealthy(true);
      let state = useHealthStore.getState();
      expect(state.isHealthy).toBe(true);

      store.setHealthy(false);
      state = useHealthStore.getState();
      expect(state.isHealthy).toBe(false);
    });

    it('should set health score', () => {
      store.setHealthScore(85);
      let state = useHealthStore.getState();
      expect(state.healthScore).toBe(85);

      store.setHealthScore(0);
      state = useHealthStore.getState();
      expect(state.healthScore).toBe(0);

      store.setHealthScore(100);
      state = useHealthStore.getState();
      expect(state.healthScore).toBe(100);
    });

    it('should update health score with delta', () => {
      store.setHealthScore(50);
      store.updateHealthScore(20);
      let state = useHealthStore.getState();
      expect(state.healthScore).toBe(70);

      store.updateHealthScore(-30);
      state = useHealthStore.getState();
      expect(state.healthScore).toBe(40);
    });

    it('should clamp health score to valid range', () => {
      store.setHealthScore(50);
      store.updateHealthScore(100); // Should be clamped to 100
      let state = useHealthStore.getState();
      expect(state.healthScore).toBe(100);

      store.updateHealthScore(-200); // Should be clamped to 0
      state = useHealthStore.getState();
      expect(state.healthScore).toBe(0);
    });

    it('should set connection quality', () => {
      store.setConnectionQuality('excellent');
      let state = useHealthStore.getState();
      expect(state.connectionQuality).toBe('excellent');

      store.setConnectionQuality('good');
      state = useHealthStore.getState();
      expect(state.connectionQuality).toBe('good');

      store.setConnectionQuality('poor');
      state = useHealthStore.getState();
      expect(state.connectionQuality).toBe('poor');

      store.setConnectionQuality('unstable');
      state = useHealthStore.getState();
      expect(state.connectionQuality).toBe('unstable');
    });
  });

  describe('Heartbeat Management', () => {
    it('should update last heartbeat', () => {
      const now = new Date();
      store.updateHeartbeat(now);
      
      const state = useHealthStore.getState();
      expect(state.lastHeartbeat).toEqual(now);
    });

    it('should set heartbeat interval', () => {
      store.setHeartbeatInterval(15000);
      
      const state = useHealthStore.getState();
      expect(state.heartbeatInterval).toBe(15000);
    });

    it('should increment missed heartbeats', () => {
      store.incrementMissedHeartbeats();
      let state = useHealthStore.getState();
      expect(state.missedHeartbeats).toBe(1);

      store.incrementMissedHeartbeats();
      state = useHealthStore.getState();
      expect(state.missedHeartbeats).toBe(2);
    });

    it('should reset missed heartbeats', () => {
      store.incrementMissedHeartbeats();
      store.incrementMissedHeartbeats();
      store.resetMissedHeartbeats();
      
      const state = useHealthStore.getState();
      expect(state.missedHeartbeats).toBe(0);
    });

    it('should set max missed heartbeats', () => {
      store.setMaxMissedHeartbeats(5);
      
      const state = useHealthStore.getState();
      expect(state.maxMissedHeartbeats).toBe(5);
    });

    it('should check if heartbeat is overdue', () => {
      const oldTime = new Date(Date.now() - 40000); // 40 seconds ago
      store.updateHeartbeat(oldTime);
      store.setHeartbeatInterval(30000); // 30 second interval
      
      expect(store.isHeartbeatOverdue()).toBe(true);

      const recentTime = new Date(Date.now() - 10000); // 10 seconds ago
      store.updateHeartbeat(recentTime);
      
      expect(store.isHeartbeatOverdue()).toBe(false);
    });
  });

  describe('Quality Metrics Management', () => {
    it('should set latency', () => {
      store.setLatency(50);
      let state = useHealthStore.getState();
      expect(state.latency).toBe(50);

      store.setLatency(null);
      state = useHealthStore.getState();
      expect(state.latency).toBeNull();
    });

    it('should set packet loss', () => {
      store.setPacketLoss(0.5);
      let state = useHealthStore.getState();
      expect(state.packetLoss).toBe(0.5);

      store.setPacketLoss(null);
      state = useHealthStore.getState();
      expect(state.packetLoss).toBeNull();
    });

    it('should set jitter', () => {
      store.setJitter(10);
      let state = useHealthStore.getState();
      expect(state.jitter).toBe(10);

      store.setJitter(null);
      state = useHealthStore.getState();
      expect(state.jitter).toBeNull();
    });

    it('should update all quality metrics at once', () => {
      const metrics = {
        latency: 25,
        packetLoss: 0.1,
        jitter: 5
      };
      
      store.updateQualityMetrics(metrics);
      
      const state = useHealthStore.getState();
      expect(state.latency).toBe(25);
      expect(state.packetLoss).toBe(0.1);
      expect(state.jitter).toBe(5);
    });
  });

  describe('Health History Management', () => {
    it('should add health snapshot to history', () => {
      const timestamp = new Date();
      const snapshot = {
        timestamp,
        score: 85,
        quality: 'good'
      };
      
      store.addHealthSnapshot(snapshot);
      
      const state = useHealthStore.getState();
      expect(state.healthHistory).toHaveLength(1);
      expect(state.healthHistory[0]).toEqual(snapshot);
    });

    it('should limit health history size', () => {
      store.setMaxHistorySize(3);
      
      // Add more snapshots than the limit
      for (let i = 0; i < 5; i++) {
        store.addHealthSnapshot({
          timestamp: new Date(),
          score: i * 20,
          quality: 'good'
        });
      }
      
      const state = useHealthStore.getState();
      expect(state.healthHistory).toHaveLength(3);
      // Should keep the most recent ones
      expect(state.healthHistory[0].score).toBe(80);
      expect(state.healthHistory[1].score).toBe(60);
      expect(state.healthHistory[2].score).toBe(40);
    });

    it('should clear health history', () => {
      // Add some snapshots
      store.addHealthSnapshot({
        timestamp: new Date(),
        score: 85,
        quality: 'good'
      });
      
      store.clearHealthHistory();
      
      const state = useHealthStore.getState();
      expect(state.healthHistory).toEqual([]);
    });

    it('should set max history size', () => {
      store.setMaxHistorySize(50);
      
      const state = useHealthStore.getState();
      expect(state.maxHistorySize).toBe(50);
    });
  });

  describe('Health Assessment', () => {
    it('should assess health based on score', () => {
      store.setHealthScore(95);
      expect(store.assessHealth()).toBe('excellent');

      store.setHealthScore(80);
      expect(store.assessHealth()).toBe('good');

      store.setHealthScore(60);
      expect(store.assessHealth()).toBe('poor');

      store.setHealthScore(30);
      expect(store.assessHealth()).toBe('unstable');
    });

    it('should assess health based on missed heartbeats', () => {
      store.setHealthScore(80);
      store.setMaxMissedHeartbeats(3);
      
      // No missed heartbeats - should be good
      expect(store.assessHealth()).toBe('good');
      
      // Some missed heartbeats - should degrade
      store.incrementMissedHeartbeats();
      store.incrementMissedHeartbeats();
      expect(store.assessHealth()).toBe('poor');
      
      // Max missed heartbeats - should be unstable
      store.incrementMissedHeartbeats();
      expect(store.assessHealth()).toBe('unstable');
    });

    it('should assess health based on latency', () => {
      store.setHealthScore(80);
      
      // Low latency - should be good
      store.setLatency(20);
      expect(store.assessHealth()).toBe('good');
      
      // High latency - should degrade
      store.setLatency(200);
      expect(store.assessHealth()).toBe('poor');
    });

    it('should assess health based on packet loss', () => {
      store.setHealthScore(80);
      
      // Low packet loss - should be good
      store.setPacketLoss(0.1);
      expect(store.assessHealth()).toBe('good');
      
      // High packet loss - should degrade
      store.setPacketLoss(5.0);
      expect(store.assessHealth()).toBe('unstable');
    });

    it('should update connection quality based on assessment', () => {
      store.setHealthScore(90);
      store.updateConnectionQuality();
      
      const state = useHealthStore.getState();
      expect(state.connectionQuality).toBe('excellent');
    });
  });

  describe('Health Monitoring', () => {
    it('should start health monitoring', () => {
      store.startHealthMonitoring();
      
      const state = useHealthStore.getState();
      expect(state.isMonitoring).toBe(true);
    });

    it('should stop health monitoring', () => {
      store.startHealthMonitoring();
      store.stopHealthMonitoring();
      
      const state = useHealthStore.getState();
      expect(state.isMonitoring).toBe(false);
    });

    it('should record health event', () => {
      const event = {
        type: 'heartbeat_received',
        timestamp: new Date(),
        data: { latency: 25 }
      };
      
      store.recordHealthEvent(event);
      
      const state = useHealthStore.getState();
      expect(state.healthEvents).toHaveLength(1);
      expect(state.healthEvents[0]).toEqual(event);
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.setHealthy(true);
      store.setHealthScore(85);
      store.setConnectionQuality('good');
      store.setLatency(25);
      store.addHealthSnapshot({
        timestamp: new Date(),
        score: 85,
        quality: 'good'
      });
      
      // Reset
      store.reset();
      
      const state = useHealthStore.getState();
      expect(state.isHealthy).toBe(false);
      expect(state.healthScore).toBe(0);
      expect(state.connectionQuality).toBe('unstable');
      expect(state.latency).toBeNull();
      expect(state.healthHistory).toEqual([]);
    });
  });

  describe('Health Queries', () => {
    it('should get average health score', () => {
      store.addHealthSnapshot({
        timestamp: new Date(),
        score: 80,
        quality: 'good'
      });
      store.addHealthSnapshot({
        timestamp: new Date(),
        score: 90,
        quality: 'excellent'
      });
      
      const average = store.getAverageHealthScore();
      expect(average).toBe(85);
    });

    it('should get health trend', () => {
      const now = new Date();
      store.addHealthSnapshot({
        timestamp: new Date(now.getTime() - 60000),
        score: 70,
        quality: 'good'
      });
      store.addHealthSnapshot({
        timestamp: new Date(now.getTime() - 30000),
        score: 80,
        quality: 'good'
      });
      store.addHealthSnapshot({
        timestamp: now,
        score: 90,
        quality: 'excellent'
      });
      
      const trend = store.getHealthTrend();
      expect(trend).toBe('improving');
    });

    it('should get time since last heartbeat', () => {
      const heartbeatTime = new Date(Date.now() - 15000); // 15 seconds ago
      store.updateHeartbeat(heartbeatTime);
      
      const timeSince = store.getTimeSinceLastHeartbeat();
      expect(timeSince).toBeGreaterThanOrEqual(15000);
      expect(timeSince).toBeLessThan(16000); // Allow some tolerance
    });
  });
});
