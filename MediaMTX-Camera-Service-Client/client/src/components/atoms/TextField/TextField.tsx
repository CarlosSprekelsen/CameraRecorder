/**
 * TextField Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Input component for form fields
 */

import React from 'react';

export interface TextFieldProps {
  label?: string;
  value?: string | number;
  onChange?: (value: string) => void;
  type?: 'text' | 'number' | 'email' | 'password';
  placeholder?: string;
  disabled?: boolean;
  error?: boolean;
  helperText?: string;
  className?: string;
  fullWidth?: boolean;
  select?: boolean;
  children?: React.ReactNode;
}

export const TextField: React.FC<TextFieldProps> = ({
  label,
  value,
  onChange,
  type = 'text',
  placeholder,
  disabled = false,
  error = false,
  helperText,
  className = '',
  fullWidth = false,
  select = false,
  children,
  ...props
}) => {
  const baseClasses = 'block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500';
  const errorClasses = error ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300';
  const disabledClasses = disabled ? 'bg-gray-50 cursor-not-allowed' : 'bg-white';
  const widthClasses = fullWidth ? 'w-full' : 'w-auto';
  
  const classes = [
    baseClasses,
    errorClasses,
    disabledClasses,
    widthClasses,
    className
  ].filter(Boolean).join(' ');

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    if (onChange) {
      onChange(e.target.value);
    }
  };

  return (
    <div className={`text-field ${fullWidth ? 'w-full' : ''}`}>
      {label && (
        <label className="block text-sm font-medium text-gray-700 mb-1">
          {label}
        </label>
      )}
      {select ? (
        <select
          value={value}
          onChange={handleChange}
          disabled={disabled}
          className={classes}
          {...props}
        >
          {children}
        </select>
      ) : (
        <input
          type={type}
          value={value}
          onChange={handleChange}
          placeholder={placeholder}
          disabled={disabled}
          className={classes}
          {...props}
        />
      )}
      {helperText && (
        <p className={`mt-1 text-sm ${error ? 'text-red-600' : 'text-gray-500'}`}>
          {helperText}
        </p>
      )}
    </div>
  );
};
