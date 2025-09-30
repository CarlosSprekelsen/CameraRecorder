/**
 * Paper Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Paper component for elevated surfaces
 */

import React from 'react';

export interface PaperProps {
  children: React.ReactNode;
  elevation?: number;
  variant?: 'elevation' | 'outlined';
  className?: string;
}

export const Paper: React.FC<PaperProps> = ({
  children,
  elevation = 1,
  variant = 'elevation',
  className = '',
  ...props
}) => {
  const elevationClasses = {
    0: 'shadow-none',
    1: 'shadow-sm',
    2: 'shadow',
    3: 'shadow-md',
    4: 'shadow-lg',
    5: 'shadow-xl',
  };

  const variantClasses = {
    elevation: elevationClasses[elevation as keyof typeof elevationClasses] || 'shadow-sm',
    outlined: 'border border-gray-200',
  };

  return (
    <div
      className={`paper bg-white rounded ${variantClasses[variant]} ${className}`}
      {...props}
    >
      {children}
    </div>
  );
};
