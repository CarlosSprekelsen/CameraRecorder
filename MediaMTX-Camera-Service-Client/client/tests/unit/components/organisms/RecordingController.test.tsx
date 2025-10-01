/**
 * RecordingController Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-RECORDING-001: RecordingController renders with recording controls
 * - REQ-RECORDING-002: RecordingController handles start recording
 * - REQ-RECORDING-003: RecordingController handles stop recording
 * - REQ-RECORDING-004: RecordingController displays recording status
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { RecordingController } from '../../../../src/components/organisms/RecordingController/RecordingController';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('RecordingController Component', () => {
  test('REQ-RECORDING-001: RecordingController renders with recording controls', () => {
    const component = renderWithProviders(
      <RecordingController />,
      { 
        withStores: true,
        initialStoreState: {
          recordingStore: { activeRecordings: {} }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Recording Controls', 'Start Recording', 'Stop Recording']
    });
  });

  test('REQ-RECORDING-002: RecordingController handles start recording', () => {
    const component = renderWithProviders(
      <RecordingController />,
      { 
        withStores: true,
        initialStoreState: {
          recordingStore: { 
            activeRecordings: {},
            loading: false 
          }
        }
      }
    );
    
    const startButton = component.getByRole('button', { name: /start recording/i });
    expect(startButton).toBeInTheDocument();
    expect(startButton).not.toBeDisabled();
  });

  test('REQ-RECORDING-003: RecordingController handles stop recording', () => {
    const component = renderWithProviders(
      <RecordingController />,
      { 
        withStores: true,
        initialStoreState: {
          recordingStore: { 
            activeRecordings: { 'camera0': { id: 'rec1', status: 'recording' } }
          }
        }
      }
    );
    
    const stopButton = component.getByRole('button', { name: /stop recording/i });
    expect(stopButton).toBeInTheDocument();
    expect(stopButton).not.toBeDisabled();
  });

  test('REQ-RECORDING-004: RecordingController displays recording status', () => {
    const component = renderWithProviders(
      <RecordingController device="camera0" />,
      { 
        withStores: true,
        initialStoreState: {
          recordingStore: { 
            activeRecordings: { 
              'camera0': { 
                device: 'camera0',
                status: 'RECORDING',
                startTime: '2025-01-15T14:30:00Z'
              } 
            }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Status: RECORDING', 'camera0']
    });
  });
});
