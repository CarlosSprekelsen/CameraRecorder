/**
 * Centralized Logging Service for MediaMTX Camera Service Client
 * Provides consistent logging across the application
 * 
 * Architecture Pattern: Centralized Logging
 * - Replaces scattered console.log/error/warn statements
 * - Provides structured logging with context
 * - Supports different log levels and formatting
 * - Integrates with error tracking and monitoring
 * 
 * Usage:
 * ```typescript
 * import { logger } from '../services/loggerService';
 * logger.info('User action', { userId: '123', action: 'connect' });
 * logger.error('Connection failed', error, { context: 'websocket' });
 * ```
 */

/**
 * Log levels in order of severity
 */
export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3
}

/**
 * Log entry structure
 */
export interface LogEntry {
  timestamp: Date;
  level: LogLevel;
  message: string;
  context?: string;
  data?: any;
  error?: Error;
  component?: string;
  userId?: string;
  sessionId?: string;
}

/**
 * Logger configuration
 */
export interface LoggerConfig {
  level: LogLevel;
  enableConsole: boolean;
  enableRemote: boolean;
  remoteEndpoint?: string;
  maxEntries: number;
  enablePerformance: boolean;
}

/**
 * Logger service class
 */
export class LoggerService {
  private static instance: LoggerService;
  private config: LoggerConfig;
  private logEntries: LogEntry[] = [];
  private performanceEntries: Map<string, number> = new Map();

  private constructor(config: Partial<LoggerConfig> = {}) {
    this.config = {
      level: process.env.NODE_ENV === 'development' ? LogLevel.DEBUG : LogLevel.INFO,
      enableConsole: true,
      enableRemote: false,
      maxEntries: 1000,
      enablePerformance: process.env.NODE_ENV === 'development',
      ...config
    };
  }

  /**
   * Get singleton instance
   */
  public static getInstance(config?: Partial<LoggerConfig>): LoggerService {
    if (!LoggerService.instance) {
      LoggerService.instance = new LoggerService(config);
    }
    return LoggerService.instance;
  }

  /**
   * Log debug message
   */
  public debug(message: string, data?: any, context?: string): void {
    this.log(LogLevel.DEBUG, message, data, context);
  }

  /**
   * Log info message
   */
  public info(message: string, data?: any, context?: string): void {
    this.log(LogLevel.INFO, message, data, context);
  }

  /**
   * Log warning message
   */
  public warn(message: string, data?: any, context?: string): void {
    this.log(LogLevel.WARN, message, data, context);
  }

  /**
   * Log error message
   */
  public error(message: string, error?: Error, context?: string, data?: any): void {
    this.log(LogLevel.ERROR, message, { ...data, error: error?.message, stack: error?.stack }, context);
  }

  /**
   * Core logging method
   */
  private log(level: LogLevel, message: string, data?: any, context?: string): void {
    // Check if we should log this level
    if (level < this.config.level) {
      return;
    }

    const entry: LogEntry = {
      timestamp: new Date(),
      level,
      message,
      context,
      data,
      component: this.getCallerComponent()
    };

    // Add to log entries
    this.logEntries.push(entry);
    this.trimLogEntries();

    // Console output
    if (this.config.enableConsole) {
      this.logToConsole(entry);
    }

    // Remote logging (if enabled)
    if (this.config.enableRemote && this.config.remoteEndpoint) {
      this.logToRemote(entry);
    }
  }

  /**
   * Log to console with appropriate formatting
   */
  private logToConsole(entry: LogEntry): void {
    const timestamp = entry.timestamp.toISOString();
    const levelName = LogLevel[entry.level];
    const contextStr = entry.context ? `[${entry.context}]` : '';
    const componentStr = entry.component ? `[${entry.component}]` : '';
    
    const prefix = `${timestamp} ${levelName} ${contextStr}${componentStr}`;
    
    switch (entry.level) {
      case LogLevel.DEBUG:
        console.debug(`${prefix} ${entry.message}`, entry.data);
        break;
      case LogLevel.INFO:
        console.info(`${prefix} ${entry.message}`, entry.data);
        break;
      case LogLevel.WARN:
        console.warn(`${prefix} ${entry.message}`, entry.data);
        break;
      case LogLevel.ERROR:
        console.error(`${prefix} ${entry.message}`, entry.data);
        break;
    }
  }

