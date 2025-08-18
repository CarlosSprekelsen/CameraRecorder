import { hello, add } from '../../src/index';

describe('index', () => {
  describe('hello', () => {
    it('should return a greeting message', () => {
      const result = hello('World');
      expect(result).toBe('Hello, World!');
    });
  });

  describe('add', () => {
    it('should add two numbers correctly', () => {
      const result = add(2, 3);
      expect(result).toBe(5);
    });
  });
}); 