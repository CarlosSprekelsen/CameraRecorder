/**
 * Error Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around ErrorHandlerService
 * following the modular store pattern established in connection/
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';

interface ErrorStoreState {
  errors: any[];
  isLoading: boolean;
  error: string | null;
}

interface ErrorStoreActions {
  getErrors: () => Promise<void>;
  clearErrors: () => void;
  clearError: () => void;
  setError: (error: string) => void;
}

type ErrorStore = ErrorStoreState & ErrorStoreActions;

const initialState: ErrorStoreState = {
  errors: [],
  isLoading: false,
  error: null,
};

export const useErrorStore = create<ErrorStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      getErrors: async () => {
        set({ isLoading: true, error: null });
        try {
          // TODO: Implement with ErrorHandlerService
          set({ errors: [], isLoading: false });
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to get errors';
          set({ error: errorMessage, isLoading: false });
        }
      },
      
      clearErrors: () => set({ errors: [] }),
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    { name: 'error-store' }
  )
);
