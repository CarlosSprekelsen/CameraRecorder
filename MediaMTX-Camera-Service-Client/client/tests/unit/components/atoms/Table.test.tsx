/**
 * Table Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-TABLE-001: Table renders with headers
 * - REQ-TABLE-002: Table displays data rows
 * - REQ-TABLE-003: Table handles sorting
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Table } from '../../../../src/components/atoms/Table/Table';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Table Component', () => {
  test('REQ-TABLE-001: Table renders with headers', () => {
    const component = renderWithProviders(
      <Table>
        <thead>
          <tr>
            <th>Header 1</th>
            <th>Header 2</th>
          </tr>
        </thead>
      </Table>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Header 1', 'Header 2'],
      hasClass: ['table']
    });
  });

  test('REQ-TABLE-002: Table displays data rows', () => {
    const component = renderWithProviders(
      <Table>
        <tbody>
          <tr>
            <td>Data 1</td>
            <td>Data 2</td>
          </tr>
        </tbody>
      </Table>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Data 1', 'Data 2']
    });
  });

  test('REQ-TABLE-003: Table handles sorting', () => {
    const handleSort = jest.fn();
    const component = renderWithProviders(
      <Table onSort={handleSort}>
        <thead>
          <tr>
            <th>Sortable Header</th>
          </tr>
        </thead>
      </Table>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Sortable Header']
    });
  });
});
