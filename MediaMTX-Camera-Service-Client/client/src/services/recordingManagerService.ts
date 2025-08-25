import { WebSocketService } from './websocket';
import { HTTPPollingService } from './httpPollingService';
import { errorRecoveryService } from './errorRecoveryService';
import type {
  RecordingSession,
  RecordingProgress,
  RecordingStatus,
  CameraDevice,
  ConfigValidationResult
} from '../types/camera';
import { RPC_METHODS, ERROR_CODES } from '../types/rpc';

/**
 * Recording Manager Service
 * 
 * Manages recording state, conflicts, and session tracking for enhanced
 * recording management capabilities.
 */
class RecordingManagerService {
  private activeSessions: Map<string, RecordingSession> = new Map();
  private progressCallbacks: Set<(progress: RecordingProgress) => void> = new Set();

  /**
   * Start recording for a camera
   */
  async startRecording(cameraId: string): Promise<RecordingSession> {
    // Validate storage before starting
    const validation = await this.validateRecordingOperation(cameraId);
    if (!validation.valid) {
      throw new Error(`Cannot start recording: ${validation.reason}`);
    }

    // Check for existing recording
    if (this.activeSessions.has(cameraId)) {
      const existingSession = this.activeSessions.get(cameraId)!;
      throw new Error(`Camera ${cameraId} is already recording (Session: ${existingSession.session_id})`);
    }

    try {
      const session = await errorRecoveryService.executeWithRetry(
        async () => {
          const response = await wsService.call(RPC_METHODS.START_RECORDING, { device: cameraId });
          return response as RecordingSession;
        },
        'startRecording'
      );

      // Update state
      this.activeSessions.set(cameraId, session);
      this.updateRecordingState(cameraId, {
        isRecording: true,
        sessionId: session.session_id,
        startTime: new Date(),
        currentFile: session.filename,
        elapsedTime: 0
      });

      return session;
    } catch (error: any) {
      if (error.code === ERROR_CODES.CAMERA_ALREADY_RECORDING) {
        // Handle recording conflict
        const conflict: RecordingConflict = {
          device: cameraId,
          session_id: error.data?.session_id || 'unknown',
          message: error.message
        };
        this.notifyConflictCallbacks(conflict);
      }
      throw error;
    }
  }

  /**
   * Stop recording for a camera
   */
  async stopRecording(cameraId: string): Promise<void> {
    const session = this.activeSessions.get(cameraId);
    if (!session) {
      throw new Error(`No active recording session for camera ${cameraId}`);
    }

    try {
      await errorRecoveryService.executeWithRetry(
        async () => {
          await wsService.call(RPC_METHODS.STOP_RECORDING, { device: cameraId });
        },
        'stopRecording'
      );

      // Update state
      this.activeSessions.delete(cameraId);
      this.updateRecordingState(cameraId, {
        isRecording: false,
        sessionId: null,
        startTime: null,
        currentFile: null,
        elapsedTime: 0
      });
    } catch (error) {
      throw new Error(`Failed to stop recording: ${error}`);
    }
  }

  /**
   * Get current recording state for a camera
   */
  getRecordingState(cameraId: string): RecordingState | null {
    return this.recordingState.get(cameraId) || null;
  }

  /**
   * Get all active recording sessions
   */
  getActiveSessions(): RecordingSession[] {
    return Array.from(this.activeSessions.values());
  }

  /**
   * Check if camera is currently recording
   */
  isRecording(cameraId: string): boolean {
    return this.activeSessions.has(cameraId);
  }

  /**
   * Get recording progress for a camera
   */
  getRecordingProgress(cameraId: string): RecordingProgress | null {
    const state = this.recordingState.get(cameraId);
    if (!state || !state.isRecording) {
      return null;
    }

    const elapsedTime = state.startTime ? 
      Math.floor((Date.now() - state.startTime.getTime()) / 1000) : 0;

    return {
      camera_id: cameraId,
      session_id: state.sessionId!,
      elapsed_time: elapsedTime,
      current_file: state.currentFile || '',
      is_active: true
    };
  }

