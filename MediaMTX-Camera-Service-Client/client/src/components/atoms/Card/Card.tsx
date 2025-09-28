/**
 * Card Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Basic building block for all card-like containers
 */

import React from 'react';

export interface CardProps {
  children: React.ReactNode;
  title?: string;
  subtitle?: string;
  variant?: 'default' | 'elevated' | 'outlined';
  padding?: 'none' | 'small' | 'medium' | 'large';
  className?: string;
  onClick?: () => void;
}

export const Card: React.FC<CardProps> = ({
  children,
  title,
  subtitle,
  variant = 'default',
  padding = 'medium',
  className = '',
  onClick,
}) => {
  const baseClasses = 'rounded-lg border transition-shadow';
  const clickableClasses = onClick ? 'cursor-pointer hover:shadow-md' : '';
  
  const variantClasses = {
    default: 'bg-white border-gray-200',
    elevated: 'bg-white border-gray-200 shadow-md',
    outlined: 'bg-white border-gray-300 border-2',
  };
  
  const paddingClasses = {
    none: '',
    small: 'p-3',
    medium: 'p-4',
    large: 'p-6',
  };

  return (
    <div
      className={`${baseClasses} ${variantClasses[variant]} ${paddingClasses[padding]} ${clickableClasses} ${className}`}
      onClick={onClick}
    >
      {(title || subtitle) && (
        <div className="mb-3">
          {title && (
            <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
          )}
          {subtitle && (
            <p className="text-sm text-gray-600">{subtitle}</p>
          )}
        </div>
      )}
      {children}
    </div>
  );
};
