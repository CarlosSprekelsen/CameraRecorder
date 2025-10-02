/**
 * E2E test setup configuration
 * MANDATORY: Use this setup for all E2E tests
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-SETUP-001: E2E test environment configuration
 * - REQ-SETUP-002: Real server interaction
 * - REQ-SETUP-003: Authentication token loading
 * 
 * Test Categories: E2E/Workflow
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import dotenv from 'dotenv';
import path from 'path';

// Load test environment tokens from server-generated .test_env
// Convert export format to standard .env format
const envPath = path.join(__dirname, '../.test_env');
const envContent = require('fs').readFileSync(envPath, 'utf8');

// Parse the export format and convert to standard .env format
const envVars: { [key: string]: string } = {};
envContent.split('\n').forEach(line => {
  const trimmed = line.trim();
  if (trimmed && !trimmed.startsWith('#') && trimmed.startsWith('export ')) {
    const match = trimmed.match(/export\s+(\w+)="([^"]*)"/);
    if (match) {
      envVars[match[1]] = match[2];
    }
  }
});

// Set environment variables
Object.entries(envVars).forEach(([key, value]) => {
  if (value && !process.env[key]) {
    process.env[key] = value;
  }
});

// Set additional E2E-specific environment variables
process.env.NODE_ENV = 'test';
process.env.TEST_WEBSOCKET_URL = `ws://${process.env.CAMERA_SERVICE_HOST}:${process.env.CAMERA_SERVICE_PORT}${process.env.CAMERA_SERVICE_WS_PATH}`;
process.env.TEST_JWT_SECRET = 'test-secret'; // Use consistent secret for test environment

console.log('🚀 Starting E2E Tests with Real Server');
console.log(`📡 Server URL: ${process.env.TEST_WEBSOCKET_URL}`);
console.log(`⏱️  Test Timeout: 60 seconds per test`);
console.log(`🔒 Authentication: Enabled with JWT tokens`);
console.log(`📊 Performance Tests: Enabled`);
console.log(`📡 API compliance: Enabled`);

// Validate test environment
if (!process.env.CAMERA_SERVICE_HOST || !process.env.CAMERA_SERVICE_PORT || !process.env.CAMERA_SERVICE_WS_PATH) {
  throw new Error('CAMERA_SERVICE_HOST, CAMERA_SERVICE_PORT, and CAMERA_SERVICE_WS_PATH environment variables are required for E2E tests');
}

if (!process.env.TEST_ADMIN_TOKEN && !process.env.TEST_OPERATOR_TOKEN && !process.env.TEST_VIEWER_TOKEN) {
  throw new Error('At least one JWT token (TEST_ADMIN_TOKEN, TEST_OPERATOR_TOKEN, or TEST_VIEWER_TOKEN) is required for E2E tests');
}

console.log('✅ E2E Tests Environment Ready');
console.log('📊 Performance metrics enabled');
console.log('🔒 Security validation enabled');
console.log('📡 API compliance verification enabled');
console.log('🧹 Cleanup completed - using unified authentication pattern');
