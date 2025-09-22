/**
 * Mock Logger Service for Error Boundary Tests
 */

const mockLogger = {
  error: jest.fn(),
  warn: jest.fn(),
  info: jest.fn(),
  debug: jest.fn(),
};

const mockLoggers = {
  component: {
    error: jest.fn(),
  },
  service: {
    error: jest.fn(),
  },
};

module.exports = {
  logger: mockLogger,
  loggers: mockLoggers,
};
