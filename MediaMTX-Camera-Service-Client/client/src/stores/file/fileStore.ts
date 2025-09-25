import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { FileService } from '../../services/file/FileService';

export interface FileInfo {
  filename: string;
  file_size: number;
  modified_time: string;
  download_url: string;
  duration?: number;
  format?: string;
  device?: string;
}

export interface FileState {
  recordings: FileInfo[];
  snapshots: FileInfo[];
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
    limit: 20,
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
            if (!fileService) {
              set({ error: 'File service not initialized' });
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
            }
          },

          loadSnapshots: async (limit = 20, offset = 0) => {
            if (!fileService) {
              set({ error: 'File service not initialized' });
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
            }
          },

          getRecordingInfo: async (filename: string) => {
            if (!fileService) {
              set({ error: 'File service not initialized' });
              return null;
            }
            try {
              const info = await fileService.getRecordingInfo(filename);
              return info;
            } catch (error) {
              set({ error: error instanceof Error ? error.message : 'Failed to get recording info' });
              return null;
            }
          },

          getSnapshotInfo: async (filename: string) => {
            if (!fileService) {
              set({ error: 'File service not initialized' });
              return null;
            }
            try {
              const info = await fileService.getSnapshotInfo(filename);
              return info;
            } catch (error) {
              set({ error: error instanceof Error ? error.message : 'Failed to get snapshot info' });
              return null;
            }
          },

          downloadFile: async (downloadUrl: string, filename: string) => {
            if (!fileService) {
              set({ error: 'File service not initialized' });
              return;
            }
            try {
              await fileService.downloadFile(downloadUrl, filename);
            } catch (error) {
              set({ error: error instanceof Error ? error.message : 'Download failed' });
            }
          },

          deleteRecording: async (filename: string) => {
            if (!fileService) {
              set({ error: 'File service not initialized' });
              return false;
            }
            set({ loading: true, error: null });
            try {
              const response = await fileService.deleteRecording(filename);
              if (response.success) {
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
              return false;
            }
          },

          deleteSnapshot: async (filename: string) => {
            if (!fileService) {
              set({ error: 'File service not initialized' });
              return false;
            }
            set({ loading: true, error: null });
            try {
              const response = await fileService.deleteSnapshot(filename);
              if (response.success) {
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
              ? selectedFiles.filter(f => f !== filename)
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

          reset: () => set(initialState),
        };
      },
      {
        name: 'file-store',
        partialize: (state) => ({
          pagination: state.pagination,
          currentTab: state.currentTab,
        }),
      }
    ),
    {
      name: 'file-store',
    }
  )
);
