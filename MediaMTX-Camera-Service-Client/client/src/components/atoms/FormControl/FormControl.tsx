/**
 * FormControl Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Form control wrapper for consistent form styling
 */

import React from 'react';

export interface FormControlProps {
  children: React.ReactNode;
  fullWidth?: boolean;
  className?: string;
}

export const FormControl: React.FC<FormControlProps> = ({
  children,
  fullWidth = false,
  className = '',
  ...props
}) => {
  return (
    <div className={`form-control ${fullWidth ? 'w-full' : ''} ${className}`} {...props}>
      {children}
    </div>
  );
};

export interface InputLabelProps {
  children: React.ReactNode;
  htmlFor?: string;
  className?: string;
}

export const InputLabel: React.FC<InputLabelProps> = ({
  children,
  htmlFor,
  className = '',
  ...props
}) => {
  return (
    <label htmlFor={htmlFor} className={`input-label block text-sm font-medium text-gray-700 mb-1 ${className}`} {...props}>
      {children}
    </label>
  );
};
