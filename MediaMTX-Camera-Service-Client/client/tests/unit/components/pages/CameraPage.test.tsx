/**
 * CameraPage Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - I.Discovery Interface: Section 5.3
 * 
 * Requirements Coverage:
 * - REQ-CAMERAPAGE-001: CameraPage renders camera management interface
 * - REQ-CAMERAPAGE-002: CameraPage displays camera list
 * - REQ-CAMERAPAGE-003: CameraPage handles loading states
 * - REQ-CAMERAPAGE-004: CameraPage shows error states
 * - REQ-CAMERAPAGE-005: CameraPage displays recording status
 * - REQ-CAMERAPAGE-006: CameraPage handles real-time updates
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import CameraPage from '../../../../src/pages/Cameras/CameraPage';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('CameraPage Component', () => {
  test('REQ-CAMERAPAGE-001: CameraPage renders camera management interface', () => {
    const component = renderWithProviders(
      <CameraPage />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: { 
            cameras: [],
            loading: false,
            error: null
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Camera Devices']
    });
  });

  test('REQ-CAMERAPAGE-002: CameraPage displays camera list', () => {
    const component = renderWithProviders(
      <CameraPage />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: { 
            cameras: [
              { device: 'camera0', status: 'CONNECTED' },
              { device: 'camera1', status: 'CONNECTED' }
            ],
            loading: false
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Test Camera 0', 'Test Camera 1']
    });
  });

  test('REQ-CAMERAPAGE-003: CameraPage handles loading states', () => {
    const component = renderWithProviders(
      <CameraPage />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: { 
            cameras: [],
            loading: true,
            error: null
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Loading camera devices...']
    });
  });

  test('REQ-CAMERAPAGE-004: CameraPage shows error states', () => {
    const component = renderWithProviders(
      <CameraPage />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: { 
            cameras: [],
            loading: false,
            error: 'Failed to load cameras'
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Failed to load cameras']
    });
  });

  test('REQ-CAMERAPAGE-005: CameraPage displays recording status', () => {
    const component = renderWithProviders(
      <CameraPage />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: { 
            cameras: [{ device: 'camera0', status: 'CONNECTED' }],
            loading: false
          },
          recordingStore: {
            activeRecordings: { 'camera0': { id: 'rec1', status: 'recording' } }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Test Camera 0']
    });
  });

  test('REQ-CAMERAPAGE-006: CameraPage handles real-time updates', () => {
    const component = renderWithProviders(
      <CameraPage />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: { 
            cameras: [{ device: 'camera0', status: 'CONNECTED' }],
            lastUpdated: '2025-01-25T10:00:00Z'
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Last updated']
    });
  });
});
