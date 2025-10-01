/**
 * Switch Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Toggle switch component for boolean inputs
 */

import React from 'react';

export interface SwitchProps {
  checked?: boolean;
  onChange?: (checked: boolean) => void;
  disabled?: boolean;
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'error';
  size?: 'small' | 'medium' | 'large';
  className?: string;
}

export const Switch: React.FC<SwitchProps> = ({
  checked = false,
  onChange,
  disabled = false,
  color = 'primary',
  size = 'medium',
  className = '',
  ...props
}) => {
  const sizeClasses = {
    small: 'w-8 h-4',
    medium: 'w-11 h-6',
    large: 'w-14 h-8'
  };

  const colorClasses = {
    primary: checked ? 'bg-blue-600' : 'bg-gray-200',
    secondary: checked ? 'bg-gray-600' : 'bg-gray-200',
    success: checked ? 'bg-green-600' : 'bg-gray-200',
    warning: checked ? 'bg-yellow-600' : 'bg-gray-200',
    error: checked ? 'bg-red-600' : 'bg-gray-200'
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (onChange && !disabled) {
      onChange(e.target.checked);
    }
  };

  return (
    <label className={`relative inline-flex items-center cursor-pointer ${disabled ? 'cursor-not-allowed' : ''} ${className}`}>
      <input
        type="checkbox"
        role="switch"
        checked={checked}
        onChange={handleChange}
        disabled={disabled}
        className="sr-only"
        {...props}
      />
      <div className={`${sizeClasses[size]} ${colorClasses[color]} rounded-full transition-colors duration-200 ease-in-out`}>
        <div className={`absolute top-0.5 left-0.5 bg-white rounded-full transition-transform duration-200 ease-in-out ${
          size === 'small' ? 'w-3 h-3' : size === 'medium' ? 'w-5 h-5' : 'w-7 h-7'
        } ${checked ? (size === 'small' ? 'translate-x-4' : size === 'medium' ? 'translate-x-5' : 'translate-x-6') : 'translate-x-0'}`} />
      </div>
    </label>
  );
};
