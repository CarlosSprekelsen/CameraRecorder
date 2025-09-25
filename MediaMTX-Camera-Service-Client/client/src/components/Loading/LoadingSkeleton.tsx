import React from 'react';
import { Skeleton, Box, Card, CardContent } from '@mui/material';

interface LoadingSkeletonProps {
  variant?: 'table' | 'card' | 'list' | 'custom';
  rows?: number;
  columns?: number;
  height?: number;
  width?: number;
}

/**
 * LoadingSkeleton - Loading skeleton components for better UX
 * Implements performance optimizations from Sprint 5
 */
const LoadingSkeleton: React.FC<LoadingSkeletonProps> = ({
  variant = 'table',
  rows = 5,
  columns = 4,
  height = 20,
  width = '100%',
}) => {
  const renderTableSkeleton = () => (
    <Box>
      {/* Table header skeleton */}
      <Box display="flex" gap={2} mb={2}>
        {Array.from({ length: columns }).map((_, index) => (
          <Skeleton key={`header-${index}`} variant="rectangular" width={120} height={40} />
        ))}
      </Box>
      
      {/* Table rows skeleton */}
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <Box key={`row-${rowIndex}`} display="flex" gap={2} mb={1}>
          {Array.from({ length: columns }).map((_, colIndex) => (
            <Skeleton
              key={`cell-${rowIndex}-${colIndex}`}
              variant="rectangular"
              width={colIndex === 0 ? 200 : 120}
              height={height}
            />
          ))}
        </Box>
      ))}
    </Box>
  );

  const renderCardSkeleton = () => (
    <Card>
      <CardContent>
        <Skeleton variant="text" width="60%" height={32} />
        <Skeleton variant="text" width="40%" height={24} sx={{ mt: 1 }} />
        <Skeleton variant="rectangular" width="100%" height={200} sx={{ mt: 2 }} />
        <Box display="flex" gap={1} mt={2}>
          <Skeleton variant="rectangular" width={80} height={32} />
          <Skeleton variant="rectangular" width={80} height={32} />
        </Box>
      </CardContent>
    </Card>
  );

  const renderListSkeleton = () => (
    <Box>
      {Array.from({ length: rows }).map((_, index) => (
        <Box key={`list-item-${index}`} display="flex" alignItems="center" mb={2}>
          <Skeleton variant="circular" width={40} height={40} sx={{ mr: 2 }} />
          <Box flex={1}>
            <Skeleton variant="text" width="70%" height={24} />
            <Skeleton variant="text" width="50%" height={20} />
          </Box>
          <Skeleton variant="rectangular" width={60} height={32} />
        </Box>
      ))}
    </Box>
  );

  const renderCustomSkeleton = () => (
    <Skeleton variant="rectangular" width={width} height={height} />
  );

  switch (variant) {
    case 'table':
      return renderTableSkeleton();
    case 'card':
      return renderCardSkeleton();
    case 'list':
      return renderListSkeleton();
    case 'custom':
      return renderCustomSkeleton();
    default:
      return renderTableSkeleton();
  }
};

export default LoadingSkeleton;
