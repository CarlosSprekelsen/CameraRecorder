/**
 * File Management → Delete Workflow Test
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - I.FileCatalog Interface: Section 5.3
 * 
 * Requirements Coverage:
 * - REQ-WORKFLOW-016: Complete file management workflow
 * - REQ-WORKFLOW-017: File deletion operations
 * - REQ-WORKFLOW-018: File listing and pagination
 * - REQ-WORKFLOW-019: File information retrieval
 * - REQ-WORKFLOW-020: Storage cleanup validation
 * 
 * Test Categories: E2E/Workflow
 */

import { executeUserWorkflow, assertWorkflowResult } from '../../utils/workflow-test-helper';

describe('File Management → Delete Workflow', () => {
  test('REQ-WORKFLOW-016: Complete file management workflow', async () => {
    const workflowSteps = [
      {
        action: 'authenticate_user',
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
        action: 'list_snapshots',
        method: 'list_snapshots',
        params: { limit: 10, offset: 0 },
        validator: (result) => Array.isArray(result.files)
      },
      {
        action: 'get_file_info',
        method: 'get_recording_info',
        params: { filename: 'test_recording.mp4' },
        validator: (result) => result.filename !== undefined
      },
      {
        action: 'delete_file',
        method: 'delete_recording',
        params: { filename: 'test_recording.mp4' },
        validator: (result) => result.success === true
      },
      {
        action: 'verify_deletion',
        method: 'get_recording_info',
        params: { filename: 'test_recording.mp4' },
        validator: (result) => result.error !== undefined
      }
    ];

    const result = await executeUserWorkflow(workflowSteps, 'admin');
    
    assertWorkflowResult(result, {
      shouldSucceed: true,
      expectedSteps: 6,
      maxDuration: 25000
    });
  });

  test('REQ-WORKFLOW-017: File deletion operations', async () => {
    const workflowSteps = [
      {
        action: 'authenticate',
        method: 'authenticate',
        params: { auth_token: 'test-admin-token' },
        validator: (result) => result.authenticated === true
      },
      {
        action: 'delete_recording',
        method: 'delete_recording',
        params: { filename: 'old_recording.mp4' },
        validator: (result) => result.success === true
      },
      {
        action: 'delete_snapshot',
        method: 'delete_snapshot',
        params: { filename: 'old_snapshot.jpg' },
        validator: (result) => result.success === true
      }
    ];

    const result = await executeUserWorkflow(workflowSteps, 'admin');
    
    assertWorkflowResult(result, {
      shouldSucceed: true,
      expectedSteps: 3,
      maxDuration: 15000
    });
  });

  test('REQ-WORKFLOW-018: File listing and pagination', async () => {
    const workflowSteps = [
      {
        action: 'authenticate',
        method: 'authenticate',
        params: { auth_token: 'test-admin-token' },
        validator: (result) => result.authenticated === true
      },
      {
        action: 'list_first_page',
        method: 'list_recordings',
        params: { limit: 5, offset: 0 },
        validator: (result) => Array.isArray(result.files) && result.files.length <= 5
      },
      {
        action: 'list_second_page',
        method: 'list_recordings',
        params: { limit: 5, offset: 5 },
        validator: (result) => Array.isArray(result.files)
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