  /**
   * Validate recording operation with enhanced checks
   */
  async validateRecordingOperation(cameraId: string): Promise<ValidationResult> {
    try {
      // Check camera status
      const cameraStatus = await errorRecoveryService.executeWithRetry(
        async () => {
          const response = await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device: cameraId });
          return response as CameraDevice;
        },
        'getCameraStatus'
      );

      if (cameraStatus.status !== 'CONNECTED') {
        return {
          valid: false,
          reason: `Camera ${cameraId} is not connected (status: ${cameraStatus.status})`
        };
      }

      // Check for existing recording (F1.4.1: Prevent multiple simultaneous recordings)
      if (this.activeSessions.has(cameraId)) {
        const existingSession = this.activeSessions.get(cameraId)!;
        return {
          valid: false,
          reason: `Camera ${cameraId} is already recording (Session: ${existingSession.session_id})`
        };
      }

      // Check if camera is already recording according to service (F1.4.1)
      if (cameraStatus.recording) {
        return {
          valid: false,
          reason: `Camera ${cameraId} is currently recording (Session: ${cameraStatus.recording_session || 'unknown'})`
        };
      }

      // Validate storage space before starting recording (F1.4.2)
      const storageValidation = await this.validateStorageForRecording();
      if (!storageValidation.valid) {
        return storageValidation;
      }

