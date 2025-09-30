/**
 * Badge Atom - Atomic Design Pattern
 */

import React from 'react';

export interface BadgeProps {
  children: React.ReactNode;
  badgeContent?: React.ReactNode;
  color?: 'default' | 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success';
  variant?: 'standard' | 'dot';
  max?: number;
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  badgeContent,
  color = 'default',
  variant = 'standard',
  max = 99,
  className = '',
  ...props
}) => {
  const colorClasses = {
    default: 'bg-gray-500 text-white',
    primary: 'bg-blue-500 text-white',
    secondary: 'bg-gray-400 text-white',
    error: 'bg-red-500 text-white',
    warning: 'bg-yellow-500 text-white',
    info: 'bg-blue-400 text-white',
    success: 'bg-green-500 text-white',
  };

  const displayContent = typeof badgeContent === 'number' && badgeContent > max 
    ? `${max}+` 
    : badgeContent;

  return (
    <span className={`badge relative inline-block ${className}`} {...props}>
      {children}
      {badgeContent !== undefined && (
        <span className={`badge-content absolute -top-2 -right-2 min-w-5 h-5 px-1 text-xs font-medium rounded-full flex items-center justify-center ${colorClasses[color]}`}>
          {variant === 'dot' ? '' : displayContent}
        </span>
      )}
    </span>
  );
};