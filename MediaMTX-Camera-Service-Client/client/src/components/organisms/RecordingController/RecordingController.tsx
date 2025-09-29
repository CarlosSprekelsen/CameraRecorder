/**
 * RecordingController - Architecture Compliance
 * 
 * Architecture requirement: "RecordingController component" (Section 5.2)
 * Provides recording control functionality following unidirectional data flow
 */

import React, { useState } from 'react';
import { Button } from '../../atoms/Button/Button';
import { Alert } from '../../atoms/Alert/Alert';
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
    <div className="p-4 border border-gray-300 rounded-lg">
      <h6 className="text-lg font-semibold mb-4">
        Recording Control - {device}
      </h6>
      
      {error && (
        <Alert variant="error" className="mb-4">
          {error}
        </Alert>
      )}

      <div className="flex gap-4 items-center">
        {!isRecording ? (
          <Button
            variant="primary"
            onClick={handleStartRecording}
            disabled={loading}
            loading={loading}
            className="min-w-[140px] flex items-center gap-2"
          >
            <PlayArrow className="h-4 w-4" />
            Start Recording
          </Button>
        ) : (
          <Button
            variant="danger"
            onClick={handleStopRecording}
            disabled={loading}
            loading={loading}
            className="min-w-[140px] flex items-center gap-2"
          >
            <Stop className="h-4 w-4" />
            Stop Recording
          </Button>
        )}
      </div>

      {currentRecording && (
        <div className="mt-4">
          <p className="text-sm text-gray-600">
            Status: {currentRecording.status}
          </p>
          <p className="text-sm text-gray-600">
            Started: {currentRecording.startTime ? new Date(currentRecording.startTime).toLocaleString() : 'Unknown'}
          </p>
        </div>
      )}
    </div>
  );
};
