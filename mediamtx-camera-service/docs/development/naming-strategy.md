# Naming Strategy Guide

**Version:** 1.0  
**Date:** 2025-01-23  
**Purpose:** Prevent naming conflicts and ensure consistent naming conventions  
**Status:** Approved  
**Related Documents:** `coding-standards.md`, `documentation-guidelines.md`

---

## 1. Overview

This document establishes systematic naming conventions to prevent duplicates and ensure code clarity across the MediaMTX Camera Service project. It applies to all code, documentation, and configuration files.

### Key Principles
- **Uniqueness:** Every identifier must be unique within its scope
- **Clarity:** Names must clearly indicate purpose and scope
- **Consistency:** Follow established patterns across the project
- **Hierarchy:** Use prefixes/suffixes to indicate scope and purpose

---

## 2. State Management Naming Strategy

### 2.1 Component State vs Store State

#### Local Component State
```typescript
// Use descriptive prefixes for local state
const [localSelectedFile, setLocalSelectedFile] = useState<FileItem | null>(null);
const [localFormData, setLocalFormData] = useState<FormData>({});
const [localSettings, setLocalSettings] = useState<Settings>({});
const [localError, setLocalError] = useState<string | null>(null);
```

#### Store State (Zustand/Global)
```typescript
// Use store-specific naming for global state
const {
  selectedFile: storeSelectedFile,
  setSelectedFile: setStoreSelectedFile,
  formData: storeFormData,
  updateFormData: updateStoreFormData
} = useStore();
```

#### Service State (API/External)
```typescript
// Use service prefixes for external state
const [serviceLoading, setServiceLoading] = useState(false);
const [serviceError, setServiceError] = useState<string | null>(null);
const [serviceData, setServiceData] = useState<ApiResponse | null>(null);
```

### 2.2 Function Naming Convention

#### Local Handlers
```typescript
// Component-specific event handlers
const handleLocalDelete = (file: FileItem) => { /* ... */ };
const handleLocalSelect = (file: FileItem) => { /* ... */ };
const handleLocalUpdate = (data: FormData) => { /* ... */ };
const handleLocalSubmit = (event: FormEvent) => { /* ... */ };
```

#### Store Actions
```typescript
// Global state actions
const handleStoreDelete = (file: FileItem) => { /* ... */ };
const dispatchStoreUpdate = (data: FormData) => { /* ... */ };
const executeStoreAction = (action: StoreAction) => { /* ... */ };
```

#### Service Calls
```typescript
// External API calls
const callServiceDelete = async (file: FileItem) => { /* ... */ };
const fetchServiceData = async (params: ApiParams) => { /* ... */ };
const submitServiceForm = async (data: FormData) => { /* ... */ };
```

---

## 3. Variable Naming Strategy

### 3.1 Scope-Based Prefixes

| Scope | Prefix | Example |
|-------|--------|---------|
| Local Component | `local` | `localSelectedFile` |
| Form State | `form` | `formData`, `formErrors` |
| Store State | `store` | `storeSelectedFile` |
| Service State | `service` | `serviceLoading` |
| UI State | `ui` | `uiModalOpen`, `uiLoading` |
| Temporary | `temp` | `tempData`, `tempFile` |

### 3.2 Type-Based Suffixes

| Type | Suffix | Example |
|------|--------|---------|
| Interface | `Interface` | `FileItemInterface` |
| Type | `Type` | `ApiResponseType` |
| Enum | `Enum` | `StatusEnum` |
| Constant | `CONST` | `API_ENDPOINTS_CONST` |
| Configuration | `Config` | `WebSocketConfig` |

### 3.3 Context-Based Naming

```typescript
// Clear context separation
interface FileManagerLocalState {
  selectedFile: FileItem | null;
  formData: FileFormData;
  errors: FileFormErrors;
}

interface FileManagerStoreState {
  selectedFile: FileItem | null;
  files: FileItem[];
  loading: boolean;
}

interface FileManagerServiceState {
  apiLoading: boolean;
  apiError: string | null;
  apiData: ApiResponse | null;
}
```

---

## 4. File Naming Strategy

### 4.1 Component Files
```
ComponentName.tsx              // Main component
ComponentName.module.css       // Component styles
ComponentName.test.tsx         // Component tests
ComponentName.stories.tsx      // Storybook stories
```

