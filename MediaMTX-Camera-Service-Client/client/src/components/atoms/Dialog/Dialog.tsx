/**
 * Dialog Atom - Atomic Design Pattern
 */

import React from 'react';

export interface DialogProps {
  open: boolean;
  onClose: () => void;
  children: React.ReactNode;
  maxWidth?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | false;
  fullWidth?: boolean;
  className?: string;
}

export const Dialog: React.FC<DialogProps> = ({
  open,
  onClose,
  children,
  maxWidth = 'sm',
  fullWidth = false,
  className = '',
  ...props
}) => {
  if (!open) return null;

  const maxWidthClasses = {
    xs: 'max-w-xs',
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
    false: 'max-w-none',
  };

  return (
    <div className="dialog fixed inset-0 z-50 overflow-y-auto">
      <div className="dialog-backdrop fixed inset-0 bg-black bg-opacity-50" onClick={onClose} />
      <div className="dialog-container flex min-h-full items-center justify-center p-4">
        <div className={`dialog-content relative bg-white rounded-lg shadow-xl ${maxWidth ? maxWidthClasses[maxWidth] : ''} ${fullWidth ? 'w-full' : ''} ${className}`} {...props}>
          <button
            className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded-full p-1"
            onClick={onClose}
            aria-label="Close dialog"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          {children}
        </div>
      </div>
    </div>
  );
};

export interface DialogTitleProps {
  children: React.ReactNode;
  className?: string;
}

export const DialogTitle: React.FC<DialogTitleProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <h2 className={`dialog-title text-lg font-medium text-gray-900 px-6 pt-6 ${className}`} {...props}>
      {children}
    </h2>
  );
};

export interface DialogContentProps {
  children: React.ReactNode;
  className?: string;
}

export const DialogContent: React.FC<DialogContentProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <div className={`dialog-content px-6 py-4 ${className}`} {...props}>
      {children}
    </div>
  );
};

export interface DialogActionsProps {
  children: React.ReactNode;
  className?: string;
}

export const DialogActions: React.FC<DialogActionsProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <div className={`dialog-actions flex justify-end space-x-2 px-6 pb-6 ${className}`} {...props}>
      {children}
    </div>
  );
};