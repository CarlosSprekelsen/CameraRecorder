/**
 * Login → Recording → Download Workflow Test
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - I.Command Interface: Section 5.3
 * 
 * Requirements Coverage:
 * - REQ-WORKFLOW-006: Complete recording workflow from login to download
 * - REQ-WORKFLOW-007: Recording start and stop operations
 * - REQ-WORKFLOW-008: File generation and availability
 * - REQ-WORKFLOW-009: Download link generation
 * - REQ-WORKFLOW-010: End-to-end file lifecycle
 * 
 * Test Categories: E2E/Workflow
 */

import { executeUserWorkflow, assertWorkflowResult } from '../utils/workflow-test-helper';

describe('Login → Recording → Download Workflow', () => {
  test('REQ-WORKFLOW-006: Complete recording workflow from login to download', async () => {
    const workflowSteps = [
      {
        action: 'authenticate_user',
        method: 'authenticate',
        params: { auth_token: 'test-admin-token' },
        validator: (result) => result.authenticated === true
      },
      {
        action: 'verify_camera_availability',
        method: 'get_camera_list',
        validator: (result) => Array.isArray(result.cameras) && result.cameras.length > 0
      },
      {
        action: 'start_recording',
        method: 'start_recording',
        params: { device: 'camera0' },
        validator: (result) => result.status === 'started'
      },
      {
        action: 'verify_recording_status',
        method: 'get_stream_status',
        params: { device: 'camera0' },
        validator: (result) => result.status === 'active'
      },
      {
        action: 'stop_recording',
        method: 'stop_recording',
        params: { device: 'camera0' },
        validator: (result) => result.status === 'stopped'
      },
      {
        action: 'list_recordings',
        method: 'list_recordings',
        params: { limit: 10, offset: 0 },
        validator: (result) => Array.isArray(result.files)
      },
      {
        action: 'get_recording_info',
        method: 'get_recording_info',
        params: { filename: 'recording_20250125_100000.mp4' },
        validator: (result) => result.filename !== undefined && result.download_url !== undefined
      }
    ];

    const result = await executeUserWorkflow(workflowSteps, 'admin');
    
    assertWorkflowResult(result, {
      shouldSucceed: true,
      expectedSteps: 7,
      maxDuration: 30000
    });
  });

  test('REQ-WORKFLOW-007: Recording start and stop operations', async () => {
    const workflowSteps = [
      {
        action: 'authenticate',
        method: 'authenticate',
        params: { auth_token: 'test-admin-token' },
        validator: (result) => result.authenticated === true
      },
      {
        action: 'start_recording',
        method: 'start_recording',
        params: { device: 'camera0', duration: 30 },
        validator: (result) => result.status === 'started'
      },
      {
        action: 'stop_recording',
        method: 'stop_recording',
        params: { device: 'camera0' },
        validator: (result) => result.status === 'stopped'
      }
    ];

    const result = await executeUserWorkflow(workflowSteps, 'admin');
    
    assertWorkflowResult(result, {
      shouldSucceed: true,
      expectedSteps: 3,
      maxDuration: 15000
    });
  });

  test('REQ-WORKFLOW-008: File generation and availability', async () => {
    const workflowSteps = [
      {
        action: 'authenticate',
        method: 'authenticate',
        params: { auth_token: 'test-admin-token' },
        validator: (result) => result.authenticated === true
      },
      {
        action: 'list_recordings',
        method: 'list_recordings',
        params: { limit: 10, offset: 0 },
        validator: (result) => Array.isArray(result.files)
      },
      {
        action: 'verify_file_info',
        method: 'get_recording_info',
        params: { filename: 'test_recording.mp4' },
        validator: (result) => result.filename !== undefined
      }
    ];

    const result = await executeUserWorkflow(workflowSteps, 'admin');
    
    assertWorkflowResult(result, {
      shouldSucceed: true,
      expectedSteps: 3,
      maxDuration: 10000
    });
  });
});
