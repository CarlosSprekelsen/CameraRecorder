import React from 'react';
import { Skeleton } from '../atoms/Skeleton/Skeleton';
import { Box } from '../atoms/Box/Box';
import { Card } from '../atoms/Card/Card';
import { CardContent } from '../atoms/CardContent/CardContent';

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
      <Box sx={{ display: 'flex', gap: 8, marginBottom: 8 }}>
        {Array.from({ length: columns }).map((_, index) => (
          <Skeleton key={`header-${index}`} variant="rectangular" width={120} height={40} />
        ))}
      </Box>

      {/* Table rows skeleton */}
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <Box key={`row-${rowIndex}`} sx={{ display: 'flex', gap: 8, marginBottom: 4 }}>
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
        <Skeleton variant="text" width="40%" height={24} className="mt-1" />
        <Skeleton variant="rectangular" width="100%" height={200} className="mt-2" />
        <Box sx={{ display: 'flex', gap: 4, marginTop: 8 }}>
          <Skeleton variant="rectangular" width={80} height={32} />
          <Skeleton variant="rectangular" width={80} height={32} />
        </Box>
      </CardContent>
    </Card>
  );

  const renderListSkeleton = () => (
    <Box>
      {Array.from({ length: rows }).map((_, index) => (
        <Box key={`list-item-${index}`} className="flex items-center mb-2">
          <Skeleton variant="circular" width={40} height={40} className="mr-2" />
          <Box className="flex-1">
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
