/**
 * Global Teardown for Integration Tests
 * 
 * Cleanup tasks that run once after all integration tests
 */

export default async function globalTeardown() {
  console.log('🧹 Cleaning up Integration Test Environment');
  
  // Get performance metrics
  const performanceMonitor = (global as any).performanceMonitor;
  if (performanceMonitor) {
    const metrics = performanceMonitor.getMetrics();
    console.log('📊 Performance Metrics:');
    console.log(`   Duration: ${metrics.duration}ms`);
    console.log(`   Operations: ${metrics.operations}`);
    console.log(`   Errors: ${metrics.errors}`);
    console.log(`   Success Rate: ${metrics.successRate.toFixed(2)}%`);
  }
  
  console.log('✅ Integration Test Environment Cleaned Up');
  console.log('🎉 All Integration Tests Completed');
}
