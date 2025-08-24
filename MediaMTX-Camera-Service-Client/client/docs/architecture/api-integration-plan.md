# API Integration Planning Document

**Version:** 1.0  
**Date:** 2025-01-23  
**Status:** Planning Complete - Ready for Implementation  
**Scope:** Enhanced Recording Management API Integration  

---

## **Executive Summary**

This document outlines the comprehensive API integration plan to align the client with the new recording management ground truth requirements. The plan covers type definitions, error handling, storage monitoring, and configuration management integration.

### **Key Integration Areas**
1. **Enhanced Error Codes**: Support for -1006, -1008, -1010 error codes
2. **Storage Monitoring**: Integration with `get_storage_info` method
3. **Recording State Management**: Enhanced recording status and conflict handling
4. **Configuration Management**: Environment variable and dynamic configuration support
5. **Real-time Notifications**: Enhanced WebSocket notifications for recording and storage

---

## **1. Type Definitions Integration**

### **1.1 Enhanced Error Codes** ✅ **COMPLETED**
- **Added to `src/types/rpc.ts`:**
  - `CAMERA_ALREADY_RECORDING: -1006` - Recording conflict detection
  - `STORAGE_SPACE_LOW: -1008` - Storage warning threshold
  - `STORAGE_SPACE_CRITICAL: -1010` - Storage critical threshold

### **1.2 Enhanced Error Response Types** ✅ **COMPLETED**
- **Added to `src/types/rpc.ts`:**
  - `RecordingConflictErrorData` - Session information for conflicts
  - `StorageErrorData` - Storage usage information for errors
  - `EnhancedJSONRPCError` - Extended error response with specific data

### **1.3 Enhanced Notification Types** ✅ **COMPLETED**
- **Added to `src/types/rpc.ts`:**
  - `StorageStatusNotification` - Real-time storage status updates
  - Updated `NotificationMessage` union type

### **1.4 Enhanced Camera Types** ✅ **COMPLETED**
- **Added to `src/types/camera.ts`:**
  - Enhanced `CameraDevice` with recording status fields
  - `StorageInfo`, `StorageUsage`, `ThresholdStatus` types
  - `RecordingConflict`, `RecordingProgress` types
  - Configuration management types (`AppConfig`, `StorageConfig`, etc.)

### **1.5 Type Exports** ✅ **COMPLETED**
- **Updated `src/types/index.ts`:**
  - All new types exported for application-wide access
  - Maintained backward compatibility with existing exports

---

## **2. Service Integration Planning**

### **2.1 WebSocket Service Enhancements** 🔄 **PLANNED**

#### **Enhanced Error Handling**
```typescript
// Enhanced error handling for new error codes
handleEnhancedError(error: EnhancedJSONRPCError): void {
  switch (error.code) {
    case ERROR_CODES.CAMERA_ALREADY_RECORDING:
      this.handleRecordingConflict(error.data as RecordingConflictErrorData);
      break;
    case ERROR_CODES.STORAGE_SPACE_LOW:
      this.handleStorageWarning(error.data as StorageErrorData);
      break;
    case ERROR_CODES.STORAGE_SPACE_CRITICAL:
      this.handleStorageCritical(error.data as StorageErrorData);
      break;
    default:
      this.handleStandardError(error);
  }
}
```

#### **Enhanced Notification Handling**
```typescript
// Enhanced notification processing
handleStorageNotification(notification: StorageStatusNotification): void {
  const { total_space, used_space, available_space, usage_percent, threshold_status } = notification.params;
  
  // Update storage state
  this.updateStorageState({
    total_space,
    used_space,
    available_space,
    usage_percent,
    threshold_status
  });
  
  // Trigger UI updates
  this.notifyStorageStatusChange(threshold_status);
}
```

### **2.2 HTTP Polling Fallback Enhancements** 🔄 **PLANNED**

#### **Storage Information Support**
```typescript
// Add get_storage_info to HTTP fallback
async getStorageInfo(): Promise<StorageInfo> {
  const response = await this.httpClient.get('/api/storage/info');
  return this.validateStorageResponse(response);
}
```

#### **Enhanced Error Response Handling**
```typescript
// Enhanced error handling in HTTP fallback
handleHTTPError(response: HTTPResponse): void {
  if (response.status === 400 && response.data?.error) {
    const error = response.data.error as EnhancedJSONRPCError;
    this.handleEnhancedError(error);
  }
}
```

---

## **3. State Management Integration Planning**

### **3.1 Enhanced App State** 🔄 **PLANNED**

#### **Storage State Management**
```typescript
interface StorageState {
  info: StorageInfo | null;
  threshold_status: ThresholdStatus;
  warnings: string[];
  isMonitoring: boolean;
  lastUpdate: number;
}

interface RecordingState {
  activeRecordings: Map<string, RecordingSession>;
  conflicts: Map<string, RecordingConflict>;
  progress: Map<string, RecordingProgress>;
  isMonitoring: boolean;
}
```

