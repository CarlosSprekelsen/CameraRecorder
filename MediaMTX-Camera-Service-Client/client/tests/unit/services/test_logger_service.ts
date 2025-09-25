/**
 * LoggerService unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-LOG-001: Log level management
 * - REQ-LOG-002: Log entry creation and storage
 * - REQ-LOG-003: Log retrieval and filtering
 * - REQ-LOG-004: Log export functionality
 * - REQ-LOG-005: Singleton pattern implementation
 * 
 * Test Categories: Unit
 * API Documentation Reference: ../mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

import { LoggerService, LogLevel, LogEntry } from '../../../src/services/logger/LoggerService';

// Mock console methods
const mockConsole = {
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
  log: jest.fn(),
};

Object.defineProperty(console, 'debug', { value: mockConsole.debug });
Object.defineProperty(console, 'info', { value: mockConsole.info });
Object.defineProperty(console, 'warn', { value: mockConsole.warn });
Object.defineProperty(console, 'error', { value: mockConsole.error });
Object.defineProperty(console, 'log', { value: mockConsole.log });

// Mock process.env
const originalEnv = process.env.NODE_ENV;

describe('LoggerService Unit Tests', () => {
  let loggerService: LoggerService;

  beforeEach(() => {
    jest.clearAllMocks();
    // Reset singleton instance
    (LoggerService as any).instance = undefined;
    loggerService = LoggerService.getInstance();
    loggerService.clearLogs();
  });

  afterEach(() => {
    process.env.NODE_ENV = originalEnv;
  });

  describe('REQ-LOG-001: Log level management', () => {
    test('should create debug log', () => {
      const message = 'Debug message';
      const context = { userId: '123' };

      loggerService.debug(message, context);

      const logs = loggerService.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].level).toBe(LogLevel.DEBUG);
      expect(logs[0].message).toBe(message);
      expect(logs[0].context).toEqual(context);
      expect(logs[0].timestamp).toBeDefined();
    });

    test('should create info log', () => {
      const message = 'Info message';
      const context = { action: 'login' };

      loggerService.info(message, context);

      const logs = loggerService.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].level).toBe(LogLevel.INFO);
      expect(logs[0].message).toBe(message);
      expect(logs[0].context).toEqual(context);
    });

    test('should create warn log', () => {
      const message = 'Warning message';
      const context = { threshold: 80 };

      loggerService.warn(message, context);

      const logs = loggerService.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].level).toBe(LogLevel.WARN);
      expect(logs[0].message).toBe(message);
      expect(logs[0].context).toEqual(context);
    });

    test('should create error log with error object', () => {
      const message = 'Error message';
      const context = { operation: 'connect' };
      const error = new Error('Connection failed');

      loggerService.error(message, context, error);

      const logs = loggerService.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].level).toBe(LogLevel.ERROR);
      expect(logs[0].message).toBe(message);
      expect(logs[0].context).toEqual(context);
      expect(logs[0].error).toBe(error);
    });

    test('should create error log without error object', () => {
      const message = 'Error message';
      const context = { operation: 'connect' };

      loggerService.error(message, context);

      const logs = loggerService.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].level).toBe(LogLevel.ERROR);
      expect(logs[0].message).toBe(message);
      expect(logs[0].context).toEqual(context);
      expect(logs[0].error).toBeUndefined();
    });
  });

  describe('REQ-LOG-002: Log entry creation and storage', () => {
    test('should store log entries with timestamp', () => {
      const message = 'Test message';
      const beforeTime = new Date().getTime();

      loggerService.info(message);

      const logs = loggerService.getLogs();
      expect(logs).toHaveLength(1);
      
      const logTime = new Date(logs[0].timestamp).getTime();
      expect(logTime).toBeGreaterThanOrEqual(beforeTime);
      expect(logTime).toBeLessThanOrEqual(new Date().getTime());
    });

    test('should store log entries with context', () => {
      const message = 'Test message';
      const context = { key1: 'value1', key2: 123 };

      loggerService.info(message, context);

      const logs = loggerService.getLogs();
      expect(logs[0].context).toEqual(context);
    });

    test('should limit log entries to maxLogs', () => {
      // Create more logs than the default maxLogs (1000)
      for (let i = 0; i < 1001; i++) {
        loggerService.info(`Message ${i}`);
      }

      const logs = loggerService.getLogs();
      expect(logs).toHaveLength(1000);
      expect(logs[0].message).toBe('Message 1'); // First log should be removed
      expect(logs[999].message).toBe('Message 1000'); // Last log should be the newest
    });

    test('should handle logs without context', () => {
      const message = 'Simple message';

      loggerService.info(message);

      const logs = loggerService.getLogs();
      expect(logs[0].context).toBeUndefined();
    });
  });

  describe('REQ-LOG-003: Log retrieval and filtering', () => {
    beforeEach(() => {
      // Create logs of different levels
      loggerService.debug('Debug message 1');
      loggerService.info('Info message 1');
      loggerService.warn('Warning message 1');
      loggerService.error('Error message 1');
      loggerService.debug('Debug message 2');
      loggerService.info('Info message 2');
    });

    test('should get all logs when no level specified', () => {
      const logs = loggerService.getLogs();
      expect(logs).toHaveLength(6);
    });

    test('should filter logs by debug level', () => {
      const logs = loggerService.getLogs(LogLevel.DEBUG);
      expect(logs).toHaveLength(2);
      logs.forEach(log => {
        expect(log.level).toBe(LogLevel.DEBUG);
      });
    });

    test('should filter logs by info level', () => {
      const logs = loggerService.getLogs(LogLevel.INFO);
      expect(logs).toHaveLength(2);
      logs.forEach(log => {
        expect(log.level).toBe(LogLevel.INFO);
      });
    });

    test('should filter logs by warn level', () => {
      const logs = loggerService.getLogs(LogLevel.WARN);
      expect(logs).toHaveLength(1);
      expect(logs[0].level).toBe(LogLevel.WARN);
    });

    test('should filter logs by error level', () => {
      const logs = loggerService.getLogs(LogLevel.ERROR);
      expect(logs).toHaveLength(1);
      expect(logs[0].level).toBe(LogLevel.ERROR);
    });

    test('should return empty array for non-existent level', () => {
      const logs = loggerService.getLogs('nonexistent' as LogLevel);
      expect(logs).toHaveLength(0);
    });
  });

  describe('REQ-LOG-004: Log export functionality', () => {
    test('should export logs as JSON string', () => {
      loggerService.info('Test message 1');
      loggerService.warn('Test message 2');

      const exportedLogs = loggerService.exportLogs();
      const parsedLogs = JSON.parse(exportedLogs);

      expect(Array.isArray(parsedLogs)).toBe(true);
      expect(parsedLogs).toHaveLength(2);
      expect(parsedLogs[0].message).toBe('Test message 1');
      expect(parsedLogs[1].message).toBe('Test message 2');
    });

    test('should export empty array when no logs', () => {
      const exportedLogs = loggerService.exportLogs();
      const parsedLogs = JSON.parse(exportedLogs);

      expect(Array.isArray(parsedLogs)).toBe(true);
      expect(parsedLogs).toHaveLength(0);
    });

    test('should export logs with proper formatting', () => {
      loggerService.info('Test message', { key: 'value' });

      const exportedLogs = loggerService.exportLogs();
      const parsedLogs = JSON.parse(exportedLogs);

      expect(parsedLogs[0]).toHaveProperty('timestamp');
      expect(parsedLogs[0]).toHaveProperty('level');
      expect(parsedLogs[0]).toHaveProperty('message');
      expect(parsedLogs[0]).toHaveProperty('context');
    });
  });

  describe('REQ-LOG-005: Singleton pattern implementation', () => {
    test('should return same instance on multiple calls', () => {
      const instance1 = LoggerService.getInstance();
      const instance2 = LoggerService.getInstance();

      expect(instance1).toBe(instance2);
    });

    test('should maintain state across singleton calls', () => {
      const logger1 = LoggerService.getInstance();
      logger1.info('Message from logger1');

      const logger2 = LoggerService.getInstance();
      const logs = logger2.getLogs();

      expect(logs).toHaveLength(1);
      expect(logs[0].message).toBe('Message from logger1');
    });

    test('should clear logs on singleton instance', () => {
      const logger1 = LoggerService.getInstance();
      logger1.info('Message 1');
      logger1.info('Message 2');

      const logger2 = LoggerService.getInstance();
      logger2.clearLogs();

      const logs = logger1.getLogs();
      expect(logs).toHaveLength(0);
    });
  });

  describe('Console output in development', () => {
    test('should output to console in development mode', () => {
      process.env.NODE_ENV = 'development';

      loggerService.info('Development message', { key: 'value' });

      expect(mockConsole.info).toHaveBeenCalledWith(
        '[INFO] Development message',
        { key: 'value' },
        undefined
      );
    });

    test('should output error to console with error object', () => {
      process.env.NODE_ENV = 'development';
      const error = new Error('Test error');

      loggerService.error('Error message', { key: 'value' }, error);

      expect(mockConsole.error).toHaveBeenCalledWith(
        '[ERROR] Error message',
        { key: 'value' },
        error
      );
    });

    test('should not output to console in production mode', () => {
      process.env.NODE_ENV = 'production';

      loggerService.info('Production message');

      expect(mockConsole.info).not.toHaveBeenCalled();
    });

    test('should handle missing console methods gracefully', () => {
      process.env.NODE_ENV = 'development';
      // Mock console with missing methods
      const originalConsole = console;
      (global as any).console = { log: jest.fn() };

      loggerService.debug('Debug message');

      expect(console.log).toHaveBeenCalledWith(
        '[DEBUG] Debug message',
        undefined,
        undefined
      );

      // Restore console
      (global as any).console = originalConsole;
    });
  });

  describe('Log entry structure validation', () => {
    test('should create valid log entry structure', () => {
      const message = 'Test message';
      const context = { key: 'value' };
      const error = new Error('Test error');

      loggerService.error(message, context, error);

      const logs = loggerService.getLogs();
      const logEntry = logs[0];

      expect(logEntry).toHaveProperty('timestamp');
      expect(logEntry).toHaveProperty('level');
      expect(logEntry).toHaveProperty('message');
      expect(logEntry).toHaveProperty('context');
      expect(logEntry).toHaveProperty('error');

      expect(typeof logEntry.timestamp).toBe('string');
      expect(typeof logEntry.level).toBe('string');
      expect(typeof logEntry.message).toBe('string');
      expect(typeof logEntry.context).toBe('object');
      expect(logEntry.error).toBeInstanceOf(Error);
    });

    test('should validate ISO timestamp format', () => {
      loggerService.info('Test message');

      const logs = loggerService.getLogs();
      const timestamp = logs[0].timestamp;

      // Should be valid ISO string
      expect(() => new Date(timestamp)).not.toThrow();
      expect(timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/);
    });

    test('should handle undefined context gracefully', () => {
      loggerService.info('Message without context');

      const logs = loggerService.getLogs();
      expect(logs[0].context).toBeUndefined();
    });
  });
});
