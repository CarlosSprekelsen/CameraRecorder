/**
 * Pagination Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Pagination component for data navigation
 */

import React from 'react';

export interface PaginationProps {
  count: number;
  page: number;
  onChange: (event: React.ChangeEvent<unknown>, page: number) => void;
  color?: 'primary' | 'secondary' | 'standard';
  size?: 'small' | 'medium' | 'large';
  showFirstButton?: boolean;
  showLastButton?: boolean;
  className?: string;
}

export const Pagination: React.FC<PaginationProps> = ({
  count,
  page,
  onChange,
  color = 'primary',
  size = 'medium',
  showFirstButton = false,
  showLastButton = false,
  className = '',
  ...props
}) => {
  const totalPages = Math.ceil(count);
  // const pages = [];
  
  // Generate page numbers with ellipsis logic
  const getPageNumbers = () => {
    const delta = 2; // Number of pages to show on each side of current page
    const range = [];
    const rangeWithDots = [];

    for (let i = Math.max(2, page - delta); i <= Math.min(totalPages - 1, page + delta); i++) {
      range.push(i);
    }

    if (page - delta > 2) {
      rangeWithDots.push(1, '...');
    } else {
      rangeWithDots.push(1);
    }

    rangeWithDots.push(...range);

    if (page + delta < totalPages - 1) {
      rangeWithDots.push('...', totalPages);
    } else {
      rangeWithDots.push(totalPages);
    }

    return rangeWithDots;
  };

  const handlePageClick = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages && newPage !== page) {
      onChange({} as React.ChangeEvent<unknown>, newPage);
    }
  };

  const sizeClasses = {
    small: 'text-xs px-2 py-1',
    medium: 'text-sm px-3 py-2',
    large: 'text-base px-4 py-3',
  };

  const colorClasses = {
    primary: 'bg-blue-500 text-white border-blue-500',
    secondary: 'bg-gray-500 text-white border-gray-500',
    standard: 'bg-white text-gray-700 border-gray-300',
  };

  if (totalPages <= 1) return null;

  return (
    <nav className={`pagination flex items-center justify-center space-x-1 ${className}`} {...props}>
      {showFirstButton && (
        <button
          onClick={() => handlePageClick(1)}
          disabled={page === 1}
          className={`pagination-button ${sizeClasses[size]} border rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100`}
        >
          First
        </button>
      )}
      
      <button
        onClick={() => handlePageClick(page - 1)}
        disabled={page === 1}
        className={`pagination-button ${sizeClasses[size]} border rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100`}
      >
        Previous
      </button>

      {getPageNumbers().map((pageNum, index) => {
        if (pageNum === '...') {
          return (
            <span key={`ellipsis-${index}`} className={`${sizeClasses[size]} text-gray-500`}>
              ...
            </span>
          );
        }

        const pageNumber = pageNum as number;
        const isCurrentPage = pageNumber === page;
        
        return (
          <button
            key={pageNumber}
            onClick={() => handlePageClick(pageNumber)}
            className={`pagination-button ${sizeClasses[size]} border rounded ${
              isCurrentPage 
                ? `${colorClasses[color]} border-current` 
                : 'hover:bg-gray-100'
            }`}
          >
            {pageNumber}
          </button>
        );
      })}

      <button
        onClick={() => handlePageClick(page + 1)}
        disabled={page === totalPages}
        className={`pagination-button ${sizeClasses[size]} border rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100`}
      >
        Next
      </button>

      {showLastButton && (
        <button
          onClick={() => handlePageClick(totalPages)}
          disabled={page === totalPages}
          className={`pagination-button ${sizeClasses[size]} border rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100`}
        >
          Last
        </button>
      )}
    </nav>
  );
};
