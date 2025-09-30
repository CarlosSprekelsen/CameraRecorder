/**
 * Container Atom - Atomic Design Pattern
 */

import React from 'react';

export interface ContainerProps {
  children: React.ReactNode;
  maxWidth?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | false;
  className?: string;
}

export const Container: React.FC<ContainerProps> = ({
  children,
  maxWidth = 'lg',
  className = '',
  ...props
}) => {
  const maxWidthClasses = {
    xs: 'max-w-xs',
    sm: 'max-w-sm', 
    md: 'max-w-md',
    lg: 'max-w-4xl',
    xl: 'max-w-6xl',
    false: ''
  };

  return (
    <div className={`container mx-auto px-4 ${maxWidth ? maxWidthClasses[maxWidth] : ''} ${className}`} {...props}>
      {children}
    </div>
  );
};
