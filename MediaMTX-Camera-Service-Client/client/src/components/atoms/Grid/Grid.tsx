import React from 'react';

export interface GridProps {
  children: React.ReactNode;
  container?: boolean;
  item?: boolean;
  xs?: number;
  sm?: number;
  md?: number;
  lg?: number;
  xl?: number;
  spacing?: number;
  className?: string;
}

export const Grid: React.FC<GridProps> = ({
  children,
  container = false,
  item = false,
  xs,
  sm,
  md,
  lg,
  xl,
  spacing,
  className = '',
}) => {
  const baseClasses = 'grid';
  const containerClasses = container ? 'grid-cols-12 gap-4' : '';
  const itemClasses = item ? 'col-span-12' : '';
  
  // Responsive classes
  const responsiveClasses = [
    xs ? `col-span-${xs}` : '',
    sm ? `sm:col-span-${sm}` : '',
    md ? `md:col-span-${md}` : '',
    lg ? `lg:col-span-${lg}` : '',
    xl ? `xl:col-span-${xl}` : '',
  ].filter(Boolean).join(' ');

  const spacingClasses = spacing ? `gap-${spacing}` : '';

  return (
    <div className={`${baseClasses} ${containerClasses} ${itemClasses} ${responsiveClasses} ${spacingClasses} ${className}`}>
      {children}
    </div>
  );
};
