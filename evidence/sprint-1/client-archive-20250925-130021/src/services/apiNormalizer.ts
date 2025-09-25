import type { FileInfo, FileListResponse } from '../types/camera';

function isObject(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

export function normalizeFileListResponse(raw: unknown): FileListResponse {
	if (!isObject(raw)) {
		return { files: [], total: 0, limit: 0, offset: 0 };
	}

	const files = Array.isArray(raw.files) ? (raw.files as FileInfo[]) : [];
	// Support both 'total' and 'total_count'
	const total = typeof raw.total === 'number' ? (raw.total as number)
		: typeof raw.total_count === 'number' ? (raw.total_count as number)
		: files.length;

	// Optional fields in server variant
	const limit = typeof raw.limit === 'number' ? (raw.limit as number) : 0;
	const offset = typeof raw.offset === 'number' ? (raw.offset as number) : 0;

	return { files, total, limit, offset };
}
