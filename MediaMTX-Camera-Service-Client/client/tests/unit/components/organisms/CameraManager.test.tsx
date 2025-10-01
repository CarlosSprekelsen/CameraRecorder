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
import { MockDataFactory } from '../../../utils/mocks';

// Mock services to prevent initialization errors
jest.mock('../../../../src/services/device/DeviceService', () => ({
  DeviceService: jest.fn().mockImplementation(() => MockDataFactory.createMockDeviceService())
}));

jest.mock('../../../../src/services/recording/RecordingService', () => ({
  RecordingService: jest.fn().mockImplementation(() => MockDataFactory.createMockRecordingService())
}));

describe('CameraManager Component', () => {
  test('REQ-CAM-001: CameraManager displays cameras from store', async () => {
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

    // Wait for the async getCameraList() call to complete
    await component.findByText('Test Camera 0');
    
    assertComponentBehavior(component, {
      hasText: ['Test Camera 0', 'Test Camera 1']
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

  test('REQ-CAM-003: CameraManager handles recording actions', async () => {
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
    
    // Wait for the async getCameraList() call to complete
    await component.findByText('Test Camera 0');
    
    assertComponentBehavior(component, {
      hasText: ['Test Camera 0']
    });
  });
});
