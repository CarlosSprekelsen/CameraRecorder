# ADR-002: State Management with Zustand

**Status:** Accepted  
**Date:** January 2025  
**Deciders:** Development Team  
**Technical Story:** [Sprint-1] State Management Foundation

## Context

The MediaMTX Camera Service Client requires predictable state management for:
- Authentication state (token, role, permissions)
- Connection state (connected, disconnected, error)
- Device state (cameras, streams, status)
- Server state (info, metrics, health)

## Decision

We will use **Zustand** as the primary state management solution.

## Rationale

### Zustand Benefits
- **Lightweight:** Minimal bundle size (~2KB)
- **TypeScript First:** Excellent TypeScript support
- **Simple API:** Easy to learn and use
- **No Boilerplate:** Less code than Redux
- **Performance:** Optimized re-renders
- **DevTools:** Built-in debugging support

### Architecture Alignment
- **Layered State:** Matches our service layer architecture
- **Domain Separation:** Each store handles specific domain
- **Predictable Updates:** Clear state mutation patterns

## Alternatives Considered

### 1. Redux Toolkit
**Rejected because:**
- High boilerplate for simple state
- Complex setup for async operations
- Larger bundle size
- Overkill for our use case

### 2. Context API + useReducer
**Rejected because:**
- Performance issues with frequent updates
- Complex provider nesting
- No built-in dev tools
- Difficult to test

### 3. Jotai/Recoil
**Rejected because:**
- Learning curve for team
- Less mature ecosystem
- Over-engineering for our needs

## Implementation Details

### Store Structure
```typescript
// Domain-based stores
const useAuthStore = create<AuthState>((set) => ({
  token: null,
  isAuthenticated: false,
  login: async (token: string) => { /* implementation */ },
  logout: () => set({ token: null, isAuthenticated: false })
}));

const useConnectionStore = create<ConnectionState>((set) => ({
  status: 'disconnected',
  setStatus: (status) => set({ status }),
  setError: (error) => set({ error })
}));
```

### State Architecture
```
stores/
├── auth/           # Authentication state
├── connection/     # WebSocket connection state  
├── device/         # Camera device state
├── recording/      # Recording operations state
├── server/         # Server information state
└── file/          # File management state
```

### Type Safety
```typescript
interface AuthState {
  token: string | null;
  isAuthenticated: boolean;
  role: 'admin' | 'operator' | 'viewer' | null;
  permissions: string[];
}

interface AuthActions {
  login: (token: string) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
}
```

## Consequences

### Positive
- **Predictable State:** Clear state mutation patterns
- **Performance:** Optimized re-renders with selectors
- **Developer Experience:** Excellent TypeScript support
- **Testing:** Easy to test with simple functions
- **Debugging:** Built-in dev tools support

### Negative
- **Learning Curve:** Team needs to learn Zustand patterns
- **State Coupling:** Stores can become tightly coupled
- **Memory Usage:** All state kept in memory

### Risks
- **State Synchronization:** Multiple stores need coordination
- **Memory Leaks:** Improper cleanup of subscriptions
- **State Corruption:** Concurrent mutations

## Mitigation Strategies

### State Coordination
- Use middleware for cross-store communication
- Implement state synchronization patterns
- Clear separation of concerns between stores

### Memory Management
- Implement proper cleanup in useEffect
- Use selectors to prevent unnecessary re-renders
- Monitor memory usage in development

### Testing
- Unit tests for each store action
- Integration tests for store interactions
- Mock stores for component testing

## Store Patterns

### Authentication Store
```typescript
export const useAuthStore = create<AuthState & AuthActions>((set, get) => ({
  // State
  token: null,
  isAuthenticated: false,
  role: null,
  
  // Actions
  login: async (token: string) => {
    // Implementation
  },
  
  logout: () => {
    set({ token: null, isAuthenticated: false, role: null });
  }
}));
```

### Connection Store
```typescript
export const useConnectionStore = create<ConnectionState & ConnectionActions>((set) => ({
  // State
  status: 'disconnected',
  lastError: null,
  reconnectAttempts: 0,
  
  // Actions
  setStatus: (status) => set({ status }),
  setError: (error) => set({ lastError: error }),
  incrementReconnectAttempts: () => set((state) => ({ 
    reconnectAttempts: state.reconnectAttempts + 1 
  }))
}));
```

## Monitoring

### Metrics to Track
- Store update frequency
- Component re-render counts
- Memory usage per store
- Action execution time

### DevTools Integration
```typescript
import { devtools } from 'zustand/middleware';

export const useAuthStore = create<AuthState>()(
  devtools(
    (set) => ({
      // Store implementation
    }),
    { name: 'auth-store' }
  )
);
```

## Related ADRs
- [ADR-001: WebSocket Communication](#) - How state updates via WebSocket
- [ADR-003: Component Architecture](#) - How components consume state

## References
- [Zustand Documentation](https://github.com/pmndrs/zustand)
- [State Management Best Practices](https://kentcdodds.com/blog/application-state-management-with-react)
- [TypeScript with Zustand](https://github.com/pmndrs/zustand#typescript)

---

**Last Updated:** January 2025  
**Review Date:** April 2025  
**Next Review:** July 2025
