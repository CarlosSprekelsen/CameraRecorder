# Naming Strategy Documentation

**Document Version:** 1.0  
**Date:** January 2025  
**Classification:** Development Reference  
**Status:** Active Implementation

---

## 1. Overview

This document establishes consistent naming conventions across the MediaMTX Camera Service Client codebase. All naming patterns are designed to enhance code readability, maintainability, and developer experience.

## 2. Store Naming Convention

### Pattern: `use[Domain]Store`

**Purpose:** Zustand store hooks for state management  
**Format:** `use` + `[Domain]` + `Store`  
**Examples:**
- `useAuthStore` - Authentication state management
- `useConnectionStore` - WebSocket connection state
- `useDeviceStore` - Camera device state management
- `useFileStore` - File management state
- `useRecordingStore` - Recording operations state
- `useServerStore` - Server information state

**Status:** ✅ **100% Compliant** - All stores follow this pattern

**Implementation Example:**
```typescript
export const useAuthStore = create<AuthStore>((set) => ({
  token: null,
  isAuthenticated: false,
  // ... actions
}));
```

## 3. Service Naming Convention

### Pattern: `[Domain]Service`

**Purpose:** Service layer classes for business logic  
**Format:** `[Domain]` + `Service`  
**Examples:**
- `AuthService` - Authentication operations
- `WebSocketService` - WebSocket communication
- `DeviceService` - Camera device operations
- `RecordingService` - Recording operations
- `LoggerService` - Logging functionality
- `ServerService` - Server status operations
- `NotificationService` - Event notifications
- `FileService` - File management operations

**Status:** ✅ **100% Compliant** - All services follow this pattern

**Implementation Example:**
```typescript
export class AuthService {
  constructor(private wsService: WebSocketService) {}
  
  async authenticate(token: string): Promise<AuthResult> {
    // Implementation
  }
}
```

## 4. Component Naming Convention

### Pattern: `[Purpose][Type]`

**Purpose:** React components for UI elements  
**Format:** `[Purpose]` + `[Type]`  
**Examples:**
- `CameraPage` - Main camera management page
- `FilesPage` - File management page
- `LoginPage` - Authentication page
- `AboutPage` - Server information page
- `CameraTable` - Device listing table
- `FileTable` - File listing table
- `CopyLinkButton` - Stream URL copy button
- `Pagination` - Page navigation component

**Status:** ✅ **100% Compliant** - All components follow this pattern

**Implementation Example:**
```typescript
const CameraPage: React.FC = memo(() => {
  // Component implementation
});

CameraPage.displayName = 'CameraPage';
```

## 5. Directory Structure Convention

### Pattern: `[domain]/[purpose]`

**Purpose:** Organized file structure for maintainability  
**Format:** `[domain]/[purpose]`  
**Examples:**
- `stores/auth/` - Authentication state management
- `services/websocket/` - WebSocket communication
- `components/cameras/` - Camera-related components
- `pages/login/` - Login page components
- `hooks/` - Custom React hooks
- `types/` - TypeScript type definitions

**Status:** ✅ **100% Compliant** - All directories follow this pattern

## 6. Interface Naming Convention

### Pattern: `I[Purpose]`

**Purpose:** TypeScript interfaces for contracts  
**Format:** `I` + `[Purpose]`  
**Examples:**
- `ICommand` - Command operations interface
- `IDiscovery` - Device discovery interface
- `IStatus` - Status operations interface
- `IFileCatalog` - File listing interface
- `IFileActions` - File operations interface

**Status:** ✅ **100% Compliant** - All interfaces follow this pattern

## 7. Hook Naming Convention

### Pattern: `use[Purpose]`

**Purpose:** Custom React hooks for reusable logic  
**Format:** `use` + `[Purpose]`  
**Examples:**
- `useKeyboardShortcuts` - Keyboard navigation
- `usePerformanceMonitor` - Performance tracking
- `usePermissions` - Role-based access control

**Status:** ✅ **100% Compliant** - All hooks follow this pattern

## 8. Type Naming Convention

### Pattern: `[Purpose]Result` or `[Purpose]State`

**Purpose:** TypeScript types for data structures  
**Format:** `[Purpose]` + `Result/State`  
**Examples:**
- `SnapshotResult` - Snapshot operation result
- `RecordingResult` - Recording operation result
- `AuthState` - Authentication state
- `ConnectionState` - Connection state

**Status:** ✅ **100% Compliant** - All types follow this pattern

## 9. File Naming Convention

### Pattern: `[purpose].[extension]`

**Purpose:** Consistent file naming across the codebase  
**Format:** `[purpose]` + `.` + `[extension]`  
**Examples:**
- `authStore.ts` - Authentication store
- `WebSocketService.ts` - WebSocket service
- `CameraPage.tsx` - Camera page component
- `useKeyboardShortcuts.ts` - Keyboard shortcuts hook

**Status:** ✅ **100% Compliant** - All files follow this pattern

## 10. Compliance Checklist

### ✅ Store Naming
- [x] All stores use `use[Domain]Store` pattern
- [x] Domain names are descriptive and clear
- [x] Consistent with Zustand conventions

### ✅ Service Naming
- [x] All services use `[Domain]Service` pattern
- [x] Domain names match business logic
- [x] Consistent with service layer architecture

### ✅ Component Naming
- [x] All components use `[Purpose][Type]` pattern
- [x] Purpose clearly describes functionality
- [x] Type indicates component category (Page, Table, Button, etc.)

### ✅ Directory Structure
- [x] All directories use `[domain]/[purpose]` pattern
- [x] Logical grouping of related files
- [x] Consistent with React/TypeScript best practices

## 11. Benefits

### Code Readability
- **Clear Intent:** Names immediately convey purpose and responsibility
- **Consistent Patterns:** Developers can predict naming conventions
- **Reduced Cognitive Load:** Familiar patterns reduce mental overhead

### Maintainability
- **Easy Refactoring:** Consistent patterns make changes predictable
- **Team Onboarding:** New developers understand conventions quickly
- **Code Reviews:** Naming standards simplify review process

### Architecture Alignment
- **Layer Separation:** Names reflect architectural layers
- **Domain Clarity:** Business domains are clearly identified
- **Interface Contracts:** Service interfaces are self-documenting

## 12. Implementation Guidelines

### For New Development
1. **Follow Established Patterns:** Use existing conventions as templates
2. **Be Descriptive:** Choose names that clearly describe purpose
3. **Consistent Casing:** Use PascalCase for classes, camelCase for functions
4. **Avoid Abbreviations:** Use full words for clarity

### For Code Reviews
1. **Check Naming Compliance:** Verify new code follows conventions
2. **Suggest Improvements:** Recommend better names when appropriate
3. **Document Exceptions:** Note any deviations with rationale

### For Refactoring
1. **Maintain Consistency:** Update related files when changing names
2. **Update Documentation:** Keep naming docs current
3. **Team Communication:** Notify team of significant naming changes

---

**Document Status:** Active Implementation  
**Classification:** Development Reference  
**Review Cycle:** Quarterly  
**Approval:** Development Team Lead

**Last Updated:** January 2025  
**Next Review:** April 2025
