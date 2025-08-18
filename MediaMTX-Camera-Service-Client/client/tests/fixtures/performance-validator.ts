/**
 * Performance Validation Utilities
 * 
 * Validates performance targets across client-server integration
 * Following the unified testing strategy performance requirements
 */

import { PERFORMANCE_TARGETS } from '../../src/types';

export interface PerformanceMeasurement {
  operation: string;
  duration: number;
  timestamp: string;
  success: boolean;
  error?: string;
}

export interface PerformanceValidationResult {
  passed: boolean;
  violations: string[];
  measurements: PerformanceMeasurement[];
  averageTime: number;
  targetTime: number;
}

/**
 * Measure performance of an async operation
 */
export async function measurePerformance<T>(
  operation: string,
  asyncFn: () => Promise<T>,
  targetTime: number
): Promise<PerformanceMeasurement> {
  const startTime = performance.now();
  const timestamp = new Date().toISOString();
  
  try {
    await asyncFn();
    const duration = performance.now() - startTime;
    
    return {
      operation,
      duration,
      timestamp,
      success: true,
    };
  } catch (error) {
    const duration = performance.now() - startTime;
    
    return {
      operation,
      duration,
      timestamp,
      success: false,
      error: error instanceof Error ? error.message : String(error),
    };
  }
}

/**
 * Validate performance against targets
 */
export function validatePerformance(
  measurements: PerformanceMeasurement[],
  targetTime: number
): PerformanceValidationResult {
  const violations: string[] = [];
  const successfulMeasurements = measurements.filter(m => m.success);
  
  if (successfulMeasurements.length === 0) {
    violations.push('No successful measurements to validate');
    return {
      passed: false,
      violations,
      measurements,
      averageTime: 0,
      targetTime,
    };
  }
  
  const averageTime = successfulMeasurements.reduce((sum, m) => sum + m.duration, 0) / successfulMeasurements.length;
  
  // Check if average time exceeds target
  if (averageTime > targetTime) {
    violations.push(`${successfulMeasurements[0].operation} average time ${averageTime.toFixed(2)}ms exceeds target ${targetTime}ms`);
  }
  
  // Check if any individual measurement exceeds target significantly
  const slowMeasurements = successfulMeasurements.filter(m => m.duration > targetTime * 1.5);
  if (slowMeasurements.length > 0) {
    violations.push(`${slowMeasurements.length} measurements exceeded target by 50% or more`);
  }
  
  return {
    passed: violations.length === 0,
    violations,
    measurements,
    averageTime,
    targetTime,
  };
}

/**
 * Validate status method performance
 */
export function validateStatusMethodPerformance(
  measurements: PerformanceMeasurement[]
): PerformanceValidationResult {
  return validatePerformance(measurements, PERFORMANCE_TARGETS.STATUS_METHODS);
}

/**
 * Validate control method performance
 */
export function validateControlMethodPerformance(
  measurements: PerformanceMeasurement[]
): PerformanceValidationResult {
  return validatePerformance(measurements, PERFORMANCE_TARGETS.CONTROL_METHODS);
}

/**
 * Validate WebSocket connection performance
 */
export function validateConnectionPerformance(
  measurements: PerformanceMeasurement[]
): PerformanceValidationResult {
  return validatePerformance(measurements, PERFORMANCE_TARGETS.CLIENT_WEBSOCKET_CONNECTION);
}

/**
 * Run performance test with multiple iterations
 */
export async function runPerformanceTest<T>(
  operation: string,
  asyncFn: () => Promise<T>,
  targetTime: number,
  iterations: number = 5
): Promise<PerformanceValidationResult> {
  const measurements: PerformanceMeasurement[] = [];
  
  for (let i = 0; i < iterations; i++) {
    const measurement = await measurePerformance(operation, asyncFn, targetTime);
    measurements.push(measurement);
    
    // Small delay between measurements
    if (i < iterations - 1) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }
  }
  
  return validatePerformance(measurements, targetTime);
}

/**
 * Performance test helper for Jest
 */
export function expectPerformanceTarget(
  result: PerformanceValidationResult,
  operation: string
): void {
  expect(result.passed).toBe(true);
  expect(result.violations).toHaveLength(0);
  expect(result.averageTime).toBeLessThan(result.targetTime);
  
  // Log performance metrics for monitoring
  console.log(`Performance Test - ${operation}:`);
  console.log(`  Average Time: ${result.averageTime.toFixed(2)}ms`);
  console.log(`  Target Time: ${result.targetTime}ms`);
  console.log(`  Success Rate: ${result.measurements.filter(m => m.success).length}/${result.measurements.length}`);
}

/**
 * Create performance test suite
 */
export function createPerformanceTestSuite(
  operation: string,
  asyncFn: () => Promise<any>,
  targetTime: number
) {
  return () => {
    it(`should meet performance target for ${operation}`, async () => {
      const result = await runPerformanceTest(operation, asyncFn, targetTime);
      expectPerformanceTarget(result, operation);
    });
  };
}
