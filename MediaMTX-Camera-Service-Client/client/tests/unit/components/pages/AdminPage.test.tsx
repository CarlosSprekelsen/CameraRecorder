/**
 * AdminPage Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Security Architecture: Section 8.3
 * 
 * Requirements Coverage:
 * - REQ-ADMINPAGE-001: AdminPage renders admin interface
 * - REQ-ADMINPAGE-002: AdminPage validates admin permissions
 * - REQ-ADMINPAGE-003: AdminPage displays system configuration
 * - REQ-ADMINPAGE-004: AdminPage handles admin actions
 * - REQ-ADMINPAGE-005: AdminPage shows access denied for non-admin
 * 
 * Test Categories: Unit/Component/Security
 */

import React from 'react';
import AdminPage from '../../../../src/pages/Admin/AdminPage';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('AdminPage Component', () => {
  test('REQ-ADMINPAGE-001: AdminPage renders admin interface', () => {
    const component = renderWithProviders(
      <AdminPage />,
      { 
        withAuth: true,
        withPermissions: ['admin'],
        withStores: true,
        initialStoreState: {
          authStore: { role: 'admin' },
          serverStore: { systemStatus: 'healthy' }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Admin Panel', 'System Configuration']
    });
  });

  test('REQ-ADMINPAGE-002: AdminPage validates admin permissions', () => {
    const component = renderWithProviders(
      <AdminPage />,
      { 
        withAuth: true,
        withPermissions: ['admin'],
        withStores: true,
        initialStoreState: {
          authStore: { role: 'admin' }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Admin Panel']
    });
  });

  test('REQ-ADMINPAGE-003: AdminPage displays system configuration', () => {
    const component = renderWithProviders(
      <AdminPage />,
      { 
        withAuth: true,
        withPermissions: ['admin'],
        withStores: true,
        initialStoreState: {
          authStore: { role: 'admin' },
          serverStore: { 
            systemStatus: 'healthy',
            storageInfo: { total: 1000000, used: 500000 }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['System Configuration', 'Storage']
    });
  });

  test('REQ-ADMINPAGE-004: AdminPage handles admin actions', () => {
    const component = renderWithProviders(
      <AdminPage />,
      { 
        withAuth: true,
        withPermissions: ['admin'],
        withStores: true,
        initialStoreState: {
          authStore: { role: 'admin' }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['System Management', 'Retention Policy']
    });
  });

  test('REQ-ADMINPAGE-005: AdminPage shows access denied for non-admin', () => {
    const component = renderWithProviders(
      <AdminPage />,
      { 
        withAuth: true,
        withPermissions: ['viewer'],
        withStores: true,
        initialStoreState: {
          authStore: { role: 'viewer' }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Access Denied', 'Admin privileges required']
    });
  });
});