      return { valid: true, reason: 'OK' };
    } catch (error) {
      return {
        valid: false,
        reason: `Validation failed: ${error}`
      };
    }
  }

  /**
   * Validate storage for recording operation (F1.4.2)
   */
  async validateStorageForRecording(): Promise<ValidationResult> {
    try {
      const storageInfo = await errorRecoveryService.executeWithRetry(
        async () => {
          const response = await wsService.call(RPC_METHODS.GET_STORAGE_INFO, {});
          return response as StorageInfo;
        },
        'getStorageInfo'
      );

      const totalSpace = storageInfo.total_space;
      const availableSpace = storageInfo.available_space;
      const usedSpace = totalSpace - availableSpace;
      const usagePercent = (usedSpace / totalSpace) * 100;

      // Check critical threshold (F1.4.2: Storage space critical)
      if (usagePercent >= 95) {
        return {
          valid: false,
          reason: `Storage space is critical (${usagePercent.toFixed(1)}% used). Recording blocked.`
        };
      }

      // Check warning threshold (F1.4.2: Storage space low)
      if (usagePercent >= 90) {
        return {
          valid: true,
          reason: `Storage space is low (${usagePercent.toFixed(1)}% used). Proceed with caution.`
        };
      }

      return { valid: true, reason: 'Storage space is adequate for recording.' };
    } catch (error) {
      return {
        valid: false,
        reason: `Storage validation failed: ${error}`
      };
    }
  }

  /**
   * Handle recording status updates from WebSocket with enhanced progress tracking (F1.4.3)
   */
  handleRecordingStatusUpdate(cameraId: string, status: Record<string, unknown>): void {
    const currentState = this.recordingState.get(cameraId);
    
    if (status.recording) {
      // Recording started or updated with comprehensive information (F1.4.3)
      this.updateRecordingState(cameraId, {
        isRecording: true,
        sessionId: status.recording_session || currentState?.sessionId || null,
        startTime: currentState?.startTime || new Date(),
        currentFile: status.current_file || currentState?.currentFile || null,
        elapsedTime: status.elapsed_time || currentState?.elapsedTime || 0,
        // Enhanced progress information (F1.4.3)
        filePath: status.file_path || currentState?.filePath || null,
        recordingQuality: status.recording_quality || currentState?.recordingQuality || 'standard',
        bitrate: status.bitrate || currentState?.bitrate || 0,
        frameRate: status.frame_rate || currentState?.frameRate || 0,
        resolution: status.resolution || currentState?.resolution || 'unknown',
        fileSize: status.file_size || currentState?.fileSize || 0
      });

      // Update comprehensive progress tracking (F1.4.3)
      this.setProgress(cameraId, {
        elapsed_time: status.elapsed_time || 0,
        file_size: status.file_size || 0,
        current_file: status.current_file || '',
        // Enhanced progress information
        file_path: status.file_path || '',
        recording_quality: status.recording_quality || 'standard',
        bitrate: status.bitrate || 0,
        frame_rate: status.frame_rate || 0,
        resolution: status.resolution || 'unknown'
      });

      // Handle file rotation seamlessly (F1.4.4)
      if (status.file_rotation_occurred) {
        console.log(`🔄 File rotation occurred for camera ${cameraId}: ${status.new_file_name}`);
        this.handleFileRotation(cameraId, status.new_file_name, status.rotation_timestamp);
      }

      // Notify progress callbacks for real-time updates (F1.4.5)
      const progress = this.getRecordingProgress(cameraId);
      if (progress) {
        this.notifyProgressCallbacks(progress);
      }
    } else {
      // Recording stopped
      this.updateRecordingState(cameraId, {
        isRecording: false,
        sessionId: null,
        startTime: null,
        currentFile: null,
        elapsedTime: 0,
        filePath: null,
        recordingQuality: null,
        bitrate: 0,
        frameRate: 0,
        resolution: null,
        fileSize: 0
      });

      // Clear progress
      this.clearProgress(cameraId);
    }

    // Notify state callbacks for real-time updates (F1.4.5)
    const state = this.recordingState.get(cameraId);
    if (state) {
      this.notifyStateCallbacks(state);
    }
  }

  /**
   * Handle file rotation seamlessly (F1.4.4)
   */
  private handleFileRotation(cameraId: string, newFileName: string, rotationTimestamp: number): void {
    const currentState = this.recordingState.get(cameraId);
    if (!currentState) return;

    // Update state with new file information while maintaining continuity
    this.updateRecordingState(cameraId, {
      currentFile: newFileName,
      fileRotationTimestamp: rotationTimestamp,
      // Maintain continuity across file rotations
      sessionId: currentState.sessionId,
      isRecording: true,
      startTime: currentState.startTime
    });

    // Update progress with new file
    const currentProgress = this.progress.get(cameraId);
    if (currentProgress) {
      this.setProgress(cameraId, {
        ...currentProgress,
        current_file: newFileName,
        file_rotation_timestamp: rotationTimestamp
      });
    }

    console.log(`✅ File rotation handled seamlessly for camera ${cameraId}`);
  }

  /**
   * Update recording state and notify callbacks
   */
  private updateRecordingState(cameraId: string, state: Partial<RecordingState>): void {
    const currentState = this.recordingState.get(cameraId) || {
      isRecording: false,
      sessionId: null,
      startTime: null,
      currentFile: null,
      elapsedTime: 0
    };

    const newState = { ...currentState, ...state };
    this.recordingState.set(cameraId, newState);
  }

  /**
   * Event handlers
   */
  onRecordingConflict(callback: (conflict: RecordingConflict) => void): void {
    this.conflictCallbacks.add(callback);
  }

  onRecordingProgress(callback: (progress: RecordingProgress) => void): void {
    this.progressCallbacks.add(callback);
  }

  onRecordingStateChange(callback: (state: RecordingState) => void): void {
    this.stateCallbacks.add(callback);
  }

  private notifyConflictCallbacks(conflict: RecordingConflict): void {
    this.conflictCallbacks.forEach(callback => callback(conflict));
  }

  private notifyProgressCallbacks(progress: RecordingProgress): void {
    this.progressCallbacks.forEach(callback => callback(progress));
  }

  private notifyStateCallbacks(state: RecordingState): void {
    this.stateCallbacks.forEach(callback => callback(state));
  }

  /**
   * Cleanup
   */
  cleanup(): void {
    this.conflictCallbacks.clear();
    this.progressCallbacks.clear();
    this.stateCallbacks.clear();
    this.recordingState.clear();
    this.activeSessions.clear();
  }
}

// Export singleton instance
export const recordingManagerService = new RecordingManagerService();
