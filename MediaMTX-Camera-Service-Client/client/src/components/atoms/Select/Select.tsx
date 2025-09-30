/**
 * Select Atom - Atomic Design Pattern
 */

import React from 'react';

export interface SelectProps {
  children: React.ReactNode;
  value: string | number;
  onChange: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  displayEmpty?: boolean;
  className?: string;
  disabled?: boolean;
}

export const Select: React.FC<SelectProps> = ({
  children,
  value,
  onChange,
  displayEmpty = false,
  className = '',
  disabled = false,
  ...props
}) => {
  return (
    <select
      value={value}
      onChange={onChange}
      disabled={disabled}
      className={`select block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed ${className}`}
      {...props}
    >
      {displayEmpty && <option value="">Select an option</option>}
      {children}
    </select>
  );
};
