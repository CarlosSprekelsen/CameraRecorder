import React, { createContext, useContext, useEffect, useState } from 'react';
import { ThemeProvider, createTheme } from '../atoms/ThemeProvider/ThemeProvider';

interface AccessibilityContextType {
  highContrast: boolean;
  reducedMotion: boolean;
  fontSize: 'small' | 'medium' | 'large';
  toggleHighContrast: () => void;
  toggleReducedMotion: () => void;
  setFontSize: (size: 'small' | 'medium' | 'large') => void;
}

const AccessibilityContext = createContext<AccessibilityContextType | undefined>(undefined);

interface AccessibilityProviderProps {
  children: React.ReactNode;
}

/**
 * AccessibilityProvider - WCAG 2.1 AA compliance provider
 * Implements accessibility from Sprint 5 requirements
 */
export const AccessibilityProvider: React.FC<AccessibilityProviderProps> = ({ children }) => {
  const [highContrast, setHighContrast] = useState(false);
  const [reducedMotion, setReducedMotion] = useState(false);
  const [fontSize, setFontSize] = useState<'small' | 'medium' | 'large'>('medium');

  // Check for user preferences
  useEffect(() => {
    // Check for reduced motion preference
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      setReducedMotion(true);
    }

    // Check for high contrast preference
    if (window.matchMedia('(prefers-contrast: high)').matches) {
      setHighContrast(true);
    }

    // Listen for preference changes
    const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    const contrastQuery = window.matchMedia('(prefers-contrast: high)');

    const handleMotionChange = (e: MediaQueryListEvent) => setReducedMotion(e.matches);
    const handleContrastChange = (e: MediaQueryListEvent) => setHighContrast(e.matches);

    motionQuery.addEventListener('change', handleMotionChange);
    contrastQuery.addEventListener('change', handleContrastChange);

    return () => {
      motionQuery.removeEventListener('change', handleMotionChange);
      contrastQuery.removeEventListener('change', handleContrastChange);
    };
  }, []);

  // Create accessible theme
  const theme = createTheme({
    palette: {
      primary: {
        main: highContrast ? '#000000' : '#1976d2',
        light: highContrast ? '#333333' : '#42a5f5',
        dark: highContrast ? '#000000' : '#1565c0',
      },
      secondary: {
        main: highContrast ? '#000000' : '#dc004e',
        light: highContrast ? '#333333' : '#ff5983',
        dark: highContrast ? '#000000' : '#9a0036',
      },
      error: {
        main: highContrast ? '#000000' : '#f44336',
        light: highContrast ? '#333333' : '#e57373',
        dark: highContrast ? '#000000' : '#d32f2f',
      },
      warning: {
        main: highContrast ? '#000000' : '#ff9800',
        light: highContrast ? '#333333' : '#ffb74d',
        dark: highContrast ? '#000000' : '#f57c00',
      },
      info: {
        main: highContrast ? '#000000' : '#2196f3',
        light: highContrast ? '#333333' : '#64b5f6',
        dark: highContrast ? '#000000' : '#1976d2',
      },
      success: {
        main: highContrast ? '#000000' : '#4caf50',
        light: highContrast ? '#333333' : '#81c784',
        dark: highContrast ? '#000000' : '#388e3c',
      },
      background: {
        default: highContrast ? '#ffffff' : '#fafafa',
        paper: highContrast ? '#ffffff' : '#ffffff',
      },
      text: {
        primary: highContrast ? '#000000' : 'rgba(0, 0, 0, 0.87)',
        secondary: highContrast ? '#000000' : 'rgba(0, 0, 0, 0.6)',
        disabled: highContrast ? '#000000' : 'rgba(0, 0, 0, 0.38)',
      },
    },
    typography: {
      fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
      fontSize: fontSize === 'small' ? 12 : fontSize === 'large' ? 16 : 14,
    },
  });

  const toggleHighContrast = () => setHighContrast(!highContrast);
  const toggleReducedMotion = () => setReducedMotion(!reducedMotion);

  const value: AccessibilityContextType = {
    highContrast,
    reducedMotion,
    fontSize,
    toggleHighContrast,
    toggleReducedMotion,
    setFontSize,
  };

  return (
    <AccessibilityContext.Provider value={value}>
      <ThemeProvider theme={theme}>{children}</ThemeProvider>
    </AccessibilityContext.Provider>
  );
};

export const useAccessibility = (): AccessibilityContextType => {
  const context = useContext(AccessibilityContext);
  if (context === undefined) {
    throw new Error('useAccessibility must be used within an AccessibilityProvider');
  }
  return context;
};
