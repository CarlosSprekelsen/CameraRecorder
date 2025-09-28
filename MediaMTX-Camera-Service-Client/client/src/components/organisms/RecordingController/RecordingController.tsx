/**
 * RecordingController - Architecture Compliance
 * 
 * Architecture requirement: "RecordingController component" (Section 5.2)
 * Provides recording control functionality following unidirectional data flow
 */

import React, { useState } from 'react';
import { Box, Typography, Button, Alert, CircularProgress } from '@mui/material';
import { PlayArrow, Stop } from '@mui/icons-material';
import { useRecordingStore } from '../../../stores/recording/recordingStore';
import { logger } from '../../../services/logger/LoggerService';
// ARCHITECTURE FIX: Logger is infrastructure - components can import it directly

interface RecordingControllerProps {
  device: string;
  // ARCHITECTURE FIX: Removed service props - components only use stores
}

export const RecordingController: React.FC<RecordingControllerProps> = ({ 
  device 
}) => {
  const { activeRecordings, startRecording, stopRecording } = useRecordingStore();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const currentRecording = activeRecordings[device];

  const handleStartRecording = async () => {
    setLoading(true);
    setError(null);
    try {
      await startRecording(device);
      logger.info(`Recording started for device: ${device}`);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to start recording';
      setError(errorMsg);
      logger.error(`Failed to start recording for ${device}:`, { error: err });
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
      logger.error(`Failed to stop recording for ${device}:`, { error: err });
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
            Started: {currentRecording.startTime ? new Date(currentRecording.startTime).toLocaleString() : 'Unknown'}
          </Typography>
        </Box>
      )}
    </Box>
  );
};
