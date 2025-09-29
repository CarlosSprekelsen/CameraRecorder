/**
 * Typography Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Text component for consistent typography
 */

import React from 'react';

export interface TypographyProps {
  children: React.ReactNode;
  variant?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6' | 'body1' | 'body2' | 'caption';
  component?: keyof JSX.IntrinsicElements;
  className?: string;
  color?: 'primary' | 'secondary' | 'text' | 'error' | 'warning' | 'success';
  align?: 'left' | 'center' | 'right' | 'justify';
}

export const Typography: React.FC<TypographyProps> = ({
  children,
  variant = 'body1',
  component,
  className = '',
  color = 'text',
  align = 'left',
  ...props
}) => {
  const Component = component || (variant.startsWith('h') ? variant as keyof JSX.IntrinsicElements : 'p');
  
  const baseClasses = {
    h1: 'text-4xl font-bold',
    h2: 'text-3xl font-bold',
    h3: 'text-2xl font-semibold',
    h4: 'text-xl font-semibold',
    h5: 'text-lg font-medium',
    h6: 'text-base font-medium',
    body1: 'text-base',
    body2: 'text-sm',
    caption: 'text-xs'
  };

  const colorClasses = {
    primary: 'text-blue-600',
    secondary: 'text-gray-600',
    text: 'text-gray-900',
    error: 'text-red-600',
    warning: 'text-yellow-600',
    success: 'text-green-600'
  };

  const alignClasses = {
    left: 'text-left',
    center: 'text-center',
    right: 'text-right',
    justify: 'text-justify'
  };

  const classes = [
    baseClasses[variant],
    colorClasses[color],
    alignClasses[align],
    className
  ].filter(Boolean).join(' ');

  return (
    <Component className={classes} {...props}>
      {children}
    </Component>
  );
};
