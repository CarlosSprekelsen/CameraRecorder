import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { FileService } from '../../services/file/FileService';

// File info for list operations (uses modified_time)
export interface FileListItem {
  filename: string;
  file_size: number;
  modified_time: string;
  download_url: string;
}

// File info for detailed operations (uses created_time)
export interface FileInfo {
  filename: string;
  file_size: number;
  created_time: string;
  download_url: string;
  duration?: number;
  format?: string;
  device?: string;
}


export interface FileState {
  recordings: FileListItem[];
  snapshots: FileListItem[];
  loading: boolean;
  error: string | null;
  pagination: {
    limit: number;
    offset: number;
    total: number;
  };
  selectedFiles: string[];
  currentTab: 'recordings' | 'snapshots';
}

export interface FileActions {
  // File catalog operations (I.FileCatalog)
  loadRecordings: (limit?: number, offset?: number) => Promise<void>;
  loadSnapshots: (limit?: number, offset?: number) => Promise<void>;
  getRecordingInfo: (filename: string) => Promise<FileInfo | null>;
  getSnapshotInfo: (filename: string) => Promise<FileInfo | null>;

  // File actions (I.FileActions)
  downloadFile: (downloadUrl: string, filename: string) => Promise<void>;
  deleteRecording: (filename: string) => Promise<boolean>;
  deleteSnapshot: (filename: string) => Promise<boolean>;
  
  // Retention policy management (I.FileActions)
  setRetentionPolicy: (policyType: 'age' | 'size' | 'manual', enabled: boolean, maxAgeDays?: number, maxSizeGb?: number) => Promise<any>;
  cleanupOldFiles: () => Promise<any>;

  // State management
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setCurrentTab: (tab: 'recordings' | 'snapshots') => void;
  setSelectedFiles: (files: string[]) => void;
  toggleFileSelection: (filename: string) => void;
  clearSelection: () => void;

  // Pagination
  setPagination: (limit: number, offset: number, total: number) => void;
  nextPage: () => void;
  prevPage: () => void;
  goToPage: (page: number) => void;

  // Service injection
  setFileService: (service: FileService) => void;

  // Reset
  reset: () => void;
}

const initialState: FileState = {
  recordings: [],
  snapshots: [],
  loading: false,
  error: null,
  pagination: {
    limit: 50,
    offset: 0,
    total: 0,
  },
  selectedFiles: [],
  currentTab: 'recordings',
};

