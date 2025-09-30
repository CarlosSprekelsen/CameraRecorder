import React from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions } from '../atoms/Dialog/Dialog';
import { Button } from '../atoms/Button/Button';
import { Typography } from '../atoms/Typography/Typography';
import { Alert } from '../atoms/Alert/Alert';

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmText?: string;
  cancelText?: string;
  severity?: 'error' | 'warning' | 'info';
}

/**
 * ConfirmDialog - Confirmation dialog for destructive actions
 * Implements architecture requirement for delete confirmations (section 13.4)
 */
const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  title,
  message,
  onConfirm,
  onCancel,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  severity = 'warning',
}) => {
  return (
    <Dialog open={open} onClose={onCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <Alert severity={severity} className="mb-2">
          {message}
        </Alert>
        <Typography variant="body2" color="secondary">
          This action cannot be undone. Please make sure you want to proceed.
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel} variant="secondary">
          {cancelText}
        </Button>
        <Button
          onClick={onConfirm}
          variant={severity === 'error' ? 'danger' : 'primary'}
        >
          {confirmText}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default ConfirmDialog;
