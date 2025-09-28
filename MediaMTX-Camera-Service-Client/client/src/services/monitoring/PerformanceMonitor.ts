/**
 * Performance Monitor - Architecture Compliance
 * 
 * Architecture requirement: "Command Ack ≤ 200ms (p95)" (Section 10.1)
 * Monitors and reports performance metrics for compliance verification
 */

import { LoggerService } from '../logger/LoggerService';

export interface PerformanceMetrics {
  commandAckTime: number;
  eventToUITime: number;
  totalRequests: number;
  successRate: number;
  averageResponseTime: number;
}

export interface PerformanceThresholds {
  commandAckMax: number; // 200ms
  eventToUIMax: number;   // 100ms
  successRateMin: number; // 99.9%
}

export class PerformanceMonitor {
  private metrics: PerformanceMetrics = {
    commandAckTime: 0,
    eventToUITime: 0,
    totalRequests: 0,
    successRate: 100,
    averageResponseTime: 0
  };

  private thresholds: PerformanceThresholds = {
    commandAckMax: 200,
    eventToUIMax: 100,
    successRateMin: 99.9
  };

  private requestTimes: number[] = [];
  private successCount: number = 0;

  constructor(private logger: LoggerService) {}

  /**
   * Start timing a command operation
   * Architecture requirement: Monitor command acknowledgment times
   */
  startCommandTimer(): () => void {
    const startTime = performance.now();
    
    return () => {
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      this.metrics.commandAckTime = duration;
      this.requestTimes.push(duration);
      this.totalRequests++;
      
      this.logger.info(`Command completed in ${duration.toFixed(2)}ms`);
      
      // Check if within architecture requirements
      if (duration > this.thresholds.commandAckMax) {
        this.logger.warn(`Command exceeded threshold: ${duration.toFixed(2)}ms > ${this.thresholds.commandAckMax}ms`);
      }
    };
  }

  /**
   * Start timing event-to-UI updates
   * Architecture requirement: "Event-to-UI ≤ 100ms (p95)"
   */
  startEventTimer(): () => void {
    const startTime = performance.now();
    
    return () => {
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      this.metrics.eventToUITime = duration;
      
      this.logger.info(`Event-to-UI update completed in ${duration.toFixed(2)}ms`);
      
      // Check if within architecture requirements
      if (duration > this.thresholds.eventToUIMax) {
        this.logger.warn(`Event-to-UI exceeded threshold: ${duration.toFixed(2)}ms > ${this.thresholds.eventToUIMax}ms`);
      }
    };
  }

  /**
   * Record successful operation
   * Architecture requirement: "Less than 0.1% transaction failure rate"
   */
  recordSuccess(): void {
    this.successCount++;
    this.updateSuccessRate();
  }

  /**
   * Record failed operation
   * Architecture requirement: "Less than 0.1% transaction failure rate"
   */
  recordFailure(): void {
    this.updateSuccessRate();
  }

  /**
   * Get current performance metrics
   * Architecture requirement: Performance monitoring and optimization
   */
  getMetrics(): PerformanceMetrics {
    this.calculateAverageResponseTime();
    return { ...this.metrics };
  }

  /**
   * Check if performance meets architecture requirements
   * Architecture requirement: Performance monitoring and optimization
   */
  checkCompliance(): {
    compliant: boolean;
    violations: string[];
  } {
    const violations: string[] = [];
    
    if (this.metrics.commandAckTime > this.thresholds.commandAckMax) {
      violations.push(`Command ACK time ${this.metrics.commandAckTime.toFixed(2)}ms exceeds ${this.thresholds.commandAckMax}ms threshold`);
    }
    
    if (this.metrics.eventToUITime > this.thresholds.eventToUIMax) {
      violations.push(`Event-to-UI time ${this.metrics.eventToUITime.toFixed(2)}ms exceeds ${this.thresholds.eventToUIMax}ms threshold`);
    }
    
    if (this.metrics.successRate < this.thresholds.successRateMin) {
      violations.push(`Success rate ${this.metrics.successRate.toFixed(1)}% below ${this.thresholds.successRateMin}% threshold`);
    }
    
    return {
      compliant: violations.length === 0,
      violations
    };
  }

  /**
   * Reset performance metrics
   * Architecture requirement: Performance monitoring and optimization
   */
  reset(): void {
    this.metrics = {
      commandAckTime: 0,
      eventToUITime: 0,
      totalRequests: 0,
      successRate: 100,
      averageResponseTime: 0
    };
    this.requestTimes = [];
    this.successCount = 0;
    this.logger.info('Performance metrics reset');
  }

  private updateSuccessRate(): void {
    if (this.totalRequests > 0) {
      this.metrics.successRate = (this.successCount / this.totalRequests) * 100;
    }
  }

  private calculateAverageResponseTime(): void {
    if (this.requestTimes.length > 0) {
      const sum = this.requestTimes.reduce((a, b) => a + b, 0);
      this.metrics.averageResponseTime = sum / this.requestTimes.length;
    }
  }
}
