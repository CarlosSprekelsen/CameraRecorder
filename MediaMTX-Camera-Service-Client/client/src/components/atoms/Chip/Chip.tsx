/**
 * Chip Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Small status indicator component
 */

import React from 'react';

export interface ChipProps {
  label: React.ReactNode;
  color?: 'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'error';
  size?: 'small' | 'medium';
  variant?: 'filled' | 'outlined';
  className?: string;
  icon?: React.ReactNode;
  onDelete?: () => void;
}

export const Chip: React.FC<ChipProps> = ({
  label,
  color = 'default',
  size = 'medium',
  variant = 'filled',
  className = '',
  icon,
  onDelete,
  ...props
}) => {
  const baseClasses = 'inline-flex items-center font-medium rounded-full border';
  
  const sizeClasses = {
    small: 'px-2 py-0.5 text-xs',
    medium: 'px-3 py-1 text-sm'
  };

  const colorClasses = {
    default: variant === 'filled' ? 'bg-gray-100 text-gray-800 border-gray-200' : 'bg-white text-gray-800 border-gray-300',
    primary: variant === 'filled' ? 'bg-blue-100 text-blue-800 border-blue-200' : 'bg-white text-blue-800 border-blue-300',
    secondary: variant === 'filled' ? 'bg-gray-100 text-gray-800 border-gray-200' : 'bg-white text-gray-800 border-gray-300',
    success: variant === 'filled' ? 'bg-green-100 text-green-800 border-green-200' : 'bg-white text-green-800 border-green-300',
    warning: variant === 'filled' ? 'bg-yellow-100 text-yellow-800 border-yellow-200' : 'bg-white text-yellow-800 border-yellow-300',
    error: variant === 'filled' ? 'bg-red-100 text-red-800 border-red-200' : 'bg-white text-red-800 border-red-300'
  };

  const classes = [
    baseClasses,
    sizeClasses[size],
    colorClasses[color],
    className
  ].filter(Boolean).join(' ');

  return (
    <span className={classes} {...props}>
      {icon && <span className="mr-1">{icon}</span>}
      {label}
      {onDelete && (
        <button
          onClick={onDelete}
          className="ml-1 hover:bg-black hover:bg-opacity-10 rounded-full p-0.5"
          type="button"
        >
          ×
        </button>
      )}
    </span>
  );
};