  /**
   * Log to remote endpoint
   */
  private async logToRemote(entry: LogEntry): Promise<void> {
    try {
      await fetch(this.config.remoteEndpoint!, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(entry)
      });
    } catch (error) {
      // Don't log remote logging errors to avoid infinite loops
      console.error('Failed to send log to remote endpoint:', error);
    }
  }

  /**
   * Get caller component name from stack trace
   */
  private getCallerComponent(): string | undefined {
    try {
      const stack = new Error().stack;
      if (!stack) return undefined;
      
      const lines = stack.split('\n');
      // Look for the first line that contains a component file
      for (const line of lines) {
        if (line.includes('/components/') && line.includes('.tsx')) {
          const match = line.match(/\/([^/]+)\.tsx/);
          return match ? match[1] : undefined;
        }
      }
    } catch (error) {
      // Ignore errors in stack trace parsing
    }
    return undefined;
  }

  /**
   * Trim log entries to prevent memory leaks
   */
  private trimLogEntries(): void {
    if (this.logEntries.length > this.config.maxEntries) {
      this.logEntries = this.logEntries.slice(-this.config.maxEntries);
    }
  }

  /**
   * Start performance timing
   */
  public startTiming(label: string): void {
    if (this.config.enablePerformance) {
      this.performanceEntries.set(label, performance.now());
    }
  }

  /**
   * End performance timing and log result
   */
  public endTiming(label: string, context?: string): number | undefined {
    if (!this.config.enablePerformance) return undefined;
    
    const startTime = this.performanceEntries.get(label);
    if (startTime === undefined) {
      this.warn(`Performance timing '${label}' was not started`, undefined, context);
      return undefined;
    }
    
    const duration = performance.now() - startTime;
    this.performanceEntries.delete(label);
    
    this.info(`Performance: ${label} took ${duration.toFixed(2)}ms`, { duration }, context);
    return duration;
  }

  /**
   * Get recent log entries
   */
  public getRecentEntries(count: number = 100): LogEntry[] {
    return this.logEntries.slice(-count);
  }

  /**
   * Get log entries by level
   */
  public getEntriesByLevel(level: LogLevel): LogEntry[] {
    return this.logEntries.filter(entry => entry.level === level);
  }

  /**
   * Clear all log entries
   */
  public clear(): void {
    this.logEntries = [];
    this.performanceEntries.clear();
  }

  /**
   * Export logs as JSON
   */
  public exportLogs(): string {
    return JSON.stringify(this.logEntries, null, 2);
  }

  /**
   * Update configuration
   */
  public updateConfig(config: Partial<LoggerConfig>): void {
    this.config = { ...this.config, ...config };
  }
}

/**
 * Singleton instance for easy access
 */
export const logger = LoggerService.getInstance();

/**
 * Convenience functions for common logging patterns
 */
export const loggers = {
  /**
   * Component lifecycle logging
   */
  component: {
    mount: (componentName: string) => logger.info(`Component mounted: ${componentName}`, undefined, 'component'),
    unmount: (componentName: string) => logger.info(`Component unmounted: ${componentName}`, undefined, 'component'),
    render: (componentName: string, props?: any) => logger.debug(`Component rendered: ${componentName}`, props, 'component'),
    error: (componentName: string, error: Error) => logger.error(`Component error: ${componentName}`, error, 'component')
  },

  /**
   * Service operation logging
   */
  service: {
    start: (serviceName: string, operation: string) => logger.info(`Service operation started: ${serviceName}.${operation}`, undefined, 'service'),
    success: (serviceName: string, operation: string, result?: any) => logger.info(`Service operation succeeded: ${serviceName}.${operation}`, result, 'service'),
    error: (serviceName: string, operation: string, error: Error) => logger.error(`Service operation failed: ${serviceName}.${operation}`, error, 'service')
  },

  /**
   * User action logging
   */
  user: {
    action: (action: string, data?: any) => logger.info(`User action: ${action}`, data, 'user'),
    error: (action: string, error: Error) => logger.error(`User action failed: ${action}`, error, 'user')
  },

  /**
   * Performance logging
   */
  performance: {
    start: (label: string) => logger.startTiming(label),
    end: (label: string, context?: string) => logger.endTiming(label, context)
  }
};
