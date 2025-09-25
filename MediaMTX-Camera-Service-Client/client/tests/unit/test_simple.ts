/**
 * Simple test to verify Jest configuration
 * 
 * Ground Truth References:
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-TEST-001: Basic test execution
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

describe('Simple Test', () => {
  test('REQ-TEST-001: Basic test execution', () => {
    expect(1 + 1).toBe(2);
  });
  
  test('REQ-TEST-002: String test', () => {
    expect('hello').toBe('hello');
  });
});
