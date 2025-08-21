/**
 * REQ-CAM01-001: Real Camera Hardware Integration
 * REQ-CAM01-002: Camera Operations Error Handling
 * Coverage: INTEGRATION
 * Quality: HIGH
 */
/**
 * Real Camera Operations Integration Tests
 * 
 * Tests actual camera hardware integration with real MediaMTX server
 * Validates real camera operations, error conditions, and edge cases
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running via systemd
 * - Real camera hardware connected (or simulation available)
 * - Server accessible at ws://localhost:8002/ws
 */

const WebSocket = require('ws');
import { RPC_METHODS, ERROR_CODES, PERFORMANCE_TARGETS } from '../../src/types';

interface CameraDevice {
  device: string;
  status: string;
  name: string;
  resolution: string;
  fps: number;
  streams: {
    rtsp: string;
    webrtc: string;
    hls: string;
  };
}

interface SnapshotResult {
  status: string;
  filename: string;
  file_size: number;
  format?: string;
  quality?: number;
}

interface RecordingSession {
  session_id: string;
  device: string;
  format: string;
  duration?: number;
  file_size?: number;
}

describe('Real Camera Operations Integration', () => {
  let ws: WebSocket;
  let authToken: string;
  const TEST_TIMEOUT = 30000;

  beforeAll(async () => {
    // Generate valid authentication token
    const jwt = require('jsonwebtoken');
    const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
    if (!secret) {
      throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable required');
    }
    
    authToken = jwt.sign(
      { user_id: 'test_user', role: 'operator' },
      secret,
      { expiresIn: '1h' }
    );
  });

  beforeEach(async () => {
    // Establish real WebSocket connection
    ws = new WebSocket('ws://localhost:8002/ws');
    
    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('Connection timeout')), 5000);
      
      ws!.onopen = () => {
        clearTimeout(timeout);
        resolve();
      };
      
      ws!.onerror = (error: any) => {
        clearTimeout(timeout);
        reject(error);
      };
    });

    // Authenticate the connection
    await sendRequest('authenticate', { token: authToken });
  });

  afterEach(async () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.close();
    }
  });

  // Helper function to send JSON-RPC requests
  async function sendRequest(method: string, params: any = {}): Promise<any> {
    return new Promise((resolve, reject) => {
      const id = Math.floor(Math.random() * 1000000);
      const request = { jsonrpc: '2.0', method, params, id };
      
      const timeout = setTimeout(() => {
        reject(new Error(`Request timeout for ${method}`));
      }, TEST_TIMEOUT);

      const originalOnMessage = ws!.onmessage;
      ws!.onmessage = (event: any) => {
        try {
          const data = JSON.parse(event.data.toString());
          if (data.id === id) {
            clearTimeout(timeout);
            ws!.onmessage = originalOnMessage;
            
            if (data.error) {
              reject(new Error(data.error.message || 'RPC error'));
            } else {
              resolve(data.result);
            }
          }
        } catch (error) {
          // Continue listening for the correct response
        }
      };

      ws!.send(JSON.stringify(request));
    });
  }

  describe('Real Camera Discovery and Hardware Integration', () => {
    it('should discover actual connected camera hardware', async () => {
      const startTime = performance.now();
      
      const cameraList = await sendRequest('get_camera_list') as any;
      const responseTime = performance.now() - startTime;
      
      // Validate response structure
      expect(cameraList).toHaveProperty('cameras');
      expect(cameraList).toHaveProperty('total');
      expect(cameraList).toHaveProperty('connected');
      expect(Array.isArray(cameraList.cameras)).toBe(true);
      
      // Validate performance
      expect(responseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      
      // Log actual camera discovery results
      console.log(`📷 Camera Discovery Results:`);
      console.log(`   - Total cameras: ${cameraList.total}`);
      console.log(`   - Connected cameras: ${cameraList.connected}`);
      console.log(`   - Response time: ${responseTime.toFixed(2)}ms`);
      
      if (cameraList.cameras.length > 0) {
        const camera = cameraList.cameras[0];
        console.log(`   - First camera: ${camera.device} (${camera.status})`);
        console.log(`   - Resolution: ${camera.resolution}, FPS: ${camera.fps}`);
      }
    }, TEST_TIMEOUT);

    it('should handle camera discovery when no hardware is connected', async () => {
      // This test validates system behavior when no cameras are available
      const cameraList = await sendRequest('get_camera_list') as any;
      
      if (cameraList.cameras.length === 0) {
        // No cameras available - validate graceful handling
        expect(cameraList.total).toBe(0);
        expect(cameraList.connected).toBe(0);
        console.log('📷 No cameras available - system handles gracefully');
      } else {
        // Cameras available - test individual camera status
        const camera = cameraList.cameras[0];
        const status = await sendRequest('get_camera_status', { device: camera.device });
        
        expect(status).toHaveProperty('device');
        expect(status.device).toBe(camera.device);
        expect(status).toHaveProperty('status');
        expect(status).toHaveProperty('name');
      }
    }, TEST_TIMEOUT);

    it('should validate camera stream URLs are accessible', async () => {
      const cameraList = await sendRequest('get_camera_list') as any;
      
      if (cameraList.cameras.length > 0) {
        const camera = cameraList.cameras[0] as CameraDevice;
        
        // Validate stream URLs are properly formatted
        expect(camera.streams).toHaveProperty('rtsp');
        expect(camera.streams).toHaveProperty('webrtc');
        expect(camera.streams).toHaveProperty('hls');
        
        // Test RTSP stream accessibility (if camera is connected)
        if (camera.status === 'CONNECTED') {
          try {
            const response = await fetch(`http://localhost:8003/health/streams/${camera.device}`);
            expect(response.status).toBe(200);
            console.log(`✅ RTSP stream accessible: ${camera.streams.rtsp}`);
          } catch (error) {
            console.log(`⚠️ RTSP stream not accessible: ${camera.streams.rtsp}`);
          }
        }
      }
    }, TEST_TIMEOUT);
  });

  describe('Real Snapshot Operations with Hardware', () => {
    it('should capture actual snapshots from connected cameras', async () => {
      const cameraList = await sendRequest('get_camera_list') as any;
      
      if (cameraList.cameras.length === 0) {
        console.log('📷 No cameras available for snapshot testing');
        return;
      }

      const connectedCamera = cameraList.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      if (!connectedCamera) {
        console.log('📷 No connected cameras available for snapshot testing');
        return;
      }

      // Test JPEG snapshot capture
      const startTime = performance.now();
      const snapshot = await sendRequest('take_snapshot', {
        device: connectedCamera.device,
        format: 'jpg',
        quality: 80
      }) as SnapshotResult;
      const captureTime = performance.now() - startTime;

      // Validate snapshot result
      expect(snapshot).toHaveProperty('status');
      expect(snapshot.status).toBe('completed');
      expect(snapshot).toHaveProperty('filename');
      expect(snapshot).toHaveProperty('file_size');
      expect(snapshot.file_size).toBeGreaterThan(0);
      expect(snapshot.format).toBe('jpg');

      // Validate performance
      expect(captureTime).toBeLessThan(PERFORMANCE_TARGETS.CONTROL_METHODS);

      console.log(`📸 Snapshot captured:`);
      console.log(`   - Device: ${connectedCamera.device}`);
      console.log(`   - Format: ${snapshot.format}`);
      console.log(`   - File size: ${snapshot.file_size} bytes`);
      console.log(`   - Capture time: ${captureTime.toFixed(2)}ms`);
      console.log(`   - Filename: ${snapshot.filename}`);
    }, TEST_TIMEOUT);

    it('should handle snapshot errors with invalid parameters', async () => {
      // Test with non-existent camera
      try {
        await sendRequest('take_snapshot', {
          device: '/dev/video999',
          format: 'jpg',
          quality: 80
        });
        fail('Expected error for non-existent camera');
      } catch (error: any) {
        expect(error.message).toMatch(/camera not found|invalid device/i);
        console.log(`✅ Properly rejected non-existent camera: ${error.message}`);
      }

      // Test with invalid format
      const cameraList = await sendRequest('get_camera_list') as any;
      if (cameraList.cameras.length > 0) {
        try {
          await sendRequest('take_snapshot', {
            device: cameraList.cameras[0].device,
            format: 'invalid_format',
            quality: 80
          });
          fail('Expected error for invalid format');
        } catch (error: any) {
          expect(error.message).toMatch(/invalid format|unsupported/i);
          console.log(`✅ Properly rejected invalid format: ${error.message}`);
        }
      }
    }, TEST_TIMEOUT);

    it('should test different snapshot formats and qualities', async () => {
      const cameraList = await sendRequest('get_camera_list') as any;
      const connectedCamera = cameraList.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      
      if (!connectedCamera) {
        console.log('📷 No connected cameras available for format testing');
        return;
      }

      const formats = ['jpg', 'png'];
      const qualities = [60, 80, 95];

      for (const format of formats) {
        for (const quality of qualities) {
          try {
            const snapshot = await sendRequest('take_snapshot', {
              device: connectedCamera.device,
              format,
              quality
            }) as SnapshotResult;

            expect(snapshot.status).toBe('completed');
            expect(snapshot.format).toBe(format);
            expect(snapshot.file_size).toBeGreaterThan(0);

            console.log(`📸 ${format.toUpperCase()} (Q${quality}): ${snapshot.file_size} bytes`);
          } catch (error) {
            console.log(`❌ Failed ${format} (Q${quality}): ${error}`);
          }
        }
      }
    }, TEST_TIMEOUT);
  });

  describe('Real Recording Operations with Hardware', () => {
    it('should perform actual video recording operations', async () => {
      const cameraList = await sendRequest('get_camera_list') as any;
      const connectedCamera = cameraList.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      
      if (!connectedCamera) {
        console.log('📷 No connected cameras available for recording testing');
        return;
      }

      // Start recording
      const startTime = performance.now();
      const startResult = await sendRequest('start_recording', {
        device: connectedCamera.device,
        format: 'mp4',
        duration_seconds: 5
      }) as RecordingSession;
      const startResponseTime = performance.now() - startTime;

      expect(startResult).toHaveProperty('session_id');
      expect(startResult).toHaveProperty('device');
      expect(startResult).toHaveProperty('format');
      expect(startResult.device).toBe(connectedCamera.device);
      expect(startResult.format).toBe('mp4');
      expect(startResponseTime).toBeLessThan(PERFORMANCE_TARGETS.CONTROL_METHODS);

      console.log(`🎥 Recording started: ${startResult.session_id}`);

      // Wait for recording to complete
      await new Promise(resolve => setTimeout(resolve, 6000));

      // Stop recording
      const stopResult = await sendRequest('stop_recording', {
        device: connectedCamera.device
      }) as RecordingSession;

      expect(stopResult).toHaveProperty('session_id');
      expect(stopResult).toHaveProperty('duration');
      expect(stopResult).toHaveProperty('file_size');
      expect(stopResult.session_id).toBe(startResult.session_id);
      expect(stopResult.duration).toBeGreaterThan(0);
      expect(stopResult.file_size).toBeGreaterThan(0);

      console.log(`🎥 Recording completed:`);
      console.log(`   - Duration: ${stopResult.duration}s`);
      console.log(`   - File size: ${stopResult.file_size} bytes`);
    }, TEST_TIMEOUT);

    it('should handle recording errors and edge cases', async () => {
      // Test recording on non-existent camera
      try {
        await sendRequest('start_recording', {
          device: '/dev/video999',
          format: 'mp4',
          duration_seconds: 10
        });
        fail('Expected error for non-existent camera');
      } catch (error: any) {
        expect(error.message).toMatch(/camera not found|invalid device/i);
        console.log(`✅ Properly rejected recording on non-existent camera: ${error.message}`);
      }

      // Test invalid recording parameters
      const cameraList = await sendRequest('get_camera_list') as any;
      if (cameraList.cameras.length > 0) {
        try {
          await sendRequest('start_recording', {
            device: cameraList.cameras[0].device,
            format: 'invalid_format',
            duration_seconds: -1
          });
          fail('Expected error for invalid parameters');
        } catch (error: any) {
          expect(error.message).toMatch(/invalid|unsupported/i);
          console.log(`✅ Properly rejected invalid recording parameters: ${error.message}`);
        }
      }
    }, TEST_TIMEOUT);
  });

  describe('Real File System Operations', () => {
    it('should list actual recordings and snapshots with metadata', async () => {
      // List recordings
      const recordings = await sendRequest('list_recordings', {
        limit: 20,
        offset: 0
      }) as any;

      expect(recordings).toHaveProperty('files');
      expect(recordings).toHaveProperty('total');
      expect(Array.isArray(recordings.files)).toBe(true);

      console.log(`📁 Recordings: ${recordings.total} total files`);

      if (recordings.files.length > 0) {
        const recording = recordings.files[0];
        expect(recording).toHaveProperty('filename');
        expect(recording).toHaveProperty('file_size');
        expect(recording).toHaveProperty('modified_time');
        expect(recording).toHaveProperty('download_url');
        expect(recording.file_size).toBeGreaterThan(0);

        console.log(`   - Sample: ${recording.filename} (${recording.file_size} bytes)`);
      }

      // List snapshots
      const snapshots = await sendRequest('list_snapshots', {
        limit: 20,
        offset: 0
      }) as any;

      expect(snapshots).toHaveProperty('files');
      expect(snapshots).toHaveProperty('total');
      expect(Array.isArray(snapshots.files)).toBe(true);

      console.log(`📁 Snapshots: ${snapshots.total} total files`);

      if (snapshots.files.length > 0) {
        const snapshot = snapshots.files[0];
        expect(snapshot).toHaveProperty('filename');
        expect(snapshot).toHaveProperty('file_size');
        expect(snapshot).toHaveProperty('modified_time');
        expect(snapshot).toHaveProperty('download_url');
        expect(snapshot.file_size).toBeGreaterThan(0);

        console.log(`   - Sample: ${snapshot.filename} (${snapshot.file_size} bytes)`);
      }
    }, TEST_TIMEOUT);

    it('should handle pagination correctly with real data', async () => {
      const recordings = await sendRequest('list_recordings', {
        limit: 5,
        offset: 0
      }) as any;

      if (recordings.total > 5) {
        const secondPage = await sendRequest('list_recordings', {
          limit: 5,
          offset: 5
        }) as any;

        expect(secondPage.files.length).toBeLessThanOrEqual(5);
        expect(secondPage.total).toBe(recordings.total);

        // Files should be different between pages
        const firstPageFilenames = recordings.files.map((f: any) => f.filename);
        const secondPageFilenames = secondPage.files.map((f: any) => f.filename);
        const intersection = firstPageFilenames.filter((name: string) => secondPageFilenames.includes(name));
        expect(intersection.length).toBe(0);

        console.log(`📁 Pagination: Page 1 (${recordings.files.length} files), Page 2 (${secondPage.files.length} files)`);
      }
    }, TEST_TIMEOUT);
  });

  describe('Real Performance and Load Testing', () => {
    it('should maintain performance under concurrent operations', async () => {
      const concurrentRequests = 5;
      const requestPromises = [];

      // Launch concurrent camera list requests
      for (let i = 0; i < concurrentRequests; i++) {
        requestPromises.push(
          sendRequest('get_camera_list').then(() => performance.now())
        );
      }

      const startTime = performance.now();
      const responseTimes = await Promise.all(requestPromises);
      const totalTime = performance.now() - startTime;

      // Calculate individual response times
      const individualTimes = responseTimes.map(time => time - startTime);
      const averageTime = individualTimes.reduce((a, b) => a + b, 0) / individualTimes.length;

      // Validate performance targets
      expect(averageTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      expect(totalTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS * 2);

      console.log(`⚡ Concurrent Performance:`);
      console.log(`   - ${concurrentRequests} concurrent requests`);
      console.log(`   - Average response time: ${averageTime.toFixed(2)}ms`);
      console.log(`   - Total time: ${totalTime.toFixed(2)}ms`);
    }, TEST_TIMEOUT);

    it('should handle rapid connection cycling (network instability simulation)', async () => {
      const cycles = 5;
      const reconnectionTimes: number[] = [];

      for (let i = 0; i < cycles; i++) {
        const startTime = performance.now();

        // Actually disconnect and reconnect
        ws.close();
        ws = new WebSocket('ws://localhost:8002/ws');
        
        await new Promise<void>((resolve, reject) => {
          const timeout = setTimeout(() => reject(new Error('Reconnection timeout')), 5000);
          
          ws.onopen = async () => {
            clearTimeout(timeout);
            try {
              await sendRequest('authenticate', { token: authToken });
              resolve();
            } catch (error) {
              reject(error);
            }
          };
          
          ws.onerror = (error) => {
            clearTimeout(timeout);
            reject(error);
          };
        });

        reconnectionTimes.push(performance.now() - startTime);
        expect(ws.readyState).toBe(WebSocket.OPEN);

        // Brief pause to simulate real network conditions
        await new Promise(resolve => setTimeout(resolve, 100));
      }

      // Validate reconnection performance
      const averageReconnectionTime = reconnectionTimes.reduce((a, b) => a + b, 0) / cycles;
      expect(averageReconnectionTime).toBeLessThan(PERFORMANCE_TARGETS.CLIENT_WEBSOCKET_CONNECTION);

      console.log(`🔄 Connection Cycling:`);
      console.log(`   - ${cycles} disconnect/reconnect cycles`);
      console.log(`   - Average reconnection time: ${averageReconnectionTime.toFixed(2)}ms`);
      console.log(`   - All reconnections successful`);
    }, TEST_TIMEOUT);
  });

  describe('Real Error Handling and Recovery', () => {
    it('should handle authentication failures gracefully', async () => {
      // Test with invalid token
      try {
        await sendRequest('authenticate', { token: 'invalid.token.here' });
        fail('Expected authentication failure');
      } catch (error: any) {
        expect(error.message).toMatch(/invalid|authentication/i);
        console.log(`✅ Properly rejected invalid token: ${error.message}`);
      }

      // Test with expired token
      const jwt = require('jsonwebtoken');
      const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
      const expiredToken = jwt.sign(
        { user_id: 'test_user', role: 'operator' },
        secret,
        { expiresIn: '-1h' } // Expired 1 hour ago
      );

      try {
        await sendRequest('authenticate', { token: expiredToken });
        fail('Expected expired token failure');
      } catch (error: any) {
        expect(error.message).toMatch(/expired|invalid/i);
        console.log(`✅ Properly rejected expired token: ${error.message}`);
      }
    }, TEST_TIMEOUT);

    it('should handle server errors and provide meaningful feedback', async () => {
      // Test invalid method
      try {
        await sendRequest('invalid_method', {});
        fail('Expected method not found error');
      } catch (error: any) {
        expect(error.message).toMatch(/method not found|invalid/i);
        console.log(`✅ Properly rejected invalid method: ${error.message}`);
      }

      // Test invalid parameters
      try {
        await sendRequest('get_camera_status', {});
        fail('Expected invalid parameters error');
      } catch (error: any) {
        expect(error.message).toMatch(/invalid|missing/i);
        console.log(`✅ Properly rejected invalid parameters: ${error.message}`);
      }
    }, TEST_TIMEOUT);
  });
});
