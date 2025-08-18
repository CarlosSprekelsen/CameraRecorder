import { hello, add } from '../../src/index';

describe('Integration Tests', () => {
  describe('index module integration', () => {
    it('should work with multiple functions together', () => {
      const greeting = hello('Test');
      const sum = add(10, 20);
      
      expect(greeting).toBe('Hello, Test!');
      expect(sum).toBe(30);
    });
  });
}); 