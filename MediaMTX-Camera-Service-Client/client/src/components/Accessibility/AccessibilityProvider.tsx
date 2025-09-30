import React, { createContext, useContext, useEffect, useState } from 'react';
import { ThemeProvider, createTheme } from '../../atoms/ThemeProvider/ThemeProvider';

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
      mode: 'light',
      primary: {
        main: highContrast ? '#000000' : '#1976d2',
      },
      secondary: {
        main: highContrast ? '#000000' : '#dc004e',
      },
      contrastThreshold: highContrast ? 7 : 3,
    },
    typography: {
      fontSize: fontSize === 'small' ? 12 : fontSize === 'large' ? 16 : 14,
    },
    components: {
      MuiButton: {
        styleOverrides: {
          root: {
            minHeight: 44, // WCAG minimum touch target
            minWidth: 44,
          },
        },
      },
      MuiIconButton: {
        styleOverrides: {
          root: {
            minHeight: 44,
            minWidth: 44,
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            minHeight: 32,
          },
        },
      },
    },
    transitions: {
      duration: {
        shortest: reducedMotion ? 0 : 150,
        shorter: reducedMotion ? 0 : 200,
        short: reducedMotion ? 0 : 250,
        standard: reducedMotion ? 0 : 300,
        complex: reducedMotion ? 0 : 375,
        enteringScreen: reducedMotion ? 0 : 225,
        leavingScreen: reducedMotion ? 0 : 195,
      },
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
