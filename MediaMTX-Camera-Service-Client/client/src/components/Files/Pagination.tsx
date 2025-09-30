import React from 'react';
import { Box } from '../atoms/Box/Box';
import { Pagination as PaginationAtom } from '../atoms/Pagination/Pagination';
import { Typography } from '../atoms/Typography/Typography';
import { Select } from '../atoms/Select/Select';
import { FormControl, InputLabel } from '../atoms/FormControl/FormControl';

interface PaginationProps {
  pagination: {
    limit: number;
    offset: number;
    total: number;
  };
  onPageChange: (page: number) => void;
  onLimitChange?: (limit: number) => void;
}

/**
 * Pagination - Pagination controls for file lists
 * Supports page navigation and limit changes
 */
const Pagination: React.FC<PaginationProps> = ({ pagination, onPageChange, onLimitChange }) => {
  const { limit, offset, total } = pagination;
  const currentPage = Math.floor(offset / limit) + 1;
  const totalPages = Math.ceil(total / limit);
  const startItem = offset + 1;
  const endItem = Math.min(offset + limit, total);

  const handlePageChange = (_event: React.ChangeEvent<unknown>, page: number) => {
    onPageChange(page);
  };

  const handleLimitChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    if (onLimitChange) {
      onLimitChange(Number(event.target.value));
    }
  };

  if (total === 0) {
    return null;
  }

  return (
    <Box className="flex items-center gap-2 flex-wrap">
      <Typography variant="body2" color="secondary">
        Showing {startItem}-{endItem} of {total} files
      </Typography>

      {onLimitChange && (
        <FormControl className="min-w-[120px]">
          <InputLabel>Per page</InputLabel>
          <Select 
            value={limit} 
            onChange={handleLimitChange}
          >
            <option value={10}>10</option>
            <option value={20}>20</option>
            <option value={50}>50</option>
            <option value={100}>100</option>
          </Select>
        </FormControl>
      )}

      <PaginationAtom
        count={totalPages}
        page={currentPage}
        onChange={handlePageChange}
        color="primary"
        showFirstButton
        showLastButton
        size="small"
      />
    </Box>
  );
};

export default Pagination;
