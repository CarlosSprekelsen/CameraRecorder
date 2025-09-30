/**
 * AppBar Atom - Atomic Design Pattern
 */

import React from 'react';

export interface AppBarProps {
  children: React.ReactNode;
  position?: 'fixed' | 'absolute' | 'sticky' | 'static' | 'relative';
  className?: string;
}

export const AppBar: React.FC<AppBarProps> = ({
  children,
  position = 'fixed',
  className = '',
  ...props
}) => {
  const positionClasses = {
    fixed: 'fixed',
    absolute: 'absolute',
    sticky: 'sticky',
    static: 'static',
    relative: 'relative',
  };

  return (
    <header className={`appbar bg-white shadow-sm border-b border-gray-200 z-40 ${positionClasses[position]} ${className}`} {...props}>
      {children}
    </header>
  );
};

export interface ToolbarProps {
  children: React.ReactNode;
  className?: string;
}

export const Toolbar: React.FC<ToolbarProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <div className={`toolbar flex items-center justify-between px-4 py-2 min-h-16 ${className}`} {...props}>
      {children}
    </div>
  );
};
