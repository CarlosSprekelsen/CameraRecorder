import React, { useState } from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions } from '../atoms/Dialog/Dialog';
import { Button } from '../atoms/Button/Button';
import { TextField } from '../atoms/TextField/TextField';

interface TimedRecordDialogProps {
  open: boolean;
  onCancel: () => void;
  onStart: (duration: number, format: string) => void;
}

const TimedRecordDialog: React.FC<TimedRecordDialogProps> = ({ open, onCancel, onStart }) => {
  const [duration, setDuration] = useState<number>(60);
  const [format, setFormat] = useState<string>('fmp4');

  return (
    <Dialog open={open} onClose={onCancel} fullWidth>
      <DialogTitle>Timed Recording</DialogTitle>
      <DialogContent>
        <TextField
          label="Duration (seconds)"
          type="number"
          value={duration}
          onChange={(value) => setDuration(Number(value))}
          fullWidth
          className="mb-4"
          min={1}
          max={86400}
        />
        <TextField
          label="Format"
          value={format}
          onChange={(value) => setFormat(value)}
          fullWidth
          className="mb-4"
          options={[
            { value: 'fmp4', label: 'fMP4 (default)' },
            { value: 'mp4', label: 'MP4' },
            { value: 'mkv', label: 'MKV' }
          ]}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <Button onClick={() => onStart(duration, format)} variant="contained">
          Start
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default TimedRecordDialog;