#### **Configuration State Management**
```typescript
interface ConfigurationState {
  recording: RecordingConfig;
  storage: StorageConfig;
  environment: EnvironmentConfig;
  validation: ConfigValidationResult;
  isLoaded: boolean;
}
```

### **3.2 Store Integration** 🔄 **PLANNED**

#### **Storage Store**
```typescript
class StorageStore {
  // Storage monitoring
  startStorageMonitoring(): void;
  stopStorageMonitoring(): void;
  getStorageInfo(): Promise<StorageInfo>;
  
  // Threshold management
  checkStorageThresholds(): Promise<ThresholdStatus>;
  validateStorageForOperation(operation: string): Promise<StorageValidationResult>;
  
  // Event handling
  onStorageThresholdExceeded(callback: (status: ThresholdStatus) => void): void;
  onStorageCritical(callback: (status: StorageStatus) => void): void;
}
```

#### **Recording State Store**
```typescript
class RecordingStateStore {
  // Session management
  startRecording(device: string): Promise<void>;
  stopRecording(device: string): Promise<void>;
  isRecording(device: string): boolean;
  getRecordingSession(device: string): RecordingSession | null;
  
  // Conflict prevention
  canStartRecording(device: string): boolean;
  validateRecordingRequest(device: string): ValidationResult;
  
  // Real-time updates
  onRecordingStatusChange(callback: (status: RecordingStatus) => void): void;
  onRecordingConflict(callback: (conflict: RecordingConflict) => void): void;
}
```

#### **Configuration Store**
```typescript
class ConfigurationStore {
  // Environment variables
  loadEnvironmentConfig(): Promise<EnvironmentConfig>;
  validateConfiguration(): Promise<ConfigValidationResult>;
  
  // Dynamic updates
  updateConfiguration(config: Partial<AppConfig>): void;
  reloadConfiguration(): Promise<void>;
  
  // Default values
  getDefaultConfiguration(): AppConfig;
}
```

---

## **4. Component Integration Planning**

### **4.1 Enhanced Error Handling Components** 🔄 **PLANNED**

#### **Recording Conflict Handler**
```typescript
// Component for handling recording conflicts
const RecordingConflictHandler: React.FC<{ conflict: RecordingConflict }> = ({ conflict }) => {
  return (
    <Alert severity="warning">
      <Typography variant="h6">Recording Conflict Detected</Typography>
      <Typography>
        Camera {conflict.device} is currently recording (Session: {conflict.session_id})
      </Typography>
      <Typography>
        Current file: {conflict.current_file} (Elapsed: {formatDuration(conflict.elapsed_time)})
      </Typography>
      <Button onClick={() => stopRecording(conflict.device)}>
        Stop Current Recording
      </Button>
    </Alert>
  );
};
```

#### **Storage Warning Handler**
```typescript
// Component for handling storage warnings
const StorageWarningHandler: React.FC<{ storageInfo: StorageInfo }> = ({ storageInfo }) => {
  const getWarningMessage = () => {
    if (storageInfo.threshold_status === 'critical') {
      return 'Storage space is critical. Recording operations are blocked.';
    }
    return 'Storage space is low. Consider freeing up space.';
  };
  
  return (
    <Alert severity={storageInfo.threshold_status === 'critical' ? 'error' : 'warning'}>
      <Typography variant="h6">Storage Warning</Typography>
      <Typography>{getWarningMessage()}</Typography>
      <Typography>
        Available: {formatBytes(storageInfo.available_space)} / {formatBytes(storageInfo.total_space)} 
        ({storageInfo.usage_percent.toFixed(1)}% used)
      </Typography>
    </Alert>
  );
};
```

### **4.2 Storage Monitoring Components** 🔄 **PLANNED**

#### **Storage Monitor Component**
```typescript
// Real-time storage monitoring component
const StorageMonitor: React.FC = () => {
  const { storageInfo, thresholdStatus, isMonitoring } = useStorageStore();
  
  return (
    <Card>
      <CardHeader title="Storage Status" />
      <CardContent>
        <StorageUsageDisplay storageInfo={storageInfo} />
        <StorageThresholdIndicator thresholdStatus={thresholdStatus} />
        <StorageMonitoringControls isMonitoring={isMonitoring} />
      </CardContent>
    </Card>
  );
};
```

#### **Recording Manager Component**
```typescript
// Enhanced recording management component
const RecordingManager: React.FC = () => {
  const { activeRecordings, conflicts, progress } = useRecordingStateStore();
  
  return (
    <Card>
      <CardHeader title="Recording Management" />
      <CardContent>
        <ActiveRecordingsList recordings={activeRecordings} />
        <RecordingConflictsList conflicts={conflicts} />
        <RecordingProgressList progress={progress} />
      </CardContent>
    </Card>
  );
};
```

