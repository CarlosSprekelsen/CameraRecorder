/**
 * IconButton Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Icon button component for icon-only actions
 */

import React from 'react';

export interface IconButtonProps {
  children: React.ReactNode;
  onClick?: (event: React.MouseEvent<HTMLButtonElement>) => void;
  disabled?: boolean;
  color?: 'default' | 'primary' | 'secondary' | 'inherit';
  size?: 'small' | 'medium' | 'large';
  className?: string;
}

export const IconButton: React.FC<IconButtonProps> = ({
  children,
  onClick,
  disabled = false,
  color = 'default',
  size = 'medium',
  className = '',
  ...props
}) => {
  const sizeClasses = {
    small: 'p-1',
    medium: 'p-2',
    large: 'p-3',
  };

  const colorClasses = {
    default: 'text-gray-600 hover:text-gray-800 hover:bg-gray-100',
    primary: 'text-blue-600 hover:text-blue-800 hover:bg-blue-50',
    secondary: 'text-gray-500 hover:text-gray-700 hover:bg-gray-50',
    inherit: 'text-inherit hover:bg-gray-100',
  };

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`icon-button inline-flex items-center justify-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed ${sizeClasses[size]} ${colorClasses[color]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
};
