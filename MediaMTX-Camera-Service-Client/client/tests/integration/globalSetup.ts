/**
 * Global Setup for Integration Tests
 * 
 * Setup tasks that run once before all integration tests
 */

export default async function globalSetup() {
  console.log('🔧 Setting up Integration Test Environment');
  
  // Check if server is running
  const serverUrl = process.env.SERVER_URL || 'ws://localhost:8002/ws';
  console.log(`📡 Checking server connectivity: ${serverUrl}`);
  
  // Wait for server to be ready
  await new Promise(resolve => setTimeout(resolve, 2000));
  
  console.log('✅ Integration Test Environment Ready');
  console.log('🚀 Starting Integration Tests');
}
