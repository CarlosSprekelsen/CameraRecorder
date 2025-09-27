/**
 * RecordingService unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-UNIT-001: RecordingService snapshot functionality
 * - REQ-UNIT-002: RecordingService recording functionality
 * - REQ-UNIT-003: Error handling
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { RecordingService } from '../../../src/services/recording/RecordingService';
import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MockDataFactory } from '../../utils/mocks';

// Use centralized mocks - eliminates duplication
const mockWebSocketService = MockDataFactory.createMockWebSocketService();
const mockLoggerService = MockDataFactory.createMockLoggerService();

describe('RecordingService Unit Tests', () => {
  let recordingService: RecordingService;

  beforeEach(() => {
    // Use centralized mocks - eliminates duplication

    // Create service instance
    recordingService = new RecordingService(mockWebSocketService, mockLoggerService);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  test('REQ-UNIT-001: takeSnapshot should call WebSocket service with correct parameters', async () => {
    const device = 'camera0';
    const filename = 'test-snapshot.jpg';
    const expectedResult = {
      device,
      filename,
      status: 'SUCCESS',
      timestamp: new Date().toISOString()
    };

    mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

    const result = await recordingService.takeSnapshot(device, filename);

    expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('take_snapshot', { device, filename });
    expect(mockLoggerService.info).toHaveBeenCalledWith('take_snapshot request', { device, filename });
    expect(result).toEqual(expectedResult);
  });

  test('REQ-UNIT-002: takeSnapshot should handle errors correctly', async () => {
    const device = 'camera0';
    const error = new Error('WebSocket connection failed');

    mockWebSocketService.sendRPC.mockRejectedValue(error);

    await expect(recordingService.takeSnapshot(device)).rejects.toThrow('WebSocket connection failed');
    expect(mockLoggerService.error).toHaveBeenCalledWith('take_snapshot failed', error);
  });

  test('REQ-UNIT-003: startRecording should call WebSocket service with correct parameters', async () => {
    const device = 'camera0';
    const duration = 60;
    const format = 'mp4';
    const expectedResult = {
      device,
      status: 'RECORDING',
      start_time: new Date().toISOString()
    };

    mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

    const result = await recordingService.startRecording(device, duration, format);

    expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('start_recording', { device, duration, format });
    expect(mockLoggerService.info).toHaveBeenCalledWith('start_recording request', { device, duration, format });
    expect(result).toEqual(expectedResult);
  });

  test('REQ-UNIT-004: startRecording should handle errors correctly', async () => {
    const device = 'camera0';
    const error = new Error('Recording failed');

    mockWebSocketService.sendRPC.mockRejectedValue(error);

    await expect(recordingService.startRecording(device)).rejects.toThrow('Recording failed');
    expect(mockLoggerService.error).toHaveBeenCalledWith('start_recording failed', error);
  });

  test('REQ-UNIT-005: stopRecording should call WebSocket service with correct parameters', async () => {
    const device = 'camera0';
    const expectedResult = {
      device,
      status: 'STOPPED',
      end_time: new Date().toISOString()
    };

    mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

    const result = await recordingService.stopRecording(device);

    expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('stop_recording', { device });
    expect(mockLoggerService.info).toHaveBeenCalledWith('stop_recording request', { device });
    expect(result).toEqual(expectedResult);
  });

  test('REQ-UNIT-006: stopRecording should handle errors correctly', async () => {
    const device = 'camera0';
    const error = new Error('Stop recording failed');

    mockWebSocketService.sendRPC.mockRejectedValue(error);

    await expect(recordingService.stopRecording(device)).rejects.toThrow('Stop recording failed');
    expect(mockLoggerService.error).toHaveBeenCalledWith('stop_recording failed', error);
  });
});
