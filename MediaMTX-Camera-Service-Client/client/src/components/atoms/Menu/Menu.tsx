/**
 * Menu Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Menu component for dropdown actions
 */

import React from 'react';

export interface MenuProps {
  children: React.ReactNode;
  open: boolean;
  onClose: () => void;
  anchorEl?: HTMLElement | null;
  className?: string;
}

export const Menu: React.FC<MenuProps> = ({
  children,
  open,
  onClose,
  anchorEl,
  className = '',
  ...props
}) => {
  if (!open) return null;

  const position = anchorEl ? {
    top: anchorEl.getBoundingClientRect().bottom + window.scrollY,
    left: anchorEl.getBoundingClientRect().left + window.scrollX,
  } : { top: 0, left: 0 };

  return (
    <div
      className={`menu fixed bg-white border border-gray-200 rounded-md shadow-lg z-50 min-w-48 open ${className}`}
      style={{ top: position.top, left: position.left }}
      {...props}
    >
      <button
        className="absolute top-2 right-2 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded-full p-1"
        onClick={onClose}
        aria-label="Close menu"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
      {children}
    </div>
  );
};

export interface MenuItemProps {
  children: React.ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  className?: string;
}

export const MenuItem: React.FC<MenuItemProps> = ({
  children,
  onClick,
  disabled = false,
  className = '',
  ...props
}) => {
  return (
    <button
      className={`menu-item w-full text-left px-4 py-2 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed ${className}`}
      onClick={onClick}
      disabled={disabled}
      {...props}
    >
      {children}
    </button>
  );
};

export interface ListItemIconProps {
  children: React.ReactNode;
  className?: string;
}

export const ListItemIcon: React.FC<ListItemIconProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <span className={`list-item-icon mr-3 ${className}`} {...props}>
      {children}
    </span>
  );
};

export interface ListItemTextProps {
  children: React.ReactNode;
  className?: string;
}

export const ListItemText: React.FC<ListItemTextProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <span className={`list-item-text ${className}`} {...props}>
      {children}
    </span>
  );
};
