/**
 * Store Test Helper - Lean utility for Zustand store testing
 * 
 * REUSES existing utilities to minimize code bloat
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-STORE-001: Store mock creation
 * - REQ-STORE-002: State transition testing
 * - REQ-STORE-003: Action validation
 * 
 * Test Categories: Unit/Store
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

// Imports removed - not used in actual code

export interface StoreTestConfig<T> {
  initialState: T;
  actions: Record<string, any>;
}

// createMockStore removed - use MockDataFactory.createMockStore directly

/**
 * Test store state transition after action execution
 * REUSES: Standard Jest assertions for state validation
 */
export const testStoreTransition = <T>(
  store: { state: T; actions: Record<string, any> },
  actionName: string,
  actionParams: any[] = [],
  expectedState: Partial<T>
): void => {
  // Execute action
  const action = store.actions[actionName];
  if (typeof action === 'function') {
    action(...actionParams);
  }

  // Validate state transition
  Object.entries(expectedState).forEach(([key, expectedValue]) => {
    expect(store.state[key as keyof T]).toEqual(expectedValue);
  });
};
