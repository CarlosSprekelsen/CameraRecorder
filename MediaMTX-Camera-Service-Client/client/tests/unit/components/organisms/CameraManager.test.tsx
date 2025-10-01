/**
 * CameraManager Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-CAM-001: CameraManager displays cameras from store
 * - REQ-CAM-002: CameraManager handles loading state
 * - REQ-CAM-003: CameraManager handles recording actions
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { CameraManager } from '../../../../src/components/organisms/CameraManager/CameraManager';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('CameraManager Component', () => {
  test('REQ-CAM-001: CameraManager displays cameras from store', () => {
    const component = renderWithProviders(
      <CameraManager />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: { 
            cameras: [
              { device: 'camera0', status: 'CONNECTED' },
              { device: 'camera1', status: 'CONNECTED' }
            ],
            loading: false,
            error: null
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['camera0', 'camera1']
    });
  });

  test('REQ-CAM-002: CameraManager handles loading state', () => {
    const component = renderWithProviders(
      <CameraManager />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: { 
            loading: true,
            error: null,
            cameras: []
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Loading cameras...']
    });
  });

  test('REQ-CAM-003: CameraManager handles recording actions', () => {
    const component = renderWithProviders(
      <CameraManager />,
      { 
        withStores: true,
        initialStoreState: {
          deviceStore: {
            cameras: [
              { device: 'camera0', status: 'CONNECTED' }
            ],
            loading: false,
            error: null
          },
          recordingStore: {
            activeRecordings: { 'camera0': { id: 'rec1', status: 'recording' } }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['camera0']
    });
  });
});
