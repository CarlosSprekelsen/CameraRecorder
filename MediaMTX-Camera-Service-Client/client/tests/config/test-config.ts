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
 * Get File URL for specific operation
 * Note: File downloads are handled via WebSocket server, not separate HTTP endpoints
 */
export function getFileUrl(operation: string): string {
  // Files are served through the WebSocket server, not separate HTTP endpoints
  return `${TEST_CONFIG.websocket.url.replace('ws://', 'http://').replace('/ws', '')}${TEST_CONFIG.endpoints.files[operation as keyof typeof TEST_CONFIG.endpoints.files]}`;
}

export default TEST_CONFIG;
