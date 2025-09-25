import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  MenuItem,
} from '@mui/material';

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
          onChange={(e) => setDuration(Number(e.target.value))}
          fullWidth
          margin="normal"
          inputProps={{ min: 1, max: 86400 }}
        />
        <TextField
          select
          label="Format"
          value={format}
          onChange={(e) => setFormat(e.target.value)}
          fullWidth
          margin="normal"
        >
          <MenuItem value="fmp4">fMP4 (default)</MenuItem>
          <MenuItem value="mp4">MP4</MenuItem>
          <MenuItem value="mkv">MKV</MenuItem>
        </TextField>
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
