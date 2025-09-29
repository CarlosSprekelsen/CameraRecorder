/**
 * @fileoverview AboutPage component for server information display
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React from 'react';
import { Grid } from '../../components/atoms/Grid/Grid';
import { Card } from '../../components/atoms/Card/Card';
import { Alert } from '../../components/atoms/Alert/Alert';
import {
  Info as InfoIcon,
  Storage as StorageIcon,
  MonitorHeart as HealthIcon,
} from '@mui/icons-material';
import { useServerStore } from '../../stores/server/serverStore';

const AboutPage: React.FC = () => {
  const { info: serverInfo, status, storage, loading, error, loadServerInfo } = useServerStore();

  React.useEffect(() => {
    loadServerInfo();
  }, [loadServerInfo]);

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="error" className="m-4">
        Failed to load server information: {error}
      </Alert>
    );
  }

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold mb-4">
        Server Information
      </h1>

      <Grid container spacing={3}>
        {/* Server Info */}
        <Grid item xs={12} md={6}>
          <Card>
            <div className="p-4">
              <div className="flex items-center mb-4">
                <InfoIcon className="mr-2 text-blue-600" />
                <h2 className="text-xl font-semibold">Server Details</h2>
              </div>

              {serverInfo ? (
                <div className="space-y-2">
                  <div>
                    <span className="text-sm text-gray-600">Version:</span>
                    <span className="ml-2 font-medium">{serverInfo.version}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">Name:</span>
                    <span className="ml-2 font-medium">{serverInfo.name}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">Architecture:</span>
                    <span className="ml-2 font-medium">{serverInfo.architecture}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">Go Version:</span>
                    <span className="ml-2 font-medium">{serverInfo.go_version}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">Build Date:</span>
                    <span className="ml-2 font-medium">{serverInfo.build_date}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">Architecture:</span>
                    <span className="ml-2 font-medium">{serverInfo.architecture}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">Max Cameras:</span>
                    <span className="ml-2 font-medium">{serverInfo.max_cameras}</span>
                  </div>

                  <div className="mt-4">
                    <span className="text-sm text-gray-600">Capabilities:</span>
                    <div className="flex flex-wrap gap-1 mt-1">
                      {serverInfo.capabilities?.map((capability: string, index: number) => (
                        <span key={index} className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded">
                          {capability}
                        </span>
                      ))}
                    </div>
                  </div>

                  <div className="mt-4">
                    <span className="text-sm text-gray-600">Supported Formats:</span>
                    <div className="flex flex-wrap gap-1 mt-1">
                      {serverInfo.supported_formats?.map((format: string, index: number) => (
                        <span key={index} className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded">
                          {format}
                        </span>
                      ))}
                    </div>
                  </div>
                </div>
              ) : (
                <p className="text-gray-500">No server information available</p>
              )}
            </div>
          </Card>
        </Grid>

        {/* System Status */}
        <Grid item xs={12} md={6}>
          <Card>
            <div className="p-4">
              <div className="flex items-center mb-4">
                <HealthIcon className="mr-2 text-green-600" />
                <h2 className="text-xl font-semibold">System Status</h2>
              </div>

              {status ? (
                <div className="space-y-2">
                  <div className="flex items-center mb-2">
                    <span className="text-sm text-gray-600 mr-2">Status:</span>
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      status.status === 'HEALTHY' ? 'bg-green-100 text-green-800' :
                      status.status === 'DEGRADED' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-red-100 text-red-800'
                    }`}>
                      {status.status}
                    </span>
                  </div>

                  <div>
                    <span className="text-sm text-gray-600">Uptime:</span>
                    <span className="ml-2 font-medium">{status.uptime}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">Version:</span>
                    <span className="ml-2 font-medium">{status.version}</span>
                  </div>

                  <div className="mt-4">
                    <span className="text-sm text-gray-600">Components:</span>
                    <div className="mt-2 space-y-1">
                      <div className="flex justify-between">
                        <span className="text-sm">WebSocket Server:</span>
                        <span className={`text-xs px-2 py-1 rounded ${
                          status.components.websocket_server === 'RUNNING' ? 'bg-green-100 text-green-800' :
                          status.components.websocket_server === 'ERROR' ? 'bg-red-100 text-red-800' :
                          'bg-yellow-100 text-yellow-800'
                        }`}>
                          {status.components.websocket_server}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-sm">Camera Monitor:</span>
                        <span className={`text-xs px-2 py-1 rounded ${
                          status.components.camera_monitor === 'RUNNING' ? 'bg-green-100 text-green-800' :
                          status.components.camera_monitor === 'ERROR' ? 'bg-red-100 text-red-800' :
                          'bg-yellow-100 text-yellow-800'
                        }`}>
                          {status.components.camera_monitor}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-sm">MediaMTX:</span>
                        <span className={`text-xs px-2 py-1 rounded ${
                          status.components.mediamtx === 'RUNNING' ? 'bg-green-100 text-green-800' :
                          status.components.mediamtx === 'ERROR' ? 'bg-red-100 text-red-800' :
                          'bg-yellow-100 text-yellow-800'
                        }`}>
                          {status.components.mediamtx}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <p className="text-gray-500">No status information available</p>
              )}
            </div>
          </Card>
        </Grid>

        {/* Storage Info */}
        {storage && (
          <Grid item xs={12}>
            <Card>
              <div className="p-4">
                <div className="flex items-center mb-4">
                  <StorageIcon className="mr-2 text-blue-600" />
                  <h2 className="text-xl font-semibold">Storage Information</h2>
                </div>

                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6} md={3}>
                    <div className="text-center p-3 bg-gray-50 rounded">
                      <div className="text-2xl font-bold text-blue-600">{storage.total_space}</div>
                      <div className="text-sm text-gray-600">Total Space</div>
                    </div>
                  </Grid>
                  <Grid item xs={12} sm={6} md={3}>
                    <div className="text-center p-3 bg-gray-50 rounded">
                      <div className="text-2xl font-bold text-green-600">{storage.used_space}</div>
                      <div className="text-sm text-gray-600">Used Space</div>
                    </div>
                  </Grid>
                  <Grid item xs={12} sm={6} md={3}>
                    <div className="text-center p-3 bg-gray-50 rounded">
                      <div className="text-2xl font-bold text-yellow-600">{storage.available_space}</div>
                      <div className="text-sm text-gray-600">Available Space</div>
                    </div>
                  </Grid>
                  <Grid item xs={12} sm={6} md={3}>
                    <div className="text-center p-3 bg-gray-50 rounded">
                      <div className="text-2xl font-bold text-purple-600">{storage.usage_percentage}%</div>
                      <div className="text-sm text-gray-600">Usage</div>
                    </div>
                  </Grid>
                </Grid>

                <div className="mt-4 grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <span className="text-sm text-gray-600">Recordings Size:</span>
                    <span className="ml-2 font-medium">{storage.recordings_size}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">Snapshots Size:</span>
                    <span className="ml-2 font-medium">{storage.snapshots_size}</span>
                  </div>
                </div>

                {storage.low_space_warning && (
                  <Alert variant="warning" className="mt-4">
                    Low storage space warning is active
                  </Alert>
                )}
              </div>
            </Card>
          </Grid>
        )}
      </Grid>
    </div>
  );
};

export default AboutPage;