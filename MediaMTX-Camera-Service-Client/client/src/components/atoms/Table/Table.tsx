/**
 * Table Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Table component for data display
 */

import React from 'react';

export interface TableProps {
  children: React.ReactNode;
  className?: string;
}

export const Table: React.FC<TableProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <table className={`table w-full border-collapse ${className}`} {...props}>
      {children}
    </table>
  );
};

export interface TableContainerProps {
  children: React.ReactNode;
  className?: string;
}

export const TableContainer: React.FC<TableContainerProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <div className={`table-container overflow-auto ${className}`} {...props}>
      {children}
    </div>
  );
};

export interface TableHeadProps {
  children: React.ReactNode;
  className?: string;
}

export const TableHead: React.FC<TableHeadProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <thead className={`table-head bg-gray-50 ${className}`} {...props}>
      {children}
    </thead>
  );
};

export interface TableBodyProps {
  children: React.ReactNode;
  className?: string;
}

export const TableBody: React.FC<TableBodyProps> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <tbody className={`table-body ${className}`} {...props}>
      {children}
    </tbody>
  );
};

export interface TableRowProps {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
}

export const TableRow: React.FC<TableRowProps> = ({
  children,
  className = '',
  hover = false,
  ...props
}) => {
  const hoverClass = hover ? 'hover:bg-gray-50' : '';
  return (
    <tr className={`table-row border-b border-gray-200 ${hoverClass} ${className}`} {...props}>
      {children}
    </tr>
  );
};

export interface TableCellProps {
  children: React.ReactNode;
  className?: string;
  align?: 'left' | 'center' | 'right';
  component?: keyof JSX.IntrinsicElements;
}

export const TableCell: React.FC<TableCellProps> = ({
  children,
  className = '',
  align = 'left',
  component: Component = 'td',
  ...props
}) => {
  const alignClasses = {
    left: 'text-left',
    center: 'text-center',
    right: 'text-right'
  };

  return (
    <Component className={`table-cell px-4 py-2 ${alignClasses[align]} ${className}`} {...props}>
      {children}
    </Component>
  );
};
