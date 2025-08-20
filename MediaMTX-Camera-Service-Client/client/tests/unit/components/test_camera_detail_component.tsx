/**
 * REQ-UNIT01-001: Camera information display must be clear and accessible
 * REQ-UNIT01-002: Camera controls must be functional and responsive
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for CameraDetail component
 * Tests component rendering and user interactions in isolation
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '../../../src/theme';
import CameraDetail from '../../../src/components/CameraDetail/CameraDetail';
import { useCameraStore } from '../../../src/stores/cameraStore';

// Mock the camera store
jest.mock('../../../src/stores/cameraStore');
const mockUseCameraStore = useCameraStore as jest.MockedFunction<typeof useCameraStore>;

// Mock react-router-dom useParams
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useParams: () => ({ deviceId: 'test-camera-1' }),
  Navigate: ({ to }: { to: string }) => <div data-testid="navigate" data-to={to} />
}));

// Mock camera data matching server API format
const mockCamera = {
  device: 'test-camera-1',
  status: 'CONNECTED' as const,
  name: 'Test Camera',
  resolution: '1920x1080',
  fps: 30,
  streams: {
    rtsp: 'rtsp://localhost:8554/test-camera-1',
    webrtc: 'webrtc://localhost:8889/test-camera-1',
    hls: 'http://localhost:8888/test-camera-1/index.m3u8'
  },
  metrics: {
    bytes_sent: 1024000,
    readers: 2,
    uptime: 3600
  }
};

const mockStore = {
  cameras: [mockCamera],
  selectedCamera: 'test-camera-1',
  activeRecordings: new Map(),
  getCameraStatus: jest.fn(),
  startRecording: jest.fn(),
  stopRecording: jest.fn(),
  takeSnapshot: jest.fn(),
  selectCamera: jest.fn(),
  isConnected: true,
  error: null
};

const renderWithProviders = (component: React.ReactElement) => {
  return render(
    <ThemeProvider theme={theme}>
      <BrowserRouter>
        {component}
      </BrowserRouter>
    </ThemeProvider>
  );
};

describe('CameraDetail Component', () => {
  beforeEach(() => {
    mockUseCameraStore.mockReturnValue(mockStore);
    jest.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render camera information correctly', () => {
      renderWithProviders(<CameraDetail />);

      // Test camera name display - REQ-UNIT01-001
      expect(screen.getByText('Camera: Test Camera')).toBeInTheDocument();
      expect(screen.getByText('Device: test-camera-1')).toBeInTheDocument();
      expect(screen.getByText('Resolution: 1920x1080 | FPS: 30')).toBeInTheDocument();
      expect(screen.getByText('CONNECTED')).toBeInTheDocument();
    });

    it('should display camera metrics when available', () => {
      renderWithProviders(<CameraDetail />);

      // Test metrics display with raw numbers - REQ-UNIT01-001
      expect(screen.getByText('Bytes Sent: 1024000')).toBeInTheDocument();
      expect(screen.getByText('Readers: 2')).toBeInTheDocument();
      expect(screen.getByText('Uptime: 3600s')).toBeInTheDocument();
    });

    it('should display stream URLs when available', () => {
      renderWithProviders(<CameraDetail />);

      // Test stream URL display - REQ-UNIT01-001
      expect(screen.getByText('RTSP:')).toBeInTheDocument();
      expect(screen.getByText('WebRTC:')).toBeInTheDocument();
      expect(screen.getByText('HLS:')).toBeInTheDocument();
      expect(screen.getByText('rtsp://localhost:8554/test-camera-1')).toBeInTheDocument();
      expect(screen.getByText('webrtc://localhost:8889/test-camera-1')).toBeInTheDocument();
      expect(screen.getByText('http://localhost:8888/test-camera-1/index.m3u8')).toBeInTheDocument();
    });

    it('should show warning when camera not found', () => {
      mockUseCameraStore.mockReturnValue({
        ...mockStore,
        cameras: []
      });

      renderWithProviders(<CameraDetail />);

      expect(screen.getByText('Camera not found. Please check the camera connection.')).toBeInTheDocument();
    });
  });

  describe('Snapshot Controls', () => {
    it('should render snapshot controls with default values', () => {
      renderWithProviders(<CameraDetail />);

      expect(screen.getByText('Snapshot Controls')).toBeInTheDocument();
      expect(screen.getByDisplayValue('jpg')).toBeInTheDocument(); // Default format
      expect(screen.getByDisplayValue('80')).toBeInTheDocument(); // Default quality
      expect(screen.getByText('Take Snapshot')).toBeInTheDocument();
    });

    it('should call takeSnapshot when button is clicked', async () => {
      mockStore.takeSnapshot.mockResolvedValue({
        success: true,
        file_path: '/snapshots/test.jpg',
        format: 'jpg',
        quality: 80,
        size: 1024
      });

      renderWithProviders(<CameraDetail />);

      const snapshotButton = screen.getByText('Take Snapshot');
      fireEvent.click(snapshotButton);

      await waitFor(() => {
        expect(mockStore.takeSnapshot).toHaveBeenCalledWith(
          'test-camera-1',
          'jpg',
          80
        );
      });
    });

    it('should disable snapshot button when camera is disconnected', () => {
      mockUseCameraStore.mockReturnValue({
        ...mockStore,
        isConnected: false
      });

      renderWithProviders(<CameraDetail />);

      const snapshotButton = screen.getByText('Take Snapshot');
      expect(snapshotButton).toBeDisabled();
    });

    it('should show loading state during snapshot capture', async () => {
      mockStore.takeSnapshot.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

      renderWithProviders(<CameraDetail />);

      const snapshotButton = screen.getByText('Take Snapshot');
      fireEvent.click(snapshotButton);

      // Look for any CircularProgress component
      expect(screen.getAllByRole('progressbar').length).toBeGreaterThan(0);
    });
  });

  describe('Recording Controls', () => {
    it('should render recording controls with default values', () => {
      renderWithProviders(<CameraDetail />);

      expect(screen.getByText('Recording Controls')).toBeInTheDocument();
      expect(screen.getByDisplayValue('mp4')).toBeInTheDocument(); // Default format
      expect(screen.getByText('Unlimited Duration')).toBeInTheDocument();
      expect(screen.getByText('Start Recording')).toBeInTheDocument();
      expect(screen.getByText('Stop Recording')).toBeInTheDocument();
    });

    it('should toggle unlimited duration mode', () => {
      renderWithProviders(<CameraDetail />);

      const unlimitedSwitch = screen.getByRole('checkbox');
      expect(unlimitedSwitch).not.toBeChecked();

      fireEvent.click(unlimitedSwitch);
      expect(unlimitedSwitch).toBeChecked();
    });

    it('should call startRecording when button is clicked', async () => {
      mockStore.startRecording.mockResolvedValue({
        device: 'test-camera-1',
        session_id: 'session-123',
        filename: 'test-recording.mp4',
        status: 'STARTED',
        start_time: '2024-01-01T00:00:00Z',
        format: 'mp4'
      });

      renderWithProviders(<CameraDetail />);

      const startButton = screen.getByText('Start Recording');
      fireEvent.click(startButton);

      await waitFor(() => {
        expect(mockStore.startRecording).toHaveBeenCalledWith(
          'test-camera-1',
          undefined, // duration (undefined for unlimited)
          'mp4'
        );
      });
    });

    it('should disable start recording when already recording', () => {
      mockUseCameraStore.mockReturnValue({
        ...mockStore,
        activeRecordings: new Map([['test-camera-1', {}]])
      });

      renderWithProviders(<CameraDetail />);

      const startButton = screen.getByText('Start Recording');
      expect(startButton).toBeDisabled();
    });

    it('should disable stop recording when not recording', () => {
      renderWithProviders(<CameraDetail />);

      const stopButton = screen.getByText('Stop Recording');
      expect(stopButton).toBeDisabled();
    });

    it('should show recording status when active', () => {
      mockUseCameraStore.mockReturnValue({
        ...mockStore,
        activeRecordings: new Map([['test-camera-1', {}]])
      });

      renderWithProviders(<CameraDetail />);

      expect(screen.getByText('Recording Active')).toBeInTheDocument();
      expect(screen.getByText('Recording in progress...')).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should display error message when snapshot fails', async () => {
      mockStore.takeSnapshot.mockRejectedValue(new Error('Snapshot failed'));

      renderWithProviders(<CameraDetail />);

      const snapshotButton = screen.getByText('Take Snapshot');
      fireEvent.click(snapshotButton);

      await waitFor(() => {
        expect(screen.getByText('Snapshot failed')).toBeInTheDocument();
      });
    });

    it('should display error message when recording fails', async () => {
      mockStore.startRecording.mockRejectedValue(new Error('Recording failed'));

      renderWithProviders(<CameraDetail />);

      const startButton = screen.getByText('Start Recording');
      fireEvent.click(startButton);

      await waitFor(() => {
        expect(screen.getByText('Recording failed')).toBeInTheDocument();
      });
    });

    it('should clear error when new operation starts', async () => {
      mockStore.takeSnapshot.mockRejectedValueOnce(new Error('Snapshot failed'));
      mockStore.takeSnapshot.mockResolvedValueOnce({
        success: true,
        file_path: '/snapshots/test.jpg',
        format: 'jpg',
        quality: 80,
        size: 1024
      });

      renderWithProviders(<CameraDetail />);

      const snapshotButton = screen.getByText('Take Snapshot');
      
      // First click - should show error
      fireEvent.click(snapshotButton);
      await waitFor(() => {
        expect(screen.getByText('Snapshot failed')).toBeInTheDocument();
      });

      // Second click - should clear error and succeed
      fireEvent.click(snapshotButton);
      await waitFor(() => {
        expect(screen.queryByText('Snapshot failed')).not.toBeInTheDocument();
      });
    });
  });
});