export const useFileStore = create<FileState & FileActions>()(
  devtools(
    persist(
      (set, get) => {
        let fileService: FileService | null = null;

        return {
          ...initialState,

          setFileService: (service: FileService) => {
            fileService = service;
          },

          loadRecordings: async (limit = 20, offset = 0) => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!fileService) {
              set({ error: 'File service not initialized', loading: false });
              return;
            }

            set({ loading: true, error: null });
            try {
              const response = await fileService.listRecordings(limit, offset);
              set({
                recordings: response.files,
                pagination: {
                  limit,
                  offset,
                  total: response.total,
                },
                loading: false,
              });
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to load recordings',
              });
              // No re-throw - graceful degradation per ADR-002
            }
          },

          loadSnapshots: async (limit = 20, offset = 0) => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!fileService) {
              set({ error: 'File service not initialized', loading: false });
              return;
            }

            set({ loading: true, error: null });
            try {
              const response = await fileService.listSnapshots(limit, offset);
              set({
                snapshots: response.files,
                pagination: {
                  limit,
                  offset,
                  total: response.total,
                },
                loading: false,
              });
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to load snapshots',
              });
              // No re-throw - graceful degradation per ADR-002
            }
          },

          getRecordingInfo: async (filename: string) => {
            if (!fileService) throw new Error('File service not initialized');
            try {
              const info = await fileService.getRecordingInfo(filename);
              return info;
            } catch (error) {
              set({
                error: error instanceof Error ? error.message : 'Failed to get recording info',
              });
              throw error;
            }
          },

          getSnapshotInfo: async (filename: string) => {
            if (!fileService) throw new Error('File service not initialized');
            try {
              const info = await fileService.getSnapshotInfo(filename);
              return info;
            } catch (error) {
              set({
                error: error instanceof Error ? error.message : 'Failed to get snapshot info',
              });
              throw error;
            }
          },

          downloadFile: async (downloadUrl: string, filename: string) => {
            if (!fileService) throw new Error('File service not initialized');
            try {
              await fileService.downloadFile(downloadUrl, filename);
            } catch (error) {
              set({ error: error instanceof Error ? error.message : 'Download failed' });
              throw error;
            }
          },

          deleteRecording: async (filename: string) => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!fileService) {
              set({ error: 'File service not initialized', loading: false });
              return false;
            }

            set({ loading: true, error: null });
            try {
              const response = await fileService.deleteRecording(filename);
              if (response.deleted) {
                // Reload recordings after successful deletion
                const { pagination } = get();
                await get().loadRecordings(pagination.limit, pagination.offset);
                set({ loading: false });
                return true;
              } else {
                set({ loading: false, error: response.message });
                return false;
              }
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Delete failed',
              });
              // No re-throw - graceful degradation per ADR-002
              return false;
            }
          },

          deleteSnapshot: async (filename: string) => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!fileService) {
              set({ error: 'File service not initialized', loading: false });
              return false;
            }

            set({ loading: true, error: null });
            try {
              const response = await fileService.deleteSnapshot(filename);
              if (response.deleted) {
                // Reload snapshots after successful deletion
                const { pagination } = get();
                await get().loadSnapshots(pagination.limit, pagination.offset);
                set({ loading: false });
                return true;
              } else {
                set({ loading: false, error: response.message });
                return false;
              }
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Delete failed',
              });
              // No re-throw - graceful degradation per ADR-002
              return false;
            }
          },

          setLoading: (loading: boolean) => set({ loading }),
          setError: (error: string | null) => set({ error }),
          setCurrentTab: (tab: 'recordings' | 'snapshots') => set({ currentTab: tab }),
          setSelectedFiles: (files: string[]) => set({ selectedFiles: files }),
          clearSelection: () => set({ selectedFiles: [] }),

          toggleFileSelection: (filename: string) => {
            const { selectedFiles } = get();
            const newSelection = selectedFiles.includes(filename)
              ? selectedFiles.filter((f) => f !== filename)
              : [...selectedFiles, filename];
            set({ selectedFiles: newSelection });
          },

          setPagination: (limit: number, offset: number, total: number) =>
            set({
              pagination: { limit, offset, total },
            }),

          nextPage: () => {
            const { pagination } = get();
            const newOffset = pagination.offset + pagination.limit;
            if (newOffset < pagination.total) {
              set({ pagination: { ...pagination, offset: newOffset } });
            }
          },

          prevPage: () => {
            const { pagination } = get();
            const newOffset = Math.max(0, pagination.offset - pagination.limit);
            set({ pagination: { ...pagination, offset: newOffset } });
          },

          goToPage: (page: number) => {
            const { pagination } = get();
            const newOffset = (page - 1) * pagination.limit;
            set({ pagination: { ...pagination, offset: newOffset } });
          },

          setRetentionPolicy: async (policyType: 'age' | 'size' | 'manual', enabled: boolean, maxAgeDays?: number, maxSizeGb?: number) => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!fileService) {
              set({ error: 'File service not initialized', loading: false });
              return undefined;
            }

            set({ loading: true, error: null });
            try {
              const response = await fileService.setRetentionPolicy(policyType, enabled, maxAgeDays, maxSizeGb);
              set({ loading: false });
              return response;
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to set retention policy',
              });
              // No re-throw - graceful degradation per ADR-002
              return undefined;
            }
          },

          cleanupOldFiles: async () => {
            // Synchronous guard - graceful error handling per ADR-002
            if (!fileService) {
              set({ error: 'File service not initialized', loading: false });
              return undefined;
            }

            set({ loading: true, error: null });
            try {
              const response = await fileService.cleanupOldFiles();
              set({ loading: false });
              return response;
            } catch (error) {
              set({
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to cleanup old files',
              });
              // No re-throw - graceful degradation per ADR-002
              return undefined;
            }
          },

          reset: () => set(initialState),
        };
      },
      {
        name: 'file-store',
        partialize: (state) => ({
          pagination: state.pagination,
          currentTab: state.currentTab,
        }),
      },
    ),
    {
      name: 'file-store',
    },
  ),
);
