/**
 * AdminPanel Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-ADMIN-001: AdminPanel renders with admin controls
 * - REQ-ADMIN-002: AdminPanel handles retention policy configuration
 * - REQ-ADMIN-003: AdminPanel handles system management actions
 * - REQ-ADMIN-004: AdminPanel validates admin permissions
 * 
 * Test Categories: Unit/Component/Security
 */

import React from 'react';
import { AdminPanel } from '../../../../src/components/organisms/AdminPanel/AdminPanel';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('AdminPanel Component', () => {
  test('REQ-ADMIN-001: AdminPanel renders with admin controls', () => {
    const component = renderWithProviders(
      <AdminPanel />,
      { 
        withAuth: true, 
        withPermissions: ['admin'],
        initialStoreState: {
          authStore: { role: 'admin' }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['System Configuration', 'Retention Policies']
    });
  });

  test('REQ-ADMIN-002: AdminPanel handles retention policy configuration', () => {
    const component = renderWithProviders(
      <AdminPanel />,
      { 
        withAuth: true, 
        withPermissions: ['admin'],
        initialStoreState: {
          fileStore: { retentionPolicy: { days: 30 } }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Retention Policy', 'Days']
    });
  });

  test('REQ-ADMIN-003: AdminPanel handles system management actions', () => {
    const component = renderWithProviders(
      <AdminPanel />,
      { 
        withAuth: true, 
        withPermissions: ['admin'],
        initialStoreState: {
          serverStore: { systemStatus: 'healthy' }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['System Management', 'Health Status']
    });
  });

  test('REQ-ADMIN-004: AdminPanel validates admin permissions', () => {
    const component = renderWithProviders(
      <AdminPanel />,
      { 
        withAuth: true, 
        withPermissions: ['viewer'] // Non-admin user
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Access Denied', 'Admin privileges required']
    });
  });
});