### 4.2 Store Files
```
componentNameStore.ts          // Zustand store
componentNameStore.test.ts     // Store tests
componentNameStore.types.ts    // Store types
```

### 4.3 Service Files
```
componentNameService.ts        // API service
componentNameService.test.ts   // Service tests
componentNameService.types.ts  // Service types
```

### 4.4 Type Files
```
componentNameTypes.ts          // Component types
apiTypes.ts                   // API types
commonTypes.ts                // Shared types
```

---

## 5. Import Aliasing Strategy

### 5.1 Conflict Resolution
```typescript
// Always alias conflicting imports
import { setSelectedFile as setStoreSelectedFile } from '../stores/fileStore';
import { FileItem as StoreFileItem } from '../types/store';
import { FileItem as UIFileItem } from '../types/ui';
import { FileItem as ApiFileItem } from '../types/api';
```

### 5.2 Namespace Aliasing
```typescript
// Use namespace aliases for complex imports
import * as FileStore from '../stores/fileStore';
import * as FileTypes from '../types/file';
import * as FileUtils from '../utils/file';
```

### 5.3 Default Import Aliasing
```typescript
// Alias default imports for clarity
import FileManagerComponent from './FileManager';
import FileManagerStore from '../stores/fileStore';
import FileManagerService from '../services/fileService';
```

---

## 6. Interface and Type Naming

### 6.1 Interface Naming Convention
```typescript
// Use descriptive prefixes for interfaces
interface LocalComponentState {
  selectedFile: FileItem | null;
  formData: FormData;
}

interface StoreState {
  selectedFile: FileItem | null;
  files: FileItem[];
}

interface ServiceState {
  loading: boolean;
  error: string | null;
  data: ApiResponse | null;
}

interface ApiResponse {
  success: boolean;
  data: any;
  error?: string;
}
```

### 6.2 Type Naming Convention
```typescript
// Use descriptive suffixes for types
type FileItemType = {
  id: string;
  name: string;
  size: number;
};

type ApiResponseType = {
  success: boolean;
  data: any;
  error?: string;
};

type FormDataType = {
  name: string;
  description: string;
};
```

---

## 7. Constant Naming Strategy

### 7.1 Global Constants
```typescript
// Use UPPERCASE with descriptive names
const API_ENDPOINTS = {
  WEBSOCKET: 'ws://localhost:8002/ws',
  HEALTH: 'http://localhost:8003',
  FILES: 'http://localhost:8003/files'
};

const RPC_METHODS = {
  GET_CAMERAS: 'get_cameras',
  GET_STREAMS: 'get_streams',
  START_RECORDING: 'start_recording'
};
```

### 7.2 Module Constants
```typescript
// Use camelCase for module-specific constants
const defaultSettings = {
  autoRefresh: true,
  refreshInterval: 5000
};

const fileTypes = {
  RECORDING: 'recording',
  SNAPSHOT: 'snapshot'
};
```

---

## 8. Error Handling Naming

### 8.1 Error State Variables
```typescript
// Use descriptive error naming
const [localError, setLocalError] = useState<string | null>(null);
const [storeError, setStoreError] = useState<string | null>(null);
const [serviceError, setServiceError] = useState<string | null>(null);
const [validationError, setValidationError] = useState<string | null>(null);
```

### 8.2 Error Handler Functions
```typescript
// Use descriptive error handler naming
const handleLocalError = (error: Error) => { /* ... */ };
const handleStoreError = (error: StoreError) => { /* ... */ };
const handleServiceError = (error: ApiError) => { /* ... */ };
const handleValidationError = (error: ValidationError) => { /* ... */ };
```

---

## 9. Testing Naming Strategy

### 9.1 Test File Naming
```
ComponentName.test.tsx         // Component tests
componentNameStore.test.ts     // Store tests
componentNameService.test.ts   // Service tests
componentNameUtils.test.ts     // Utility tests
```

### 9.2 Test Function Naming
```typescript
// Use descriptive test function names
describe('FileManager Component', () => {
  it('should render file list correctly', () => { /* ... */ });
  it('should handle file selection', () => { /* ... */ });
  it('should handle file deletion', () => { /* ... */ });
  it('should show error on API failure', () => { /* ... */ });
});
```

