# Phase 1 Migration Guide: Foundation & Deprecation

**Date**: January 15, 2025  
**Version**: Phase 1 Implementation  
**Status**: In Progress  

## Overview

This document outlines the Phase 1 migration changes that establish the foundation for architectural improvements. Phase 1 focuses on deprecation of duplicate components and introduction of service layer abstractions.

## Changes Implemented

### 1. Deprecated Components

#### ConnectionStatus Component Deprecation

**Deprecated**: `/components/common/ConnectionStatus.tsx`  
**Replacement**: `/components/ConnectionStatus/ConnectionStatus.tsx`  

**Migration Steps**:
1. Update imports in affected files:
   ```typescript
   // OLD (deprecated)
   import ConnectionStatus from '../common/ConnectionStatus';
   
   // NEW (current)
   import ConnectionStatus from '../ConnectionStatus/ConnectionStatus';
   ```

2. Props interface remains the same - no breaking changes
3. Component will be removed in v2.0

**Files Affected**: Check for imports of the deprecated component

### 2. Service Layer Abstraction

#### New ConnectionService

**File**: `/services/connectionService.ts`  
**Purpose**: Abstraction layer between components and connection store  

**Usage**:
```typescript
import { connectionService } from '../services/connectionService';

// Instead of direct store access
const handleConnect = async () => {
  try {
    await connectionService.connect();
  } catch (error) {
    // Handle error
  }
};
```

**Benefits**:
- Encapsulates connection logic
- Provides consistent API
- Enables easier testing and mocking
- Follows architectural patterns

#### New LoggerService

**File**: `/services/loggerService.ts`  
**Purpose**: Centralized logging to replace scattered console statements  

**Usage**:
```typescript
import { logger, loggers } from '../services/loggerService';

// Basic logging
logger.info('User connected', { userId: '123' });
logger.error('Connection failed', error, 'websocket');

// Convenience functions
loggers.service.start('ConnectionService', 'connect');
loggers.user.action('camera_connect', { deviceId: 'camera0' });
loggers.performance.start('connection_time');
```

**Benefits**:
- Consistent log formatting
- Structured logging with context
- Performance timing support
- Easy log export and analysis

### 3. Updated Components

#### ConnectionStatus Component

**File**: `/components/ConnectionStatus/ConnectionStatus.tsx`  
**Changes**:
- Added service layer integration
- Replaced direct store access with service calls
- Added structured logging
- Improved error handling

**Before**:
```typescript
const handleConnect = async () => {
  try {
    await connect(); // Direct store method
  } catch (error) {
    setLocalError(error.message);
  }
};
```

**After**:
```typescript
const handleConnect = async () => {
  loggers.service.start('ConnectionService', 'connect');
  try {
    await connectionService.connect(); // Service layer
    loggers.service.success('ConnectionService', 'connect');
  } catch (error) {
    loggers.service.error('ConnectionService', 'connect', error);
    setLocalError(error.message);
  }
};
```

## Migration Checklist

### For Developers

- [ ] Update imports from deprecated ConnectionStatus component
- [ ] Replace direct store access with service layer calls
- [ ] Add structured logging to new components
- [ ] Test service layer integration
- [ ] Update error handling to use new patterns

### For Components

- [ ] Replace `useConnectionStore.getState()` calls with service methods
- [ ] Add logging to user actions and service operations
- [ ] Implement proper error boundaries
- [ ] Use service layer for all external operations

### For Services

- [ ] Follow singleton pattern for service classes
- [ ] Implement proper error handling with custom error classes
- [ ] Add configuration interfaces
- [ ] Include comprehensive JSDoc documentation

## Testing Strategy

### Unit Tests
- Test service layer methods independently
- Mock store dependencies
- Verify error handling and logging

### Integration Tests
- Test service-store integration
- Verify component-service interaction
- Test error propagation

### Manual Testing
- Verify deprecated component warnings
- Test service layer functionality
- Check logging output

## Rollback Plan

If issues arise during migration:

1. **Immediate**: Revert to previous component versions
2. **Service Layer**: Disable service layer, fall back to direct store access
3. **Logging**: Disable centralized logging, revert to console statements

## Next Steps

### Phase 2 Preparation
- Monitor service layer usage patterns
- Identify additional components for migration
- Plan store refactoring strategy

### Documentation Updates
- Update component documentation
- Create service layer usage examples
- Document architectural patterns

## Support

For questions or issues during migration:
- Check this migration guide
- Review service layer documentation
- Test in development environment first
- Report issues with detailed logs

## Success Metrics

- [ ] All deprecated component imports updated
- [ ] Service layer used in 100% of new components
- [ ] Centralized logging implemented
- [ ] No direct store access in components
- [ ] All tests passing
- [ ] Performance maintained or improved
