/**
 * Infrastructure Layer - Logger Service
 *
 * Provides comprehensive logging functionality for the MediaMTX client application.
 * Implements structured logging with multiple levels, context support, and development
 * console output. Singleton pattern ensures consistent logging across the application.
 *
 * @fileoverview Logger service implementation
 * @author MediaMTX Development Team
 * @version 1.0.0
 */


/**
 * Log Level Enumeration
 *
 * Defines the available logging levels in order of severity.
 *
 * @enum {string}
 */
export enum LogLevel {
  DEBUG = 'debug',
  INFO = 'info',
  WARN = 'warn',
  ERROR = 'error',
}

/**
 * Log Entry Interface
 *
 * Defines the structure for individual log entries with timestamp, level,
 * message, optional context, and error information.
 *
 * @interface LogEntry
 */
export interface LogEntry {
  /** ISO timestamp when the log entry was created */
  timestamp: string;
  /** Log level indicating severity */
  level: LogLevel;
  /** Log message content */
  message: string;
  /** Optional context data for debugging */
  context?: Record<string, unknown>;
  /** Optional error object for error logs */
  error?: Error;
}

/**
 * Logger Service - Centralized logging system
 *
 * Singleton service providing structured logging with multiple levels, context support,
 * and development console output. Manages log entries with automatic rotation and
 * provides methods for different log levels with optional context and error information.
 *
 * @class LoggerService
 *
 * @example
 * ```typescript
 * const logger = LoggerService.getInstance();
 *
 * // Basic logging
 * logger.info('User logged in', { userId: '123' });
 * logger.error('Connection failed', error, { endpoint: '/api/connect' });
 *
 * // Get logs
 * const logs = logger.getLogs(LogLevel.ERROR);
 * ```
 *
 * @see {@link LogLevel} Available log levels
 * @see {@link LogEntry} Log entry structure
 */
export class LoggerService {
  private static instance: LoggerService;
  private logs: LogEntry[] = [];
  private maxLogs = 1000;

  private constructor() {}

  static getInstance(): LoggerService {
    if (!LoggerService.instance) {
      LoggerService.instance = new LoggerService();
    }
    return LoggerService.instance;
  }

  private log(
    level: LogLevel,
    message: string,
    context?: Record<string, unknown>,
    error?: Error,
  ): void {
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      message,
      context,
      error,
    };

    this.logs.push(entry);

    // Keep only the last maxLogs entries
    if (this.logs.length > this.maxLogs) {
      this.logs = this.logs.slice(-this.maxLogs);
    }

    // Console output for development
    if (process.env.NODE_ENV === 'development') {
      const logMethod = console[level] || console.log;
      logMethod(`[${level.toUpperCase()}] ${message}`, context, error);
    }
  }

  debug(message: string, context?: Record<string, unknown>): void {
    this.log(LogLevel.DEBUG, message, context);
  }

  info(message: string, context?: Record<string, unknown>): void {
    this.log(LogLevel.INFO, message, context);
  }

  warn(message: string, context?: Record<string, unknown>): void {
    this.log(LogLevel.WARN, message, context);
  }

  error(message: string, context?: Record<string, unknown>, error?: Error): void {
    this.log(LogLevel.ERROR, message, context, error);
  }

  getLogs(level?: LogLevel): LogEntry[] {
    if (level) {
      return this.logs.filter((log) => log.level === level);
    }
    return [...this.logs];
  }

  clearLogs(): void {
    this.logs = [];
  }

  exportLogs(): string {
    return JSON.stringify(this.logs, null, 2);
  }
}

// Export singleton instance
export const logger = LoggerService.getInstance();
