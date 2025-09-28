/**
 * Camera Manager Organism - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Combines molecules to create complete camera management functionality
 */

import React, { useEffect } from 'react';
import { CameraCard } from '../../molecules/CameraCard/CameraCard';
import { Button } from '../../atoms/Button/Button';
import { useDeviceStore } from '../../../stores/device/deviceStore';
import { useRecordingStore } from '../../../stores/recording/recordingStore';
import { logger } from '../../../services/logger/LoggerService';
// ARCHITECTURE FIX: Logger is infrastructure - components can import it directly
// ARCHITECTURE FIX: Removed DeviceActions import - this doesn't exist

export interface CameraManagerProps {
  // ARCHITECTURE FIX: Removed service props - components only use stores
}

export const CameraManager: React.FC<CameraManagerProps> = () => {
  const {
    cameras,
    loading: deviceLoading,
    error: deviceError,
    getCameraList,
  } = useDeviceStore();
  
  const { activeRecordings } = useRecordingStore();

  useEffect(() => {
    // Load cameras on component mount
    getCameraList();
  }, [getCameraList]);

  const handleStartRecording = (deviceId: string) => {
    // Business logic moved to actions
    logger.info(`Starting recording for device: ${deviceId}`);
    // TODO: Implement recording actions
  };

  const handleStopRecording = (deviceId: string) => {
    // Business logic moved to actions
    logger.info(`Stopping recording for device: ${deviceId}`);
    // TODO: Implement recording actions
  };

  const handleTakeSnapshot = (deviceId: string) => {
    // Business logic moved to actions
    logger.info(`Taking snapshot for device: ${deviceId}`);
    // TODO: Implement snapshot actions
  };

  const handleRefresh = () => {
    getCameraList();
  };

  if (deviceLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading cameras...</p>
        </div>
      </div>
    );
  }

  if (deviceError) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <p className="text-red-600 mb-4">{deviceError}</p>
          <Button onClick={handleRefresh} variant="primary">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Camera Management</h2>
        <Button onClick={handleRefresh} variant="secondary">
          Refresh
        </Button>
      </div>

      {cameras.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-500 text-lg">No cameras found</p>
          <p className="text-gray-400 text-sm mt-2">Connect a camera to get started</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {cameras.map((camera) => (
            <CameraCard
              key={camera.device}
              camera={camera}
              onSelect={() => {}} // TODO: Implement camera selection
              onStartRecording={handleStartRecording}
              onStopRecording={handleStopRecording}
              onTakeSnapshot={handleTakeSnapshot}
              isSelected={false} // TODO: Implement camera selection state
              isRecording={!!activeRecordings[camera.device]}
            />
          ))}
        </div>
      )}

      <div className="mt-8 p-4 bg-gray-50 rounded-lg">
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Camera Status Summary</h3>
        <div className="grid grid-cols-3 gap-4 text-center">
          <div>
            <p className="text-2xl font-bold text-green-600">
              {cameras.filter(c => c.status === 'CONNECTED').length}
            </p>
            <p className="text-sm text-gray-600">Connected</p>
          </div>
          <div>
            <p className="text-2xl font-bold text-gray-600">
              {cameras.filter(c => c.status === 'DISCONNECTED').length}
            </p>
            <p className="text-sm text-gray-600">Disconnected</p>
          </div>
          <div>
            <p className="text-2xl font-bold text-red-600">
              {cameras.filter(c => c.status === 'ERROR').length}
            </p>
            <p className="text-sm text-gray-600">Error</p>
          </div>
        </div>
      </div>
    </div>
  );
};
