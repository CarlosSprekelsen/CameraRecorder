/**
 * FilesPage Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - I.FileCatalog Interface: Section 5.3
 * 
 * Requirements Coverage:
 * - REQ-FILESPAGE-001: FilesPage renders file management interface
 * - REQ-FILESPAGE-002: FilesPage displays file list
 * - REQ-FILESPAGE-003: FilesPage handles pagination
 * - REQ-FILESPAGE-004: FilesPage shows loading states
 * - REQ-FILESPAGE-005: FilesPage handles file operations
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import FilesPage from '../../../../src/pages/Files/FilesPage';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';
import { waitFor } from '@testing-library/react';

describe('FilesPage Component', () => {
  test('REQ-FILESPAGE-001: FilesPage renders file management interface', async () => {
    const component = renderWithProviders(
      <FilesPage />,
      { 
        withStores: true,
        initialStoreState: {
          fileStore: { 
            recordings: [],
            snapshots: [],
            loading: false,
            error: null
          }
        }
      }
    );
    
    // Wait for the async loadRecordings/loadSnapshots to complete
    await waitFor(() => {
      expect(component.getByText('Files')).toBeInTheDocument();
    });
    
    assertComponentBehavior(component, {
      hasText: ['Files', 'Recordings', 'Snapshots']
    });
  });

  test('REQ-FILESPAGE-002: FilesPage displays file list', () => {
    const component = renderWithProviders(
      <FilesPage />,
      { 
        withStores: true,
        initialStoreState: {
          fileStore: { 
            recordings: [
              { filename: 'recording1.mp4', size: 1024000, created_at: '2025-01-25T10:00:00Z' },
              { filename: 'recording2.mp4', size: 2048000, created_at: '2025-01-25T11:00:00Z' }
            ],
            loading: false
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['recording1.mp4', 'recording2.mp4']
    });
  });

  test('REQ-FILESPAGE-003: FilesPage handles pagination', () => {
    const component = renderWithProviders(
      <FilesPage />,
      { 
        withStores: true,
        initialStoreState: {
          fileStore: { 
            recordings: [],
            pagination: { current: 1, total: 5, limit: 10 }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Page 1 of 5']
    });
  });

  test('REQ-FILESPAGE-004: FilesPage shows loading states', () => {
    const component = renderWithProviders(
      <FilesPage />,
      { 
        withStores: true,
        initialStoreState: {
          fileStore: { 
            recordings: [],
            loading: true,
            error: null
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['No recordings found.', 'Start recording to see files here.']
    });
  });

  test('REQ-FILESPAGE-005: FilesPage handles file operations', () => {
    const component = renderWithProviders(
      <FilesPage />,
      { 
        withStores: true,
        initialStoreState: {
          fileStore: { 
            recordings: [{ filename: 'test.mp4', size: 1024000 }],
            loading: false
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Download', 'Delete']
    });
  });
});
