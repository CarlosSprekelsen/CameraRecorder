/**
 * Camera Card Molecule - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Combines atoms to create camera display functionality
 */

import React from 'react';
import { Button } from '../../atoms/Button/Button';
import { Camera } from '../../../stores/device/deviceStore';

export interface CameraCardProps {
  camera: Camera;
  onSelect: (deviceId: string) => void;
  onStartRecording: (deviceId: string) => void;
  onStopRecording: (deviceId: string) => void;
  onTakeSnapshot: (deviceId: string) => void;
  isSelected?: boolean;
  isRecording?: boolean;
}

export const CameraCard: React.FC<CameraCardProps> = ({
  camera,
  onSelect,
  onStartRecording,
  onStopRecording,
  onTakeSnapshot,
  isSelected = false,
  isRecording = false,
}) => {
  const handleCardClick = () => {
    onSelect(camera.device);
  };

  const handleRecordingClick = () => {
    if (isRecording) {
      onStopRecording(camera.device);
    } else {
      onStartRecording(camera.device);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'CONNECTED':
        return 'text-green-600 bg-green-100';
      case 'DISCONNECTED':
        return 'text-gray-600 bg-gray-100';
      case 'ERROR':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };

  return (
    <div
      className={`border rounded-lg p-4 cursor-pointer transition-all ${
        isSelected ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'
      }`}
      onClick={handleCardClick}
    >
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-lg font-semibold text-gray-900">{camera.name || camera.device}</h3>
        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(camera.status)}`}>
          {camera.status}
        </span>
      </div>

      {camera.resolution && (
        <p className="text-sm text-gray-600 mb-2">Resolution: {camera.resolution}</p>
      )}

      {camera.fps && (
        <p className="text-sm text-gray-600 mb-3">FPS: {camera.fps}</p>
      )}

      <div className="flex space-x-2">
        <Button
          variant={isRecording ? 'danger' : 'primary'}
          size="small"
          onClick={(e) => {
            e.stopPropagation();
            handleRecordingClick();
          }}
        >
          {isRecording ? 'Stop Recording' : 'Start Recording'}
        </Button>

        <Button
          variant="secondary"
          size="small"
          onClick={(e) => {
            e.stopPropagation();
            onTakeSnapshot(camera.device);
          }}
        >
          Take Snapshot
        </Button>
      </div>

      {camera.streams && (
        <div className="mt-3 pt-3 border-t border-gray-200">
          <p className="text-xs text-gray-500 mb-1">Stream URLs:</p>
          <div className="space-y-1">
            {camera.streams.rtsp && (
              <p className="text-xs text-blue-600 break-all">RTSP: {camera.streams.rtsp}</p>
            )}
            {camera.streams.hls && (
              <p className="text-xs text-blue-600 break-all">HLS: {camera.streams.hls}</p>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
