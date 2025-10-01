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
  message?: string;
  autoHideDuration?: number;
  children?: React.ReactNode;
  className?: string;
}

export const Snackbar: React.FC<SnackbarProps> = ({
  open,
  onClose,
  message,
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
      <div className="bg-gray-800 text-white px-4 py-3 rounded-lg shadow-lg flex items-center justify-between min-w-80">
        <span>{message || children}</span>
        <button
          onClick={onClose}
          className="ml-4 text-white hover:text-gray-300 focus:outline-none focus:ring-2 focus:ring-white rounded-full p-1"
          aria-label="Close notification"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>
  );
};
