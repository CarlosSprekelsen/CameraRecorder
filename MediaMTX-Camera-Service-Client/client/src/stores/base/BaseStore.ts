/**
 * Base Store Interface - Standardized Store-Service Integration Pattern
 * 
 * Architecture requirement: "Unidirectional data flow" (ADR-002)
 * Provides consistent service injection and error handling patterns across all stores
 */

export interface BaseStoreState {
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
}

export interface BaseStoreActions {
  // Standardized service injection
  setService: <T>(service: T) => void;
  
  // Standardized state management
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setLastUpdated: (timestamp: string | null) => void;
  
  // Standardized error handling
  handleError: (error: unknown, context: string) => void;
  
  // Reset functionality
  reset: () => void;
}

/**
 * Standardized error handling utility
 * Provides consistent error processing across all stores
 */
export const createErrorHandler = (set: (state: any) => void) => {
  return (error: unknown, context: string) => {
    const errorMessage = error instanceof Error 
      ? error.message 
      : `Unknown error in ${context}`;
    
    set({ 
      loading: false, 
      error: errorMessage,
      lastUpdated: new Date().toISOString()
    });
  };
};

/**
 * Standardized loading state management
 * Provides consistent loading state handling across all stores
 */
export const createLoadingHandler = (set: (state: any) => void) => {
  return {
    startLoading: () => set({ loading: true, error: null }),
    stopLoading: () => set({ loading: false }),
    stopLoadingWithError: (error: string) => set({ 
      loading: false, 
      error,
      lastUpdated: new Date().toISOString()
    })
  };
};

/**
 * Standardized service injection pattern
 * Provides consistent service injection across all stores
 */
export const createServiceInjection = <T>() => {
  let service: T | null = null;
  
  return {
    setService: (newService: T) => {
      service = newService;
    },
    getService: () => service,
    requireService: (serviceName: string) => {
      if (!service) {
        throw new Error(`${serviceName} not initialized`);
      }
      return service;
    }
  };
};
