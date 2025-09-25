/**
 * useKeyboardShortcuts hook unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-HOOK-001: Keyboard shortcut registration and handling
 * - REQ-HOOK-002: Navigation shortcuts functionality
 * - REQ-HOOK-003: Event listener management
 * - REQ-HOOK-004: Shortcut matching logic
 * - REQ-HOOK-005: Logger integration
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { renderHook } from '@testing-library/react';
import { useKeyboardShortcuts } from '../../../src/hooks/useKeyboardShortcuts';

// Mock react-router-dom
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate
}));

// Mock logger service
jest.mock('../../../src/services/logger/LoggerService', () => ({
  logger: {
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn()
  }
}));

describe('useKeyboardShortcuts Hook Unit Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset DOM
    document.body.innerHTML = '';
  });

  test('REQ-HOOK-001: Should register keyboard shortcuts correctly', () => {
    // Arrange & Act
    const { result } = renderHook(() => useKeyboardShortcuts());

    // Assert
    expect(result.current.shortcuts).toHaveLength(6);
    expect(result.current.shortcuts[0]).toEqual({
      key: 'h',
      ctrlKey: true,
      action: expect.any(Function),
      description: 'Go to Cameras page'
    });
    expect(result.current.shortcuts[1]).toEqual({
      key: 'f',
      ctrlKey: true,
      action: expect.any(Function),
      description: 'Go to Files page'
    });
    expect(result.current.shortcuts[2]).toEqual({
      key: 'a',
      ctrlKey: true,
      action: expect.any(Function),
      description: 'Go to About page'
    });
    expect(result.current.shortcuts[3]).toEqual({
      key: 'r',
      ctrlKey: true,
      action: expect.any(Function),
      description: 'Reload page'
    });
    expect(result.current.shortcuts[4]).toEqual({
      key: 'Escape',
      action: expect.any(Function),
      description: 'Close dialogs/menus'
    });
    expect(result.current.shortcuts[5]).toEqual({
      key: 'F1',
      action: expect.any(Function),
      description: 'Show help'
    });
  });

  test('REQ-HOOK-002: Should handle Ctrl+H navigation shortcut', () => {
    // Arrange
    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'h',
      ctrlKey: true,
      bubbles: true
    });

    // Act
    document.dispatchEvent(event);

    // Assert
    expect(mockNavigate).toHaveBeenCalledWith('/cameras');
  });

  test('REQ-HOOK-003: Should handle Ctrl+F navigation shortcut', () => {
    // Arrange
    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'f',
      ctrlKey: true,
      bubbles: true
    });

    // Act
    document.dispatchEvent(event);

    // Assert
    expect(mockNavigate).toHaveBeenCalledWith('/files');
  });

  test('REQ-HOOK-004: Should handle Ctrl+A navigation shortcut', () => {
    // Arrange
    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'a',
      ctrlKey: true,
      bubbles: true
    });

    // Act
    document.dispatchEvent(event);

    // Assert
    expect(mockNavigate).toHaveBeenCalledWith('/about');
  });

  test('REQ-HOOK-005: Should handle Escape key shortcut', () => {
    // Arrange
    const mockElement = document.createElement('input');
    mockElement.focus = jest.fn();
    mockElement.blur = jest.fn();
    Object.defineProperty(document, 'activeElement', {
      value: mockElement,
      writable: true
    });

    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'Escape',
      bubbles: true
    });

    // Act
    document.dispatchEvent(event);

    // Assert
    expect(mockElement.blur).toHaveBeenCalled();
  });

  test('REQ-HOOK-006: Should not trigger shortcuts without correct modifiers', () => {
    // Arrange
    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'h',
      ctrlKey: false,
      bubbles: true
    });

    // Act
    document.dispatchEvent(event);

    // Assert
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  test('REQ-HOOK-007: Should not trigger shortcuts for non-matching keys', () => {
    // Arrange
    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'x',
      ctrlKey: true,
      bubbles: true
    });

    // Act
    document.dispatchEvent(event);

    // Assert
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  test('REQ-HOOK-008: Should handle multiple modifier keys correctly', () => {
    // Arrange
    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'h',
      ctrlKey: true,
      altKey: true,
      shiftKey: true,
      metaKey: true,
      bubbles: true
    });

    // Act
    document.dispatchEvent(event);

    // Assert
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  test('REQ-HOOK-009: Should prevent default and stop propagation for matching shortcuts', () => {
    // Arrange
    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'h',
      ctrlKey: true,
      bubbles: true
    });
    const preventDefaultSpy = jest.spyOn(event, 'preventDefault');
    const stopPropagationSpy = jest.spyOn(event, 'stopPropagation');

    // Act
    document.dispatchEvent(event);

    // Assert
    expect(preventDefaultSpy).toHaveBeenCalled();
    expect(stopPropagationSpy).toHaveBeenCalled();
  });

  test('REQ-HOOK-010: Should clean up event listeners on unmount', () => {
    // Arrange
    const removeEventListenerSpy = jest.spyOn(document, 'removeEventListener');
    const { unmount } = renderHook(() => useKeyboardShortcuts());

    // Act
    unmount();

    // Assert
    expect(removeEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
  });

  test('REQ-HOOK-011: Should handle activeElement without blur method', () => {
    // Arrange
    const mockElement = document.createElement('div');
    Object.defineProperty(document, 'activeElement', {
      value: mockElement,
      writable: true
    });

    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'Escape',
      bubbles: true
    });

    // Act & Assert (should not throw)
    expect(() => document.dispatchEvent(event)).not.toThrow();
  });

  test('REQ-HOOK-012: Should handle null activeElement', () => {
    // Arrange
    Object.defineProperty(document, 'activeElement', {
      value: null,
      writable: true
    });

    renderHook(() => useKeyboardShortcuts());
    const event = new KeyboardEvent('keydown', {
      key: 'Escape',
      bubbles: true
    });

    // Act & Assert (should not throw)
    expect(() => document.dispatchEvent(event)).not.toThrow();
  });
});
