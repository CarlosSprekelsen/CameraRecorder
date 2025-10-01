# BUG-022: Component CSS Class Mismatches

## Summary
Several atomic components have test expectations that don't match the actual CSS classes used in the implementation. Tests are expecting Material-UI style classes but components use Tailwind CSS.

## Affected Components
- **AppBar**: Expects `bg-blue-600` but component uses `bg-white`
- **TextField**: Missing `for` attribute on labels, causing accessibility issues
- **PermissionGate**: Component not rendering properly in tests

## Test Failures
```
● AppBar Component › REQ-APPBAR-001: AppBar renders with correct styling
  Expected the element to have class: bg-blue-600
  Received: appbar bg-white shadow-sm border-b border-gray-200 z-40 fixed

● TextField Component › REQ-TEXTFIELD-003: TextField handles disabled state
  Found a label with the text of: Disabled Field, however no form control was found associated to that label
```

## Root Cause
- Tests were written expecting Material-UI styling but components use Tailwind CSS
- Missing accessibility attributes in component implementations
- Test expectations not updated after architecture change to Atomic Design pattern

## Expected Behavior
- Components should use consistent Tailwind CSS classes
- Tests should expect the actual classes used in components
- Accessibility attributes should be properly implemented

## Priority
**MEDIUM** - Affects component styling and accessibility

## Assignee
**UI/UX Team**

## Files to Fix
- `tests/unit/components/atoms/AppBar.test.tsx` - Update CSS expectations
- `src/components/atoms/TextField/TextField.tsx` - Add accessibility attributes
- `tests/unit/components/atoms/TextField.test.tsx` - Fix accessibility tests
- `tests/unit/components/organisms/PermissionGate.test.tsx` - Fix rendering issues
