/**
 * SystemStatusMonitor Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-SYSTEM-001: SystemStatusMonitor renders system health status
 * - REQ-SYSTEM-002: SystemStatusMonitor displays camera availability
 * - REQ-SYSTEM-003: SystemStatusMonitor shows discovery status
 * - REQ-SYSTEM-004: SystemStatusMonitor handles real-time updates
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { SystemStatusMonitor } from '../../../../src/components/organisms/SystemStatusMonitor/SystemStatusMonitor';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('SystemStatusMonitor Component', () => {
  test('REQ-SYSTEM-001: SystemStatusMonitor renders system health status', () => {
    const component = renderWithProviders(
      <SystemStatusMonitor />,
      { 
        withStores: true,
        initialStoreState: {
          serverStore: { 
            systemStatus: { 
              status: 'ready',
              available_cameras: ['camera0', 'camera1']
            }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['System Status', 'Ready']
    });
  });

  test('REQ-SYSTEM-002: SystemStatusMonitor displays camera availability', () => {
    const component = renderWithProviders(
      <SystemStatusMonitor />,
      { 
        withStores: true,
        initialStoreState: {
          serverStore: { 
            systemStatus: { 
              status: 'ready',
              available_cameras: ['camera0', 'camera1', 'camera2']
            }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Available Cameras', '3']
    });
  });

  test('REQ-SYSTEM-003: SystemStatusMonitor shows discovery status', () => {
    const component = renderWithProviders(
      <SystemStatusMonitor />,
      { 
        withStores: true,
        initialStoreState: {
          serverStore: { 
            systemStatus: { 
              discovery_active: true,
              discovery_interval: 30
            }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Discovery Active', '30s']
    });
  });

  test('REQ-SYSTEM-004: SystemStatusMonitor handles real-time updates', () => {
    const component = renderWithProviders(
      <SystemStatusMonitor />,
      { 
        withStores: true,
        initialStoreState: {
          serverStore: { 
            systemStatus: { 
              status: 'starting',
              last_updated: '2025-01-25T10:00:00Z'
            }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['System Status', 'Starting']
    });
  });
});
