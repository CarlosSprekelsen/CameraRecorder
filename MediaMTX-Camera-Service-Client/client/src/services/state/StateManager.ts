/**
 * StateManager Service - Architecture Compliance
 * 
 * Architecture requirement: "StateManager service" (Section 5.2)
 * Centralized state management coordinating existing Zustand stores
 */

import { useDeviceStore } from '../../stores/device/deviceStore';
import { useRecordingStore } from '../../stores/recording/recordingStore';
import { useServerStore } from '../../stores/server/serverStore';
import { useFileStore } from '../../stores/file/fileStore';
import { useStreamingStore } from '../../stores/streaming/streamingStore';
import { useAuthStore } from '../../stores/auth/authStore';
import { useConnectionStore } from '../../stores/connection/connectionStore';

export type StoreName = 
  | 'device' 
  | 'recording' 
  | 'server' 
  | 'file' 
  | 'streaming' 
  | 'auth' 
  | 'connection';

export interface StateManagerConfig {
  enableLogging?: boolean;
  enablePersistence?: boolean;
}

/**
 * StateManager Service
 * Architecture requirement: Centralized state store with unidirectional data flow (ADR-002)
 */
export class StateManager {
  private stores: Map<StoreName, any> = new Map();
  private config: StateManagerConfig;

  constructor(config: StateManagerConfig = {}) {
    this.config = {
      enableLogging: false,
      enablePersistence: false,
      ...config
    };
    
    this.initializeStores();
  }

  /**
   * Initialize all stores
   * Leverage existing Zustand store implementations
   */
  private initializeStores(): void {
    this.stores.set('device', useDeviceStore);
    this.stores.set('recording', useRecordingStore);
    this.stores.set('server', useServerStore);
    this.stores.set('file', useFileStore);
    this.stores.set('streaming', useStreamingStore);
    this.stores.set('auth', useAuthStore);
    this.stores.set('connection', useConnectionStore);
  }

  /**
   * Get state from specified store
   * Architecture requirement: getState()
   */
  getState(storeName: StoreName): any {
    const store = this.stores.get(storeName);
    if (!store) {
      throw new Error(`Store '${storeName}' not found`);
    }
    
    if (this.config.enableLogging) {
      console.log(`StateManager: Getting state from ${storeName}`);
    }
    
    return store.getState();
  }

  /**
   * Dispatch action to specified store
   * Architecture requirement: dispatch()
   */
  dispatch(storeName: StoreName, action: any): void {
    const store = this.stores.get(storeName);
    if (!store) {
      throw new Error(`Store '${storeName}' not found`);
    }
    
    if (this.config.enableLogging) {
      console.log(`StateManager: Dispatching action to ${storeName}`, action);
    }
    
    // Dispatch to store - Zustand stores handle actions through their methods
    if (typeof action === 'function') {
      action(store.getState());
    } else {
      // Handle object-based actions
      const { type, payload } = action;
      const storeState = store.getState();
      
      if (typeof storeState[type] === 'function') {
        storeState[type](payload);
      } else {
        throw new Error(`Action '${type}' not found in store '${storeName}'`);
      }
    }
  }

  /**
   * Subscribe to store changes
   * Architecture requirement: subscribe()
   */
  subscribe(storeName: StoreName, callback: (state: any) => void): () => void {
    const store = this.stores.get(storeName);
    if (!store) {
      throw new Error(`Store '${storeName}' not found`);
    }
    
    if (this.config.enableLogging) {
      console.log(`StateManager: Subscribing to ${storeName}`);
    }
    
    // Subscribe to store changes
    return store.subscribe(callback);
  }

  /**
   * Get all store states
   * Convenience method for debugging and monitoring
   */
  getAllStates(): Record<StoreName, any> {
    const states: Record<string, any> = {};
    
    for (const [storeName, store] of this.stores) {
      states[storeName] = store.getState();
    }
    
    return states as Record<StoreName, any>;
  }

  /**
   * Reset all stores
   * Useful for testing and error recovery
   */
  resetAllStores(): void {
    for (const [storeName, store] of this.stores) {
      const storeState = store.getState();
      if (typeof storeState.reset === 'function') {
        storeState.reset();
      }
    }
    
    if (this.config.enableLogging) {
      console.log('StateManager: All stores reset');
    }
  }

  /**
   * Get store names
   * Utility method for introspection
   */
  getStoreNames(): StoreName[] {
    return Array.from(this.stores.keys());
  }
}
