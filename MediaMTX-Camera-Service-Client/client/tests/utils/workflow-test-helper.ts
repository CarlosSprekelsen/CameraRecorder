/**
 * Workflow Test Helper - Lean utility for end-to-end workflow testing
 * 
 * REUSES existing utilities to minimize code bloat:
 * - TestAPIClient for API calls
 * - AuthHelper for authentication
 * - APIResponseValidator for response validation
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-WORKFLOW-001: Workflow execution
 * - REQ-WORKFLOW-002: Result validation
 * - REQ-WORKFLOW-003: Error handling
 * 
 * Test Categories: E2E/Workflow
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { TestAPIClient } from './api-client';
import { AuthHelper } from './auth-helper';
import { APIResponseValidator } from './validators';
import { AuthService } from '../../src/services/auth/AuthService';
import { LoggerService } from '../../src/services/logger/LoggerService';

export interface WorkflowStep {
  action: string;
  method: string;
  params?: Record<string, any>;
  expectedResult?: any;
  validator?: (result: any) => boolean;
}

export interface WorkflowResult {
  success: boolean;
  steps: Array<{
    step: WorkflowStep;
    result: any;
    success: boolean;
    error?: string;
  }>;
  totalDuration: number;
}

/**
 * Execute complete user workflow with multiple API calls
 * REUSES: TestAPIClient for API calls, AuthHelper for authentication
 */
export const executeUserWorkflow = async (
  steps: WorkflowStep[],
  role: 'admin' | 'operator' | 'viewer' = 'admin'
): Promise<WorkflowResult> => {
  const startTime = Date.now();
  const results: WorkflowResult['steps'] = [];

  // REUSE: TestAPIClient for API communication
  const apiClient = new TestAPIClient({ mockMode: false });
  await apiClient.connect();

  // REUSE: AuthHelper for authentication
  const token = AuthHelper.generateTestToken(role);
  
  // Create AuthService following architectural pattern
  const authService = new AuthService(apiClient, LoggerService.getInstance());
  await authService.authenticate(token);

  let workflowSuccess = true;

  for (const step of steps) {
    try {
      // Execute API call
      const result = await apiClient.call(step.method, step.params || {});
      
      // Validate result if validator provided
      const stepSuccess = step.validator ? step.validator(result) : true;
      
      results.push({
        step,
        result,
        success: stepSuccess,
        error: stepSuccess ? undefined : 'Validation failed'
      });

      if (!stepSuccess) {
        workflowSuccess = false;
        break;
      }
    } catch (error) {
      results.push({
        step,
        result: null,
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error'
      });
      workflowSuccess = false;
      break;
    }
  }

  await apiClient.disconnect();

  return {
    success: workflowSuccess,
    steps: results,
    totalDuration: Date.now() - startTime
  };
};

/**
 * Assert workflow result matches expected outcome
 * REUSES: APIResponseValidator for response validation
 */
export const assertWorkflowResult = (
  result: WorkflowResult,
  expectedOutcome: {
    shouldSucceed: boolean;
    expectedSteps?: number;
    maxDuration?: number;
  }
): void => {
  const { shouldSucceed, expectedSteps, maxDuration = 30000 } = expectedOutcome;

  expect(result.success).toBe(shouldSucceed);
  
  if (expectedSteps) {
    expect(result.steps).toHaveLength(expectedSteps);
  }
  
  expect(result.totalDuration).toBeLessThan(maxDuration);
  
  if (shouldSucceed) {
    expect(result.steps.every(step => step.success)).toBe(true);
  }
};
