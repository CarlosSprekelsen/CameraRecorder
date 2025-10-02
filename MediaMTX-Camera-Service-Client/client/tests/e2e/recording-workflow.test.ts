/**
 * Complete Recording Workflow E2E Test
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-WORKFLOW-001: Complete recording workflow (login → record → download)
 * - REQ-WORKFLOW-002: Error handling during workflow
 * - REQ-WORKFLOW-003: Performance validation
 * 
 * Test Categories: E2E/Workflow
 */

import { executeUserWorkflow, assertWorkflowResult } from '../utils/workflow-test-helper';

describe('Recording Workflow E2E Tests', () => {
  test('REQ-WORKFLOW-001: Complete recording workflow', async () => {
    const workflowSteps = [
      {
        action: 'get_cameras',
        method: 'get_camera_list',
        validator: (result) => Array.isArray(result.cameras)
      },
      {
        action: 'start_recording',
        method: 'start_recording',
        params: { device: 'camera0' },
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
      maxDuration: 30000
    });
  });

  test('REQ-WORKFLOW-002: Error handling during workflow', async () => {
    const workflowSteps = [
      {
        action: 'authenticate',
        method: 'authenticate',
        params: { auth_token: 'invalid-token' },
        validator: (result) => result.authenticated === true
      }
    ];

    const result = await executeUserWorkflow(workflowSteps, 'admin');
    
    assertWorkflowResult(result, {
      shouldSucceed: false,
      expectedSteps: 1,
      maxDuration: 10000
    });
  });
});
