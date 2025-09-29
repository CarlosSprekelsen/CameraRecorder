/**
 * Box Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Basic container component for layout and spacing
 */

import React from 'react';

export interface BoxProps {
  children?: React.ReactNode;
  className?: string;
  sx?: React.CSSProperties;
  component?: keyof JSX.IntrinsicElements;
}

export const Box: React.FC<BoxProps> = ({
  children,
  className = '',
  sx = {},
  component: Component = 'div',
  ...props
}) => {
  const style = { ...sx };
  
  return (
    <Component 
      className={`box ${className}`}
      style={style}
      {...props}
    >
      {children}
    </Component>
  );
};