### 9.3 Test Variable Naming
```typescript
// Use descriptive test variable names
const mockFileData = { id: '1', name: 'test.mp4' };
const mockApiResponse = { success: true, data: mockFileData };
const mockErrorResponse = { success: false, error: 'Not found' };
```

---

## 10. Documentation Naming Strategy

### 10.1 Document File Naming
```
component-name-guide.md        // Component guides
api-reference.md              // API documentation
architecture-overview.md      // Architecture docs
deployment-guide.md           // Deployment docs
```

### 10.2 Section Naming
```markdown
## Component Overview
## API Reference
## Architecture
## Deployment
## Troubleshooting
```

---

## 11. Configuration Naming Strategy

### 11.1 Environment Variables
```bash
# Use descriptive environment variable names
CAMERA_SERVICE_WEBSOCKET_URL=ws://localhost:8002/ws
CAMERA_SERVICE_HEALTH_URL=http://localhost:8003
CAMERA_SERVICE_LOG_LEVEL=INFO
CAMERA_SERVICE_MAX_CONNECTIONS=50
```

### 11.2 Configuration Objects
```typescript
// Use descriptive configuration object names
const websocketConfig = {
  url: 'ws://localhost:8002/ws',
  reconnectInterval: 5000,
  maxReconnectAttempts: 10
};

const apiConfig = {
  baseUrl: 'http://localhost:8003',
  timeout: 30000,
  retryAttempts: 3
};
```

---

## 12. Validation and Enforcement

### 12.1 Linting Rules
```json
{
  "rules": {
    "naming-convention": [
      "error",
      {
        "selector": "variable",
        "format": ["camelCase", "UPPER_CASE"],
        "prefix": ["local", "store", "service", "ui", "temp"]
      }
    ]
  }
}
```

### 12.2 Code Review Checklist
- [ ] All variables have unique names within their scope
- [ ] Local state uses appropriate prefixes
- [ ] Store state uses store-specific naming
- [ ] Service state uses service-specific naming
- [ ] Import conflicts are resolved with aliases
- [ ] File names follow established conventions
- [ ] Interface and type names are descriptive
- [ ] Constants use appropriate naming conventions

---

## 13. Examples

### 13.1 Complete Component Example
```typescript
// FileManager.tsx
import { useState } from 'react';
import { useFileStore } from '../stores/fileStore';
import { useFileService } from '../services/fileService';

interface FileManagerLocalState {
  selectedFile: FileItem | null;
  formData: FileFormData;
  localError: string | null;
}

export const FileManager = () => {
  // Local state with prefixes
  const [localSelectedFile, setLocalSelectedFile] = useState<FileItem | null>(null);
  const [localFormData, setLocalFormData] = useState<FileFormData>({});
  const [localError, setLocalError] = useState<string | null>(null);

  // Store state with aliases
  const {
    selectedFile: storeSelectedFile,
    setSelectedFile: setStoreSelectedFile,
    files: storeFiles
  } = useFileStore();

  // Service state with prefixes
  const {
    loading: serviceLoading,
    error: serviceError,
    data: serviceData
  } = useFileService();

  // Local handlers
  const handleLocalSelect = (file: FileItem) => {
    setLocalSelectedFile(file);
    setStoreSelectedFile(file);
  };

  const handleLocalDelete = async (file: FileItem) => {
    try {
      await callServiceDelete(file);
      setLocalError(null);
    } catch (error) {
      setLocalError(error.message);
    }
  };

  return (
    <div>
      {/* Component JSX */}
    </div>
  );
};
```

---

## 14. Maintenance and Updates

### 14.1 Review Schedule
- **Weekly:** Review naming conventions in new code
- **Monthly:** Update naming strategy based on project evolution
- **Per Release:** Validate naming consistency across the project

### 14.2 Update Triggers
- **New Patterns:** Add new naming patterns as they emerge
- **Conflicts:** Resolve naming conflicts and update guidelines
- **Technology Changes:** Update naming strategy for new technologies
- **Team Feedback:** Incorporate team feedback on naming conventions

---

**Naming Strategy Status: ✅ ESTABLISHED**

This naming strategy provides a comprehensive framework for preventing duplicates and ensuring consistent naming conventions across the MediaMTX Camera Service project. All team members should follow these guidelines to maintain code clarity and prevent naming conflicts. 