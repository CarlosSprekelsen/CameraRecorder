import { useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { logger } from '../services/logger/LoggerService';

interface KeyboardShortcut {
  key: string;
  ctrlKey?: boolean;
  altKey?: boolean;
  shiftKey?: boolean;
  metaKey?: boolean;
  action: () => void;
  description: string;
}

/**
 * useKeyboardShortcuts - Keyboard navigation and shortcuts
 * Implements accessibility from Sprint 5 requirements
 */
export const useKeyboardShortcuts = () => {
  const navigate = useNavigate();

  const shortcuts: KeyboardShortcut[] = [
    {
      key: 'h',
      ctrlKey: true,
      action: () => navigate('/cameras'),
      description: 'Go to Cameras page',
    },
    {
      key: 'f',
      ctrlKey: true,
      action: () => navigate('/files'),
      description: 'Go to Files page',
    },
    {
      key: 'a',
      ctrlKey: true,
      action: () => navigate('/about'),
      description: 'Go to About page',
    },
    {
      key: 'r',
      ctrlKey: true,
      action: () => window.location.reload(),
      description: 'Reload page',
    },
    {
      key: 'Escape',
      action: () => {
        // Close any open dialogs or menus
        const activeElement = document.activeElement as HTMLElement;
        if (activeElement && activeElement.blur) {
          activeElement.blur();
        }
      },
      description: 'Close dialogs/menus',
    },
    {
      key: 'F1',
      action: () => {
        // Show help/shortcuts
        logger.info('Keyboard shortcuts help requested');
        // TODO: Implement help modal
      },
      description: 'Show help',
    },
  ];

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      const matchingShortcut = shortcuts.find((shortcut) => {
        return (
          shortcut.key === event.key &&
          !!shortcut.ctrlKey === event.ctrlKey &&
          !!shortcut.altKey === event.altKey &&
          !!shortcut.shiftKey === event.shiftKey &&
          !!shortcut.metaKey === event.metaKey
        );
      });

      if (matchingShortcut) {
        event.preventDefault();
        event.stopPropagation();

        logger.info('Keyboard shortcut triggered', {
          shortcut: matchingShortcut.description,
          key: event.key,
          modifiers: {
            ctrl: event.ctrlKey,
            alt: event.altKey,
            shift: event.shiftKey,
            meta: event.metaKey,
          },
        });

        matchingShortcut.action();
      }
    },
    [shortcuts],
  );

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [handleKeyDown]);

  return {
    shortcuts,
  };
};
