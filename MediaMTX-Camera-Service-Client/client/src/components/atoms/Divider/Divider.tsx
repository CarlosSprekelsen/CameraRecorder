/**
 * Divider Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Visual separator component
 */

import React from 'react';

export interface DividerProps {
  orientation?: 'horizontal' | 'vertical';
  variant?: 'fullWidth' | 'inset' | 'middle';
  className?: string;
}

export const Divider: React.FC<DividerProps> = ({
  orientation = 'horizontal',
  variant = 'fullWidth',
  className = '',
  ...props
}) => {
  const baseClasses = 'border-gray-300';
  
  const orientationClasses = {
    horizontal: 'w-full border-t',
    vertical: 'h-full border-l'
  };

  const variantClasses = {
    fullWidth: '',
    inset: orientation === 'horizontal' ? 'mx-4' : 'my-4',
    middle: orientation === 'horizontal' ? 'mx-8' : 'my-8'
  };

  const classes = [
    baseClasses,
    orientationClasses[orientation],
    variantClasses[variant],
    className
  ].filter(Boolean).join(' ');

  return (
    <hr className={classes} {...props} />
  );
};
