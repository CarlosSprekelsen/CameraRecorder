/**
 * RecordingController - Architecture Compliance
 * 
 * Architecture requirement: "RecordingController component" (Section 5.2)
 * Provides recording control functionality following unidirectional data flow
 */

import React, { useState } from 'react';
import { Box, Typography, Button, Alert, CircularProgress } from '@mui/material';
import { PlayArrow, Stop, Pause } from '@mui/icons-material';
import { useUnifiedStore } from '../../../stores/UnifiedStateStore';
import { APIClient } from '../../../services/abstraction/APIClient';
import { LoggerService } from '../../../services/logger/LoggerService';

interface RecordingControllerProps {
  device: string;
  apiClient: APIClient;
  logger: LoggerService;
}

export const RecordingController: React.FC<RecordingControllerProps> = ({ 
  device, 
  apiClient, 
  logger 
}) => {
  const { recordings, startRecording, stopRecording, setRecordingError } = useUnifiedStore();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const currentRecording = recordings.activeRecordings[device];

  const handleStartRecording = async () => {
    setLoading(true);
    setError(null);
    try {
      await startRecording(device);
      logger.info(`Recording started for device: ${device}`);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to start recording';
      setError(errorMsg);
      setRecordingError(device, errorMsg);
      logger.error(`Failed to start recording for ${device}:`, err);
    } finally {
      setLoading(false);
    }
  };

  const handleStopRecording = async () => {
    setLoading(true);
    setError(null);
    try {
      await stopRecording(device);
      logger.info(`Recording stopped for device: ${device}`);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to stop recording';
      setError(errorMsg);
      setRecordingError(device, errorMsg);
      logger.error(`Failed to stop recording for ${device}:`, err);
    } finally {
      setLoading(false);
    }
  };

  const isRecording = currentRecording?.status === 'RECORDING';

  return (
    <Box sx={{ p: 2, border: 1, borderColor: 'grey.300', borderRadius: 1 }}>
      <Typography variant="h6" gutterBottom>
        Recording Control - {device}
      </Typography>
      
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
        {!isRecording ? (
          <Button
            variant="contained"
            color="primary"
            startIcon={<PlayArrow />}
            onClick={handleStartRecording}
            disabled={loading}
            sx={{ minWidth: 140 }}
          >
            {loading ? <CircularProgress size={20} /> : 'Start Recording'}
          </Button>
        ) : (
          <Button
            variant="contained"
            color="error"
            startIcon={<Stop />}
            onClick={handleStopRecording}
            disabled={loading}
            sx={{ minWidth: 140 }}
          >
            {loading ? <CircularProgress size={20} /> : 'Stop Recording'}
          </Button>
        )}
      </Box>

      {currentRecording && (
        <Box sx={{ mt: 2 }}>
          <Typography variant="body2" color="text.secondary">
            Status: {currentRecording.status}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Started: {new Date(currentRecording.start_time).toLocaleString()}
          </Typography>
        </Box>
      )}
    </Box>
  );
};
