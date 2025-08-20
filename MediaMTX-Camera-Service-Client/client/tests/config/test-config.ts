/**
 * Centralized Test Configuration
 * Based on working quarantine test configuration
 */

export const TEST_CONFIG = {
  // WebSocket Server (JSON-RPC operations)
  websocket: {
    url: process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws',
    port: 8002,
    timeout: 10000,
  },
  
  // Health Server (REST operations)
  health: {
    url: process.env.TEST_HEALTH_URL || 'http://localhost:8003',
    port: 8003,
    timeout: 5000,
  },
  
  // Test Configuration
  test: {
    timeout: 10000,
    retries: 3,
    delay: 1000,
  },
  
  // Authentication
  auth: {
    jwtSecret: process.env.CAMERA_SERVICE_JWT_SECRET,
    tokenExpiry: 3600, // 1 hour
  },
  
  // Endpoints
  endpoints: {
    // WebSocket operations (camera control, file management)
    websocket: {
      ping: '/ws',
      camera_list: '/ws',
      camera_status: '/ws',
      take_snapshot: '/ws',
      start_recording: '/ws',
      stop_recording: '/ws',
      list_recordings: '/ws',
      list_snapshots: '/ws',
    },
    
    // Health operations (system status, monitoring)
    health: {
      system: '/health/system',
      cameras: '/health/cameras',
      mediamtx: '/health/mediamtx',
      ready: '/health/ready',
    },
    
    // File download operations
    files: {
      recordings: '/files/recordings',
      snapshots: '/files/snapshots',
    },
  },
};

/**
 * Test environment validation
 */
export function validateTestEnvironment(): boolean {
  const required = [
    'CAMERA_SERVICE_JWT_SECRET',
  ];
  
  const missing = required.filter(env => !process.env[env]);
  
  if (missing.length > 0) {
    console.error('❌ Missing required environment variables:', missing);
    console.error('💡 Run ./set-test-env.sh before executing tests');
    return false;
  }
  
  return true;
}

/**
 * Get WebSocket URL for specific operation
 */
export function getWebSocketUrl(): string {
  return TEST_CONFIG.websocket.url;
}

/**
 * Get Health URL for specific endpoint
 */
export function getHealthUrl(endpoint: string): string {
  return `${TEST_CONFIG.health.url}${endpoint}`;
}

/**
 * Get File URL for specific operation
 */
export function getFileUrl(operation: string): string {
  return `${TEST_CONFIG.health.url}${TEST_CONFIG.endpoints.files[operation as keyof typeof TEST_CONFIG.endpoints.files]}`;
}

export default TEST_CONFIG;
