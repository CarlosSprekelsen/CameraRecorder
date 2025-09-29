/**
 * CircularProgress Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Loading spinner component
 */

import React from 'react';

export interface CircularProgressProps {
  size?: number | string;
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'error';
  thickness?: number;
  className?: string;
}

export const CircularProgress: React.FC<CircularProgressProps> = ({
  size = 24,
  color = 'primary',
  thickness = 3.6,
  className = '',
  ...props
}) => {
  const sizeValue = typeof size === 'number' ? `${size}px` : size;
  
  const colorClasses = {
    primary: 'text-blue-600',
    secondary: 'text-gray-600',
    success: 'text-green-600',
    warning: 'text-yellow-600',
    error: 'text-red-600'
  };

  const classes = [
    'animate-spin',
    colorClasses[color],
    className
  ].filter(Boolean).join(' ');

  return (
    <div
      className={classes}
      style={{
        width: sizeValue,
        height: sizeValue,
        border: `${thickness}px solid currentColor`,
        borderTop: `${thickness}px solid transparent`,
        borderRadius: '50%'
      }}
      {...props}
    />
  );
};
