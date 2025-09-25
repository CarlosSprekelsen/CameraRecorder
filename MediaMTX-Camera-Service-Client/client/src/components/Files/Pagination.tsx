import React from 'react';
import {
  Box,
  Pagination as MuiPagination,
  Typography,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
} from '@mui/material';

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

  const handleLimitChange = (event: React.ChangeEvent<HTMLInputElement> | any) => {
    const newLimit = parseInt(event.target.value, 10);
    if (onLimitChange) {
      onLimitChange(newLimit);
    }
  };

  if (total === 0) {
    return null;
  }

  return (
    <Box display="flex" alignItems="center" gap={2} flexWrap="wrap">
      <Typography variant="body2" color="text.secondary">
        Showing {startItem}-{endItem} of {total} files
      </Typography>

      {onLimitChange && (
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Per page</InputLabel>
          <Select value={limit} label="Per page" onChange={handleLimitChange}>
            <MenuItem value={10}>10</MenuItem>
            <MenuItem value={20}>20</MenuItem>
            <MenuItem value={50}>50</MenuItem>
            <MenuItem value={100}>100</MenuItem>
          </Select>
        </FormControl>
      )}

      <MuiPagination
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
