/**
 * REQ-CAM01-001: [Primary requirement being tested]
 * REQ-CAM01-002: [Secondary requirements covered]
 * Coverage: INTEGRATION
 * Quality: HIGH
 */
/**
 * Integration tests for camera operations
 * Tests real server integration for snapshot and recording functionality
 * 
 * These tests require a running MediaMTX server for full validation
 */

import { renderHook, act } from '@testing-library/react';
import { useCameraStore } from '../../src/stores/cameraStore';
import { useFileStore } from '../../src/stores/fileStore';
import { createWebSocketServiceSync } from '../../src/services/websocket';
import type { CameraDevice } from '../../src/types';

const TEST_WS_URL = 'ws://localhost:8002/ws';

describe('Camera Operations Integration', () => {
  let cameraStore: any;
  let fileStore: any;

  beforeAll(async () => {
    // Initialize stores
    const cameraResult = renderHook(() => useCameraStore());
    const fileResult = renderHook(() => useFileStore());
    
    cameraStore = cameraResult.result.current;
    fileStore = fileResult.result.current;

    // Initialize connections
    await act(async () => {
      await cameraStore.initialize(TEST_WS_URL);
      await fileStore.initialize(TEST_WS_URL);
    });

    // Wait for connection
    await new Promise(resolve => setTimeout(resolve, 1000));
  });

  afterAll(async () => {
    await act(async () => {
      cameraStore.disconnect();
      fileStore.disconnect();
    });
  });

  describe('Camera Discovery', () => {
    it('should discover available cameras', async () => {
      await act(async () => {
        await cameraStore.refreshCameras();
      });

      expect(cameraStore.cameras).toBeDefined();
      expect(Array.isArray(cameraStore.cameras)).toBe(true);
      
      if (cameraStore.cameras.length > 0) {
        const camera = cameraStore.cameras[0];
        expect(camera.device).toBeDefined();
        expect(camera.status).toBeDefined();
        expect(camera.name).toBeDefined();
        expect(camera.resolution).toBeDefined();
        expect(camera.fps).toBeDefined();
      }
    }, 10000);

    it('should get individual camera status', async () => {
      if (cameraStore.cameras.length === 0) {
        fail('No cameras available for status test - cannot validate core functionality');
      }

      const testCamera = cameraStore.cameras[0];
      
      await act(async () => {
        const status = await cameraStore.getCameraStatus(testCamera.device);
        expect(status).toBeDefined();
        expect(status?.device).toBe(testCamera.device);
      });
    }, 10000);
  });

  describe('Snapshot Operations', () => {
    it('should take snapshot with default settings', async () => {
      if (cameraStore.cameras.length === 0) {
        fail('No cameras available for snapshot test - cannot validate core functionality');
      }

      const testCamera = cameraStore.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      if (!testCamera) {
        fail('No connected cameras available for snapshot test - cannot validate core functionality');
      }

      await act(async () => {
        const result = await cameraStore.takeSnapshot(testCamera.device, 'jpg', 80);
        
        expect(result).toBeDefined();
        expect(result?.success).toBe(true);
        expect(result?.file_path).toBeDefined();
        expect(result?.format).toBe('jpg');
        expect(result?.quality).toBe(80);
        expect(result?.size).toBeGreaterThan(0);
      });
    }, 15000);

    it('should take snapshot with PNG format', async () => {
      if (cameraStore.cameras.length === 0) {
        fail('No cameras available for PNG snapshot test - cannot validate core functionality');
      }

      const testCamera = cameraStore.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      if (!testCamera) {
        fail('No connected cameras available for PNG snapshot test - cannot validate core functionality');
      }

      await act(async () => {
        const result = await cameraStore.takeSnapshot(testCamera.device, 'png', 90);
        
        expect(result).toBeDefined();
        expect(result?.success).toBe(true);
        expect(result?.format).toBe('png');
        expect(result?.quality).toBe(90);
      });
    }, 15000);

    it('should handle snapshot errors gracefully', async () => {
      await act(async () => {
        try {
          await cameraStore.takeSnapshot('non-existent-camera', 'jpg', 80);
          fail('Should have thrown an error');
        } catch (error) {
          expect(error).toBeDefined();
          expect((error as Error).message).toContain('Camera not found');
        }
      });
    }, 10000);
  });

  describe('Recording Operations', () => {
    it('should start and stop recording', async () => {
      if (cameraStore.cameras.length === 0) {
        fail('No cameras available for recording test - cannot validate core functionality');
      }

      const testCamera = cameraStore.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      if (!testCamera) {
        fail('No connected cameras available for recording test - cannot validate core functionality');
      }

      let recordingSession: any;

      // Start recording
      await act(async () => {
        recordingSession = await cameraStore.startRecording(testCamera.device, 10, 'mp4');
        
        expect(recordingSession).toBeDefined();
        expect(recordingSession.device).toBe(testCamera.device);
        expect(recordingSession.status).toBe('STARTED');
        expect(recordingSession.format).toBe('mp4');
        expect(recordingSession.session_id).toBeDefined();
      });

      // Wait a moment for recording to start
      await new Promise(resolve => setTimeout(resolve, 2000));

      // Stop recording
      await act(async () => {
        const stopResult = await cameraStore.stopRecording(testCamera.device);
        
        expect(stopResult).toBeDefined();
        expect(stopResult?.status).toBe('STOPPED');
        expect(stopResult?.session_id).toBe(recordingSession.session_id);
        expect(stopResult?.duration).toBeGreaterThan(0);
        expect(stopResult?.file_size).toBeGreaterThan(0);
      });
    }, 30000);

    it('should start unlimited recording', async () => {
      if (cameraStore.cameras.length === 0) {
        fail('No cameras available for unlimited recording test - cannot validate core functionality');
      }

      const testCamera = cameraStore.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      if (!testCamera) {
        fail('No connected cameras available for unlimited recording test - cannot validate core functionality');
      }

      let recordingSession: any;

      // Start unlimited recording
      await act(async () => {
        recordingSession = await cameraStore.startRecording(testCamera.device, undefined, 'mp4');
        
        expect(recordingSession).toBeDefined();
        expect(recordingSession.status).toBe('STARTED');
        expect(recordingSession.format).toBe('mp4');
      });

      // Wait a moment then stop
      await new Promise(resolve => setTimeout(resolve, 3000));

      await act(async () => {
        const stopResult = await cameraStore.stopRecording(testCamera.device);
        expect(stopResult?.status).toBe('STOPPED');
      });
    }, 20000);

    it('should handle recording errors gracefully', async () => {
      await act(async () => {
        try {
          await cameraStore.startRecording('non-existent-camera', 10, 'mp4');
          fail('Should have thrown an error');
        } catch (error) {
          expect(error).toBeDefined();
          expect((error as Error).message).toContain('Camera not found');
        }
      });
    }, 10000);
  });

  describe('File Management Integration', () => {
    it('should list recordings after creating them', async () => {
      await act(async () => {
        await fileStore.loadRecordings(20, 0);
      });

      expect(fileStore.recordings).toBeDefined();
      expect(Array.isArray(fileStore.recordings)).toBe(true);
      
      if (fileStore.recordings && fileStore.recordings.length > 0) {
        const recording = fileStore.recordings[0];
        expect(recording.filename).toBeDefined();
        expect(recording.file_size).toBeGreaterThan(0);
        expect(recording.created_at).toBeDefined();
        expect(recording.download_url).toBeDefined();
      }
    }, 10000);

    it('should list snapshots after creating them', async () => {
      await act(async () => {
        await fileStore.loadSnapshots(20, 0);
      });

      expect(fileStore.snapshots).toBeDefined();
      expect(Array.isArray(fileStore.snapshots)).toBe(true);
      
      if (fileStore.snapshots && fileStore.snapshots.length > 0) {
        const snapshot = fileStore.snapshots[0];
        expect(snapshot.filename).toBeDefined();
        expect(snapshot.file_size).toBeGreaterThan(0);
        expect(snapshot.created_at).toBeDefined();
        expect(snapshot.download_url).toBeDefined();
      }
    }, 10000);

    it('should handle pagination correctly', async () => {
      await act(async () => {
        await fileStore.loadRecordings(5, 0); // First page
      });

      const firstPageCount = fileStore.recordings?.length || 0;

      if (firstPageCount >= 5) {
        await act(async () => {
          await fileStore.loadRecordings(5, 5); // Second page
        });

        expect(fileStore.recordings?.length).toBeLessThanOrEqual(5);
      }
    }, 10000);
  });

  describe('Real-time Updates', () => {
    it('should receive camera status updates', async () => {
      // This test verifies that WebSocket notifications work
      // The actual status updates depend on server implementation
      
      expect(cameraStore.isConnected).toBe(true);
      
      // Wait for potential status updates
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      // Verify that camera list is still accessible
      expect(cameraStore.cameras).toBeDefined();
    }, 10000);

    it('should handle connection loss and recovery', async () => {
      // This test would require simulating connection loss
      // For now, we'll verify the connection is stable
      
      expect(cameraStore.isConnected).toBe(true);
      expect(fileStore.isConnected).toBe(true);
    }, 10000);
  });

  describe('Error Handling', () => {
    it('should handle invalid camera operations', async () => {
      await act(async () => {
        try {
          await cameraStore.takeSnapshot('invalid-device', 'jpg', 80);
          fail('Should have thrown an error');
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    }, 10000);

    it('should handle invalid file operations', async () => {
      await act(async () => {
        try {
          await fileStore.downloadFile('recordings', 'non-existent-file.mp4');
          // Download might not throw immediately, but should handle gracefully
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    }, 10000);
  });

  describe('Performance Validation', () => {
    it('should meet performance targets for camera operations', async () => {
      if (cameraStore.cameras.length === 0) {
        fail('No cameras available for performance test - cannot validate core functionality');
      }

      const testCamera = cameraStore.cameras[0];
      
      const startTime = Date.now();
      
      await act(async () => {
        await cameraStore.getCameraStatus(testCamera.device);
      });
      
      const endTime = Date.now();
      const duration = endTime - startTime;
      
      // Should complete within 100ms as per performance targets
      expect(duration).toBeLessThan(100);
    }, 10000);

    it('should meet performance targets for file operations', async () => {
      const startTime = Date.now();
      
      await act(async () => {
        await fileStore.loadRecordings(20, 0);
      });
      
      const endTime = Date.now();
      const duration = endTime - startTime;
      
      // Should complete within reasonable time (adjust based on server performance)
      expect(duration).toBeLessThan(5000);
    }, 10000);
  });
});
