/**
 * Snackbar Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Snackbar component for notifications
 */

import React, { useEffect } from 'react';

export interface SnackbarProps {
  open: boolean;
  onClose: () => void;
  autoHideDuration?: number;
  children: React.ReactNode;
  className?: string;
}

export const Snackbar: React.FC<SnackbarProps> = ({
  open,
  onClose,
  autoHideDuration = 2500,
  children,
  className = '',
  ...props
}) => {
  useEffect(() => {
    if (open && autoHideDuration > 0) {
      const timer = setTimeout(onClose, autoHideDuration);
      return () => clearTimeout(timer);
    }
  }, [open, autoHideDuration, onClose]);

  if (!open) return null;

  return (
    <div
      className={`snackbar fixed bottom-4 left-1/2 transform -translate-x-1/2 z-50 ${className}`}
      {...props}
    >
      {children}
    </div>
  );
};
