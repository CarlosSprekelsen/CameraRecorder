/**
 * Basic Jest test to validate test infrastructure
 * This test validates that Jest is working without React dependencies
 */

describe('Basic Jest Infrastructure Test', () => {
  it('should run basic Jest tests', () => {
    expect(1 + 1).toBe(2);
    expect('hello').toBe('hello');
    expect(true).toBe(true);
  });

  it('should handle async operations', async () => {
    const result = await Promise.resolve('async result');
    expect(result).toBe('async result');
  });

  it('should handle object operations', () => {
    const obj = { name: 'test', value: 42 };
    expect(obj.name).toBe('test');
    expect(obj.value).toBe(42);
  });

  it('should handle array operations', () => {
    const arr = [1, 2, 3, 4, 5];
    expect(arr.length).toBe(5);
    expect(arr[0]).toBe(1);
    expect(arr.includes(3)).toBe(true);
  });
});
