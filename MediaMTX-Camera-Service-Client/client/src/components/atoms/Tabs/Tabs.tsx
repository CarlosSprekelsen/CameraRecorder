/**
 * Tabs Atom - Atomic Design Pattern
 */

import React, { useState } from 'react';

export interface TabsProps {
  children: React.ReactNode;
  value?: number;
  onChange?: (event: React.SyntheticEvent, newValue: number) => void;
  className?: string;
}

export const Tabs: React.FC<TabsProps> = ({
  children,
  value = 0,
  onChange,
  className = '',
  ...props
}) => {
  return (
    <div className={`tabs border-b border-gray-200 ${className}`} {...props}>
      <nav className="flex space-x-8">
        {React.Children.map(children, (child, index) => {
          if (React.isValidElement(child)) {
            return React.cloneElement(child, {
              ...child.props,
              selected: index === value,
              onClick: (e: React.MouseEvent) => {
                onChange?.(e, index);
              }
            });
          }
          return child;
        })}
      </nav>
    </div>
  );
};

export interface TabProps {
  label: React.ReactNode;
  selected?: boolean;
  onClick?: (event: React.MouseEvent) => void;
  className?: string;
}

export const Tab: React.FC<TabProps> = ({
  label,
  selected = false,
  onClick,
  className = '',
  ...props
}) => {
  return (
    <button
      className={`tab px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
        selected 
          ? 'border-blue-500 text-blue-600' 
          : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
      } ${className}`}
      onClick={onClick}
      {...props}
    >
      {label}
    </button>
  );
};
