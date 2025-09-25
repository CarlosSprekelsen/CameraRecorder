import { useEffect, useCallback } from 'react';
import { logger } from '../services/logger/LoggerService';

// Performance API types for proper typing
interface WindowWithGtag extends Window {
  gtag?: (command: string, targetId: string, config: Record<string, unknown>) => void;
}

interface PerformanceEntryWithInput extends PerformanceEntry {
  processingStart?: number;
  hadRecentInput?: boolean;
  value?: number;
}

// Performance metrics interface for future use
// interface PerformanceMetrics {
//   lcp?: number; // Largest Contentful Paint
//   fid?: number; // First Input Delay
//   cls?: number; // Cumulative Layout Shift
//   fcp?: number; // First Contentful Paint
//   ttfb?: number; // Time to First Byte
// }

/**
 * usePerformanceMonitor - Performance monitoring hook for Core Web Vitals
 * Implements performance monitoring from architecture section 10.1
 */
export const usePerformanceMonitor = () => {
  const trackMetric = useCallback((name: string, value: number, delta: number) => {
    logger.info('Performance metric', {
      metric: name,
      value,
      delta,
      timestamp: Date.now(),
    });

    // Send to analytics service if available
    if (typeof window !== 'undefined' && 'gtag' in window) {
      (window as WindowWithGtag).gtag?.('event', name, {
        value: Math.round(name === 'CLS' ? delta * 1000 : delta),
        event_category: 'Web Vitals',
        event_label: name,
        non_interaction: true,
      });
    }
  }, []);

  const trackCustomMetric = useCallback(
    (name: string, value: number, metadata?: Record<string, unknown>) => {
      logger.info('Custom performance metric', {
        metric: name,
        value,
        metadata,
        timestamp: Date.now(),
      });
    },
    [],
  );

  useEffect(() => {
    // Track Core Web Vitals
    const trackWebVitals = () => {
      // Largest Contentful Paint
      if ('PerformanceObserver' in window) {
        try {
          const lcpObserver = new PerformanceObserver((list) => {
            const entries = list.getEntries();
            const lastEntry = entries[entries.length - 1];
            if (lastEntry) {
              trackMetric('LCP', lastEntry.startTime, lastEntry.startTime);
            }
          });
          lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] });
        } catch (error) {
          logger.warn('Failed to observe LCP', error as Error);
        }

        // First Input Delay
        try {
          const fidObserver = new PerformanceObserver((list) => {
            const entries = list.getEntries();
            entries.forEach((entry: PerformanceEntryWithInput) => {
              if (entry.processingStart) {
                trackMetric(
                  'FID',
                  entry.processingStart - entry.startTime,
                  entry.processingStart - entry.startTime,
                );
              }
            });
          });
          fidObserver.observe({ entryTypes: ['first-input'] });
        } catch (error) {
          logger.warn('Failed to observe FID', error as Error);
        }

        // Cumulative Layout Shift
        try {
          const clsObserver = new PerformanceObserver((list) => {
            let clsValue = 0;
            const entries = list.getEntries();
            entries.forEach((entry: PerformanceEntryWithInput) => {
              if (!entry.hadRecentInput && entry.value !== undefined) {
                clsValue += entry.value;
              }
            });
            trackMetric('CLS', clsValue, clsValue);
          });
          clsObserver.observe({ entryTypes: ['layout-shift'] });
        } catch (error) {
          logger.warn('Failed to observe CLS', error as Error);
        }
      }
    };

    // Track page load performance
    const trackPageLoad = () => {
      if ('performance' in window) {
        const navigation = performance.getEntriesByType(
          'navigation',
        )[0] as PerformanceNavigationTiming;
        if (navigation) {
          trackCustomMetric('TTFB', navigation.responseStart - navigation.requestStart, {
            type: 'page_load',
          });
          trackCustomMetric(
            'DOM_LOAD',
            navigation.domContentLoadedEventEnd - navigation.fetchStart,
            {
              type: 'page_load',
            },
          );
          trackCustomMetric('PAGE_LOAD', navigation.loadEventEnd - navigation.fetchStart, {
            type: 'page_load',
          });
        }
      }
    };

    // Track WebSocket connection performance (for future use)
    // const trackWebSocketPerformance = () => {
    //   const startTime = Date.now();
    //   return {
    //     markConnectionStart: () => {
    //       logger.info('WebSocket connection started', { timestamp: startTime });
    //     },
    //     markConnectionEnd: () => {
    //       const duration = Date.now() - startTime;
    //       trackCustomMetric('WS_CONNECTION_TIME', duration, {
    //         type: 'websocket',
    //       });
    //     },
    //   };
    // };

    trackWebVitals();
    trackPageLoad();

    return () => {
      // Cleanup if needed
    };
  }, [trackMetric, trackCustomMetric]);

  return {
    trackMetric,
    trackCustomMetric,
  };
};