---

## **5. Configuration Management Planning**

### **5.1 Environment Variable Integration** 🔄 **PLANNED**

#### **Configuration Loading**
```typescript
// Load configuration from environment variables
const loadEnvironmentConfig = (): EnvironmentConfig => {
  return {
    RECORDING_ROTATION_MINUTES: process.env.RECORDING_ROTATION_MINUTES,
    STORAGE_WARN_PERCENT: process.env.STORAGE_WARN_PERCENT,
    STORAGE_BLOCK_PERCENT: process.env.STORAGE_BLOCK_PERCENT,
  };
};
```

#### **Configuration Validation**
```typescript
// Validate configuration values
const validateConfiguration = (config: AppConfig): ConfigValidationResult => {
  const errors: string[] = [];
  const warnings: string[] = [];
  
  // Validate storage thresholds
  if (config.storage.warn_percent >= config.storage.block_percent) {
    errors.push('Warning threshold must be less than block threshold');
  }
  
  // Validate recording rotation
  if (config.recording.rotation_minutes < 1 || config.recording.rotation_minutes > 1440) {
    errors.push('Recording rotation must be between 1 and 1440 minutes');
  }
  
  return {
    isValid: errors.length === 0,
    errors,
    warnings,
    config
  };
};
```

### **5.2 Dynamic Configuration Updates** 🔄 **PLANNED**

#### **Configuration Store Methods**
```typescript
// Update configuration dynamically
const updateConfiguration = (updates: Partial<AppConfig>): void => {
  const newConfig = { ...currentConfig, ...updates };
  const validation = validateConfiguration(newConfig);
  
  if (validation.isValid) {
    setConfiguration(newConfig);
    notifyConfigurationChange(newConfig);
  } else {
    notifyConfigurationError(validation.errors);
  }
};
```

---

## **6. Implementation Priority**

### **Phase 1: Core Type Integration** ✅ **COMPLETED**
- [x] Enhanced error codes and types
- [x] Storage information types
- [x] Recording state management types
- [x] Configuration management types
- [x] Type exports and compatibility

### **Phase 2: Service Layer Integration** 🔄 **NEXT**
- [ ] Enhanced WebSocket service error handling
- [ ] HTTP polling fallback for storage methods
- [ ] Enhanced notification processing
- [ ] Service integration testing

### **Phase 3: State Management Integration** 🔄 **PLANNED**
- [ ] Storage store implementation
- [ ] Recording state store implementation
- [ ] Configuration store implementation
- [ ] Store integration and testing

### **Phase 4: Component Integration** 🔄 **PLANNED**
- [ ] Enhanced error handling components
- [ ] Storage monitoring components
- [ ] Recording management components
- [ ] Configuration management components

### **Phase 5: Integration Testing** 🔄 **PLANNED**
- [ ] End-to-end API integration testing
- [ ] Error handling validation
- [ ] Storage monitoring validation
- [ ] Configuration management validation

---

## **7. Testing Strategy**

### **7.1 API Integration Testing**
- **Error Code Testing**: Validate all new error codes (-1006, -1008, -1010)
- **Storage Method Testing**: Test `get_storage_info` integration
- **Notification Testing**: Validate enhanced WebSocket notifications
- **Configuration Testing**: Test environment variable loading and validation

### **7.2 Error Handling Testing**
- **Recording Conflict Scenarios**: Test conflict detection and resolution
- **Storage Warning Scenarios**: Test threshold-based warnings and blocking
- **Configuration Error Scenarios**: Test invalid configuration handling

### **7.3 Component Testing**
- **Storage Monitor**: Test real-time storage monitoring
- **Recording Manager**: Test recording state management
- **Error Handlers**: Test user-friendly error display

---

## **8. Risk Assessment**

### **8.1 Technical Risks**
- **Type Compatibility**: Risk of breaking existing functionality during type updates
- **Service Integration**: Risk of service layer complexity increase
- **Performance Impact**: Risk of monitoring overhead affecting performance

### **8.2 Mitigation Strategies**
- **Incremental Implementation**: Implement changes in phases with testing
- **Backward Compatibility**: Maintain compatibility with existing functionality
- **Performance Monitoring**: Monitor and optimize performance impact

---

## **9. Success Criteria**

### **9.1 Functional Requirements**
- [ ] All new error codes properly handled and displayed
- [ ] Storage monitoring fully integrated and functional
- [ ] Recording state management working correctly
- [ ] Configuration management operational

### **9.2 Non-Functional Requirements**
- [ ] No performance degradation from new features
- [ ] All existing functionality remains intact
- [ ] User experience improvements for error handling
- [ ] Configuration flexibility and validation

---

**Document Status**: Planning Complete - Ready for Implementation  
**Next Actions**: Begin Phase 2 Service Layer Integration  
**Ground Truth**: Aligned with server recording management requirements
